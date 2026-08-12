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

func extractCookieValueFromSetCookieHeaders(setCookieHeaders [][]byte, cookieName string) string {
	prefix := cookieName + "="
	for _, header := range setCookieHeaders {
		parts := strings.Split(string(header), ";")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if strings.HasPrefix(trimmed, prefix) {
				return strings.TrimPrefix(trimmed, prefix)
			}
		}
	}
	return ""
}

func buildCookieHeaderFromSetCookieHeaders(setCookieHeaders [][]byte) string {
	cookies := make([]string, 0, len(setCookieHeaders))
	for _, header := range setCookieHeaders {
		parts := strings.SplitN(string(header), ";", 2)
		if len(parts) == 0 {
			continue
		}
		cookiePair := strings.TrimSpace(parts[0])
		if cookiePair == "" || !strings.Contains(cookiePair, "=") {
			continue
		}
		cookies = append(cookies, cookiePair)
	}

	return strings.Join(cookies, "; ")
}

var (
	// tokenRefreshLock prevents concurrent TV object modifications from race conditions
	tokenRefreshLock sync.Mutex
	// nextCredentialValidationTime tracks when token validity should be checked next.
	// This prevents rechecking/re-refreshing on every request.
	nextCredentialValidationTime time.Time

	// drmMpdCache caches the DrmMpdOutput for a short duration to prevent double API calls
	// when IPTV apps request both the MPD and Key sequentially.
	drmMpdCache    sync.Map
	drmMpdCacheTTL = 30 * time.Second
)

// The proxied Broadpeak MPDs timestamp their segment timeline with the CDN's
// own clock (publishTime), which can differ from the machine clock by minutes.
// /dashtime is the UTCTiming clock source players sync to, so it must report
// the CDN's clock - not the machine's - or players compute the live edge far
// from where the segments actually are and playback stalls for the duration
// of the skew. These hold the last observed CDN publishTime so
// DASHTimeHandler can serve an extrapolated CDN clock. Note the cache is
// global across channels: if two channels' CDNs run different clocks, an
// interleaved fetch of one could briefly serve the wrong time to the other.
// This self-corrects because each MPD fetch refreshes the cache and players
// resolve /dashtime immediately after their own MPD fetch.
var (
	cdnClockMu          sync.Mutex
	cdnPublishTime      time.Time
	cdnPublishFetchedAt time.Time
)

// publishTimeRe extracts the publishTime attribute from live MPD bodies.
var publishTimeRe = regexp.MustCompile(`publishTime="([^"]+)"`)

// recordCdnPublishTime caches the upstream CDN clock observed from an MPD fetch.
func recordCdnPublishTime(publishTime time.Time) {
	cdnClockMu.Lock()
	defer cdnClockMu.Unlock()
	cdnPublishTime = publishTime
	cdnPublishFetchedAt = time.Now()
}

// cdnNow returns the CDN's current clock, extrapolated from the last observed
// publishTime, and whether any observation exists yet.
func cdnNow() (time.Time, bool) {
	cdnClockMu.Lock()
	defer cdnClockMu.Unlock()
	if cdnPublishFetchedAt.IsZero() {
		return time.Time{}, false
	}
	elapsed := time.Since(cdnPublishFetchedAt)
	if elapsed < 0 {
		elapsed = 0
	}
	return cdnPublishTime.Add(elapsed), true
}

// extractPublishTime parses the publishTime attribute from a DASH MPD body.
func extractPublishTime(body []byte) (time.Time, bool) {
	m := publishTimeRe.FindSubmatch(body)
	if len(m) < 2 {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339Nano, string(m[1]))
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

type drmMpdCacheEntry struct {
	Output    *DrmMpdOutput
	UpdatedAt time.Time
}

func getCachedDrmMpd(key string) *DrmMpdOutput {
	entryRaw, ok := drmMpdCache.Load(key)
	if !ok {
		return nil
	}
	entry, ok := entryRaw.(drmMpdCacheEntry)
	if !ok {
		drmMpdCache.Delete(key)
		return nil
	}
	if time.Since(entry.UpdatedAt) > drmMpdCacheTTL {
		drmMpdCache.Delete(key)
		return nil
	}
	return entry.Output
}

func setCachedDrmMpd(key string, output *DrmMpdOutput) {
	drmMpdCache.Store(key, drmMpdCacheEntry{Output: output, UpdatedAt: time.Now()})
}

// EnsureFreshCredentials refreshes tokens proactively before they expire
// This function prevents 403 errors by keeping credentials always fresh
// Returns true if tokens are fresh (either just refreshed or cached)
func EnsureFreshCredentials() bool {
	tokenRefreshLock.Lock()
	defer tokenRefreshLock.Unlock()

	now := time.Now()
	if !nextCredentialValidationTime.IsZero() && now.Before(nextCredentialValidationTime) {
		return true
	}

	credentials, err := utils.GetJIOTVCredentials()
	if err != nil || credentials == nil {
		nextCredentialValidationTime = now.Add(credentialRefreshRetryBackoff)
		if os.Getenv("JIOTV_DEBUG") == "true" && err != nil {
			utils.Log.Printf("[DEBUG] EnsureFreshCredentials: failed to load credentials: %v", err)
		}
		return true
	}

	refreshAccessToken := credentials.AccessToken != "" && credentials.RefreshToken != "" && shouldRefreshToken(
		credentials.AccessToken,
		credentials.LastTokenRefreshTime,
		jwtTokenRefreshLeadTime,
		accessTokenFallbackTTL,
		accessTokenFallbackLeadTime,
		now,
	)

	refreshSSOToken := credentials.SSOToken != "" && credentials.UniqueID != "" && shouldRefreshToken(
		credentials.SSOToken,
		credentials.LastSSOTokenRefreshTime,
		jwtTokenRefreshLeadTime,
		ssoTokenFallbackTTL,
		ssoTokenFallbackLeadTime,
		now,
	)

	if !refreshAccessToken && !refreshSSOToken {
		nextCredentialValidationTime = calculateNextCredentialValidationTime(credentials, now)
		return true
	}

	return performTokenRefresh(refreshAccessToken, refreshSSOToken, now)
}

// ForceRefreshCredentials bypasses proactive validity checks and forces immediate refresh
// Use this only in error recovery paths when we know tokens have failed
func ForceRefreshCredentials() bool {
	tokenRefreshLock.Lock()
	defer tokenRefreshLock.Unlock()
	now := time.Now()

	if os.Getenv("JIOTV_DEBUG") == "true" {
		utils.Log.Printf("[DEBUG] FORCED token refresh (bypassing expiry checks)")
	}

	credentials, err := utils.GetJIOTVCredentials()
	if err != nil || credentials == nil {
		nextCredentialValidationTime = now.Add(credentialRefreshRetryBackoff)
		if os.Getenv("JIOTV_DEBUG") == "true" && err != nil {
			utils.Log.Printf("[DEBUG] ForceRefreshCredentials: failed to load credentials: %v", err)
		}
		return false
	}

	refreshAccessToken := credentials.AccessToken != "" && credentials.RefreshToken != ""
	refreshSSOToken := credentials.SSOToken != "" && credentials.UniqueID != ""

	if !refreshAccessToken && !refreshSSOToken {
		nextCredentialValidationTime = now.Add(credentialRefreshRetryBackoff)
		return false
	}

	return performTokenRefresh(refreshAccessToken, refreshSSOToken, now)
}

// performTokenRefresh does the actual token refresh work (must be called with lock held)
func performTokenRefresh(refreshAccessToken, refreshSSOToken bool, now time.Time) bool {
	var accessTokenErr error
	var ssoTokenErr error
	var refreshed bool

	// CRITICAL REFRESH #1: Refresh AccessToken
	if refreshAccessToken {
		accessTokenErr = LoginRefreshAccessToken()
		if accessTokenErr != nil {
			if os.Getenv("JIOTV_DEBUG") == "true" {
				utils.Log.Printf("[DEBUG] AccessToken refresh error: %v", accessTokenErr)
			}
		} else {
			refreshed = true
			if os.Getenv("JIOTV_DEBUG") == "true" {
				utils.Log.Printf("[DEBUG] AccessToken refreshed successfully")
			}
		}
	}

	// CRITICAL REFRESH #2: Refresh SSOToken
	if refreshSSOToken {
		ssoTokenErr = LoginRefreshSSOToken()
		if ssoTokenErr != nil {
			if os.Getenv("JIOTV_DEBUG") == "true" {
				utils.Log.Printf("[DEBUG] SSOToken refresh error: %v", ssoTokenErr)
			}
		} else {
			refreshed = true
			if os.Getenv("JIOTV_DEBUG") == "true" {
				utils.Log.Printf("[DEBUG] SSOToken refreshed successfully")
			}
		}
	}

	if refreshed {
		freshCreds, freshErr := utils.GetJIOTVCredentials()
		if freshErr == nil && freshCreds != nil {
			nextCredentialValidationTime = calculateNextCredentialValidationTime(freshCreds, now)
		} else {
			nextCredentialValidationTime = now.Add(minCredentialValidationInterval)
		}
		if os.Getenv("JIOTV_DEBUG") == "true" {
			utils.Log.Printf("[DEBUG] Token refresh cycle completed. Next validation at %s", nextCredentialValidationTime.Format(time.RFC3339))
		}
		return true
	}

	nextCredentialValidationTime = now.Add(credentialRefreshRetryBackoff)

	// Both refreshes failed - log comprehensive error
	if os.Getenv("JIOTV_DEBUG") == "true" {
		utils.Log.Printf("[DEBUG] CRITICAL: Token refresh failed! AccessToken error: %v, SSOToken error: %v",
			accessTokenErr, ssoTokenErr)
	}
	return false
}

func calculateNextCredentialValidationTime(credentials *utils.JIOTV_CREDENTIALS, now time.Time) time.Time {
	nextChecks := make([]time.Time, 0, 2)

	if nextCheck, ok := nextTokenValidationCheck(
		credentials.AccessToken,
		credentials.LastTokenRefreshTime,
		jwtTokenRefreshLeadTime,
		accessTokenFallbackTTL,
		accessTokenFallbackLeadTime,
		now,
	); ok {
		nextChecks = append(nextChecks, nextCheck)
	}

	if nextCheck, ok := nextTokenValidationCheck(
		credentials.SSOToken,
		credentials.LastSSOTokenRefreshTime,
		jwtTokenRefreshLeadTime,
		ssoTokenFallbackTTL,
		ssoTokenFallbackLeadTime,
		now,
	); ok {
		nextChecks = append(nextChecks, nextCheck)
	}

	if len(nextChecks) == 0 {
		return now.Add(credentialRefreshRetryBackoff)
	}

	closest := nextChecks[0]
	for _, checkTime := range nextChecks[1:] {
		if checkTime.Before(closest) {
			closest = checkTime
		}
	}

	return closest
}

// getDrmMpd returns required properties for rendering DRM MPD
func getDrmMpd(channelID, quality string) (*DrmMpdOutput, error) {
	cacheKey := channelID + "_" + quality
	if cached := getCachedDrmMpd(cacheKey); cached != nil {
		return cached, nil
	}

	// Get live stream URL from JioTV API
	liveResult, err := TV.Live(channelID)
	if err != nil {
		return nil, err
	}
	if refreshedResult, refreshErr := refreshLiveResultIfNeeded(channelID, liveResult); refreshErr == nil && refreshedResult != nil {
		liveResult = refreshedResult
	}
	return buildDrmMpdOutput(liveResult, channelID, quality)
}

// buildDrmMpdOutput turns a playback response into the properties needed to
// render the DRM player. It is shared by live channels and premium provider
// content, which return the same payload shape.
func buildDrmMpdOutput(liveResult *television.LiveURLOutput, channelID, quality string) (*DrmMpdOutput, error) {
	cacheKey := channelID + "_" + quality

	// ResolvedBitrates covers both shapes: live channels nest URLs under
	// mpd.bitrates, premium content returns a single mpd.auto.
	bitrates := liveResult.Mpd.ResolvedBitrates()

	tv_url := internalUtils.SelectQuality(quality, bitrates.Auto, bitrates.High, bitrates.Medium, bitrates.Low)

	// If quality selection fails (empty), try to fallback to any available quality
	if tv_url == "" {
		if bitrates.High != "" {
			tv_url = bitrates.High
		} else if bitrates.Auto != "" {
			tv_url = bitrates.Auto
		} else if bitrates.Medium != "" {
			tv_url = bitrates.Medium
		} else if bitrates.Low != "" {
			tv_url = bitrates.Low
		}
	}

	if tv_url == "" {
		tv_url = liveResult.Mpd.Result
	}
	if tv_url == "" {
		output := &DrmMpdOutput{
			IsDRM:       liveResult.IsDRM,
			PlayUrl:     "",
			LicenseUrl:  "",
			Tv_url_host: "",
			Tv_url_path: "",
		}
		setCachedDrmMpd(cacheKey, output)
		return output, nil
	}

	channel_enc_url, err := secureurl.EncryptURL(tv_url)
	if err != nil {
		utils.Log.Panicln(err)
		return nil, err
	}

	// ResolvedLicenseURL covers both sources: live channels carry the license in
	// mpd.key, premium provider content returns a top-level keyUrl.
	licenseUrl := ""
	if resolvedLicenseURL := liveResult.ResolvedLicenseURL(); resolvedLicenseURL != "" {
		enc_key, encErr := secureurl.EncryptURL(resolvedLicenseURL)
		if encErr != nil {
			utils.Log.Panicln(encErr)
			return nil, encErr
		}
		licenseUrl = "/drm?auth=" + enc_key + "&channel_id=" + channelID + "&channel=" + channel_enc_url
	}

	// Quick fix for timesplay channels.
	if liveResult.AlgoName == "timesplay" {
		output := &DrmMpdOutput{
			IsDRM:       liveResult.IsDRM,
			PlayUrl:     tv_url,
			LicenseUrl:  licenseUrl,
			Tv_url_host: "",
			Tv_url_path: "",
		}
		setCachedDrmMpd(cacheKey, output)
		return output, nil
	}

	parsedTvUrl, err := url.Parse(tv_url)
	if err != nil {
		utils.Log.Panicln(err)
		return nil, err
	}
	tv_url_split := strings.Split(parsedTvUrl.Path, "/")
	tv_url_path, err := secureurl.EncryptURLDeterministic(strings.Join(tv_url_split[:len(tv_url_split)-1], "/") + "/")
	if err != nil {
		utils.Log.Panicln(err)
		return nil, err
	}

	tv_url_host, err := secureurl.EncryptURLDeterministic(parsedTvUrl.Host)
	if err != nil {
		utils.Log.Panicln(err)
		return nil, err
	}

	output := &DrmMpdOutput{
		IsDRM:       liveResult.IsDRM,
		PlayUrl:     "/render.mpd?auth=" + channel_enc_url + "&channel_id=" + channelID + "&q=" + quality,
		LicenseUrl:  licenseUrl,
		Tv_url_host: tv_url_host,
		Tv_url_path: tv_url_path,
	}
	setCachedDrmMpd(cacheKey, output)
	return output, nil
}

// LiveMpdHandler handles live stream routes /mpd/:channelID
func LiveMpdHandler(c *fiber.Ctx) error {
	// Get channel ID from URL
	channelID := c.Params("channelID")
	quality := c.Query("q")
	playerMode := c.Query("pm") // "hd" (force Shaka) or "auto" (try Shaka, fallback HLS)
	if quality == "" {
		quality = "auto"
	}
	if playerMode != "hd" && playerMode != "auto" {
		playerMode = "auto"
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

	// Get cookies from all Set-Cookie headers and forward them as a Cookie header.
	setCookieHeaders := resp.Header.PeekAll("Set-Cookie")
	if cookieHeader := buildCookieHeaderFromSetCookieHeaders(setCookieHeaders); cookieHeader != "" {
		c.Request().Header.Set("Cookie", cookieHeader)
	}

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

	// Capture the local server URL before proxying: proxy.Do rewrites the
	// request URI to the upstream CDN, after which c.Hostname()/c.Protocol()
	// would return the CDN's host and scheme instead of ours.
	localServerURL := c.Protocol() + "://" + c.Hostname()

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
					utils.Log.Println(err)
					return err
				}
			}
		}
	}

	proxyHost := parsedUrl.Host
	pathParts := strings.Split(parsedUrl.Path, "/")
	basePath := strings.Join(pathParts[:len(pathParts)-1], "/") + "/"
	encProxyHost, err := secureurl.EncryptURLDeterministic(proxyHost)
	if err != nil {
		utils.Log.Panicln(err)
		return err
	}
	encProxyPath, err := secureurl.EncryptURLDeterministic(basePath)
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
		encHDNEA, encErr := secureurl.EncryptURLDeterministic("__hdnea__=" + cachedHDNEA)
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

	// Handle 403/401 auth failures by stripping HDNEA and retrying.
	// HDNEA tokens are CDN-managed and expire independently of the JioTV
	// session tokens, so the first retry drops the stale HDNEA and lets the
	// CDN issue fresh auth. A full credential refresh is only attempted if
	// that still fails, avoiding a blocking token refresh on every expired
	// manifest request.
	statusCode := c.Response().StatusCode()
	if statusCode == fiber.StatusForbidden || statusCode == fiber.StatusUnauthorized {
		if os.Getenv("JIOTV_DEBUG") == "true" {
			utils.Log.Printf("[DEBUG] MpdHandler got %d response - stripping HDNEA and retrying", statusCode)
		}

		// Reset response to allow retry
		c.Response().Reset()

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

		// If stripping HDNEA did not help, the JioTV session tokens may have
		// expired; refresh them and retry once more.
		if retryStatus := c.Response().StatusCode(); retryStatus == fiber.StatusForbidden || retryStatus == fiber.StatusUnauthorized {
			if os.Getenv("JIOTV_DEBUG") == "true" {
				utils.Log.Printf("[DEBUG] MpdHandler retry still %d - forcing credential refresh", retryStatus)
			}
			c.Response().Reset()
			ForceRefreshCredentials()
			if err := proxy.Do(c, strippedUrl, TV.Client); err != nil {
				if os.Getenv("JIOTV_DEBUG") == "true" {
					utils.Log.Printf("[DEBUG] MpdHandler forced-refresh retry failed: %v", err)
				}
				return err
			}
		}

		if os.Getenv("JIOTV_DEBUG") == "true" {
			utils.Log.Printf("[DEBUG] MpdHandler retry - new status: %d", c.Response().StatusCode())
		}
	}

	c.Response().Header.Del(fiber.HeaderServer)

	// Extract __hdnea__ from upstream response for injecting into dashBaseURL
	upstreamHDNEA := ""

	setCookieHeaders := c.Response().Header.PeekAll("Set-Cookie")
	upstreamHDNEA = extractCookieValueFromSetCookieHeaders(setCookieHeaders, "__hdnea__")

	// If we got a fresh __hdnea__ from upstream, update dashBaseURL with it
	if upstreamHDNEA != "" {
		encHDNEA, encErr := secureurl.EncryptURLDeterministic("__hdnea__=" + upstreamHDNEA)
		if encErr == nil {
			dashBaseURL = fmt.Sprintf("/render.dash/host/%s/path/%s/hdnea/%s", encProxyHost, encProxyPath, encHDNEA)
		}
	}

	// Delete Domain from cookies
	if len(setCookieHeaders) > 0 {
		c.Response().Header.Del("Set-Cookie")
		domainBytes := []byte("Domain=" + proxyHost + ";")

		for _, rawCookie := range setCookieHeaders {
			cookie := append([]byte(nil), rawCookie...)
			cookie = bytes.Replace(cookie, domainBytes, []byte(""), 1)
			// Modify path in cookies
			cookie = bytes.Replace(cookie, []byte("path=/"), []byte("path=/render.dash"), 1)

			// Restore all Set-Cookie headers after normalization.
			c.Response().Header.Add("Set-Cookie", string(cookie))
		}
	}
	resBody := c.Response().Body()

	// Record the upstream CDN clock so /dashtime can report it: the segment
	// timeline in this MPD is stamped in CDN time, and players must sync to
	// that clock to find the live edge.
	if pt, ok := extractPublishTime(resBody); ok {
		recordCdnPublishTime(pt)
	}

	// Inject a UTCTiming element pointing at the local /dashtime clock source.
	// JioTV's Broadpeak MPDs carry no UTCTiming element and use
	// availabilityStartTime="1970-01-01T00:00:00Z", so without a clock source
	// players compute the live edge from the client clock. A skewed client
	// clock makes them hunt for the live edge and pre-roll a large chunk of
	// the DVR window before playback can start. Serving the clock from the
	// server fixes this for every client (Shaka, IPTV apps, ExoPlayer...),
	// not just the browser player.
	if !bytes.Contains(resBody, []byte("<UTCTiming")) {
		utcTiming := fmt.Sprintf(
			`<UTCTiming schemeIdUri="urn:mpeg:dash:utc:http-xsdate:2014" value="%s/dashtime"/>`,
			localServerURL,
		)
		mpdTagPattern := regexp.MustCompile(`<MPD[^>]*>`)
		resBody = mpdTagPattern.ReplaceAllFunc(resBody, func(match []byte) []byte {
			return append(append([]byte{}, match...), []byte(utcTiming)...)
		})
	}

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

	// Handle 403/401 auth failures by clearing the stale HDNEA cookie and
	// retrying. Segment requests carry only the CDN's __hdnea__ cookie (not
	// the JioTV session tokens), so an expired token is a CDN auth problem:
	// dropping the cookie lets the CDN issue fresh auth without the cost of a
	// blocking token refresh on every expired segment request. A full refresh
	// is attempted only if the retry still fails.
	statusCode := c.Response().StatusCode()
	if statusCode == fiber.StatusForbidden || statusCode == fiber.StatusUnauthorized {
		if os.Getenv("JIOTV_DEBUG") == "true" {
			utils.Log.Printf("[DEBUG] DashHandler got %d response - clearing HDNEA cookie and retrying", statusCode)
		}

		// Reset response to allow retry
		c.Response().Reset()

		// Clear HDNEA cookie - expired token causes 403
		// CDN will provide fresh HDNEA in the response
		c.Request().Header.DelCookie("__hdnea__")

		if err := proxy.Do(c, proxyUrl, TV.Client); err != nil {
			if os.Getenv("JIOTV_DEBUG") == "true" {
				utils.Log.Printf("[DEBUG] DashHandler retry failed: %v", err)
			}
			return err
		}

		// If clearing the cookie did not help, the JioTV session tokens may
		// have expired; refresh them and retry once more.
		if retryStatus := c.Response().StatusCode(); retryStatus == fiber.StatusForbidden || retryStatus == fiber.StatusUnauthorized {
			if os.Getenv("JIOTV_DEBUG") == "true" {
				utils.Log.Printf("[DEBUG] DashHandler retry still %d - forcing credential refresh", retryStatus)
			}
			c.Response().Reset()
			ForceRefreshCredentials()
			if err := proxy.Do(c, proxyUrl, TV.Client); err != nil {
				if os.Getenv("JIOTV_DEBUG") == "true" {
					utils.Log.Printf("[DEBUG] DashHandler forced-refresh retry failed: %v", err)
				}
				return err
			}
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

// LiveManifestMpdHandler handles the IPTV M3U route for MPD manifests: /live/mpd/:channelID
func LiveManifestMpdHandler(c *fiber.Ctx) error {
	channelID := c.Params("channelID")
	quality := c.Query("q")
	if quality == "" {
		quality = "auto"
	}

	EnsureFreshCredentials()

	drmMpdOutput, err := getDrmMpd(channelID, quality)
	if err != nil {
		utils.Log.Printf("Error getting DRM MPD: %v", err)
		return internalUtils.InternalServerError(c, err.Error())
	}

	if drmMpdOutput.PlayUrl == "" {
		return internalUtils.NotFoundError(c, "No MPD URL found for channel "+channelID)
	}

	return c.Redirect(drmMpdOutput.PlayUrl, fiber.StatusFound)
}

// LiveManifestKeyHandler acts as a proxy for the Widevine license request for IPTV clients: /live/key/:channelID
func LiveManifestKeyHandler(c *fiber.Ctx) error {
	channelID := c.Params("channelID")
	quality := c.Query("q")
	if quality == "" {
		quality = "auto"
	}

	// The MPD handler was likely called just milliseconds ago,
	// so getDrmMpd will instantly return the cached result.
	drmMpdOutput, err := getDrmMpd(channelID, quality)
	if err != nil {
		utils.Log.Printf("Error getting DRM Key info: %v", err)
		return internalUtils.InternalServerError(c, err.Error())
	}

	if drmMpdOutput.LicenseUrl == "" {
		return internalUtils.NotFoundError(c, "No License URL found for channel "+channelID)
	}

	parsedUrl, err := url.Parse(drmMpdOutput.LicenseUrl)
	if err != nil {
		return internalUtils.InternalServerError(c, err.Error())
	}

	// Inject query parameters into the context so DRMKeyHandler can read them
	c.Request().URI().SetQueryString(parsedUrl.RawQuery)

	return DRMKeyHandler(c)
}
