package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	internalUtils "github.com/jiotv-go/jiotv_go/v3/internal/utils"
	"github.com/jiotv-go/jiotv_go/v3/pkg/secureurl"
	"github.com/jiotv-go/jiotv_go/v3/pkg/television"
	pkgUtils "github.com/jiotv-go/jiotv_go/v3/pkg/utils"
	"github.com/valyala/fasthttp"
)

const (
	catchupEPGURL   = "https://jiotvapi.cdn.jio.com/apis/v1.3/getepg/get?offset=%d&channel_id=%s&langId=%d"
	okhttpUserAgent = "okhttp/4.12.13"
	defaultLangID   = 6
	epochThreshold  = 100000000000
)

// catchupSupport reports the channel's display name and whether the channel
// list marks it as offering catchup. known is false when the channel list is
// unavailable or the ID is absent from it, in which case callers should not
// block: an API hiccup should not disable a working feature.
func catchupSupport(channelID string) (name string, supported bool, known bool) {
	channelList, err := television.Channels()
	if err != nil {
		pkgUtils.Log.Printf("Unable to check catchup availability for %s: %v", channelID, err)
		return channelID, false, false
	}
	for _, channel := range channelList.Result {
		if channel.ID == channelID {
			return channel.Name, channel.IsCatchupAvailable, true
		}
	}
	return channelID, false, false
}

func CatchupHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	offsetStr := c.Query("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
		pkgUtils.Log.Printf("Invalid offset query parameter, defaulting to 0: %v", err)
	}

	// Channels without catchup still list programmes, but every stream request
	// for them is refused with an empty HTTP 403, which reaches the player as
	// an unexplained format/demuxer error. Say so instead of offering dead
	// links. A paced survey across ~40 channels found isCatchupAvailable to be
	// a perfect negative predictor: 0/8 sampled channels with the flag unset
	// were ever playable.
	if channelName, supported, known := catchupSupport(id); known && !supported {
		return c.Render("views/catchup", fiber.Map{
			"Title":       Title,
			"Error":       "Catchup is not available for " + channelName + ". This channel only offers a live stream.",
			"Channel":     channelName,
			"LivePlayURL": "/play/" + id + "?live=true",
		})
	}

	epgData, err := getCatchupEPG(id, offset)
	if err != nil {
		pkgUtils.Log.Println("Error fetching catchup EPG:", err)
		return c.Render("views/catchup", fiber.Map{
			"Title":       Title,
			"Error":       "Could not fetch catchup data",
			"Channel":     id,
			"LivePlayURL": "/play/" + id + "?live=true",
		})
	}

	currentTime := time.Now().UnixMilli()
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		loc = time.FixedZone("IST", 5*3600+30*60)
	}

	var pastEpgData []map[string]interface{}
	for _, p := range epgData {
		if start, ok := p["startEpoch"].(int64); ok {
			if start < epochThreshold {
				start = start * 1000
			}
			if start > currentTime {
				continue
			}
			startTime := time.UnixMilli(start).In(loc)
			p["showtime"] = startTime.Format("03:04 PM")
			if end, ok := p["endEpoch"].(int64); ok {
				if end < epochThreshold {
					end = end * 1000
				}
				endTime := time.UnixMilli(end).In(loc)
				p["endtime"] = endTime.Format("03:04 PM")
				if start <= currentTime && end > currentTime {
					p["IsLive"] = true
				}
			}
		}
		pastEpgData = append(pastEpgData, p)
	}

	currentDate := time.Now().In(loc).AddDate(0, 0, offset).Format("02/01/2006")
	showNext := offset < 0
	showPrev := offset > -7

	return c.Render("views/catchup", fiber.Map{
		"Title":       Title,
		"Data":        pastEpgData,
		"Channel":     id,
		"Offset":      offset,
		"NextOffset":  offset + 1,
		"PrevOffset":  offset - 1,
		"CurrentDate": currentDate,
		"ShowNext":    showNext,
		"ShowPrev":    showPrev,
	})
}

func CatchupStreamHandler(c *fiber.Ctx) error {
	id := strings.TrimSuffix(c.Params("id"), ".m3u8")
	start := c.Query("start")
	end := c.Query("end")

	if start == "" || end == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing start or end time")
	}

	if err := EnsureFreshTokens(); err != nil {
		pkgUtils.Log.Printf("Failed to ensure fresh tokens: %v", err)
	}

	srno := c.Query("srno")
	if srno == "" {
		pkgUtils.Log.Println("Warning: srno is missing for catchup request")
	}

	startInt, errStart := strconv.ParseInt(start, 10, 64)
	endInt, errEnd := strconv.ParseInt(end, 10, 64)
	if errStart == nil && errEnd == nil {
		start = time.UnixMilli(startInt).UTC().Format("20060102T150405")
		end = time.UnixMilli(endInt).UTC().Format("20060102T150405")
	}

	pkgUtils.Log.Printf("Fetching catchup URL for channel %s, start: %s, end: %s, srno: %s", id, start, end, srno)
	catchupResult, err := TV.GetCatchupURL(id, srno, start, end)
	if err != nil {
		pkgUtils.Log.Printf("Error fetching catchup URL: %v", err)
		return internalUtils.InternalServerError(c, err)
	}

	targetURL := catchupResult.Bitrates.Auto
	if targetURL == "" {
		targetURL = catchupResult.Result
	}
	pkgUtils.Log.Printf("Catchup Target URL: %s", targetURL)

	if targetURL == "" {
		return internalUtils.InternalServerError(c, fmt.Errorf("failed to get catchup URL from API"))
	}

	codedUrl, err := secureurl.EncryptURL(targetURL)
	if err != nil {
		return internalUtils.InternalServerError(c, err)
	}

	redirectURL := fmt.Sprintf("/render.m3u8?auth=%s&channel_key_id=%s", codedUrl, id)
	// Only append the token when the target URL does not already carry one.
	// The check must cover "__hdnea__=" as well: that is the spelling the
	// catchup API actually returns, and it does not contain the substring
	// "hdnea=", so a narrower check appends a second, conflicting token,
	// which JioTV's CDN then rejects with HTTP 400.
	if catchupResult.Hdnea != "" && !strings.Contains(targetURL, "hdnea=") && !strings.Contains(targetURL, "__hdnea__=") {
		redirectURL += "&hdnea=" + url.QueryEscape(catchupResult.Hdnea)
	}
	return c.Redirect(redirectURL, fiber.StatusFound)
}

func CatchupPlayerHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	start := c.Query("start")
	end := c.Query("end")
	srno := c.Query("srno")
	showName := c.Query("showname", "Catchup Show")
	description := c.Query("description", "No description available")
	episodePoster := c.Query("poster", "")
	showTime := c.Query("showtime", "")

	playerURL := fmt.Sprintf("/catchup/render/%s?start=%s&end=%s&srno=%s&v=6", id, start, end, srno)

	return c.Render("views/catchup_player", fiber.Map{
		"Title":         Title,
		"ChannelID":     id,
		"ShowName":      showName,
		"Description":   description,
		"EpisodePoster": episodePoster,
		"ShowTime":      showTime,
		"player_url":    playerURL,
	})
}

func CatchupRenderPlayerHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	start := c.Query("start")
	end := c.Query("end")
	srno := c.Query("srno")
	quality := c.Query("q", "")
	qualityForDrm := quality
	if qualityForDrm == "" {
		qualityForDrm = "high"
	}

	playURL := fmt.Sprintf("/catchup/stream/%s.m3u8?start=%s&end=%s&srno=%s", id, start, end, srno)
	if quality != "" {
		playURL += "&q=" + quality
	}

	startFmt := start
	endFmt := end
	startInt, errStart := strconv.ParseInt(start, 10, 64)
	endInt, errEnd := strconv.ParseInt(end, 10, 64)
	if errStart == nil && errEnd == nil {
		startFmt = time.UnixMilli(startInt).UTC().Format("20060102T150405")
		endFmt = time.UnixMilli(endInt).UTC().Format("20060102T150405")
	}

	if err := EnsureFreshTokens(); err != nil {
		pkgUtils.Log.Printf("Failed to ensure fresh tokens: %v", err)
	}

	catchupResult, err := TV.GetCatchupURL(id, srno, startFmt, endFmt)
	if err == nil && catchupResult != nil && catchupResult.IsDRM {
		mpdURL := internalUtils.SelectQuality(qualityForDrm, catchupResult.Mpd.Bitrates.Auto, catchupResult.Mpd.Bitrates.High, catchupResult.Mpd.Bitrates.Medium, catchupResult.Mpd.Bitrates.Low)
		if mpdURL == "" {
			candidates := []string{
				catchupResult.Mpd.Bitrates.High,
				catchupResult.Mpd.Bitrates.Auto,
				catchupResult.Mpd.Bitrates.Medium,
				catchupResult.Mpd.Bitrates.Low,
				catchupResult.Mpd.Result,
			}
			for _, candidate := range candidates {
				if candidate != "" {
					mpdURL = candidate
					break
				}
			}
		}

		if mpdURL != "" {
			encMpdUrl, encErr := secureurl.EncryptURL(mpdURL)
			if encErr == nil {
				licenseUrl := ""
				if catchupResult.Mpd.Key != "" {
					encKey, keyErr := secureurl.EncryptURL(catchupResult.Mpd.Key)
					if keyErr == nil {
						licenseUrl = "/drm?auth=" + encKey + "&channel_id=" + id + "&channel=" + encMpdUrl
					}
				}

				if catchupResult.AlgoName == "timesplay" {
					return c.Render("views/player_drm", fiber.Map{
						"play_url":     mpdURL,
						"license_url":  licenseUrl,
						"channel_host": "",
						"channel_path": "",
					})
				}

				parsedTvUrl, parseErr := url.Parse(mpdURL)
				if parseErr == nil {
					tvUrlSplit := strings.Split(parsedTvUrl.Path, "/")
					if len(tvUrlSplit) > 1 {
						tvUrlPath, pathErr := secureurl.EncryptURL(strings.Join(tvUrlSplit[:len(tvUrlSplit)-1], "/") + "/")
						tvUrlHost, hostErr := secureurl.EncryptURL(parsedTvUrl.Host)
						if pathErr == nil && hostErr == nil {
							return c.Render("views/player_drm", fiber.Map{
								"play_url":     "/render.mpd?auth=" + encMpdUrl,
								"license_url":  licenseUrl,
								"channel_host": tvUrlHost,
								"channel_path": tvUrlPath,
							})
						}
					}
				}
			}
		}
	}

	return c.Render("views/player_hls", fiber.Map{
		"play_url":   playURL,
		"is_catchup": true,
	})
}

func getCatchupEPG(id string, offset int) ([]map[string]interface{}, error) {
	url := fmt.Sprintf(catchupEPGURL, offset, id, defaultLangID)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(url)
	req.Header.SetMethod("GET")
	req.Header.Set("Host", "jiotvapi.cdn.jio.com")
	req.Header.Set("user-agent", okhttpUserAgent)
	req.Header.Set("Accept-Encoding", "gzip")

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	client := &fasthttp.Client{}
	if err := client.Do(req, resp); err != nil {
		return nil, err
	}

	var body []byte
	var err error

	contentEncoding := resp.Header.Peek("Content-Encoding")
	if bytes.Contains(contentEncoding, []byte("gzip")) {
		body, err = resp.BodyGunzip()
		if err != nil {
			return nil, err
		}
	} else {
		body = resp.Body()
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if epg, ok := result["epg"].([]interface{}); ok {
		epgList := make([]map[string]interface{}, len(epg))
		for i, v := range epg {
			if m, ok := v.(map[string]interface{}); ok {
				if start, ok := m["startEpoch"].(float64); ok {
					m["startEpoch"] = int64(start)
				}
				if end, ok := m["endEpoch"].(float64); ok {
					m["endEpoch"] = int64(end)
				}
				if srno, ok := m["srno"].(float64); ok {
					m["srno"] = fmt.Sprintf("%.0f", srno)
				}
				epgList[len(epg)-1-i] = m
			}
		}
		return epgList, nil
	}

	return nil, fmt.Errorf("epg field not found or not a list")
}
