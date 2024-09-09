package handlers

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/rabilrbl/jiotv_go/v3/pkg/secureurl"
	"github.com/rabilrbl/jiotv_go/v3/pkg/utils"
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

	channel, err := secureurl.EncryptURL(tv_url)
	if err != nil {
		utils.Log.Panicln(err)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": err,
		})
	}

	tv_url = strings.Replace(tv_url, "https://jiotvmblive.cdn.jio.com", "", 1)
	return c.Render("views/flow_player_drm", fiber.Map{
		"play_url":    tv_url,
		"license_url": "/drm?auth=" + enc_key + "&channel_id=" + channelID + "&channel=" + channel,
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

	// Print ALL request headers
	utils.Log.Println("Request headers:", c.Request().Header.String())

	if err := proxy.Do(c, decoded_url, TV.Client); err != nil {
		return err
	}

	c.Response().Header.Del(fiber.HeaderServer)
	return nil
}

// BpkProxyHandler handles BPK proxy routes /bpk/:channelID
func BpkProxyHandler(c *fiber.Ctx) error {
	c.Request().Header.Set("Host", "jiotvmblive.cdn.jio.com")
	c.Request().Header.Set("User-Agent", "plaYtv/7.1.3 (Linux;Android 13) ExoPlayerLib/2.11.7")

	// Request path with query params
	url := "https://jiotvmblive.cdn.jio.com" + c.Path() + "?" + string(c.Request().URI().QueryString())
	if url[len(url)-1:] == "?" {
		url = url[:len(url)-1]
	}

	if err := proxy.Do(c, url, TV.Client); err != nil {
		return err
	}
	c.Response().Header.Del(fiber.HeaderServer)

	// Delete Domain from cookies
	if c.Response().Header.Peek("Set-Cookie") != nil {
		cookies := c.Response().Header.Peek("Set-Cookie")
		c.Response().Header.Del("Set-Cookie")

		cookies = bytes.Replace(cookies, []byte("Domain=jiotvmblive.cdn.jio.com;"), []byte(""), 1)
		// Modify path in cookies
		cookies = bytes.Replace(cookies, []byte("path=/"), []byte("path=/bpk-tv/"), 1)

		// Modify Set-Cookie header
		c.Response().Header.SetBytesV("Set-Cookie", cookies)
	}

	return nil
}
