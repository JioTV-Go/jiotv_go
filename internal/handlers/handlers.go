package handlers

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jiotv-go/jiotv_go/v3/internal/config"
	"github.com/jiotv-go/jiotv_go/v3/internal/constants/headers"
	"github.com/jiotv-go/jiotv_go/v3/internal/constants/urls"
	internalUtils "github.com/jiotv-go/jiotv_go/v3/internal/utils"
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
	renderHDNEACache sync.Map
)

const (
	REFRESH_TOKEN_URL     = urls.RefreshTokenURL
	REFRESH_SSO_TOKEN_URL = urls.RefreshSSOTokenURL
	PLAYER_USER_AGENT     = headers.UserAgentPlayTV
	REQUEST_USER_AGENT    = headers.UserAgentOkHttp
	hdneaCacheTTL         = 60 * time.Second // Aggressive TTL: 60 seconds (tokens expire ~90-120s, keep cache short)
	hdneaRefreshLeadTime  = 20 * time.Second
)

type hdneaCacheEntry struct {
	Token     string
	UpdatedAt time.Time
}

// truncateToken returns first 10 and last 10 chars of token for logging
func truncateToken(token string) string {
	if len(token) == 0 {
		return "(empty)"
	}
	if len(token) <= 20 {
		return token
	}
	return token[:10] + "..." + token[len(token)-10:]
}

// Init initializes the necessary operations required for the handlers to work.
func Init() {
	if config.Cfg.Title != "" {
		Title = config.Cfg.Title
	} else {
		Title = "JioTV Go"
	}
	DisableTSHandler = config.Cfg.DisableTSHandler
	isLogoutDisabled = config.Cfg.DisableLogout
	EnableDRM = true // DRM is enabled by default, only channels that support DRM will use it
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
		return internalUtils.InternalServerError(c, err.Error())
	}
	return nil
}

// isCustomChannel checks if a given channel ID is a custom channel
func isCustomChannel(channelID string) bool {
	if config.Cfg.CustomChannelsFile == "" {
		return false
	}

	// Check direct lookup with the provided ID
	if _, exists := television.GetCustomChannelByID(channelID); exists {
		return true
	}

	return false
}

// IndexHandler handles the index page for `/` route
func IndexHandler(c *fiber.Ctx) error {
	// Get all channels
	channels, err := television.Channels()
	if err != nil {
		return ErrorMessageHandler(c, err)
	}

	// Get language and category from query params
	language := c.Query("language")
	category := c.Query("category")

	// Process logo URLs for all channels
	hostURL := c.Protocol() + "://" + c.Hostname()
	for i, channel := range channels.Result {
		if strings.HasPrefix(channel.LogoURL, "http://") || strings.HasPrefix(channel.LogoURL, "https://") {
			// Custom channel with full URL, use as-is
			channels.Result[i].LogoURL = channel.LogoURL
		} else {
			// Regular channel with relative path, add proxy prefix
			channels.Result[i].LogoURL = hostURL + "/jtvimage/" + channel.LogoURL
		}
	}

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

	// Filter channels by query params if provided
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

	// If no query parameters are provided, use default config filtering
	if len(config.Cfg.DefaultCategories) > 0 || len(config.Cfg.DefaultLanguages) > 0 {
		channels_list := television.FilterChannelsByDefaults(channels.Result, config.Cfg.DefaultCategories, config.Cfg.DefaultLanguages)
		indexContext["Channels"] = channels_list
		return c.Render("views/index", indexContext)
	}

	// If no query params and no default config, return all channels
	indexContext["Channels"] = channels.Result
	return c.Render("views/index", indexContext)
}

// checkFieldExist checks if the field is provided in the request.
// If not, send a bad request response
func checkFieldExist(field string, check bool, c *fiber.Ctx) error {
	return internalUtils.CheckFieldExist(c, field, check)
}

func isLikelyHLSURL(streamURL string) bool {
	if streamURL == "" {
		return false
	}
	urlLower := strings.ToLower(streamURL)
	return strings.Contains(urlLower, ".m3u8")
}

func isAbsoluteHTTPURL(streamURL string) bool {
	if streamURL == "" {
		return false
	}
	urlLower := strings.ToLower(streamURL)
	if !(strings.HasPrefix(urlLower, "http://") || strings.HasPrefix(urlLower, "https://")) {
		return false
	}
	parsed, err := url.Parse(streamURL)
	return err == nil && parsed.Scheme != "" && parsed.Host != ""
}

func absoluteBaseFromLiveResult(liveResult *television.LiveURLOutput) string {
	if liveResult == nil {
		return ""
	}

	candidates := []string{
		liveResult.Bitrates.Auto,
		liveResult.Bitrates.High,
		liveResult.Bitrates.Medium,
		liveResult.Bitrates.Low,
		liveResult.Result,
		liveResult.Mpd.Result,
		liveResult.Mpd.Bitrates.Auto,
		liveResult.Mpd.Bitrates.High,
		liveResult.Mpd.Bitrates.Medium,
		liveResult.Mpd.Bitrates.Low,
	}

	for _, candidate := range candidates {
		if !isAbsoluteHTTPURL(candidate) {
			continue
		}
		parsed, err := url.Parse(candidate)
		if err == nil && parsed.Scheme != "" && parsed.Host != "" {
			return parsed.Scheme + "://" + parsed.Host
		}
	}

	return ""
}

func toAbsoluteStreamURL(streamURL string, liveResult *television.LiveURLOutput) string {
	if streamURL == "" {
		return ""
	}
	if isAbsoluteHTTPURL(streamURL) {
		return streamURL
	}
	if strings.HasPrefix(streamURL, "//") {
		return "https:" + streamURL
	}

	// Handle host without scheme: jiotv.example.com/path/file.m3u8
	firstPart := strings.SplitN(streamURL, "/", 2)[0]
	if strings.Contains(firstPart, ".") && !strings.HasPrefix(streamURL, "/") {
		return "https://" + streamURL
	}

	if !strings.HasPrefix(streamURL, "/") {
		streamURL = "/" + streamURL
	}

	base := absoluteBaseFromLiveResult(liveResult)
	if base == "" {
		base = "https://" + urls.JioTVCDNDomain
	}

	return base + streamURL
}

func stripHDNEAFromURL(streamURL string) string {
	if streamURL == "" {
		return streamURL
	}
	parsed, err := url.Parse(streamURL)
	if err != nil {
		return streamURL
	}
	query := parsed.Query()
	query.Del("hdnea")
	query.Del("__hdnea__")
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func extractHDNEAFromURL(streamURL string) string {
	if streamURL == "" {
		return ""
	}
	parsed, err := url.Parse(streamURL)
	if err != nil {
		return ""
	}
	query := parsed.Query()
	if token := query.Get("__hdnea__"); token != "" {
		return token
	}
	if token := query.Get("hdnea"); token != "" {
		return token
	}
	return ""
}

func hdneaRemainingLifetime(token string) (time.Duration, bool) {
	if token == "" {
		return 0, false
	}

	parts := strings.Split(token, "~")
	for _, part := range parts {
		if strings.HasPrefix(part, "exp=") {
			expirationStr := strings.TrimPrefix(part, "exp=")
			expirationUnix, err := strconv.ParseInt(expirationStr, 10, 64)
			if err != nil {
				return 0, false
			}
			expirationTime := time.Unix(expirationUnix, 0)
			return time.Until(expirationTime), true
		}
	}

	return 0, false
}

func extractLiveResultHDNEA(liveResult *television.LiveURLOutput) string {
	if liveResult == nil {
		return ""
	}

	candidates := []string{
		liveResult.Bitrates.Auto,
		liveResult.Bitrates.High,
		liveResult.Bitrates.Medium,
		liveResult.Bitrates.Low,
		liveResult.Result,
		liveResult.Mpd.Bitrates.Auto,
		liveResult.Mpd.Bitrates.High,
		liveResult.Mpd.Bitrates.Medium,
		liveResult.Mpd.Bitrates.Low,
		liveResult.Mpd.Result,
	}

	for _, candidate := range candidates {
		if token := extractHDNEAFromURL(candidate); token != "" {
			return token
		}
	}

	return liveResult.Hdnea
}

func liveResultNeedsRefresh(liveResult *television.LiveURLOutput) bool {
	hdneaToken := extractLiveResultHDNEA(liveResult)
	remaining, ok := hdneaRemainingLifetime(hdneaToken)
	return ok && remaining <= hdneaRefreshLeadTime
}

func refreshLiveResultIfNeeded(channelID string, liveResult *television.LiveURLOutput) (*television.LiveURLOutput, error) {
	if channelID == "" || liveResult == nil || !liveResultNeedsRefresh(liveResult) {
		return liveResult, nil
	}

	utils.Log.Printf("HDNEA token is near expiry for channel %s; refreshing live URL", channelID)
	refreshedResult, err := TV.Live(channelID)
	if err != nil {
		return liveResult, err
	}

	if refreshedResult == nil {
		return liveResult, nil
	}

	return refreshedResult, nil
}

func getCachedHDNEA(channelID string) string {
	if channelID == "" {
		return ""
	}
	entryRaw, ok := renderHDNEACache.Load(channelID)
	if !ok {
		return ""
	}
	entry, ok := entryRaw.(hdneaCacheEntry)
	if !ok {
		renderHDNEACache.Delete(channelID)
		return ""
	}
	if entry.Token == "" || time.Since(entry.UpdatedAt) > hdneaCacheTTL {
		renderHDNEACache.Delete(channelID)
		return ""
	}
	return entry.Token
}

func setCachedHDNEA(channelID, token string) {
	if channelID == "" || token == "" {
		return
	}
	renderHDNEACache.Store(channelID, hdneaCacheEntry{Token: token, UpdatedAt: time.Now()})
}

func selectBestLiveHLSURL(liveResult *television.LiveURLOutput, quality string) string {
	if liveResult == nil {
		return ""
	}

	// Try requested quality first.
	selected := internalUtils.SelectQuality(quality, liveResult.Bitrates.Auto, liveResult.Bitrates.High, liveResult.Bitrates.Medium, liveResult.Bitrates.Low)
	if selected != "" {
		return selected
	}

	// Then try any other HLS bitrate that is available.
	for _, candidate := range []string{liveResult.Bitrates.High, liveResult.Bitrates.Auto, liveResult.Bitrates.Medium, liveResult.Bitrates.Low} {
		if candidate != "" {
			return candidate
		}
	}

	// Some newer Jio channels return playable HLS in result instead of bitrates.
	if isLikelyHLSURL(liveResult.Result) {
		return liveResult.Result
	}

	// Safety fallback when MPD block contains an HLS URL (rare, but seen in API drift cases).
	if isLikelyHLSURL(liveResult.Mpd.Result) {
		return liveResult.Mpd.Result
	}

	return ""
}

func selectBestLiveMPDURL(liveResult *television.LiveURLOutput, quality string) string {
	if liveResult == nil {
		return ""
	}

	selected := internalUtils.SelectQuality(quality, liveResult.Mpd.Bitrates.Auto, liveResult.Mpd.Bitrates.High, liveResult.Mpd.Bitrates.Medium, liveResult.Mpd.Bitrates.Low)
	if selected != "" {
		return selected
	}

	for _, candidate := range []string{liveResult.Mpd.Bitrates.High, liveResult.Mpd.Bitrates.Auto, liveResult.Mpd.Bitrates.Medium, liveResult.Mpd.Bitrates.Low} {
		if candidate != "" {
			return candidate
		}
	}

	return liveResult.Mpd.Result
}

// LiveHandler handles the live channel stream route `/live/:id.m3u8`.
func LiveHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	// remove suffix .m3u8 if exists
	id = strings.Replace(id, ".m3u8", "", 1)

	// Check if this is a custom channel - serve directly for custom channels
	if isCustomChannel(id) {
		channel, exists := television.GetCustomChannelByID(id)
		if !exists {
			utils.Log.Printf("Custom channel with ID %s not found", id)
			return internalUtils.NotFoundError(c, fmt.Sprintf("Custom channel with ID %s not found", id))
		}
		// For custom channels, redirect directly to the m3u8 URL (no render pipeline needed)
		return c.Redirect(channel.URL, fiber.StatusFound)
	}

	// For regular JioTV channels, ensure tokens are fresh before making API call
	if err := EnsureFreshTokens(); err != nil {
		utils.Log.Printf("Failed to ensure fresh tokens: %v", err)
		// Continue with the request - tokens might still work
	}

	liveResult, err := TV.Live(id)
	if err != nil {
		utils.Log.Println(err)
		return internalUtils.InternalServerError(c, err)
	}

	liveURL := selectBestLiveHLSURL(liveResult, "auto")
	if liveURL == "" {
		error_message := "No stream found for channel id: " + id + "Status: " + liveResult.Message
		utils.Log.Println(error_message)
		utils.Log.Println(liveResult)
		return internalUtils.NotFoundError(c, error_message)
	}
	liveURL = toAbsoluteStreamURL(liveURL, liveResult)
	if liveResult.Hdnea != "" {
		setCachedHDNEA(id, liveResult.Hdnea)
	}
	// quote url as it will be passed as a query parameter
	// It is required to quote the url as it may contain special characters like ? and &

	coded_url, err := secureurl.EncryptURL(liveURL)
	if err != nil {
		utils.Log.Println(err)
		return internalUtils.ForbiddenError(c, err)
	}
	redirectURL := "/render.m3u8?auth=" + coded_url + "&channel_key_id=" + id
	return c.Redirect(redirectURL, fiber.StatusFound)
}

// LiveQualityHandler handles the live channel stream route `/live/:quality/:id.m3u8`.
func LiveQualityHandler(c *fiber.Ctx) error {
	quality := c.Params("quality")
	id := c.Params("id")
	// remove suffix .m3u8 if exists
	id = strings.Replace(id, ".m3u8", "", 1)

	// Check if this is a custom channel - serve directly for custom channels
	if isCustomChannel(id) {
		channel, exists := television.GetCustomChannelByID(id)
		if !exists {
			utils.Log.Printf("Custom channel with ID %s not found", id)
			return internalUtils.NotFoundError(c, fmt.Sprintf("Custom channel with ID %s not found", id))
		}
		// For custom channels, redirect directly to the m3u8 URL (no render pipeline needed)
		return c.Redirect(channel.URL, fiber.StatusFound)
	}

	// For regular JioTV channels, ensure tokens are fresh before making API call
	if err := EnsureFreshTokens(); err != nil {
		utils.Log.Printf("Failed to ensure fresh tokens: %v", err)
		// Continue with the request - tokens might still work
	}

	liveResult, err := TV.Live(id)
	if err != nil {
		utils.Log.Println(err)
		return internalUtils.InternalServerError(c, err)
	}
	// Channels with following IDs output audio only m3u8 when quality level is enforced
	if id == "1349" || id == "1322" {
		quality = "auto"
	}

	// select quality level based on query parameter and API fallbacks.
	liveURL := selectBestLiveHLSURL(liveResult, quality)
	if liveURL == "" {
		error_message := "No stream found for channel id: " + id + "Status: " + liveResult.Message
		utils.Log.Println(error_message)
		utils.Log.Println(liveResult)
		return internalUtils.NotFoundError(c, error_message)
	}
	liveURL = toAbsoluteStreamURL(liveURL, liveResult)
	if liveResult.Hdnea != "" {
		setCachedHDNEA(id, liveResult.Hdnea)
	}

	// quote url as it will be passed as a query parameter
	coded_url, err := secureurl.EncryptURL(liveURL)
	if err != nil {
		utils.Log.Println(err)
		return internalUtils.ForbiddenError(c, err)
	}
	redirectURL := "/render.m3u8?auth=" + coded_url + "&channel_key_id=" + id + "&q=" + quality
	return c.Redirect(redirectURL, fiber.StatusFound)
}

// RenderHandler handles M3U8 file for modification
// This handler shall replace JioTV server URLs with our own server URLs
func RenderHandler(c *fiber.Ctx) error {
	// URL to be rendered
	auth := c.Query("auth")
	if err := internalUtils.ValidateRequiredParam("auth", auth); err != nil {
		return err
	}
	// Channel ID to be used for key rendering
	channel_id := c.Query("channel_key_id")
	if err := internalUtils.ValidateRequiredParam("channel_key_id", channel_id); err != nil {
		return err
	}
	// decrypt url
	decoded_url, err := secureurl.DecryptURL(auth)
	if err != nil {
		utils.Log.Println(err)
		return err
	}

	decoded_url = toAbsoluteStreamURL(decoded_url, nil)

	// Always prefer a freshly cached HDNEA token if available to prevent 403s on expired URL tokens
	cachedHDNEA := getCachedHDNEA(channel_id)
	urlToken := extractHDNEAFromURL(decoded_url)

	renderURL := decoded_url
	if cachedHDNEA != "" {
		// We have a freshly fetched token from a recent recovery, use it instead of the potentially expired URL token
		renderURL = stripHDNEAFromURL(decoded_url)
	} else if urlToken != "" {
		cachedHDNEA = urlToken
	}

	// DEBUG: Log token selection
	if os.Getenv("JIOTV_DEBUG") == "true" {
		sourceStr := "cache"
		if cachedHDNEA == "" {
			sourceStr = "none"
		} else if cachedHDNEA == urlToken {
			sourceStr = "URL"
		}
		utils.Log.Printf("[DEBUG] Token selection - URL token: %s | Cached token: %s | Using: %s (source: %s)",
			truncateToken(urlToken), truncateToken(getCachedHDNEA(channel_id)), truncateToken(cachedHDNEA), sourceStr)
	}
	renderResult, statusCode, newHdnea := TV.Render(renderURL, cachedHDNEA)

	// DEBUG: Log token extraction and response
	if os.Getenv("JIOTV_DEBUG") == "true" {
		utils.Log.Printf("[DEBUG] Render response - Status: %d | Token from response: %s", statusCode, truncateToken(newHdnea))
	}

	// Always cache fresh token from response for fallback on next request
	if newHdnea != "" {
		setCachedHDNEA(channel_id, newHdnea)
		cachedHDNEA = newHdnea
	}

	// On authentication failure or 404, unify the retry logic by fetching a fresh stream URL
	if statusCode == fiber.StatusForbidden || statusCode == fiber.StatusUnauthorized || statusCode == fiber.StatusNotFound {
		// Clear the stale cached token
		if statusCode != fiber.StatusNotFound {
			renderHDNEACache.Delete(channel_id)
		}

		if os.Getenv("JIOTV_DEBUG") == "true" {
			utils.Log.Printf("[DEBUG] Auth failure or not found (Status %d) - fetching fresh live URL and auth", statusCode)
		}

		if channel_id != "" {
			retryQuality := c.Query("q")
			if retryQuality == "" {
				retryQuality = "auto"
			}

			if refreshedLiveResult, refreshErr := TV.Live(channel_id); refreshErr == nil && refreshedLiveResult != nil {
				if freshToken := extractLiveResultHDNEA(refreshedLiveResult); freshToken != "" {
					setCachedHDNEA(channel_id, freshToken)
					cachedHDNEA = freshToken
				}

				qualityCandidates := []string{retryQuality, "auto", "high", "medium", "low"}
				triedURL := map[string]bool{renderURL: true}

				for _, candidateQuality := range qualityCandidates {
					candidateURL := selectBestLiveHLSURL(refreshedLiveResult, candidateQuality)
					candidateURL = toAbsoluteStreamURL(candidateURL, refreshedLiveResult)
					if candidateURL == "" || triedURL[candidateURL] {
						continue
					}
					triedURL[candidateURL] = true

					if os.Getenv("JIOTV_DEBUG") == "true" {
						utils.Log.Printf("[DEBUG] RenderHandler recovery - trying quality=%s for channel=%s", candidateQuality, channel_id)
					}

					renderURL = candidateURL
					renderResult, statusCode, newHdnea = TV.Render(renderURL, cachedHDNEA)
					if newHdnea != "" {
						setCachedHDNEA(channel_id, newHdnea)
						cachedHDNEA = newHdnea
					}

					if statusCode == fiber.StatusOK {
						break
					}
				}
			}
		}
	}
	// No client cookie: if upstream rotated __hdnea__, we'll embed the fresh token into rewritten URLs below

	// baseUrl is the part of the url excluding suffix file.m3u8 and params is the part of the url after the suffix
	split_url_by_params := strings.Split(renderURL, "?")
	baseStringUrl := split_url_by_params[0]
	// Pattern to match file names ending with .m3u8
	pattern := `[a-z0-9=\_\-A-Z\.]*\.m3u8`
	re := regexp.MustCompile(pattern)
	// Add baseUrl to all the file names ending with .m3u8
	baseUrl := []byte(re.ReplaceAllString(baseStringUrl, ""))
	params := ""
	if len(split_url_by_params) > 1 {
		params = split_url_by_params[1]
	}
	if params != "" {
		if parsedParams, parseErr := url.ParseQuery(params); parseErr == nil {
			parsedParams.Del("hdnea")
			parsedParams.Del("__hdnea__")
			encodedParams := parsedParams.Encode()
			if cachedHDNEA != "" {
				if encodedParams != "" {
					params = encodedParams + "&__hdnea__=" + cachedHDNEA
				} else {
					params = "__hdnea__=" + cachedHDNEA
				}
			} else {
				params = encodedParams
			}
		}
	} else if cachedHDNEA != "" {
		params = "__hdnea__=" + cachedHDNEA
	}

	// replacer replaces all the file names ending with .m3u8 and .ts with our own server URLs
	// More info: https://golang.org/pkg/regexp/#Regexp.ReplaceAllFunc
	replacer := func(match []byte) []byte {
		switch {
		case bytes.HasSuffix(match, []byte(".m3u8")):
			return television.ReplaceM3U8(baseUrl, match, params, channel_id, c.Query("q"))
		case bytes.HasSuffix(match, []byte(".ts")):
			return television.ReplaceTS(baseUrl, match, params, channel_id)
		case bytes.HasSuffix(match, []byte(".aac")):
			return television.ReplaceAAC(baseUrl, match, params, channel_id)
		default:
			return match
		}
	}

	// Match media URIs with optional query strings so catchup params like
	// ?vbegin=... are consumed as part of the replacement target.
	pattern = `[a-z0-9=\_\-A-Z\/\.]*\.(m3u8|ts|aac)(\?[^\s"']*)?`
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
	internalUtils.SetMustRevalidateHeader(c, 3)
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
	internalUtils.SetPlayerHeaders(c, PLAYER_USER_AGENT)
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
	// parse incoming hdnea query and set as request cookie only for upstream call (no client cookie)
	if hdnea := c.Query("hdnea"); hdnea != "" {
		c.Request().Header.SetCookie("__hdnea__", hdnea)
	}
	// decode url
	decoded_url, err := internalUtils.DecryptURLParam("auth", auth)
	if err != nil {
		return err
	}

	parsedURL, parseErr := url.Parse(decoded_url)
	if parseErr == nil {
		queryValues := parsedURL.Query()
		for key, values := range queryValues {
			if len(values) > 0 {
				c.Request().Header.SetCookie(key, values[0])
			}
		}
		if hdnea := queryValues.Get("hdnea"); hdnea != "" {
			c.Request().Header.SetCookie("__hdnea__", hdnea)
		} else if hdnea := queryValues.Get("__hdnea__"); hdnea != "" {
			c.Request().Header.SetCookie("__hdnea__", hdnea)
		}
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
	// Ensure tokens are fresh before proxying TS segments
	if err := EnsureFreshTokens(); err != nil {
		utils.Log.Printf("Failed to ensure fresh tokens before TS proxy: %v", err)
	}

	channelID := c.Query("channel_key_id")
	auth := c.Query("auth")
	// parse incoming hdnea query and set as request cookie only for upstream call (no client cookie)
	if hdnea := c.Query("hdnea"); hdnea != "" {
		c.Request().Header.SetCookie("__hdnea__", hdnea)
	}
	// decode url
	decoded_url, err := internalUtils.DecryptURLParam("auth", auth)
	if err != nil {
		utils.Log.Panicln(err)
		return err
	}

	// Always prefer a freshly cached HDNEA token if available
	cachedHDNEA := getCachedHDNEA(channelID)
	if cachedHDNEA != "" {
		c.Request().Header.SetCookie("__hdnea__", cachedHDNEA)
		// We should also replace the token in the URL if it's there
		decoded_url = stripHDNEAFromURL(decoded_url)
	} else if len(c.Request().Header.Cookie("__hdnea__")) == 0 && strings.Contains(decoded_url, "hdnea=") {
		// Check if decoded_url has hdnea or __hdnea__ and set cookie if not already set
		// This is crucial when hdnea is embedded in the encrypted auth URL but not in the request query params
		qIdx := strings.Index(decoded_url, "?")
		if qIdx != -1 {
			params := decoded_url[qIdx+1:]
			for _, p := range strings.Split(params, "&") {
				if strings.HasPrefix(p, "hdnea=") {
					c.Request().Header.SetCookie("__hdnea__", strings.TrimPrefix(p, "hdnea="))
					break
				}
				if strings.HasPrefix(p, "__hdnea__=") {
					c.Request().Header.SetCookie("__hdnea__", strings.TrimPrefix(p, "__hdnea__="))
					break
				}
			}
		}
	}

	if err := internalUtils.ProxyRequest(c, decoded_url, TV.Client, PLAYER_USER_AGENT); err != nil {
		return err
	}

	statusCode := c.Response().StatusCode()
	if statusCode == fiber.StatusForbidden || statusCode == fiber.StatusUnauthorized {
		if os.Getenv("JIOTV_DEBUG") == "true" {
			utils.Log.Printf("[DEBUG] RenderTSHandler got %d response - forcing refresh and retrying", statusCode)
		}

		c.Response().Reset()
		c.Request().Header.DelCookie("__hdnea__")

		retryUrl := stripHDNEAFromURL(decoded_url)
		if channelID != "" {
			if refreshedResult, refreshErr := TV.Live(channelID); refreshErr == nil && refreshedResult != nil {
				if refreshedHDNEA := extractLiveResultHDNEA(refreshedResult); refreshedHDNEA != "" {
					setCachedHDNEA(channelID, refreshedHDNEA)
					c.Request().Header.SetCookie("__hdnea__", refreshedHDNEA)
				}
			}
		}

		if err := internalUtils.ProxyRequest(c, retryUrl, TV.Client, PLAYER_USER_AGENT); err != nil {
			return err
		}
	}

	return nil
}

// ChannelsHandler fetch all channels from JioTV API
// Also to generate M3U playlist
func ChannelsHandler(c *fiber.Ctx) error {

	quality := strings.TrimSpace(c.Query("q"))
	splitCategory := strings.TrimSpace(c.Query("c"))
	languages := strings.TrimSpace(c.Query("l"))
	skipGenres := strings.TrimSpace(c.Query("sg"))
	apiResponse, err := television.Channels()
	if err != nil {
		return ErrorMessageHandler(c, err)
	}
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
			var channelLogoURL string
			if strings.HasPrefix(channel.LogoURL, "http://") || strings.HasPrefix(channel.LogoURL, "https://") {
				// Custom channel with full URL
				channelLogoURL = channel.LogoURL
			} else {
				// Regular channel with relative path
				channelLogoURL = fmt.Sprintf("%s/%s", logoURL, channel.LogoURL)
			}
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
	if quality == "" {
		quality = "auto"
	}

	// Ensure tokens are fresh before making API call for DRM channels
	if err := EnsureFreshTokens(); err != nil {
		utils.Log.Printf("Failed to ensure fresh tokens: %v", err)
		// Continue with the request - tokens might still work or it might be a custom channel
	}

	var player_url string
	if EnableDRM {
		// Sony channels should always use DRM player for consistency
		// This avoids routing issues and 403 errors from mixed player usage
		if utils.ContainsString(id, SONY_LIST) {
			player_url = "/mpd/" + id + "?q=" + quality
		} else if isCustomChannel(id) {
			player_url = "/player/" + id + "?q=" + quality
		} else {
			player_url = "/mpd/" + id + "?q=" + quality
		}
	} else {
		player_url = "/player/" + id + "?q=" + quality
	}
	internalUtils.SetCacheHeader(c, 3600)
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
	play_url := utils.BuildHLSPlayURL(quality, id)
	internalUtils.SetCacheHeader(c, 3600)
	return c.Render("views/player_hls", fiber.Map{
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
	return internalUtils.ProxyRequest(c, url, TV.Client, REQUEST_USER_AGENT)
}

func DASHTimeHandler(c *fiber.Ctx) error {
	return c.SendString(time.Now().UTC().Format("2006-01-02T15:04:05.000Z"))
}
