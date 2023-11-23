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

	"github.com/rabilrbl/jiotv_go/v2/pkg/television"
	"github.com/rabilrbl/jiotv_go/v2/pkg/utils"
	"github.com/valyala/fasthttp"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
)

var (
	TV               *television.Television
	DisableTSHandler bool
	isLogoutDisabled  bool
)

// Init initializes the necessary operations required for the handlers to work.
func Init() {
	DisableTSHandler = os.Getenv("JIOTV_DISABLE_TS_HANDLER") == "true"
	isLogoutDisabled = os.Getenv("JIOTV_LOGOUT") == "false"
	if DisableTSHandler {
		utils.Log.Println("TS Handler disabled!. All TS video requests will be served directly from JioTV servers.")
	}
	// Get credentials from file
	credentials, err := utils.GetJIOTVCredentials()
	// Initialize TV object with nil credentials
	TV = television.New(nil)
	if err != nil {
		utils.Log.Println("Login error!", err)
	} else {
		// If AccessToken is present, check for its validity and schedule a refresh if required
		if credentials.AccessToken != "" {
			// Check validity of credentials
			go RefreshTokenIfExpired(credentials)
		}
		// Initialize TV object with credentials
		TV = television.New(credentials)
	}
}

// ErrorMessageHandler handles error messages
// Responds with 500 status code and error message
func ErrorMessageHandler(c *fiber.Ctx, err error) error {
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return nil
}

// IndexHandler handles the index page for `/` route
func IndexHandler(c *fiber.Ctx) error {
	// Get all channels
	channels := television.Channels()

	// Get language and category from query params
	language := c.Query("language")
	category := c.Query("category")

	// Context data for index page
	indexContext := fiber.Map{
		"Channels":      nil,
		"IsNotLoggedIn": !utils.CheckLoggedIn(),
		"Categories":    television.CategoryMap,
		"Languages":     television.LanguageMap,
		"Qualities": map[string]string{
			"auto":   "Quality (Auto)",
			"high":   "High",
			"medium": "Medium",
			"low":    "Low",
		},
	}

	// Filter channels by language and category if provided
	if language != "" || category != "" {
		language_int, err := strconv.Atoi(language)
		if err != nil {
			return ErrorMessageHandler(c, err)
		}
		category_int, err := strconv.Atoi(category)
		if err != nil {
			return ErrorMessageHandler(c, err)
		}
		channels_list := television.FilterChannels(channels.Result, language_int, category_int)
		indexContext["Channels"] = channels_list
		return c.Render("views/index", indexContext)
	}
	// If language and category are not provided, return all channels
	indexContext["Channels"] = channels.Result
	return c.Render("views/index", indexContext)
}

// checkFieldExist checks if the field is provided in the request.
// If not, send a bad request response
func checkFieldExist(field string, check bool, c *fiber.Ctx) error {
	if !check {
		utils.Log.Println(field + " not provided")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": field + " not provided",
		})
	}
	return nil
}

// LiveHandler handles the live channel stream route `/live/:id.m3u8`.
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
	// quote url as it will be passed as a query parameter
	// It is required to quote the url as it may contain special characters like ? and &
	coded_url := url.QueryEscape(liveResult.Auto)
	return c.Redirect("/render.m3u8?auth="+coded_url+"&channel_key_id="+id, fiber.StatusFound)
}

// LiveQualityHandler handles the live channel stream route `/live/:quality/:id.m3u8`.
func LiveQualityHandler(c *fiber.Ctx) error {
	quality := c.Params("quality")
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
	// Channels with following IDs output audio only m3u8 when quality level is enforced
	if id == "1349" || id == "1322" {
		quality = "auto"
	}
	var liveURL string
	// select quality level based on query parameter
	switch quality {
	case "high", "h":
		liveURL = liveResult.High
	case "medium", "med", "m":
		liveURL = liveResult.Medium
	case "low", "l":
		liveURL = liveResult.Low
	default:
		liveURL = liveResult.Auto
	}
	// quote url as it will be passed as a query parameter
	// It is required to quote the url as it may contain special characters like ? and &
	coded_url := url.QueryEscape(liveURL)
	return c.Redirect("/render.m3u8?auth="+coded_url+"&channel_key_id="+id, fiber.StatusFound)
}

// RenderHandler handles M3U8 file for modification
// This handler shall replace JioTV server URLs with our own server URLs
func RenderHandler(c *fiber.Ctx) error {
	// URL to be rendered
	auth := c.Query("auth")
	if auth == "" {
		utils.Log.Println("Auth not provided")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Auth not provided",
		})
	}
	// Channel ID to be used for key rendering
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
	// Pattern to match file names ending with .m3u8
	pattern := `[a-z0-9=\_\-A-Z]*\.m3u8`
	re := regexp.MustCompile(pattern)
	// Add baseUrl to all the file names ending with .m3u8
	baseUrl = re.ReplaceAllString(baseUrl, "")
	params := split_url_by_params[1]

	// replacer replaces all the file names ending with .m3u8 and .ts with our own server URLs
	// More info: https://golang.org/pkg/regexp/#Regexp.ReplaceAllFunc
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

	// Pattern to match file names ending with .m3u8 and .ts
	pattern = `[a-z0-9=\_\-A-Z\/]*\.(m3u8|ts)`
	re = regexp.MustCompile(pattern)
	// Execute replacer function on renderResult
	renderResult = re.ReplaceAllFunc(renderResult, replacer)

	// replacer_key replaces all the URLs ending with .key and .pkey with our own server URLs
	replacer_key := func(match []byte) []byte {
		switch {
		case bytes.HasSuffix(match, []byte(".key")) || bytes.HasSuffix(match, []byte(".pkey")):
			return []byte("/render.key?auth=" + url.QueryEscape(string(match)+"?"+params) + "&channel_key_id=" + channel_id)
		default:
			return match
		}
	}

	// Pattern to match URLs ending with .key and .pkey
	pattern_key := `http[\S]+\.(pkey|key)`
	re_key := regexp.MustCompile(pattern_key)
	// Execute replacer_key function on renderResult
	renderResult = re_key.ReplaceAllFunc(renderResult, replacer_key)

	return c.Send(renderResult)
}

// RenderKeyHandler requests m3u8 key from JioTV server
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

// RenderTSHandler loads TS file from JioTV server
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

// ChannelsHandler fetch all channels from JioTV API
// Also to generate M3U playlist
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

// PlayHandler loads HTML Page with video player iframe embedded with video URL
// URL is generated from the channel ID
func PlayHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	quality := c.Query("q")
	player_url := "/player/" + id + "?q=" + quality
	return c.Render("views/play", fiber.Map{
		"player_url": player_url,
	})
}

// PlayerHandler loads Web Player to stream live TV
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

// ClapprHandler is previous (old) Web Player to stream live TV
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

// FaviconHandler Responds for favicon.ico request
func FaviconHandler(c *fiber.Ctx) error {
	return c.Redirect("/static/favicon.ico", fiber.StatusMovedPermanently)
}

// PlaylistHandler is the route for generating M3U playlist only
// For user convenience, redirect to /channels?type=m3u
func PlaylistHandler(c *fiber.Ctx) error {
	quality := c.Query("q")
	return c.Redirect("/channels?type=m3u&q="+quality, fiber.StatusMovedPermanently)
}

// ImageHandler loads image from JioTV server
func ImageHandler(c *fiber.Ctx) error {
	url := "http://jiotv.catchup.cdn.jio.com/dare_images/images/" + c.Params("file")
	if err := proxy.Do(c, url, TV.Client); err != nil {
		return err
	}
	c.Response().Header.Del(fiber.HeaderServer)
	return nil
}

// LoginSendOTPHandler sends OTP for login
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

// LoginVerifyOTPHandler verifies OTP and login
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

// LoginPasswordHandler is used to login with password
func LoginPasswordHandler(c *fiber.Ctx) error {
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
	Init()
	return c.JSON(result)
}

// LogoutHandler is used to logout
func LogoutHandler(c *fiber.Ctx) error {
	if !isLogoutDisabled {
		err := utils.Logout()
		if err != nil {
			utils.Log.Println(err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Internal server error",
			})
		}
		Init()
	}
	return c.Redirect("/", fiber.StatusFound)
}

// EPGHandler handles EPG requests
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

// LoginRefreshAccessToken Function is used to refresh AccessToken
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
		TV = television.New(tokenData)
		go RefreshTokenIfExpired(tokenData)
		return nil
	} else {
		return fmt.Errorf("AccessToken not found in response")
	}
}

// RefreshTokenIfExpired Function is used to handle AccessToken refresh
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
