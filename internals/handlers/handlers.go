package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rabilrbl/jiotv_go/internals/television"
	"github.com/rabilrbl/jiotv_go/internals/utils"
	"github.com/valyala/fasthttp"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
)

var (
	TV               *television.Television
	DisableTSHandler bool
)

func InitLogin() {
	DisableTSHandler = os.Getenv("JIOTV_DISABLE_TS_HANDLER") == "true"
	if DisableTSHandler {
		utils.Log.Println("TS Handler disabled!. All TS video requests will be served directly from JioTV servers.")
	}
	credentials, err := utils.GetJIOTVCredentials()
	if err != nil {
		utils.Log.Println("Login error!", err)
	} else {
		if credentials.AccessToken != "" {
			// Check validity of credentials
			go RefreshTokenIfExpired(credentials)
		}
		TV = television.NewTelevision(credentials)
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
			"Qualities": map[string]string{
				"auto":   "Quality (Auto)",
				"high":   "High",
				"medium": "Medium",
				"low":    "Low",
			},
		})
	} else {
		return c.Render("views/index", fiber.Map{
			"Channels":      channels.Result,
			"IsNotLoggedIn": !utils.CheckLoggedIn(),
			"Categories":    categoryMap,
			"Languages":     languageMap,
			"Qualities": map[string]string{
				"auto":   "Quality (Auto)",
				"high":   "High",
				"medium": "Medium",
				"low":    "Low",
			},
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
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": err,
			})
		}
	}
	// quote url
	coded_url := url.QueryEscape(liveResult.Auto)
	return c.Redirect("/render.m3u8?auth="+coded_url+"&channel_key_id="+id, fiber.StatusFound)
}

func LiveHighHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	// remove suffix .m3u8 if exists
	id = strings.Replace(id, ".m3u8", "", 1)
	liveResult, err := TV.Live(id)
	if err != nil {
		utils.Log.Println(err)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": err,
			})
		}
	}
	// quote url
	coded_url := url.QueryEscape(liveResult.High)
	return c.Redirect("/render.m3u8?auth="+coded_url+"&channel_key_id="+id, fiber.StatusFound)
}

func LiveMediumHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	// remove suffix .m3u8 if exists
	id = strings.Replace(id, ".m3u8", "", 1)
	liveResult, err := TV.Live(id)
	if err != nil {
		utils.Log.Println(err)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": err,
			})
		}
	}
	// quote url
	coded_url := url.QueryEscape(liveResult.Medium)
	return c.Redirect("/render.m3u8?auth="+coded_url+"&channel_key_id="+id, fiber.StatusFound)
}

func LiveLowHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	// remove suffix .m3u8 if exists
	id = strings.Replace(id, ".m3u8", "", 1)
	liveResult, err := TV.Live(id)
	if err != nil {
		utils.Log.Println(err)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": err,
			})
		}
	}
	// quote url
	coded_url := url.QueryEscape(liveResult.Low)
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
			if DisableTSHandler {
				return []byte(baseUrl + string(match) + "?" + params)
			}
			return []byte("/render.ts?auth=" + url.QueryEscape(baseUrl+string(match)+"?"+params))
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

func RenderTSHandler(c *fiber.Ctx) error {
	auth := c.Query("auth")
	// decode url
	decoded_url, err := url.QueryUnescape(auth)
	if err != nil {
		utils.Log.Panicln(err)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if err := proxy.Do(c, decoded_url, TV.Client); err != nil {
		return err
	}
	c.Response().Header.Del(fiber.HeaderServer)
	return nil
}

func ChannelsHandler(c *fiber.Ctx) error {
	quality := strings.TrimSpace(c.Query("q"))
	apiResponse := television.Channels()
	// hostUrl should be request URL like http://localhost:5001
	hostURL := strings.ToLower(c.Protocol()) + "://" + c.Hostname()

	// Check if the query parameter "type" is set to "m3u"
	if c.Query("type") == "m3u" {
		// Create an M3U playlist
		m3uContent := "#EXTM3U x-tvg-url=\"" + hostURL + "/epg.xml.gz\"\n"
		logoURL := hostURL + "/jtvimage"
		for _, channel := range apiResponse.Result {
			var channelURL string
			if quality != "" {
				channelURL = fmt.Sprintf("%s/live/%s/%d.m3u8", hostURL, quality, channel.ID)
			} else {
				channelURL = fmt.Sprintf("%s/live/%d.m3u8", hostURL, channel.ID)
			}
			channelLogoURL := fmt.Sprintf("%s/%s", logoURL, channel.LogoURL)
			m3uContent += fmt.Sprintf("#EXTINF:-1 tvg-id=%d tvg-name=%q tvg-logo=%q tvg-language=%q tvg-type=%q group-title=%q, %s\n%s\n",
				channel.ID, channel.Name, channelLogoURL, television.LanguageMap[channel.Language], television.CategoryMap[channel.Category], television.CategoryMap[channel.Category], channel.Name, channelURL)
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
	quality := c.Query("q")
	player_url := "/player/" + id + "?q=" + quality
	return c.Render("views/play", fiber.Map{
		"player_url": player_url,
	})
}

func PlayerHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	quality := c.Query("q")
	var play_url string
	if quality != "" {
		play_url = "/live/" + quality + "/" + id + ".m3u8"
	} else {
		play_url = "/live/" + id + ".m3u8"
	}
	return c.Render("views/flow_player", fiber.Map{
		"play_url": play_url,
	})
}

func ClapprHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	quality := c.Query("q")
	var play_url string
	if quality != "" {
		play_url = "/live/" + quality + "/" + id + ".m3u8"
	} else {
		play_url = "/live/" + id + ".m3u8"
	}
	return c.Render("views/clappr", fiber.Map{
		"play_url": play_url,
	})
}

func FaviconHandler(c *fiber.Ctx) error {
	return c.Redirect("/static/favicon.ico", fiber.StatusMovedPermanently)
}

func PlaylistHandler(c *fiber.Ctx) error {
	quality := c.Query("q")
	return c.Redirect("/channels?type=m3u&q="+quality, fiber.StatusMovedPermanently)
}

func ImageHandler(c *fiber.Ctx) error {
	url := "http://jiotv.catchup.cdn.jio.com/dare_images/images/" + c.Params("file")
	if err := proxy.Do(c, url, TV.Client); err != nil {
		return err
	}
	c.Response().Header.Del(fiber.HeaderServer)
	return nil
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
	InitLogin()
	return c.JSON(result)
}

func LoginHandler(c *fiber.Ctx) error {
	var username, password string
	if c.Method() == "GET" {
		username = c.Query("username")
		checkFieldExist("Username", username != "", c)
		password = c.Query("password")
		checkFieldExist("Password", password != "", c)
	} else if c.Method() == "POST" {
		formBody := new(LoginRequestBodyData)
		err := c.BodyParser(&formBody)
		if err != nil {
			utils.Log.Println(err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid JSON",
			})
		}
		username = formBody.Username
		checkFieldExist("Username", username != "", c)
		password = formBody.Password
		checkFieldExist("Password", password != "", c)
	}

	result, err := utils.Login(username, password)
	if err != nil {
		utils.Log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}
	InitLogin()
	return c.JSON(result)
}

func EPGHandler(c *fiber.Ctx) error {
	// if epg.xml.gz exists, return it
	if _, err := os.Stat("epg.xml.gz"); err == nil {
		return c.SendFile("epg.xml.gz", true)
	} else {
		err_message := "EPG not found. Please restart the server after setting the environment variable JIOTV_EPG to true."
		fmt.Println(err_message)
		return c.Status(fiber.StatusNotFound).SendString(err_message)
	}
}

func LoginRefreshAccessToken() error {
	utils.Log.Println("Refreshing AccessToken...")
	tokenData, err := utils.GetJIOTVCredentials()
	if err != nil {
		return err
	}

	// Prepare the request body
	requestBody := map[string]string{
		"appName":      "RJIL_JioTV",
		"deviceId":     "6fcadeb7b4b10d77",
		"refreshToken": tokenData.RefreshToken,
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		utils.Log.Fatalln(err)
		return err
	}

	// Prepare the request
	req := fasthttp.AcquireRequest()
	req.SetRequestURI("https://auth.media.jio.com/tokenservice/apis/v1/refreshtoken?langId=6")
	req.Header.SetMethod("POST")
	req.Header.Set("devicetype", "phone")
	req.Header.Set("versionCode", "315")
	req.Header.Set("os", "android")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Host", "auth.media.jio.com")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("User-Agent", "okhttp/4.2.2")
	req.Header.Set("accessToken", tokenData.AccessToken)
	req.SetBody(requestBodyJSON)

	// Send the request
	resp := fasthttp.AcquireResponse()
	client := utils.GetRequestClient()
	if err := client.Do(req, resp); err != nil {
		utils.Log.Fatalln(err)
		return err
	}

	// Check the response
	if resp.StatusCode() != fasthttp.StatusOK {
		utils.Log.Fatalln("Request failed with status code:", resp.StatusCode())
		return fmt.Errorf("Request failed with status code: %d", resp.StatusCode())
	}

	// Parse the response body
	respBody, err := resp.BodyGunzip()
	if err != nil {
		utils.Log.Fatalln(err)
		return err
	}
	var response RefreshTokenResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		utils.Log.Fatalln(err)
		return err
	}

	// Update tokenData
	if response.AccessToken != "" {
		tokenData.AccessToken = response.AccessToken
		tokenData.LastTokenRefreshTime = strconv.FormatInt(time.Now().Unix(), 10)
		err := utils.WriteJIOTVCredentials(tokenData)
		if err != nil {
			utils.Log.Fatalln(err)
			return err
		}
		TV = television.NewTelevision(tokenData)
		go RefreshTokenIfExpired(tokenData)
		return nil
	} else {
		return fmt.Errorf("AccessToken not found in response")
	}
}

func RefreshTokenIfExpired(credentials *utils.JIOTV_CREDENTIALS) {
	utils.Log.Println("Checking if AccessToken is expired...")
	lastTokenRefreshTime, err := strconv.ParseInt(credentials.LastTokenRefreshTime, 10, 64)
	if err != nil {
		utils.Log.Fatal(err)
	}
	lastTokenRefreshTimeUnix := time.Unix(lastTokenRefreshTime, 0)
	thresholdTime := lastTokenRefreshTimeUnix.Add(1*time.Hour + 50*time.Minute)

	if thresholdTime.Before(time.Now()) {
		LoginRefreshAccessToken()
	} else {
		utils.Log.Println("Refreshing AccessToken after", time.Until(thresholdTime).Truncate(time.Second))
		go utils.ScheduleFunctionCall(func() { RefreshTokenIfExpired(credentials) }, thresholdTime)
	}
}
