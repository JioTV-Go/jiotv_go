package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

func indexHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Success",
	})
}

func loginHandler(c *gin.Context) {
	username, check := c.GetQuery("username")
	if !check {
		Log.Println("Username not provided")	
		c.JSON(400, gin.H{
			"message": "Username not provided",
		})
		return
	}
	password, check := c.GetQuery("password")
	if !check {
		Log.Println("Password not provided")	
		c.JSON(400, gin.H{
			"message": "Password not provided",
		})
		return
	}
	result, err := Login(username, password)
	if err != nil {
		Log.Println(err)
		return
	}
	c.JSON(200, result)

}

func liveHandler(c *gin.Context) {
	id := c.Param("id")
	tv := getTV()
	liveResult := tv.live(id)
	// quote url
	coded_url := url.QueryEscape(liveResult)
	c.Redirect(302, "/render?auth="+coded_url+"&channel_key_id="+id)
}

func renderHandler(c *gin.Context) {
	auth, check := c.GetQuery("auth")
	if !check {
		Log.Println("Auth not provided")
		c.JSON(400, gin.H{
			"message": "Auth not provided",
		})
		return
	}
	channel_id, check := c.GetQuery("channel_key_id")
	if !check {
		Log.Println("Channel ID not provided")
		c.JSON(400, gin.H{
			"message": "Channel ID not provided",
		})
		return
	}
	// unquote url
	decoded_url, err := url.QueryUnescape(auth)
	if err != nil {
		Log.Println(err)
		return
	}
	tv := getTV()
	renderResult := tv.render(decoded_url)
	// baseUrl is the part of the url excluding suffix file.m3u8 and params is the part of the url after the suffix
	split_url_by_params := strings.Split(decoded_url, "?")
	baseUrl := split_url_by_params[0]
	pattern := `[a-z0-9=\_\-A-Z]*\.m3u8`
	re := regexp.MustCompile(pattern)
	baseUrl = re.ReplaceAllString(baseUrl, "")
	params := split_url_by_params[1]

	replacer := func(match []byte) []byte {
		if bytes.HasSuffix(match, []byte("-iframes.m3u8")) {
			return match // Skip replacements for matches with "-iframes.m3u8" suffix
		}
		switch {
		case bytes.HasSuffix(match, []byte(".m3u8")):
			return []byte("/render?auth=" + url.QueryEscape(baseUrl + string(match) + "?" + params) + "&channel_key_id=" + channel_id)
		case bytes.HasSuffix(match, []byte(".ts")):
			return []byte(baseUrl + string(match) + "?" + params)
		default:
			return match
		}
	}

	pattern = `[a-z0-9=\_\-A-Z]*\.(m3u8|ts)`
	re = regexp.MustCompile(pattern)
	renderResult = re.ReplaceAllFunc(renderResult, replacer)

	replacer_key := func(match []byte) []byte {
		switch {
		case bytes.HasSuffix(match, []byte(".key")) || bytes.HasSuffix(match, []byte(".pkey")):
			return []byte("/renderKey?auth=" + url.QueryEscape((string(match))) + "&channel_key_id=" + channel_id)
		default:
			return match
		}
	}

	pattern_key := `http[\S]+\.(pkey|key)`
	re_key := regexp.MustCompile(pattern_key)
	renderResult = re_key.ReplaceAllFunc(renderResult, replacer_key)

	c.Data(200, "application/vnd.apple.mpegurl", renderResult)
}

func renderKeyHandler(c *gin.Context) {
	channel_id, _ := c.GetQuery("channel_key_id")
	auth, _ := c.GetQuery("auth")
	// decode url
	decoded_url, err := url.QueryUnescape(auth)
	if err != nil {
		Log.Println(err)
		return
	}
	tv := getTV()
	keyResult, status := tv.renderKey(decoded_url, channel_id)
	c.Data(status, "application/octet-stream", keyResult)
}

func channelsHandler(c *gin.Context) {
	tv := getTV()
	apiResponse := tv.channels()

	// Check if the query parameter "type" is set to "m3u"
	if c.Query("type") == "m3u" {
		// Create an M3U playlist
		m3uContent := "#EXTM3U\n"
		hostURL := "http://localhost:5001"
		for _, channel := range apiResponse.Result {
			channelURL := fmt.Sprintf("%s/live/%d", hostURL, channel.ID)
			m3uContent += fmt.Sprintf("#EXTINF:-1,%s\n%s\n", channel.Name, channelURL)
		}

		// Set the Content-Disposition header for file download
		c.Header("Content-Disposition", "attachment; filename=jiotv_playlist.m3u")
		c.Header("Content-Type", "application/vnd.apple.mpegurl") // Set the video M3U MIME type
		c.String(http.StatusOK, m3uContent)
		return
	}

	c.JSON(http.StatusOK, apiResponse)
}
