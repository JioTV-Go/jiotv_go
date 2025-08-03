package handlers

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jiotv-go/jiotv_go/v3/internal/config"
	"github.com/jiotv-go/jiotv_go/v3/internal/constants/headers"
	"github.com/jiotv-go/jiotv_go/v3/internal/constants/urls"
	"github.com/jiotv-go/jiotv_go/v3/pkg/secureurl"
	"github.com/jiotv-go/jiotv_go/v3/pkg/television"
	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
)

var (
	TV               *television.Television
	DisableTSHandler bool
	isLogoutDisabled bool
	Title            string
	EnableDRM        bool
	SONY_LIST        = []string{"154", "155", "162", "289", "291", "471", "474", "476", "483", "514", "524", "525", "697", "872", "873", "874", "891", "892", "1146", "1393", "1772", "1773", "1774", "1775"}
)

const (
	REFRESH_TOKEN_URL     = urls.RefreshTokenURL
	REFRESH_SSO_TOKEN_URL = urls.RefreshSSOTokenURL
	PLAYER_USER_AGENT     = headers.UserAgentPlayTV
	REQUEST_USER_AGENT    = headers.UserAgentOkHttp
)

// Init initializes the necessary operations required for the handlers to work.
func Init() {
	if config.Cfg.Title != "" {
		Title = config.Cfg.Title
	} else {
		Title = "JioTV Go"
	}
	DisableTSHandler = config.Cfg.DisableTSHandler
	isLogoutDisabled = config.Cfg.DisableLogout
	EnableDRM = config.Cfg.DRM
	if DisableTSHandler {
		utils.Log.Println("TS Handler disabled!. All TS video requests will be served directly from JioTV servers.")
	}
	if !EnableDRM {
		utils.Log.Println("If you're not using IPTV Client. We strongly recommend enabling DRM for accessing channels without any issues! Either enable by setting environment variable JIOTV_DRM=true or by setting DRM: true in config. For more info Read https://telegram.me/jiotv_go/128")
	}
	// Generate a new device ID if not present
	utils.GetDeviceID()
	// Get credentials from file
	credentials, err := utils.GetJIOTVCredentials()
	// Initialize TV object with nil credentials initially
	TV = television.New(nil)
	if err != nil {
		utils.Log.Println("Login error!", err)
	} else {
		// If AccessToken is present, validate on first use
		if credentials.AccessToken != "" && credentials.RefreshToken == "" {
			utils.Log.Println("Warning: AccessToken present but RefreshToken is missing. Token refresh may fail.")
		}
		// If SsoToken is present, validate on first use  
		if credentials.SSOToken != "" && credentials.UniqueID == "" {
			utils.Log.Println("Warning: SSOToken present but UniqueID is missing. Token refresh may fail.")
		}
		// Initialize TV object with credentials
		TV = television.New(credentials)
	}

	// Initialize custom channels at startup if configured
	television.InitCustomChannels()
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
		"Title":         Title,
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
	
	// Ensure tokens are fresh before making API call
	if err := EnsureFreshTokens(); err != nil {
		utils.Log.Printf("Failed to ensure fresh tokens: %v", err)
		// Continue with the request - tokens might still work or it might be a custom channel
	}
	
	liveResult, err := TV.Live(id)
	if err != nil {
		utils.Log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err,
		})
	}

	// Check if liveResult.Bitrates.Auto is empty
	if liveResult.Bitrates.Auto == "" {
		error_message := "No stream found for channel id: " + id + "Status: " + liveResult.Message
		utils.Log.Println(error_message)
		utils.Log.Println(liveResult)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": error_message,
		})
	}
	// quote url as it will be passed as a query parameter
	// It is required to quote the url as it may contain special characters like ? and &
	coded_url, err := secureurl.EncryptURL(liveResult.Bitrates.Auto)
	if err != nil {
		utils.Log.Println(err)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": err,
		})
	}
	return c.Redirect("/render.m3u8?auth="+coded_url+"&channel_key_id="+id, fiber.StatusFound)
}

// LiveQualityHandler handles the live channel stream route `/live/:quality/:id.m3u8`.
func LiveQualityHandler(c *fiber.Ctx) error {
	quality := c.Params("quality")
	id := c.Params("id")
	// remove suffix .m3u8 if exists
	id = strings.Replace(id, ".m3u8", "", 1)
	
	// Ensure tokens are fresh before making API call
	if err := EnsureFreshTokens(); err != nil {
		utils.Log.Printf("Failed to ensure fresh tokens: %v", err)
		// Continue with the request - tokens might still work or it might be a custom channel
	}
	
	liveResult, err := TV.Live(id)
	if err != nil {
		utils.Log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err,
		})
	}
	Bitrates := liveResult.Bitrates
	// if id[:2] == "sl" {
	// 	return sonyLivRedirect(c, liveResult)
	// }
	// Channels with following IDs output audio only m3u8 when quality level is enforced
	if id == "1349" || id == "1322" {
		quality = "auto"
	}
	var liveURL string
	// select quality level based on query parameter
	switch quality {
	case "high", "h":
		liveURL = Bitrates.High
	case "medium", "med", "m":
		liveURL = Bitrates.Medium
	case "low", "l":
		liveURL = Bitrates.Low
	default:
		liveURL = Bitrates.Auto
	}
	// quote url as it will be passed as a query parameter
	coded_url, err := secureurl.EncryptURL(liveURL)
	if err != nil {
		utils.Log.Println(err)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": err,
		})
	}
	return c.Redirect("/render.m3u8?auth="+coded_url+"&channel_key_id="+id, fiber.StatusFound)
}

// RenderHandler handles M3U8 file for modification
// This handler shall replace JioTV server URLs with our own server URLs
func RenderHandler(c *fiber.Ctx) error {
	// URL to be rendered
	auth := c.Query("auth")
	if auth == "" {
		utils.Log.Println("Auth not provided")
		return fmt.Errorf("auth not provided")
	}
	// Channel ID to be used for key rendering
	channel_id := c.Query("channel_key_id")
	if channel_id == "" {
		utils.Log.Println("Channel ID not provided")
		return fmt.Errorf("channel ID not provided")
	}
	// decrypt url
	decoded_url, err := secureurl.DecryptURL(auth)
	if err != nil {
		utils.Log.Println(err)
		return err
	}
	
	renderResult, statusCode := TV.Render(decoded_url)
	
	// If we get a 403 (Forbidden), try refreshing tokens and retry once
	if statusCode == fiber.StatusForbidden {
		if err := EnsureFreshTokens(); err != nil {
			utils.Log.Printf("Failed to refresh tokens after 403: %v", err)
		} else {
			// Retry the request once after refreshing tokens
			renderResult, statusCode = TV.Render(decoded_url)
		}
	}
	// baseUrl is the part of the url excluding suffix file.m3u8 and params is the part of the url after the suffix
	split_url_by_params := strings.Split(decoded_url, "?")
	baseStringUrl := split_url_by_params[0]
	// Pattern to match file names ending with .m3u8
	pattern := `[a-z0-9=\_\-A-Z]*\.m3u8`
	re := regexp.MustCompile(pattern)
	// Add baseUrl to all the file names ending with .m3u8
	baseUrl := []byte(re.ReplaceAllString(baseStringUrl, ""))
	params := split_url_by_params[1]

	// replacer replaces all the file names ending with .m3u8 and .ts with our own server URLs
	// More info: https://golang.org/pkg/regexp/#Regexp.ReplaceAllFunc
	replacer := func(match []byte) []byte {
		switch {
		case bytes.HasSuffix(match, []byte(".m3u8")):
			return television.ReplaceM3U8(baseUrl, match, params, channel_id)
		case bytes.HasSuffix(match, []byte(".ts")):
			return television.ReplaceTS(baseUrl, match, params)
		case bytes.HasSuffix(match, []byte(".aac")):
			return television.ReplaceAAC(baseUrl, match, params)
		default:
			return match
		}
	}

	// Pattern to match file names ending with .m3u8 and .ts
	pattern = `[a-z0-9=\_\-A-Z\/]*\.(m3u8|ts|aac)`
	re = regexp.MustCompile(pattern)
	// Execute replacer function on renderResult
	renderResult = re.ReplaceAllFunc(renderResult, replacer)

	// replacer_key replaces all the URLs ending with .key and .pkey with our own server URLs
	replacer_key := func(match []byte) []byte {
		switch {
		case bytes.HasSuffix(match, []byte(".key")) || bytes.HasSuffix(match, []byte(".pkey")):
			return television.ReplaceKey(match, params, channel_id)
		default:
			return match
		}
	}

	// Pattern to match URLs ending with .key and .pkey
	pattern_key := `http[\S]+\.(pkey|key)`
	re_key := regexp.MustCompile(pattern_key)

	// Execute replacer_key function on renderResult
	renderResult = re_key.ReplaceAllFunc(renderResult, replacer_key)

	if statusCode != fiber.StatusOK {
		utils.Log.Println("Error rendering M3U8 file")
		utils.Log.Println(string(renderResult))
	}
	c.Response().Header.Set("Cache-Control", "public, must-revalidate, max-age=3")
	return c.Status(statusCode).Send(renderResult)
}

// SLHandler proxies requests to SonyLiv CDN
func SLHandler(c *fiber.Ctx) error {
	// Request path with query params
	url := "https://lin-gd-001-cf.slivcdn.com" + c.Path() + "?" + string(c.Request().URI().QueryString())
	if url[len(url)-1:] == "?" {
		url = url[:len(url)-1]
	}
	// Delete all browser headers
	c.Request().Header.Del("Accept")
	c.Request().Header.Del("Accept-Encoding")
	c.Request().Header.Del("Accept-Language")
	c.Request().Header.Del("Origin")
	c.Request().Header.Del("Referer")
	c.Request().Header.Set("User-Agent", PLAYER_USER_AGENT)
	if err := proxy.Do(c, url, TV.Client); err != nil {
		return err
	}

	c.Response().Header.Del(fiber.HeaderServer)
	c.Response().Header.Add("Access-Control-Allow-Origin", "*")
	return nil
}

// RenderKeyHandler requests m3u8 key from JioTV server
func RenderKeyHandler(c *fiber.Ctx) error {
	channel_id := c.Query("channel_key_id")
	auth := c.Query("auth")
	// decode url
	decoded_url, err := secureurl.DecryptURL(auth)
	if err != nil {
		utils.Log.Println(err)
		return err
	}

	// extract params from url
	params := strings.Split(decoded_url, "?")[1]

	// set params as cookies as JioTV uses cookies to authenticate
	for _, param := range strings.Split(params, "&") {
		key := strings.Split(param, "=")[0]
		value := strings.Split(param, "=")[1]
		c.Request().Header.SetCookie(key, value)
	}

	// Copy headers from the Television headers map to the request
	for key, value := range TV.Headers {
		c.Request().Header.Set(key, value) // Assuming only one value for each header
	}
	c.Request().Header.Set("srno", "230203144000")
	c.Request().Header.Set("ssotoken", TV.SsoToken)
	c.Request().Header.Set("channelId", channel_id)
	c.Request().Header.Set("User-Agent", PLAYER_USER_AGENT)
	if err := proxy.Do(c, decoded_url, TV.Client); err != nil {
		return err
	}
	c.Response().Header.Del(fiber.HeaderServer)
	return nil
}

// RenderTSHandler loads TS file from JioTV server
func RenderTSHandler(c *fiber.Ctx) error {
	auth := c.Query("auth")
	// decode url
	decoded_url, err := secureurl.DecryptURL(auth)
	if err != nil {
		utils.Log.Panicln(err)
		return err
	}
	c.Request().Header.Set("User-Agent", PLAYER_USER_AGENT)
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
	splitCategory := strings.TrimSpace(c.Query("c"))
	languages := strings.TrimSpace(c.Query("l"))
	skipGenres := strings.TrimSpace(c.Query("sg"))
	apiResponse := television.Channels()
	// hostUrl should be request URL like http://localhost:5001
	hostURL := strings.ToLower(c.Protocol()) + "://" + c.Hostname()

	// Check if the query parameter "type" is set to "m3u"
	if c.Query("type") == "m3u" {
		// Create an M3U playlist
		m3uContent := "#EXTM3U x-tvg-url=\"" + hostURL + "/epg.xml.gz\"\n"
		logoURL := hostURL + "/jtvimage"
		for _, channel := range apiResponse.Result {

			if languages != "" && !utils.ContainsString(television.LanguageMap[channel.Language], strings.Split(languages, ",")) {
				continue
			}

			if skipGenres != "" && utils.ContainsString(television.CategoryMap[channel.Category], strings.Split(skipGenres, ",")) {
				continue
			}

			var channelURL string
			if quality != "" {
				channelURL = fmt.Sprintf("%s/live/%s/%s.m3u8", hostURL, quality, channel.ID)
			} else {
				channelURL = fmt.Sprintf("%s/live/%s.m3u8", hostURL, channel.ID)
			}
			channelLogoURL := fmt.Sprintf("%s/%s", logoURL, channel.LogoURL)
			var groupTitle string
			switch splitCategory {
			case "split":
				groupTitle = fmt.Sprintf("%s - %s", television.CategoryMap[channel.Category], television.LanguageMap[channel.Language])
			case "language":
				groupTitle = television.LanguageMap[channel.Language]
			default:
				groupTitle = television.CategoryMap[channel.Category]
			}
			m3uContent += fmt.Sprintf("#EXTINF:-1 tvg-id=%q tvg-name=%q tvg-logo=%q tvg-language=%q tvg-type=%q group-title=%q, %s\n%s\n",
				channel.ID, channel.Name, channelLogoURL, television.LanguageMap[channel.Language], television.CategoryMap[channel.Category], groupTitle, channel.Name, channelURL)
		}

		// Set the Content-Disposition header for file download
		c.Set("Content-Disposition", "attachment; filename=jiotv_playlist.m3u")
		c.Set("Content-Type", "application/vnd.apple.mpegurl") // Set the video M3U MIME type
		return c.SendStream(strings.NewReader(m3uContent))
	}

	for i, channel := range apiResponse.Result {
		apiResponse.Result[i].URL = fmt.Sprintf("%s/live/%s", hostURL, channel.ID)
	}

	return c.JSON(apiResponse)
}

// PlayHandler loads HTML Page with video player iframe embedded with video URL
// URL is generated from the channel ID
func PlayHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	quality := c.Query("q")

	// Ensure tokens are fresh before making API call for DRM channels
	if err := EnsureFreshTokens(); err != nil {
		utils.Log.Printf("Failed to ensure fresh tokens: %v", err)
		// Continue with the request - tokens might still work or it might be a custom channel
	}

	var player_url string
	if EnableDRM {
		// Some sonyLiv channels are DRM protected and others are not
		// In order to check, we need to make additional request to JioTV API
		// Quick dirty fix, otherwise we need to refactor entire LiveTV Handler approach
		if utils.ContainsString(id, SONY_LIST) {
			liveResult, err := TV.Live(id)
			if err != nil {
				utils.Log.Println(err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": err,
				})
			}
			// if drm is available, use DRM player
			if liveResult.IsDRM {
				player_url = "/mpd/" + id + "?q=" + quality
			} else {
				// if not, use HLS player
				player_url = "/player/" + id + "?q=" + quality
			}
		} else {
			player_url = "/mpd/" + id + "?q=" + quality
		}
	} else {
		player_url = "/player/" + id + "?q=" + quality
	}
	c.Response().Header.Set("Cache-Control", "public, max-age=3600")
	return c.Render("views/play", fiber.Map{
		"Title":      Title,
		"player_url": player_url,
		"ChannelID":  id,
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
	c.Response().Header.Set("Cache-Control", "public, max-age=3600")
	return c.Render("views/flow_player", fiber.Map{
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
	splitCategory := c.Query("c")
	languages := c.Query("l")
	skipGenres := c.Query("sg")
	return c.Redirect("/channels?type=m3u&q="+quality+"&c="+splitCategory+"&l="+languages+"&sg="+skipGenres, fiber.StatusMovedPermanently)
}

// ImageHandler loads image from JioTV server
func ImageHandler(c *fiber.Ctx) error {
	url := "https://jiotv.catchup.cdn.jio.com/dare_images/images/" + c.Params("file")
	c.Request().Header.Set("User-Agent", REQUEST_USER_AGENT)
	if err := proxy.Do(c, url, TV.Client); err != nil {
		return err
	}
	c.Response().Header.Del(fiber.HeaderServer)
	return nil
}

func DASHTimeHandler(c *fiber.Ctx) error {
	return c.SendString(time.Now().UTC().Format("2006-01-02T15:04:05.000Z"))
}
