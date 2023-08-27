package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/rabilrbl/jiotv_go/internals/utils"
	"github.com/rabilrbl/jiotv_go/internals/television"
)

var TV *television.Television

func Init() {
	credentials, err := utils.GetLoginCredentials()
	if err != nil {
		utils.Log.Println("Login error!")
	} else {
		TV = television.NewTelevision(credentials["ssoToken"], credentials["crm"], credentials["uniqueId"])	
	}
}

func IndexHandler(c *gin.Context) {
	channels := television.Channels()
	c.HTML(http.StatusOK, "index.html", gin.H{
		"Channels": channels.Result,
		"IsNotLoggedIn": !utils.CheckLoggedIn(),
	})
}

func LoginHandler(c *gin.Context) {
	username, check := c.GetQuery("username")
	if !check {
		utils.Log.Println("Username not provided")	
		c.JSON(400, gin.H{
			"message": "Username not provided",
		})
		return
	}
	password, check := c.GetQuery("password")
	if !check {
		utils.Log.Println("Password not provided")	
		c.JSON(400, gin.H{
			"message": "Password not provided",
		})
		return
	}
	
	// Strip trailing "-encoded" from password, if sent through Login form
	if strings.HasSuffix(password, "-encoded") {
		password = password[:len(password)-8]
		// Decode the decodeURI password
		decodedPassword, err := url.QueryUnescape(password)
		if err != nil {
			utils.Log.Println("Failed to decode password")
			c.JSON(400, gin.H{
				"message": "Failed to decode password",
			})
			return
		}
		password = decodedPassword
	}
	
	result, err := utils.Login(username, password)
	if err != nil {
		utils.Log.Println(err)
		c.JSON(500, gin.H{
			"message": "Internal server error",
		})
		return
	}
	Init()
	c.JSON(200, result)

}

func LiveHandler(c *gin.Context) {
	id := c.Param("id")
	// remove suffix .m3u8 if exists
	id = strings.Replace(id, ".m3u8", "", -1)
	liveResult := TV.Live(id)
	// quote url
	coded_url := url.QueryEscape(liveResult)
	c.Redirect(302, "/render?auth="+coded_url+"&channel_key_id="+id)
}

func RenderHandler(c *gin.Context) {
	auth, check := c.GetQuery("auth")
	if !check {
		utils.Log.Println("Auth not provided")
		c.JSON(400, gin.H{
			"message": "Auth not provided",
		})
		return
	}
	channel_id, check := c.GetQuery("channel_key_id")
	if !check {
		utils.Log.Println("Channel ID not provided")
		c.JSON(400, gin.H{
			"message": "Channel ID not provided",
		})
		return
	}
	// unquote url
	decoded_url, err := url.QueryUnescape(auth)
	if err != nil {
		utils.Log.Println(err)
		return
	}
	renderResult := TV.Render(decoded_url)
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

func RenderKeyHandler(c *gin.Context) {
	channel_id, _ := c.GetQuery("channel_key_id")
	auth, _ := c.GetQuery("auth")
	// decode url
	decoded_url, err := url.QueryUnescape(auth)
	if err != nil {
		utils.Log.Println(err)
		return
	}
	keyResult, status := TV.RenderKey(decoded_url, channel_id)
	c.Data(status, "application/octet-stream", keyResult)
}

func ChannelsHandler(c *gin.Context) {
	apiResponse := television.Channels()
	// hostUrl should be request URL like http://localhost:5001
	hostURL :=  strings.ToLower(c.Request.Proto[0:strings.Index(c.Request.Proto, "/")]) + "://" + c.Request.Host

	// Check if the query parameter "type" is set to "m3u"
	if c.Query("type") == "m3u" {
		// Create an M3U playlist
		m3uContent := "#EXTM3U\n"
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

	for i, channel := range apiResponse.Result {
		apiResponse.Result[i].URL = fmt.Sprintf("%s/live/%d", hostURL, channel.ID)
	}

	c.JSON(http.StatusOK, apiResponse)
}

func PlayHandler(c *gin.Context) {
	id := c.Param("id")
	player_url := "/player/" + id
	c.HTML(http.StatusOK, "play.html", gin.H{
		"player_url": player_url,
	})
}

func PlayerHandler(c *gin.Context) {
	id := c.Param("id")
	play_url := "/live/" + id + ".m3u8"
	c.HTML(http.StatusOK, "player.html", gin.H{
		"play_url": play_url,
	})
}
