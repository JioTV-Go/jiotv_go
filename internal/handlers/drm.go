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
	"github.com/gofiber/fiber/v2/middleware/proxy"
)

// LiveMpdHandler handles live stream routes /mpd/:channelID
func LiveMpdHandler(c *fiber.Ctx) error {
	// Get channel ID from URL
	channelID := c.Params("channelID")
	quality := c.Query("q")
	// Get live stream URL from JioTV API
	liveResult, err := TV.Live(channelID)
	if err != nil {
		return err
	}
	enc_key, err := secureurl.EncryptURL(liveResult.Mpd.Key)
	if err != nil {
		utils.Log.Panicln(err)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": err,
		})
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
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": err,
		})
	}

	parsedTvUrl, err := url.Parse(tv_url)
	if err != nil {
		utils.Log.Panicln(err)
		return err
	}
	tv_url_split := strings.Split(parsedTvUrl.Path, "/")
	tv_url_path, err := secureurl.EncryptURL(strings.Join(tv_url_split[:len(tv_url_split)-1], "/") + "/")
	if err != nil {
		utils.Log.Panicln(err)
		return err
	}

	tv_url_host, err := secureurl.EncryptURL(parsedTvUrl.Host)
	if err != nil {
		utils.Log.Panicln(err)
		return err
	}

	return c.Render("views/flow_player_drm", fiber.Map{
		"play_url":     "/render.mpd?auth=" + channel_enc_url,
		"license_url":  "/drm?auth=" + enc_key + "&channel_id=" + channelID + "&channel=" + channel_enc_url,
		"channel_host": tv_url_host,
		"channel_path": tv_url_path,
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

	// Perform the HTTP GET request
	if err := client.Do(req, resp); err != nil {
		utils.Log.Panic(err)
	}

	// Get the cookies from the response
	cookies := resp.Header.Peek("Set-Cookie")

	// Set the cookies in the request
	c.Request().Header.Set("Cookie", string(cookies))

	decoded_url, err := secureurl.DecryptURL(auth)
	if err != nil {
		utils.Log.Panicln(err)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": err,
		})
	}

	// Add headers to the request
	c.Request().Header.Set("accesstoken", TV.AccessToken)
	c.Request().Header.Set("Connection", "keep-alive")
	c.Request().Header.Set("os", "android")
	c.Request().Header.Set("appName", "RJIL_JioTV")
	c.Request().Header.Set("subscriberId", TV.Crm)
	c.Request().Header.Set("Host", "tv.media.jio.com")
	c.Request().Header.Set("User-Agent", "plaYtv/7.1.3 (Linux;Android 13) ExoPlayerLib/2.11.7")
	c.Request().Header.Set("ssotoken", TV.SsoToken)
	c.Request().Header.Set("x-platform", "android")
	c.Request().Header.Set("srno", generateDateTime())
	c.Request().Header.Set("crmid", TV.Crm)
	c.Request().Header.Set("channelid", channel_id)
	c.Request().Header.Set("uniqueId", TV.UniqueID)
	c.Request().Header.Set("versionCode", "330")
	c.Request().Header.Set("usergroup", "tvYR7NSNn7rymo3F")
	c.Request().Header.Set("devicetype", "phone")
	c.Request().Header.Set("Accept-Encoding", "gzip, deflate")
	c.Request().Header.Set("osVersion", "13")
	c.Request().Header.Set("deviceId", utils.GetDeviceID())
	c.Request().Header.Set("Content-Type", "application/octet-stream")

	// Delete User-Agent header from the request
	c.Request().Header.Del("Accept")
	c.Request().Header.Del("Origin")

	if err := proxy.Do(c, decoded_url, TV.Client); err != nil {
		return err
	}

	c.Response().Header.Del(fiber.HeaderServer)
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

	// proxyQuery := parsedUrl.RawQuery

	c.Request().Header.Set("Host", proxyHost)
	c.Request().Header.Set("User-Agent", "plaYtv/7.1.3 (Linux;Android 13) ExoPlayerLib/2.11.7")

	// Request path with query params
	requestUrl := decryptedUrl
	// if requestUrl[len(requestUrl)-1:] == "?" {
	// 	requestUrl = requestUrl[:len(requestUrl)-1]
	// }

	c.Request().Header.Set("User-Agent", PLAYER_USER_AGENT)
	// remove Accept-Encoding header
	c.Request().Header.Del("Accept-Encoding")
	if err := proxy.Do(c, requestUrl, TV.Client); err != nil {
		return err
	}
	c.Response().Header.Del(fiber.HeaderServer)

	// Delete Domain from cookies
	if c.Response().Header.Peek("Set-Cookie") != nil {
		cookies := c.Response().Header.Peek("Set-Cookie")
		c.Response().Header.Del("Set-Cookie")

		cookies = bytes.Replace(cookies, []byte("Domain="+proxyHost+";"), []byte(""), 1)
		// Modify path in cookies
		cookies = bytes.Replace(cookies, []byte("path=/"), []byte("path=/render.dash"), 1)

		// Modify Set-Cookie header
		c.Response().Header.SetBytesV("Set-Cookie", cookies)
	}
	resBody := c.Response().Body()
	basePathPattern := `<BaseURL>(.*)<\/BaseURL>`
	re := regexp.MustCompile(basePathPattern)
	// check for match
	if re.Match(resBody) {
		resBody = re.ReplaceAllFunc(resBody, func(match []byte) []byte {
			return []byte("<BaseURL>/render.dash/dash/</BaseURL>")
		})
	} else {
		pattern := `<Period(.*)>`
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

	c.Request().Header.Set("User-Agent", PLAYER_USER_AGENT)

	if err := proxy.Do(c, proxyUrl, TV.Client); err != nil {
		return err
	}
	c.Response().Header.Del(fiber.HeaderServer)

	return nil
}
