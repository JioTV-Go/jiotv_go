package handlers

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/rabilrbl/jiotv_go/internals/television"
	"github.com/rabilrbl/jiotv_go/internals/utils"

	"github.com/gofiber/fiber/v2"
)

var TV *television.Television

type LoginRequestBodyData struct {
	Username string `json:"username" xml:"username" form:"username"`
	Password string `json:"password" xml:"password" form:"password"`
}

type LoginSendOTPRequestBodyData struct {
	MobileNumber string `json:"number" xml:"number" form:"number"`
}

type LoginVerifyOTPRequestBodyData struct {
	MobileNumber string `json:"number" xml:"number" form:"number"`
	OTP          string `json:"otp" xml:"otp" form:"otp"`
}

func Init() {
	credentials, err := utils.GetLoginCredentials()
	if err != nil {
		utils.Log.Println("Login error!")
	} else {
		TV = television.NewTelevision(credentials["accessToken"], credentials["ssoToken"], credentials["crm"], credentials["uniqueId"])
	}
}

func IndexHandler(c *fiber.Ctx) error {
	channels := television.Channels()

	language := c.Query("language")
	category := c.Query("category")

	categoryMap := television.CategoryMap
	categoryMap[0] = "All Categories"
	languageMap := television.LanguageMap
	languageMap[0] = "All Languages"

	if language != "" || category != "" {
		language_int, _ := strconv.Atoi(language)
		category_int, _ := strconv.Atoi(category)
		channels_list := television.FilterChannels(channels.Result, language_int, category_int)
		return c.Render("views/index", fiber.Map{
			"Channels":      channels_list,
			"IsNotLoggedIn": !utils.CheckLoggedIn(),
			"Categories":    categoryMap,
			"Languages":     languageMap,
		})
	} else {
		return c.Render("views/index", fiber.Map{
			"Channels":      channels.Result,
			"IsNotLoggedIn": !utils.CheckLoggedIn(),
			"Categories":    categoryMap,
			"Languages":     languageMap,
		})
	}
}

func checkFieldExist(field string, check bool, c *fiber.Ctx) {
	if !check {
		utils.Log.Println(field + " not provided")
		c.Status(fiber.StatusBadRequest)
		c.JSON(fiber.Map{
			"message": field + " not provided",
		})
	}
}

func LiveHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	// remove suffix .m3u8 if exists
	id = strings.Replace(id, ".m3u8", "", 1)
	liveResult, err := TV.Live(id)
	if err != nil {
		utils.Log.Println(err)
		utils.LoginRefreshAccessToken()
		liveResult, err = TV.Live(id)
		if err != nil {
			utils.Log.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
	}
	// quote url
	coded_url := url.QueryEscape(liveResult)
	return c.Redirect("/render.m3u8?auth="+coded_url+"&channel_key_id="+id, fiber.StatusFound)
}

func RenderHandler(c *fiber.Ctx) error {
	auth := c.Query("auth")
	if auth == "" {
		utils.Log.Println("Auth not provided")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Auth not provided",
		})
	}
	channel_id := c.Query("channel_key_id")
	if channel_id == "" {
		utils.Log.Println("Channel ID not provided")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Channel ID not provided",
		})
	}
	// unquote url
	decoded_url, err := url.QueryUnescape(auth)
	if err != nil {
		utils.Log.Println(err)
		return c.SendStatus(fiber.StatusInternalServerError)
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
		switch {
		case bytes.HasSuffix(match, []byte(".m3u8")):
			return []byte("/render.m3u8?auth=" + url.QueryEscape(baseUrl+string(match)+"?"+params) + "&channel_key_id=" + channel_id)
		case bytes.HasSuffix(match, []byte(".ts")):
			return []byte(baseUrl + string(match) + "?" + params)
		default:
			return match
		}
	}

	pattern = `[a-z0-9=\_\-A-Z\/]*\.(m3u8|ts)`
	re = regexp.MustCompile(pattern)
	renderResult = re.ReplaceAllFunc(renderResult, replacer)

	replacer_key := func(match []byte) []byte {
		switch {
		case bytes.HasSuffix(match, []byte(".key")) || bytes.HasSuffix(match, []byte(".pkey")):
			return []byte("/render.key?auth=" + url.QueryEscape(string(match)+"?"+params) + "&channel_key_id=" + channel_id)
		default:
			return match
		}
	}

	pattern_key := `http[\S]+\.(pkey|key)`
	re_key := regexp.MustCompile(pattern_key)
	renderResult = re_key.ReplaceAllFunc(renderResult, replacer_key)

	return c.Send(renderResult)
}

func RenderKeyHandler(c *fiber.Ctx) error {
	channel_id := c.Query("channel_key_id")
	auth := c.Query("auth")
	// decode url
	decoded_url, err := url.QueryUnescape(auth)
	if err != nil {
		utils.Log.Println(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	keyResult, status := TV.RenderKey(decoded_url, channel_id)
	return c.Status(status).Send(keyResult)
}

func ChannelsHandler(c *fiber.Ctx) error {
	apiResponse := television.Channels()
	// hostUrl should be request URL like http://localhost:5001
	hostURL := strings.ToLower(c.Protocol()) + "://" + c.Hostname()

	// Check if the query parameter "type" is set to "m3u"
	if c.Query("type") == "m3u" {
		// Create an M3U playlist
		m3uContent := "#EXTM3U\n"
		logoURL := "https://jiotv.catchup.cdn.jio.com/dare_images/images"
		for _, channel := range apiResponse.Result {
			channelURL := fmt.Sprintf("%s/live/%d.m3u8", hostURL, channel.ID)
			channelLogoURL := fmt.Sprintf("%s/%s", logoURL, channel.LogoURL)
			m3uContent += fmt.Sprintf("#EXTINF:-1 tvg-name=%q tvg-logo=%q tvg-language=%q tvg-type=%q group-title=%q, %s\n%s\n",
				channel.Name, channelLogoURL, television.LanguageMap[channel.Language], television.CategoryMap[channel.Category], television.CategoryMap[channel.Category], channel.Name, channelURL)
		}

		// Set the Content-Disposition header for file download
		c.Set("Content-Disposition", "attachment; filename=jiotv_playlist.m3u")
		c.Set("Content-Type", "application/vnd.apple.mpegurl") // Set the video M3U MIME type
		return c.SendString(m3uContent)
	}

	for i, channel := range apiResponse.Result {
		apiResponse.Result[i].URL = fmt.Sprintf("%s/live/%d", hostURL, channel.ID)
	}

	return c.JSON(apiResponse)
}

func PlayHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	player_url := "/player/" + id
	return c.Render("views/play", fiber.Map{
		"player_url": player_url,
	})
}

func PlayerHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	play_url := "/live/" + id + ".m3u8"
	return c.Render("views/flow_player", fiber.Map{
		"play_url": play_url,
	})
}

func ClapprHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	play_url := "/live/" + id + ".m3u8"
	return c.Render("views/clappr", fiber.Map{
		"play_url": play_url,
	})
}

func FaviconHandler(c *fiber.Ctx) error {
	return c.Redirect("/static/favicon.ico", fiber.StatusMovedPermanently)
}

func PlaylistHandler(c *fiber.Ctx) error {
	return c.Redirect("/channels?type=m3u", fiber.StatusMovedPermanently)
}

func LoginSendOTPHandler(c *fiber.Ctx) error {
	// get mobile number from post request
	formBody := new(LoginSendOTPRequestBodyData)
	err := c.BodyParser(&formBody)
	if err != nil {
		utils.Log.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid JSON",
		})
	}
	mobileNumber := formBody.MobileNumber
	checkFieldExist("Mobile Number", mobileNumber != "", c)

	result, err := utils.LoginSendOTP(mobileNumber)
	if err != nil {
		utils.Log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err,
		})
	}
	return c.JSON(fiber.Map{
		"status": result,
	})
}

func LoginVerifyOTPHandler(c *fiber.Ctx) error {
	// get mobile number and otp from post request
	formBody := new(LoginVerifyOTPRequestBodyData)
	err := c.BodyParser(&formBody)
	if err != nil {
		utils.Log.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid JSON",
		})
	}
	mobileNumber := formBody.MobileNumber
	checkFieldExist("Mobile Number", mobileNumber != "", c)
	otp := formBody.OTP
	checkFieldExist("OTP", otp != "", c)

	result, err := utils.LoginVerifyOTP(mobileNumber, otp)
	if err != nil {
		utils.Log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}
	Init()
	return c.JSON(result)
}
