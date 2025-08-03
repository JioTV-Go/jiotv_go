package handlers

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/jiotv-go/jiotv_go/v3/pkg/secureurl"
	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
	"github.com/valyala/fasthttp"

	"github.com/gofiber/fiber/v2"
)

// getDrmMpd returns required properties for rendering DRM MPD
func getDrmMpd(channelID, quality string) (*DrmMpdOutput, error) {
	// Get live stream URL from JioTV API
	liveResult, err := TV.Live(channelID)
	if err != nil {
		return nil, err
	}
	enc_key, err := secureurl.EncryptURL(liveResult.Mpd.Key)
	if err != nil {
		utils.Log.Panicln(err)
		return nil, err
	}

	var tv_url string
	switch quality {
	case "high", "h":
		tv_url = liveResult.Mpd.Bitrates.High
	case "medium", "med", "m":
		tv_url = liveResult.Mpd.Bitrates.Medium
	case "low", "l":
		tv_url = liveResult.Mpd.Bitrates.Low
	default:
		tv_url = liveResult.Mpd.Bitrates.Auto
	}

	channel_enc_url, err := secureurl.EncryptURL(tv_url)
	if err != nil {
		utils.Log.Panicln(err)
		return nil, err
	}

	parsedTvUrl, err := url.Parse(tv_url)
	if err != nil {
		utils.Log.Panicln(err)
		return nil, err
	}
	tv_url_split := strings.Split(parsedTvUrl.Path, "/")
	tv_url_path, err := secureurl.EncryptURL(strings.Join(tv_url_split[:len(tv_url_split)-1], "/") + "/")
	if err != nil {
		utils.Log.Panicln(err)
		return nil, err
	}

	tv_url_host, err := secureurl.EncryptURL(parsedTvUrl.Host)
	if err != nil {
		utils.Log.Panicln(err)
		return nil, err
	}

	return &DrmMpdOutput{
		PlayUrl:     "/render.mpd?auth=" + channel_enc_url,
		LicenseUrl:  "/drm?auth=" + enc_key + "&channel_id=" + channelID + "&channel=" + channel_enc_url,
		Tv_url_host: tv_url_host,
		Tv_url_path: tv_url_path,
	}, nil
}

// LiveMpdHandler handles live stream routes /mpd/:channelID
func LiveMpdHandler(c *fiber.Ctx) error {
	// Get channel ID from URL
	channelID := c.Params("channelID")
	quality := c.Query("q")

	drmMpdOutput, err := getDrmMpd(channelID, quality)
	if err != nil {
		utils.Log.Panicln(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err,
		})
	}

	return c.Render("views/flow_player_drm", fiber.Map{
		"play_url":     drmMpdOutput.PlayUrl,
		"license_url":  drmMpdOutput.LicenseUrl,
		"channel_host": drmMpdOutput.Tv_url_host,
		"channel_path": drmMpdOutput.Tv_url_path,
	})
}

func generateDateTime() string {
	currentTime := time.Now()
	formattedDateTime := fmt.Sprintf("%02d%02d%02d%02d%02d%03d",
		currentTime.Year()%100, currentTime.Month(), currentTime.Day(),
		currentTime.Hour(), currentTime.Minute(),
		currentTime.Nanosecond()/1000000)
	return formattedDateTime
}

// DRMKeyHandler handles DRM key routes /drm?auth=xxx
func DRMKeyHandler(c *fiber.Ctx) error {
	// Get auth token from URL
	auth := c.Query("auth")
	channel := c.Query("channel")
	channel_id := c.Query("channel_id")

	decoded_channel, err := secureurl.DecryptURL(channel)
	if err != nil {
		utils.Log.Panicln(err)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": err,
		})
	}

	// Make a HEAD request to the decoded_channel to get the cookies
	client := utils.GetRequestClient()
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(decoded_channel)
	req.Header.SetMethod("HEAD")

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Perform the HTTP HEAD request
	if err := client.Do(req, resp); err != nil {
		utils.Log.Panic(err)
	}

	// Get the cookies from the response
	cookies := resp.Header.Peek("Set-Cookie")

	// Set the cookies in the request context
	c.Request().Header.Set("Cookie", string(cookies))

	decoded_url, err := secureurl.DecryptURL(auth)
	if err != nil {
		utils.Log.Panicln(err)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": err,
		})
	}

	// Remove headers that might interfere
	c.Request().Header.Del("Accept")
	c.Request().Header.Del("Origin")

	// Make HTTP request using Television
	resBody, statusCode, err := TV.RequestDRMKey(decoded_url, channel_id)
	if err != nil {
		return err
	}

	// If we get a 403 (Forbidden), try refreshing tokens and retry once
	if statusCode == fiber.StatusForbidden {
		if err := EnsureFreshTokens(); err != nil {
			utils.Log.Printf("Failed to refresh tokens after 403: %v", err)
		} else {
			// Retry the request once after refreshing tokens
			utils.Log.Println("Retrying DRM key request after token refresh")
			resBody, statusCode, err = TV.RequestDRMKey(decoded_url, channel_id)
			if err != nil {
				return err
			}
		}
	}

	// Set response
	c.Status(statusCode)
	c.Response().Header.Del(fiber.HeaderServer)
	c.Response().SetBody(resBody)

	return nil
}

// MpdHandler handles BPK proxy routes /bpk/:channelID
func MpdHandler(c *fiber.Ctx) error {
	proxyUrl := c.Query("auth")
	if proxyUrl == "" {
		c.Status(fiber.StatusBadRequest)
		return fmt.Errorf("auth query param is required")
	}

	decryptedUrl, err := secureurl.DecryptURL(proxyUrl)
	if err != nil {
		utils.Log.Panicln(err)
		return err
	}
	parsedUrl, err := url.Parse(decryptedUrl)
	if err != nil {
		utils.Log.Panicln(err)
		return err
	}

	proxyHost := parsedUrl.Host

	// Set Host header for the request context (used for cookies processing)
	c.Request().Header.Set("Host", proxyHost)

	// Make HTTP request using Television
	resBody, statusCode, err := TV.RequestMPD(decryptedUrl)
	if err != nil {
		return err
	}

	// If we get a 403 (Forbidden), try refreshing tokens and retry once
	if statusCode == fiber.StatusForbidden {
		if err := EnsureFreshTokens(); err != nil {
			utils.Log.Printf("Failed to refresh tokens after 403: %v", err)
		} else {
			// Retry the request once after refreshing tokens
			utils.Log.Println("Retrying MPD request after token refresh")
			resBody, statusCode, err = TV.RequestMPD(decryptedUrl)
			if err != nil {
				return err
			}
		}
	}

	// Set the status code for the response
	c.Status(statusCode)
	c.Response().Header.Del(fiber.HeaderServer)

	// Delete Domain from cookies processing would be handled here if cookies were present
	// Note: Since we're using TV.RequestMPD, cookies handling is simplified
	// but we keep the logic for processing BaseURL patterns

	basePathPattern := `<BaseURL>(.*)<\/BaseURL>`
	re := regexp.MustCompile(basePathPattern)
	// check for match
	if re.Match(resBody) {
		resBody = re.ReplaceAllFunc(resBody, func(match []byte) []byte {
			return []byte("<BaseURL>/render.dash/dash/</BaseURL>")
		})
	} else {
		pattern := `<Period(\s+[^>]*?)?\s*\/?>`
		re = regexp.MustCompile(pattern)
		resBody = re.ReplaceAllFunc(resBody, func(match []byte) []byte {
			return []byte(fmt.Sprintf("%s\n<BaseURL>/render.dash/</BaseURL>", match))
		})
	}

	c.Response().SetBody(resBody)

	return nil
}

// DashHandler
func DashHandler(c *fiber.Ctx) error {
	proxyHost := c.Query("host")
	proxyPath := c.Query("path")

	if proxyHost == "" || proxyPath == "" {
		c.Status(fiber.StatusBadRequest)
		return fmt.Errorf("host and path query params are required")
	}

	// decode the URL
	proxyHost, err := secureurl.DecryptURL(proxyHost)
	if err != nil {
		utils.Log.Panicln(err)
		return err
	}
	proxyPath, err = secureurl.DecryptURL(proxyPath)
	if err != nil {
		utils.Log.Panicln(err)
		return err
	}

	// remove render.dash from c.Request().URI().RequestURI()
	requestUri := bytes.Replace(c.Request().URI().RequestURI(), []byte("/render.dash"), []byte(""), 1)

	proxyUrl := fmt.Sprintf("https://%s%s/%s", proxyHost, proxyPath, requestUri)

	// Make HTTP request using Television
	resBody, statusCode, err := TV.RequestDashSegment(proxyUrl)
	if err != nil {
		return err
	}

	// If we get a 403 (Forbidden), try refreshing tokens and retry once
	if statusCode == fiber.StatusForbidden {
		if err := EnsureFreshTokens(); err != nil {
			utils.Log.Printf("Failed to refresh tokens after 403: %v", err)
		} else {
			// Retry the request once after refreshing tokens
			utils.Log.Println("Retrying DASH segment request after token refresh")
			resBody, statusCode, err = TV.RequestDashSegment(proxyUrl)
			if err != nil {
				return err
			}
		}
	}

	// Set response
	c.Status(statusCode)
	c.Response().Header.Del(fiber.HeaderServer)
	c.Response().SetBody(resBody)

	return nil
}
