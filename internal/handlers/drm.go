package handlers

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/jiotv-go/jiotv_go/v3/internal/constants/headers"
	internalUtils "github.com/jiotv-go/jiotv_go/v3/internal/utils"
	"github.com/jiotv-go/jiotv_go/v3/pkg/secureurl"
	"github.com/jiotv-go/jiotv_go/v3/pkg/television"
	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
	"github.com/valyala/fasthttp"
)

var (
	// tokenRefreshLock prevents concurrent TV object modifications from race conditions
	tokenRefreshLock sync.Mutex
	// lastTokenRefreshTime tracks when we last successfully refreshed
	lastTokenRefreshTime = time.Now()
)

// EnsureFreshCredentials refreshes tokens proactively before they expire
// This function prevents 403 errors by keeping credentials always fresh
// Returns true if tokens are fresh (either just refreshed or cached)
func EnsureFreshCredentials() bool {
	tokenRefreshLock.Lock()
	defer tokenRefreshLock.Unlock()

	// Only refresh if at least 30 seconds have passed since last successful refresh
	// This prevents excessive API calls while staying within token TTL (90-120s)
	timeSinceLastRefresh := time.Since(lastTokenRefreshTime)
	if timeSinceLastRefresh < 30*time.Second {
		return true // Recently refreshed, tokens are fresh
	}

	return performTokenRefresh()
}

// ForceRefreshCredentials bypasses the 30-second interval check and forces immediate refresh
// Use this only in error recovery paths when we know tokens have failed
func ForceRefreshCredentials() bool {
	tokenRefreshLock.Lock()
	defer tokenRefreshLock.Unlock()

	if os.Getenv("JIOTV_DEBUG") == "true" {
		utils.Log.Printf("[DEBUG] FORCED token refresh (bypassing 30-second interval)")
	}

	return performTokenRefresh()
}

// performTokenRefresh does the actual token refresh work (must be called with lock held)
func performTokenRefresh() bool {
	// CRITICAL REFRESH #1: Refresh AccessToken
	// LoginRefreshAccessToken() does:
	//   1. Call refresh API
	//   2. Update credentials file
	//   3. Update TV object with new token
	accessTokenErr := LoginRefreshAccessToken()
	if accessTokenErr != nil {
		if os.Getenv("JIOTV_DEBUG") == "true" {
			utils.Log.Printf("[DEBUG] AccessToken refresh error: %v", accessTokenErr)
		}
		// Don't return false yet - try SSO token refresh
	} else {
		if os.Getenv("JIOTV_DEBUG") == "true" {
			utils.Log.Printf("[DEBUG] AccessToken refreshed successfully")
		}
	}

	// CRITICAL REFRESH #2: Refresh SSOToken
	// LoginRefreshSSOToken() does:
	//   1. Call refresh API
	//   2. Update credentials file
	//   3. Update TV object with new token
	ssoTokenErr := LoginRefreshSSOToken()
	if ssoTokenErr != nil {
		if os.Getenv("JIOTV_DEBUG") == "true" {
			utils.Log.Printf("[DEBUG] SSOToken refresh error: %v", ssoTokenErr)
		}
		// Don't return false yet - check if at least one refresh succeeded
	} else {
		if os.Getenv("JIOTV_DEBUG") == "true" {
			utils.Log.Printf("[DEBUG] SSOToken refreshed successfully")
		}
	}

	// Update last refresh time if either refresh succeeded
	if accessTokenErr == nil || ssoTokenErr == nil {
		lastTokenRefreshTime = time.Now()
		if os.Getenv("JIOTV_DEBUG") == "true" {
			utils.Log.Printf("[DEBUG] Token refresh cycle completed. TV object already updated by refresh functions")
		}
		return true
	}

	// Both refreshes failed - log comprehensive error
	if os.Getenv("JIOTV_DEBUG") == "true" {
		utils.Log.Printf("[DEBUG] CRITICAL: Both token refreshes failed! AccessToken error: %v, SSOToken error: %v",
			accessTokenErr, ssoTokenErr)
	}
	return false
}

// getDrmMpd returns required properties for rendering DRM MPD
func getDrmMpd(channelID, quality string) (*DrmMpdOutput, error) {
	// Get live stream URL from JioTV API
	liveResult, err := TV.Live(channelID)
	if err != nil {
		return nil, err
	}
	if refreshedResult, refreshErr := refreshLiveResultIfNeeded(channelID, liveResult); refreshErr == nil && refreshedResult != nil {
		liveResult = refreshedResult
	}

	tv_url := internalUtils.SelectQuality(quality, liveResult.Mpd.Bitrates.Auto, liveResult.Mpd.Bitrates.High, liveResult.Mpd.Bitrates.Medium, liveResult.Mpd.Bitrates.Low)

	// If quality selection fails (empty), try to fallback to any available quality
	if tv_url == "" {
		if liveResult.Mpd.Bitrates.High != "" {
			tv_url = liveResult.Mpd.Bitrates.High
		} else if liveResult.Mpd.Bitrates.Auto != "" {
			tv_url = liveResult.Mpd.Bitrates.Auto
		} else if liveResult.Mpd.Bitrates.Medium != "" {
			tv_url = liveResult.Mpd.Bitrates.Medium
		} else if liveResult.Mpd.Bitrates.Low != "" {
			tv_url = liveResult.Mpd.Bitrates.Low
		}
	}

	if tv_url == "" {
		tv_url = liveResult.Mpd.Result
	}
	if tv_url == "" {
		return &DrmMpdOutput{
			IsDRM:       liveResult.IsDRM,
			PlayUrl:     "",
			LicenseUrl:  "",
			Tv_url_host: "",
			Tv_url_path: "",
		}, nil
	}

	channel_enc_url, err := secureurl.EncryptURL(tv_url)
	if err != nil {
		utils.Log.Panicln(err)
		return nil, err
	}

	licenseUrl := ""
	if liveResult.Mpd.Key != "" {
		enc_key, err := secureurl.EncryptURL(liveResult.Mpd.Key)
		if err != nil {
			utils.Log.Panicln(err)
			return nil, err
		}
		licenseUrl = "/drm?auth=" + enc_key + "&channel_id=" + channelID + "&channel=" + channel_enc_url
	}

	// Quick fix for timesplay channels.
	if liveResult.AlgoName == "timesplay" {
		return &DrmMpdOutput{
			IsDRM:       liveResult.IsDRM,
			PlayUrl:     tv_url,
			LicenseUrl:  licenseUrl,
			Tv_url_host: "",
			Tv_url_path: "",
		}, nil
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
		IsDRM:       liveResult.IsDRM,
		PlayUrl:     "/render.mpd?auth=" + channel_enc_url + "&channel_id=" + channelID + "&q=" + quality,
		LicenseUrl:  licenseUrl,
		Tv_url_host: tv_url_host,
		Tv_url_path: tv_url_path,
	}, nil
}

// LiveMpdHandler handles live stream routes /mpd/:channelID
func LiveMpdHandler(c *fiber.Ctx) error {
	// Get channel ID from URL
	channelID := c.Params("channelID")
	quality := c.Query("q")
	playerMode := c.Query("pm") // "hd" (force Shaka) or "auto" (try Shaka, fallback HLS)
	if quality == "" {
		quality = "high"
	}
	if playerMode == "" {
		playerMode = "hd" // Default to HD mode
	}

	if isCustomChannel(channelID) {
		channel, exists := television.GetCustomChannelByID(channelID)
		if !exists {
			utils.Log.Printf("Custom channel with ID %s not found", channelID)
			return internalUtils.NotFoundError(c, fmt.Sprintf("Custom channel with ID %s not found", channelID))
		}
		internalUtils.SetCacheHeader(c, 3600)
		return c.Render("views/player_hls", fiber.Map{
			"play_url": channel.URL,
		})
	}

	// Ensure tokens are fresh before requesting MPD
	EnsureFreshCredentials()

	drmMpdOutput, err := getDrmMpd(channelID, quality)

	// If getting DRM MPD failed, try refreshing tokens forcefully and retry with multiple attempts
	if err != nil {
		utils.Log.Printf("First attempt to get DRM MPD failed: %v. Attempting recovery with forced credentials refresh...", err)

		// Force refresh credentials (bypasses 30-second interval for error recovery)
		if ForceRefreshCredentials() {
			// Retry getDrmMpd with fresh tokens
			drmMpdOutput, err = getDrmMpd(channelID, quality)
			if err == nil {
				utils.Log.Println("Retry successful after forced token refresh")
				return nil // Early return - success
			}
		}

		// If we still have error, log the string for debugging
		errStr := fmt.Sprintf("%v", err)
		if strings.Contains(errStr, "refresh token not found") {
			utils.Log.Printf("ERROR: Session refresh token is lost or expired. User needs to re-login.")
			utils.Log.Printf("This can happen when: 1) Jio session expired  2) Logged out from another device  3) Token cache cleared")
		}
	}

	// Fallback to HLS on error or empty URL
	if err != nil {
		utils.Log.Printf("Error getting DRM MPD (falling back to HLS): %v", err)
	} else if drmMpdOutput == nil {
		utils.Log.Printf("DRM MPD output is nil (falling back to HLS)")
	} else if drmMpdOutput.PlayUrl == "" {
		utils.Log.Printf("DRM MPD PlayUrl is empty (falling back to HLS)")
	}

	if err != nil || drmMpdOutput == nil || drmMpdOutput.PlayUrl == "" {
		// Use requested quality (default high) for HLS fallback to ensure best available quality first
		play_url := utils.BuildHLSPlayURL(quality, channelID)
		internalUtils.SetCacheHeader(c, 3600)
		return c.Render("views/player_hls", fiber.Map{
			"play_url": play_url,
		})
	}

	hlsFallbackURL := utils.BuildHLSPlayURL(quality, channelID)
	hlsPlayerFallbackURL := "/player/" + channelID + "?q=" + quality + "&af=1"

	return c.Render("views/player_drm", fiber.Map{
		"play_url":                drmMpdOutput.PlayUrl,
		"license_url":             drmMpdOutput.LicenseUrl,
		"channel_host":            drmMpdOutput.Tv_url_host,
		"channel_path":            drmMpdOutput.Tv_url_path,
		"hls_fallback_url":        hlsFallbackURL,
		"hls_player_fallback_url": hlsPlayerFallbackURL,
		"player_mode":             playerMode, // Pass mode to template
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

	decoded_channel, err := internalUtils.DecryptURLParam("channel", channel)
	if err != nil {
		utils.Log.Panicln(err)
		return internalUtils.ForbiddenError(c, err)
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

	decoded_url, err := internalUtils.DecryptURLParam("auth", auth)
	if err != nil {
		utils.Log.Panicln(err)
		return internalUtils.ForbiddenError(c, err)
	}

	// Add headers to the request
	c.Request().Header.Set("accesstoken", TV.AccessToken)
	c.Request().Header.Set("Connection", "keep-alive")
	c.Request().Header.Set("os", "android")
	c.Request().Header.Set("appName", "RJIL_JioTV")
	c.Request().Header.Set("subscriberId", TV.Crm)
	c.Request().Header.Set("User-Agent", PLAYER_USER_AGENT)
	c.Request().Header.Set("ssotoken", TV.SsoToken)
	c.Request().Header.Set("x-platform", "android")
	c.Request().Header.Set("srno", generateDateTime())
	c.Request().Header.Set("crmid", TV.Crm)
	c.Request().Header.Set("channelid", channel_id)
	c.Request().Header.Set("uniqueId", TV.UniqueID)
	c.Request().Header.Set("versionCode", headers.VersionCode389)
	c.Request().Header.Set("usergroup", "tvYR7NSNn7rymo3F")
	c.Request().Header.Set("devicetype", "phone")
	c.Request().Header.Set("Accept-Encoding", "gzip, deflate")
	c.Request().Header.Set("osVersion", "13")
	c.Request().Header.Set("deviceId", utils.GetDeviceID())
	c.Request().Header.Set("Content-Type", "application/octet-stream")

	// Remove headers
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
	// CRITICAL: Refresh credentials before proxying MPD
	EnsureFreshCredentials()

	channelID := c.Query("channel_id")
	quality := c.Query("q")
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

	if channelID != "" {
		if liveResult, liveErr := TV.Live(channelID); liveErr == nil && liveResult != nil {
			if freshUrl := selectBestLiveMPDURL(liveResult, quality); freshUrl != "" {
				decryptedUrl = freshUrl
				parsedUrl, err = url.Parse(decryptedUrl)
				if err != nil {
					utils.Log.Panicln(err)
					return err
				}
			}
		}
	}

	proxyHost := parsedUrl.Host
	pathParts := strings.Split(parsedUrl.Path, "/")
	basePath := strings.Join(pathParts[:len(pathParts)-1], "/") + "/"
	encProxyHost, err := secureurl.EncryptURL(proxyHost)
	if err != nil {
		utils.Log.Panicln(err)
		return err
	}
	encProxyPath, err := secureurl.EncryptURL(basePath)
	if err != nil {
		utils.Log.Panicln(err)
		return err
	}

	// Extract channel_id and cached HDNEA if available from query params
	// This allows DashHandler to use the same auth context
	var cachedHDNEA string
	if chd := c.Query("hdnea"); chd != "" {
		cachedHDNEA = chd
	}

	dashBaseURL := fmt.Sprintf("/render.dash/host/%s/path/%s", encProxyHost, encProxyPath)
	if cachedHDNEA != "" {
		encHDNEA, encErr := secureurl.EncryptURL("__hdnea__=" + cachedHDNEA)
		if encErr == nil {
			dashBaseURL = fmt.Sprintf("/render.dash/host/%s/path/%s/hdnea/%s", encProxyHost, encProxyPath, encHDNEA)
		}
	}

	// proxyQuery := parsedUrl.RawQuery

	c.Request().Header.Set("Host", proxyHost)
	c.Request().Header.Set("User-Agent", PLAYER_USER_AGENT)

	// Request path with query params
	requestUrl := decryptedUrl

	c.Request().Header.Set("User-Agent", PLAYER_USER_AGENT)
	// remove Accept-Encoding header
	c.Request().Header.Del("Accept-Encoding")

	// AGGRESSIVE REFRESH: Make initial proxy request
	if err := proxy.Do(c, requestUrl, TV.Client); err != nil {
		return err
	}

	// Handle 403/401 auth failures by stripping HDNEA and retrying
	statusCode := c.Response().StatusCode()
	if statusCode == fiber.StatusForbidden || statusCode == fiber.StatusUnauthorized {
		if os.Getenv("JIOTV_DEBUG") == "true" {
			utils.Log.Printf("[DEBUG] MpdHandler got %d response - stripping HDNEA and retrying", statusCode)
		}

		// Reset response to allow retry
		c.Response().Reset()
		ForceRefreshCredentials()

		// Strip HDNEA token and retry - CDN will provide fresh auth
		// HDNEA tokens are CDN-managed and expire, so requesting without them
		// forces CDN to issue fresh auth
		strippedUrl := stripHDNEAFromURL(decryptedUrl)

		if os.Getenv("JIOTV_DEBUG") == "true" {
			if strippedUrl != requestUrl {
				utils.Log.Printf("[DEBUG] MpdHandler: removed HDNEA from URL, retrying")
			} else {
				utils.Log.Printf("[DEBUG] MpdHandler: retrying request (no HDNEA to strip)")
			}
		}

		if err := proxy.Do(c, strippedUrl, TV.Client); err != nil {
			if os.Getenv("JIOTV_DEBUG") == "true" {
				utils.Log.Printf("[DEBUG] MpdHandler retry failed: %v", err)
			}
			return err
		}

		if os.Getenv("JIOTV_DEBUG") == "true" {
			utils.Log.Printf("[DEBUG] MpdHandler retry - new status: %d", c.Response().StatusCode())
		}
	}

	c.Response().Header.Del(fiber.HeaderServer)

	// Extract __hdnea__ from upstream response for injecting into dashBaseURL
	upstreamHDNEA := ""

	// Try to extract from Set-Cookie header first
	setCookie := c.Response().Header.Peek("Set-Cookie")
	if setCookie != nil {
		setCookieStr := string(setCookie)
		// Parse Set-Cookie: name=value; attributes...
		// Look for __hdnea__=value
		if strings.Contains(setCookieStr, "__hdnea__=") {
			parts := strings.Split(setCookieStr, ";")
			for _, part := range parts {
				trimmed := strings.TrimSpace(part)
				if strings.HasPrefix(trimmed, "__hdnea__=") {
					upstreamHDNEA = strings.TrimPrefix(trimmed, "__hdnea__=")
					break
				}
			}
		}
	}

	// If we got a fresh __hdnea__ from upstream, update dashBaseURL with it
	if upstreamHDNEA != "" {
		encHDNEA, encErr := secureurl.EncryptURL("__hdnea__=" + upstreamHDNEA)
		if encErr == nil {
			dashBaseURL = fmt.Sprintf("/render.dash/host/%s/path/%s/hdnea/%s", encProxyHost, encProxyPath, encHDNEA)
		}
	}

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
			return []byte(fmt.Sprintf("<BaseURL>%s/dash/</BaseURL>", dashBaseURL))
		})
	} else {
		pattern := `<Period(\s+[^>]*?)?\s*\/?>`
		re = regexp.MustCompile(pattern)
		resBody = re.ReplaceAllFunc(resBody, func(match []byte) []byte {
			return []byte(fmt.Sprintf("%s\n<BaseURL>%s/</BaseURL>", match, dashBaseURL))
		})
	}

	c.Response().SetBody(resBody)

	return nil
}

// DashHandler
func DashHandler(c *fiber.Ctx) error {
	proxyHost := c.Query("host")
	proxyPath := c.Query("path")
	requestPath := string(c.Request().URI().Path())
	requestQuery := string(c.Request().URI().QueryString())

	// Extract embedded HDNEA if present
	var hdneaToken string

	if proxyHost == "" || proxyPath == "" {
		const prefix = "/render.dash/host/"
		if strings.HasPrefix(requestPath, prefix) {
			trimmed := strings.TrimPrefix(requestPath, prefix)
			parts := strings.SplitN(trimmed, "/path/", 2)
			if len(parts) == 2 {
				proxyHost = parts[0]
				remainder := parts[1]

				// Check for embedded hdnea pattern: /render.dash/host/{host}/path/{path}/hdnea/{hdnea}/{rest}
				hdneaParts := strings.SplitN(remainder, "/hdnea/", 2)
				if len(hdneaParts) == 2 {
					proxyPath = hdneaParts[0]
					// Now hdneaParts[1] contains "{encHdnea}/{rest...}"
					restParts := strings.SplitN(hdneaParts[1], "/", 2)
					encHdnea := restParts[0]

					// Decrypt HDNEA
					decHdnea, decErr := secureurl.DecryptURL(encHdnea)
					if decErr == nil && strings.HasPrefix(decHdnea, "__hdnea__=") {
						hdneaToken = strings.TrimPrefix(decHdnea, "__hdnea__=")
					}

					// Set request path to the remaining part after hdnea
					if len(restParts) == 2 {
						requestPath = "/" + restParts[1]
					} else {
						requestPath = "/"
					}
				} else {
					// No hdnea, parse normally
					pathParts := strings.SplitN(remainder, "/", 2)
					proxyPath = pathParts[0]
					if len(pathParts) == 2 {
						requestPath = "/" + pathParts[1]
					} else {
						requestPath = "/"
					}
				}
			}
		}
	}

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

	if strings.HasPrefix(requestPath, "/render.dash") {
		requestPath = strings.TrimPrefix(requestPath, "/render.dash")
		if requestPath == "" {
			requestPath = "/"
		}
	}
	requestUri := requestPath
	if requestQuery != "" {
		requestUri = requestUri + "?" + requestQuery
	}

	proxyPath = strings.TrimSuffix(proxyPath, "/")
	proxyUrl := fmt.Sprintf("https://%s%s%s", proxyHost, proxyPath, requestUri)

	c.Request().Header.Set("User-Agent", PLAYER_USER_AGENT)

	// Set HDNEA cookie if we have it
	if hdneaToken != "" {
		c.Request().Header.SetCookie("__hdnea__", hdneaToken)
	}

	// CRITICAL: Refresh credentials before proxying segments
	EnsureFreshCredentials()

	// AGGRESSIVE REFRESH: Make initial proxy request
	if err := proxy.Do(c, proxyUrl, TV.Client); err != nil {
		return err
	}

	// Handle 403/401 auth failures with retry mechanism (AGGRESSIVE REFRESH)
	statusCode := c.Response().StatusCode()
	if statusCode == fiber.StatusForbidden || statusCode == fiber.StatusUnauthorized {
		if os.Getenv("JIOTV_DEBUG") == "true" {
			utils.Log.Printf("[DEBUG] DashHandler got %d response - clearing HDNEA cookie and retrying", statusCode)
		}

		// Reset response to allow retry
		c.Response().Reset()
		ForceRefreshCredentials()

		// Clear HDNEA cookie - expired token causes 403
		// CDN will provide fresh HDNEA in the response
		c.Request().Header.DelCookie("__hdnea__")

		if err := proxy.Do(c, proxyUrl, TV.Client); err != nil {
			if os.Getenv("JIOTV_DEBUG") == "true" {
				utils.Log.Printf("[DEBUG] DashHandler retry failed: %v", err)
			}
			return err
		}

		if os.Getenv("JIOTV_DEBUG") == "true" {
			utils.Log.Printf("[DEBUG] DashHandler retry - new status: %d", c.Response().StatusCode())
		}

		if os.Getenv("JIOTV_DEBUG") == "true" {
			utils.Log.Printf("[DEBUG] DashHandler retry successful - new status: %d", c.Response().StatusCode())
		}
	}

	c.Response().Header.Del(fiber.HeaderServer)

	return nil
}
