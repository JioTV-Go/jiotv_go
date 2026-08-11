package television

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	neturl "net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/valyala/fasthttp"
	"gopkg.in/yaml.v3"

	"github.com/jiotv-go/jiotv_go/v3/internal/config"
	"github.com/jiotv-go/jiotv_go/v3/internal/constants"
	"github.com/jiotv-go/jiotv_go/v3/internal/constants/headers"
	"github.com/jiotv-go/jiotv_go/v3/internal/constants/urls"
	"github.com/jiotv-go/jiotv_go/v3/pkg/store"
	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
)

const (
	// JioTV API domain constants
	JIOTV_API_DOMAIN = urls.JioTVAPIDomain
	TV_MEDIA_DOMAIN  = urls.TVMediaDomain
	JIOTV_CDN_DOMAIN = urls.JioTVCDNDomain

	// URL for fetching channels from JioTV API
	CHANNELS_API_URL               = urls.ChannelsAPIURL
	CHANNELS_AUTH_API_URL          = urls.ChannelURL
	PLANS_API_URL                  = urls.PlansAPIURL
	ACTIVE_PLANS_API_URL           = urls.ActivePlansAPIURL
	PROVIDER_CONFIG_API_BASE_URL   = urls.ProviderConfigAPIBaseURL
	PROVIDER_METADATA_API_BASE_URL = urls.ProviderMetadataAPIBaseURL
	// Error message for unsupported custom channels file formats
	errUnsupportedChannelsFormat = constants.ErrUnsupportedChannelsFormat
	// Maximum recommended number of custom channels before performance warnings
	maxRecommendedChannels = constants.MaxRecommendedChannels
)

// logExcessiveChannelsWarning logs a comprehensive warning when the number of custom channels exceeds the recommended limit
func logExcessiveChannelsWarning(channelCount int, context string) {
	if channelCount <= maxRecommendedChannels {
		return
	}

	utils.SafeLogf("WARNING: %s %d custom channels, which exceeds the recommended limit of %d channels.", context, channelCount, maxRecommendedChannels)
	utils.SafeLog("WARNING: Large numbers of custom channels may impact performance:")
	utils.SafeLog("  - Slower channel listing and filtering operations")
	utils.SafeLog("  - Increased memory usage")
	utils.SafeLog("  - Longer startup times")
	utils.SafeLog("  - Potential UI responsiveness issues")
	utils.SafeLog("Consider splitting channels into multiple configuration files or reducing the total number.")
}

var (
	// customChannelsCacheMap holds cached custom channels indexed by ID for efficient lookups
	customChannelsCacheMap map[string]Channel
	customChannelsMu       sync.RWMutex
)

type premiumProviderLink struct {
	ProviderID  string
	DisplayName string
	URL         string
}

var premiumProviderByID = map[string]premiumProviderLink{
	"Z0177": {
		ProviderID:  "Z0177",
		DisplayName: "FanCode",
		URL:         "https://www.fancode.com/",
	},
}

var premiumProviderByPlanID = map[string]premiumProviderLink{
	"1019037": {
		ProviderID:  "Z0177",
		DisplayName: "FanCode",
		URL:         "https://www.fancode.com/",
	},
}

var premiumProviderByKeyword = []struct {
	Keyword string
	Link    premiumProviderLink
}{
	{
		Keyword: "fancode",
		Link: premiumProviderLink{
			ProviderID:  "Z0177",
			DisplayName: "FanCode",
			URL:         "https://www.fancode.com/",
		},
	},
	{
		Keyword: "jiocinema",
		Link: premiumProviderLink{
			ProviderID:  "",
			DisplayName: "JioCinema Premium",
			URL:         "https://www.jiocinema.com/",
		},
	},
}

// New function creates a new Television instance with the provided credentials
func New(credentials *utils.JIOTV_CREDENTIALS) *Television {
	// Check if credentials are provided
	if credentials == nil {
		// If credentials are not provided, set them to empty strings
		credentials = &utils.JIOTV_CREDENTIALS{
			AccessToken: "",
			SSOToken:    "",
			CRM:         "",
			UniqueID:    "",
		}
	}
	headers := map[string]string{
		"Content-type":    "application/x-www-form-urlencoded",
		"appkey":          "NzNiMDhlYzQyNjJm",
		"channel_id":      "",
		"crmid":           credentials.CRM,
		"userId":          credentials.CRM,
		"deviceId":        utils.GetDeviceID(),
		"devicetype":      "phone",
		"isott":           "false",
		"languageId":      "6",
		"lbcookie":        "1",
		"os":              "android",
		"osVersion":       "13",
		"subscriberId":    credentials.CRM,
		"uniqueId":        credentials.UniqueID,
		headers.UserAgent: headers.UserAgentOkHttp,
		"usergroup":       "tvYR7NSNn7rymo3F",
		"versionCode":     headers.VersionCode413,
	}

	// Create a fasthttp.Client
	client := utils.GetRequestClient()

	// Return a new Television instance
	return &Television{
		AccessToken: credentials.AccessToken,
		SsoToken:    credentials.SSOToken,
		Crm:         credentials.CRM,
		UniqueID:    credentials.UniqueID,
		Headers:     headers,
		Client:      client,
	}
}

// InitCustomChannels initializes custom channels at startup if configured
func InitCustomChannels() {
	if config.Cfg.CustomChannelsFile != "" {
		loadAndCacheCustomChannels()
	}
}

func ReloadCustomChannels() {
	if config.Cfg.CustomChannelsFile != "" {
		loadAndCacheCustomChannels()
	}
}

// getCustomChannelByID efficiently looks up a custom channel by ID
func getCustomChannelByID(channelID string) (Channel, bool) {
	customChannelsMu.RLock()
	defer customChannelsMu.RUnlock()

	if customChannelsCacheMap == nil {
		return Channel{}, false
	}

	channel, exists := customChannelsCacheMap[channelID]
	return channel, exists
}

// GetCustomChannelByID efficiently looks up a custom channel by ID (public version)
func GetCustomChannelByID(channelID string) (Channel, bool) {
	return getCustomChannelByID(channelID)
}

// loadAndCacheCustomChannels loads custom channels from file and caches them
func loadAndCacheCustomChannels() {
	channels, err := LoadCustomChannels(config.Cfg.CustomChannelsFile)
	next := make(map[string]Channel)
	if err != nil {
		utils.SafeLogf("Error loading custom channels: %v", err)
	} else {
		for _, channel := range channels {
			next[channel.ID] = channel
		}

		logExcessiveChannelsWarning(len(channels), "Cached")
	}

	customChannelsMu.Lock()
	customChannelsCacheMap = next
	customChannelsMu.Unlock()
}

// Live method generates m3u8 link from JioTV API with the provided channel ID
func (tv *Television) Live(channelID string) (*LiveURLOutput, error) {
	// If channelID starts with sl, then it is a Sony Channel
	if len(channelID) >= 2 && channelID[:2] == "sl" {
		return getSLChannel(channelID)
	}

	formData := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(formData)

	formData.Add("channel_id", channelID)
	formData.Add("stream_type", "Seek")
	formData.Add("begin", utils.GenerateCurrentTime())
	formData.Add("srno", utils.GenerateDate())

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	// Copy headers from the Television headers map to the request
	for key, value := range tv.Headers {
		req.Header.Set(key, value)
	}

	// Always use the v1.1 API endpoint
	url := "https://" + JIOTV_API_DOMAIN + urls.PlaybackAPIPath
	req.Header.Set(headers.AccessToken, tv.AccessToken)
	req.SetRequestURI(url)
	req.Header.SetMethod("POST")

	// Encode the form data and set it as the request body
	req.SetBody(formData.QueryString())

	req.Header.Set("channel_id", channelID)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Perform the HTTP POST request
	if err := tv.Client.Do(req, resp); err != nil {
		if strings.Contains(err.Error(), "server closed connection before returning the first response byte") {
			utils.Log.Println("Retrying the request...")
			return tv.Live(channelID)
		}
		utils.Log.Println(err)
		return nil, err
	}
	if resp.StatusCode() != fasthttp.StatusOK {
		// Store the response body as a string
		response := string(resp.Body())

		// Log headers and request data
		utils.Log.Println("Request headers:", req.Header.String())
		utils.Log.Println("Request data:", formData.String())
		utils.Log.Println("Response: ", response)

		return nil, fmt.Errorf("Request failed with status code: %d\nresponse: %s", resp.StatusCode(), response)
	}

	var result LiveURLOutput
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		utils.Log.Println(err)
		return nil, err
	}

	// Extract hdnea from any URL fields in the response (Live does not set Set-Cookie)
	extractHdneaFromURL := func(u string) string {
		if u == "" {
			return ""
		}
		idx := strings.Index(u, "hdnea=")
		if idx == -1 {
			return ""
		}
		// token starts after hdnea=
		token := u[idx+len("hdnea="):]
		if i := strings.IndexByte(token, '&'); i != -1 {
			token = token[:i]
		}
		return token
	}
	hdnea := extractHdneaFromURL(result.Bitrates.Auto)
	if hdnea == "" {
		hdnea = extractHdneaFromURL(result.Mpd.Result)
	}
	result.Hdnea = hdnea

	// If hdnea exists and URLs don't already have it, append as query param
	if hdnea != "" {
		appendHdnea := func(u string) string {
			if u == "" {
				return u
			}
			if strings.Contains(u, "hdnea=") {
				return u
			}
			sep := "?"
			if strings.Contains(u, "?") {
				sep = "&"
			}
			return u + sep + "hdnea=" + hdnea
		}
		result.Bitrates.Auto = appendHdnea(result.Bitrates.Auto)
		result.Bitrates.High = appendHdnea(result.Bitrates.High)
		result.Bitrates.Medium = appendHdnea(result.Bitrates.Medium)
		result.Bitrates.Low = appendHdnea(result.Bitrates.Low)
		result.Result = appendHdnea(result.Result)
		if result.Mpd.Result != "" {
			result.Mpd.Result = appendHdnea(result.Mpd.Result)
		}
		if result.Mpd.Key != "" {
			result.Mpd.Key = appendHdnea(result.Mpd.Key)
		}
	}

	return &result, nil
}

// Render method does HTTP GET request to the provided URL and return the response body
func (tv *Television) Render(streamURL string, hdneaToken string) ([]byte, int, string) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(streamURL)
	req.Header.SetMethod("GET")

	// Copy headers from the Television headers map to the request
	for key, value := range tv.Headers {
		req.Header.Set(key, value)
	}

	// Override User-Agent to simulate a player, as some CDNs (e.g. for channel 182) block okhttp
	req.Header.Set("User-Agent", headers.UserAgentPlayTV)

	// Prefer explicit token override from handler cache; otherwise derive from URL query.
	// When both hdnea and __hdnea__ are present, prefer __hdnea__ as the fresher token.
	if hdneaToken != "" {
		req.Header.SetCookie("__hdnea__", hdneaToken)
	} else if strings.Contains(streamURL, "hdnea=") {
		// quick parse to extract value
		q := streamURL[strings.Index(streamURL, "?")+1:]
		var parsedToken string
		for _, p := range strings.Split(q, "&") {
			if strings.HasPrefix(p, "__hdnea__=") {
				token := strings.TrimPrefix(p, "__hdnea__=")
				if decodedToken, decodeErr := neturl.QueryUnescape(token); decodeErr == nil {
					token = decodedToken
				}
				parsedToken = token
				break
			}
			if parsedToken == "" && strings.HasPrefix(p, "hdnea=") {
				token := strings.TrimPrefix(p, "hdnea=")
				if decodedToken, decodeErr := neturl.QueryUnescape(token); decodeErr == nil {
					token = decodedToken
				}
				parsedToken = token
			}
		}
		if parsedToken != "" {
			req.Header.SetCookie("__hdnea__", parsedToken)
		}
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Perform the HTTP GET request
	if err := tv.Client.Do(req, resp); err != nil {
		utils.Log.Println("Render upstream request failed:", err)
		return []byte(""), fasthttp.StatusBadGateway, ""
	}

	buf := resp.Body()
	// Capture any __hdnea__ Set-Cookie returned by upstream so caller can set cookie on client
	var newHdnea string
	for _, setCookie := range resp.Header.PeekAll("Set-Cookie") {
		setCookieStr := string(setCookie)
		if !strings.Contains(setCookieStr, "__hdnea__=") {
			continue
		}

		parts := strings.Split(setCookieStr, ";")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if strings.HasPrefix(trimmed, "__hdnea__=") {
				newHdnea = strings.TrimPrefix(trimmed, "__hdnea__=")
				break
			}
		}

		if newHdnea != "" {
			break
		}
	}

	return buf, resp.StatusCode(), newHdnea
}

// detectAndParseFormat attempts to detect the format of custom channels data and parse it
func detectAndParseFormat(data []byte, filePath string) (CustomChannelsConfig, error) {
	var customConfig CustomChannelsConfig

	// Determine file format by extension and parse accordingly, fallback to content-based detection
	if strings.HasSuffix(filePath, ".json") {
		err := json.Unmarshal(data, &customConfig)
		return customConfig, err
	}

	if strings.HasSuffix(filePath, ".yml") || strings.HasSuffix(filePath, ".yaml") {
		err := yaml.Unmarshal(data, &customConfig)
		return customConfig, err
	}

	// Fallback: try to detect format by content for unknown extensions
	trimmed := strings.TrimSpace(string(data))

	// For unsupported extensions, require non-empty content
	if trimmed == "" {
		return customConfig, errors.New(errUnsupportedChannelsFormat)
	}

	// Try JSON if content starts with '{' or '['
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		err := json.Unmarshal(data, &customConfig)
		if err == nil {
			return customConfig, nil
		}
		// If JSON parsing failed, try YAML as fallback
		err = yaml.Unmarshal(data, &customConfig)
		if err != nil {
			return customConfig, errors.New(errUnsupportedChannelsFormat)
		}
		return customConfig, nil
	}

	// Try YAML for other content
	err := yaml.Unmarshal(data, &customConfig)
	if err != nil {
		return customConfig, errors.New(errUnsupportedChannelsFormat)
	}
	return customConfig, nil
}

// LoadCustomChannels loads custom channels from configuration file
func LoadCustomChannels(filePath string) ([]Channel, error) {
	if filePath == "" {
		return []Channel{}, nil
	}

	// Check if file exists and read it
	fileResult := utils.CheckAndReadFile(filePath)
	if !fileResult.Exists {
		utils.SafeLogf("Custom channels file not found: %s", filePath)
		if isDefaultCustomChannelsPath(filePath) {
			customConfig, err := loadBuiltInCustomChannelsConfig()
			if err == nil {
				return convertCustomConfigToChannels(customConfig), nil
			}
		}
		return []Channel{}, nil
	}

	if fileResult.Error != nil {
		return nil, fileResult.Error
	}

	// Parse the file using format detection
	customConfig, err := detectAndParseFormat(fileResult.Data, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse custom channels file: %w", err)
	}

	channels := convertCustomConfigToChannels(customConfig)

	utils.SafeLogf("Loaded %d custom channels from %s", len(channels), filePath)

	// Warn user about performance implications if too many channels
	logExcessiveChannelsWarning(len(channels), "You have loaded")
	return channels, nil
}

func getCustomChannels() []Channel {
	customChannelsMu.RLock()
	defer customChannelsMu.RUnlock()

	var customChannels []Channel
	for _, channel := range customChannelsCacheMap {
		customChannels = append(customChannels, channel)
	}
	return customChannels
}

func isDefaultCustomChannelsPath(filePath string) bool {
	base := strings.ToLower(filepath.Base(filePath))
	if base != "custom-channels.json" && base != "custom_channels.json" && base != "custom-channels.yml" && base != "custom_channels.yml" && base != "custom-channels.yaml" && base != "custom_channels.yaml" {
		return false
	}
	return true
}

func loadBuiltInCustomChannelsConfig() (CustomChannelsConfig, error) {
	var customConfig CustomChannelsConfig
	if err := json.Unmarshal([]byte(builtInCustomChannelsJSON), &customConfig); err != nil {
		return CustomChannelsConfig{}, err
	}
	return customConfig, nil
}

func convertCustomConfigToChannels(customConfig CustomChannelsConfig) []Channel {
	var channels []Channel
	for _, customChannel := range customConfig.Channels {
		channelID := customChannel.ID
		if !strings.HasPrefix(channelID, "cc_") {
			channelID = "cc_" + channelID
		}

		channel := Channel{
			ID:       channelID,
			Name:     customChannel.Name,
			URL:      customChannel.URL,
			LogoURL:  customChannel.LogoURL,
			Category: customChannel.Category,
			Language: customChannel.Language,
			IsHD:     customChannel.IsHD,
		}
		channels = append(channels, channel)
	}
	return channels
}

const builtInCustomChannelsJSON = `{
  "channels": [
    {
      "id": "custom_news_1",
      "name": "Sample News Channel",
      "url": "https://example.com/news/playlist.m3u8",
      "logo_url": "https://example.com/logos/news.png",
      "category": 12,
      "language": 6,
      "is_hd": true
    },
    {
      "id": "custom_entertainment_1",
      "name": "Sample Entertainment Channel",
      "url": "https://example.com/entertainment/playlist.m3u8",
      "logo_url": "https://example.com/logos/entertainment.png",
      "category": 5,
      "language": 1,
      "is_hd": false
    }
  ]
}`

func fetchChannelsFromAPI(client *fasthttp.Client, channelAPIURL string, requestHeaders map[string]string) (ChannelsResponse, error) {
	requestConfig := utils.HTTPRequestConfig{
		URL:     channelAPIURL,
		Method:  "GET",
		Headers: requestHeaders,
	}

	resp, err := utils.MakeHTTPRequest(requestConfig, client)
	if err != nil {
		return ChannelsResponse{}, err
	}
	defer fasthttp.ReleaseResponse(resp)

	var apiResponse ChannelsResponse
	if err := utils.ParseJSONResponse(resp, &apiResponse); err != nil {
		return ChannelsResponse{}, err
	}

	return apiResponse, nil
}

func fetchPlansFromAPI(client *fasthttp.Client, plansAPIURL string, requestHeaders map[string]string) (PlansResponse, error) {
	requestConfig := utils.HTTPRequestConfig{
		URL:     plansAPIURL,
		Method:  "GET",
		Headers: requestHeaders,
	}

	resp, err := utils.MakeHTTPRequest(requestConfig, client)
	if err != nil {
		return PlansResponse{}, err
	}
	defer fasthttp.ReleaseResponse(resp)

	var apiResponse PlansResponse
	if err := utils.ParseJSONResponse(resp, &apiResponse); err != nil {
		utils.SafeLogf("Plans API returned unparseable body (status %d): %s", resp.StatusCode(), truncateForLog(resp.Body(), 600))
		return PlansResponse{}, err
	}

	if len(apiResponse.PackageInfo) == 0 && len(apiResponse.Result.Plans) == 0 {
		utils.SafeLogf("Plans API returned no packages (status %d): %s", resp.StatusCode(), truncateForLog(resp.Body(), 600))
	}

	return apiResponse, nil
}

// describePremiumProviders renders providers as "id=name" pairs for diagnostics.
func describePremiumProviders(premiumProviders []PremiumProvider) string {
	if len(premiumProviders) == 0 {
		return "none"
	}
	descriptions := make([]string, 0, len(premiumProviders))
	for _, provider := range premiumProviders {
		descriptions = append(descriptions, provider.ProviderID+"="+provider.Name)
	}
	return strings.Join(descriptions, ", ")
}

// truncateForLog renders a response body for diagnostics, capped at maxLen bytes.
func truncateForLog(body []byte, maxLen int) string {
	if len(body) <= maxLen {
		return string(body)
	}
	return string(body[:maxLen]) + "...(truncated)"
}

func buildAuthenticatedHeaders(credentials *utils.JIOTV_CREDENTIALS) map[string]string {
	authHeaders := map[string]string{
		headers.UserAgent:   headers.UserAgentOkHttp,
		headers.Accept:      headers.AcceptJSON,
		headers.DeviceType:  headers.DeviceTypePhone,
		headers.OS:          headers.OSAndroid,
		headers.VersionCode: headers.VersionCode413,
		"appkey":            "NzNiMDhlYzQyNjJm",
		"lbcookie":          "1",
		"usertype":          "JIO",
		"usergroup":         "tvYR7NSNn7rymo3F",
		"languageId":        "6",
		"isott":             "false",
	}

	if credentials == nil {
		return authHeaders
	}

	if credentials.AccessToken != "" {
		authHeaders[headers.AccessToken] = credentials.AccessToken
	}
	if credentials.SSOToken != "" {
		authHeaders["ssotoken"] = credentials.SSOToken
	}
	if credentials.CRM != "" {
		authHeaders["crmid"] = credentials.CRM
		authHeaders["subscriberId"] = credentials.CRM
		authHeaders["userId"] = credentials.CRM
	}
	if credentials.UniqueID != "" {
		authHeaders["uniqueId"] = credentials.UniqueID
	}

	return authHeaders
}

func resolvePremiumProvider(providerID, providerName string) (premiumProviderLink, bool) {
	normalizedProviderID := strings.ToUpper(strings.TrimSpace(providerID))
	if normalizedProviderID != "" {
		if providerLink, exists := premiumProviderByID[normalizedProviderID]; exists {
			return providerLink, true
		}
	}

	normalizedProviderName := strings.ToLower(strings.TrimSpace(providerName))
	for _, premiumProvider := range premiumProviderByKeyword {
		if strings.Contains(normalizedProviderName, premiumProvider.Keyword) {
			return premiumProvider.Link, true
		}
	}

	return premiumProviderLink{}, false
}

// collectPlanProviders converts provider entries from the plans API into
// PremiumProvider values, appending to premiumProviders and tracking duplicates
// in seenProviders. Every provider returned by the API is kept: the API only
// lists providers the account is actually entitled to, so filtering against a
// local allowlist would drop legitimate providers. The local registry is used
// only to enrich entries with an external URL where one is known.
func collectPlanProviders(planProviders []PlanProvider, premiumProviders *[]PremiumProvider, seenProviders map[string]struct{}) {
	for _, provider := range planProviders {
		providerID := strings.ToUpper(provider.resolvedID())
		providerName := provider.resolvedName()

		providerKey := providerID
		if providerKey == "" {
			providerKey = strings.ToLower(providerName)
		}
		if providerKey == "" {
			continue
		}
		if _, alreadySeen := seenProviders[providerKey]; alreadySeen {
			continue
		}
		seenProviders[providerKey] = struct{}{}

		providerLink, _ := resolvePremiumProvider(providerID, providerName)

		displayName := providerName
		if displayName == "" {
			displayName = providerLink.DisplayName
		}
		if displayName == "" {
			displayName = providerID
		}
		canonicalProviderID := providerID
		if canonicalProviderID == "" {
			canonicalProviderID = providerLink.ProviderID
		}

		// provider is the plans API's taxonomy name (e.g. "Fancode") and matches
		// the provider directory exactly, so it is tried before the display name.
		matchNames := make([]string, 0, 2)
		if taxonomyName := strings.TrimSpace(provider.Provider); taxonomyName != "" {
			matchNames = append(matchNames, taxonomyName)
		}
		if displayName != "" {
			matchNames = append(matchNames, displayName)
		}

		*premiumProviders = append(*premiumProviders, PremiumProvider{
			ID:         providerKey,
			ProviderID: canonicalProviderID,
			Name:       displayName,
			URL:        providerLink.URL,
			Image:      strings.TrimSpace(provider.Image),
			matchNames: matchNames,
		})
	}
}

// fetchProviderDirectory retrieves the full provider config document and maps
// each provider's short name (lowercased, e.g. "fancode") to the content
// provider ID used by the catalog API (e.g. "200169"). Requesting the document
// without a providerId returns every provider.
func fetchProviderDirectory(client *fasthttp.Client, requestHeaders map[string]string) (map[string]string, error) {
	requestConfig := utils.HTTPRequestConfig{
		URL:     strings.TrimRight(PROVIDER_CONFIG_API_BASE_URL, "/") + "/cnf/provider",
		Method:  "GET",
		Headers: requestHeaders,
	}

	resp, err := utils.MakeHTTPRequest(requestConfig, client)
	if err != nil {
		return nil, err
	}
	defer fasthttp.ReleaseResponse(resp)

	var directoryResponse PremiumProviderFilterResponse
	if err := utils.ParseJSONResponse(resp, &directoryResponse); err != nil {
		return nil, err
	}

	providerDirectory := make(map[string]string)
	for _, providerEntry := range directoryResponse.Data.Provider {
		providerName := normalizeProviderName(providerEntry.Provider)
		providerID := strings.TrimSpace(providerEntry.ProviderID)
		if providerName == "" || providerID == "" {
			continue
		}
		// The directory lists some providers more than once (for example
		// several "Simca" entries); the first entry with an ID wins.
		if _, exists := providerDirectory[providerName]; !exists {
			providerDirectory[providerName] = providerID
		}
	}

	return providerDirectory, nil
}

// resolveContentProviderIDs replaces each provider's entitlement ID with the
// content provider ID used by the catalog API, matching on provider name.
// Providers with no match keep their entitlement ID.
func resolveContentProviderIDs(premiumProviders []PremiumProvider, providerDirectory map[string]string) {
	if len(providerDirectory) == 0 {
		return
	}

	for index := range premiumProviders {
		provider := &premiumProviders[index]
		names := provider.matchNames
		if len(names) == 0 {
			names = []string{provider.Name}
		}
		if contentProviderID, matched := lookupContentProviderID(names, providerDirectory); matched {
			provider.ProviderID = contentProviderID
		}
	}
}

// lookupContentProviderID resolves the first of names that can be matched
// against the provider directory. Exact matches on every name are preferred
// over a substring match on any of them.
func lookupContentProviderID(names []string, providerDirectory map[string]string) (string, bool) {
	candidates := make([]string, 0, len(names)*2)
	for _, name := range names {
		candidates = append(candidates, providerNameCandidates(name)...)
	}

	for _, candidate := range candidates {
		if contentProviderID, exists := providerDirectory[candidate]; exists {
			return contentProviderID, true
		}
	}

	// Fall back to containment, which covers names carrying an extra qualifier
	// ("JioCinema Premium" vs "JioCinema", "LionsgatePlay" vs "LionsGate").
	// The length guard keeps very short names from matching too eagerly, and the
	// longest directory name wins so the match is both specific and stable
	// regardless of map iteration order.
	for _, candidate := range candidates {
		if len(candidate) < 4 {
			continue
		}
		bestName := ""
		bestProviderID := ""
		for directoryName, contentProviderID := range providerDirectory {
			if len(directoryName) < 4 {
				continue
			}
			if !strings.Contains(candidate, directoryName) && !strings.Contains(directoryName, candidate) {
				continue
			}
			if len(directoryName) > len(bestName) || (len(directoryName) == len(bestName) && directoryName < bestName) {
				bestName = directoryName
				bestProviderID = contentProviderID
			}
		}
		if bestProviderID != "" {
			return bestProviderID, true
		}
	}

	return "", false
}

// providerNameCandidates returns normalised lookup keys for a provider name.
// Plan names such as "Fancode (Via JioTV)" must match the directory's "Fancode",
// and "Discovery+" must match "DiscoveryPlus".
func providerNameCandidates(providerName string) []string {
	trimmed := strings.TrimSpace(providerName)
	if trimmed == "" {
		return nil
	}

	variants := []string{trimmed}
	// Drop a trailing qualifier such as "(Via JioTV)" or "(sports excluded)".
	if base, _, found := strings.Cut(trimmed, "("); found {
		if baseName := strings.TrimSpace(base); baseName != "" {
			variants = append(variants, baseName)
		}
	}

	candidates := make([]string, 0, len(variants))
	seen := make(map[string]struct{}, len(variants))
	for _, variant := range variants {
		normalized := normalizeProviderName(variant)
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		candidates = append(candidates, normalized)
	}
	return candidates
}

// normalizeProviderName lowercases a provider name and strips everything that
// varies between the plans API and the provider directory: spacing, separators
// and the "+" suffix style ("Discovery+" and "DiscoveryPlus" both become
// "discoveryplus").
func normalizeProviderName(providerName string) string {
	lowered := strings.ToLower(strings.TrimSpace(providerName))
	lowered = strings.ReplaceAll(lowered, "+", "plus")

	var builder strings.Builder
	builder.Grow(len(lowered))
	for _, character := range lowered {
		if (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') {
			builder.WriteRune(character)
		}
	}
	return builder.String()
}

func extractPremiumProviders(plansResponse PlansResponse) []PremiumProvider {
	premiumProviders := make([]PremiumProvider, 0)
	seenProviders := make(map[string]struct{})

	// Prefer plans flagged active on the account.
	for _, plan := range plansResponse.PackageInfo {
		if !plan.IsActive {
			continue
		}
		collectPlanProviders(plan.PackageDetail.Providers, &premiumProviders, seenProviders)
	}

	// Some responses omit "isactive"; fall back to every listed package.
	if len(premiumProviders) == 0 {
		for _, plan := range plansResponse.PackageInfo {
			collectPlanProviders(plan.PackageDetail.Providers, &premiumProviders, seenProviders)
		}
	}

	// Older nested response shape.
	if len(premiumProviders) == 0 {
		for _, plan := range plansResponse.Result.Plans {
			collectPlanProviders(plan.Providers, &premiumProviders, seenProviders)
		}
	}

	return premiumProviders
}

func extractPremiumProvidersFromAccessToken(accessToken string) []PremiumProvider {
	tokenParts := strings.Split(accessToken, ".")
	if len(tokenParts) < 2 {
		return []PremiumProvider{}
	}

	payload, err := base64.RawURLEncoding.DecodeString(tokenParts[1])
	if err != nil {
		return []PremiumProvider{}
	}

	var claims struct {
		Data struct {
			Extra string `json:"extra"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return []PremiumProvider{}
	}
	if claims.Data.Extra == "" {
		return []PremiumProvider{}
	}

	var tokenExtra struct {
		PlanDetails struct {
			PackageInfo []struct {
				PlanID interface{} `json:"planid"`
			} `json:"PackageInfo"`
		} `json:"plandetails"`
	}
	if err := json.Unmarshal([]byte(claims.Data.Extra), &tokenExtra); err != nil {
		return []PremiumProvider{}
	}

	premiumProviders := make([]PremiumProvider, 0)
	seenProviders := make(map[string]struct{})

	for _, packageInfo := range tokenExtra.PlanDetails.PackageInfo {
		planID := strings.TrimSpace(fmt.Sprint(packageInfo.PlanID))
		if planID == "" || planID == "<nil>" {
			continue
		}

		providerLink, exists := premiumProviderByPlanID[planID]
		if !exists {
			continue
		}
		if _, alreadySeen := seenProviders[planID]; alreadySeen {
			continue
		}
		seenProviders[planID] = struct{}{}

		premiumProviders = append(premiumProviders, PremiumProvider{
			ID:         planID,
			ProviderID: providerLink.ProviderID,
			Name:       providerLink.DisplayName,
			URL:        providerLink.URL,
		})
	}

	return premiumProviders
}

// PremiumProviders fetches premium providers enabled for the logged-in account.
func PremiumProviders() ([]PremiumProvider, error) {
	if store.KVS == nil {
		return []PremiumProvider{}, nil
	}

	credentials, err := utils.GetJIOTVCredentials()
	if err != nil {
		if errors.Is(err, store.ErrKeyNotFound) {
			return []PremiumProvider{}, nil
		}
		return nil, err
	}
	if credentials == nil || credentials.AccessToken == "" {
		return []PremiumProvider{}, nil
	}

	client := utils.GetRequestClient()

	// The active plans API lists what the account is actually subscribed to and
	// is the authoritative source for entitlements. The packs API is only a
	// catalogue of purchasable plans, so it is tried second.
	plansAPIAttempts := []struct {
		name           string
		url            string
		requestHeaders map[string]string
	}{
		{"active plans", ACTIVE_PLANS_API_URL, buildActivePlansHeaders(credentials)},
		{"subscription packs", PLANS_API_URL, buildPlansAPIHeaders(credentials)},
	}

	for _, attempt := range plansAPIAttempts {
		plansResponse, err := fetchPlansFromAPI(client, attempt.url, attempt.requestHeaders)
		if err != nil {
			utils.SafeLogf("Unable to fetch premium providers from %s API: %v", attempt.name, err)
			continue
		}

		plansBasedProviders := extractPremiumProviders(plansResponse)
		if len(plansBasedProviders) == 0 {
			utils.SafeLogf("%s API returned %d package(s), no premium providers", attempt.name, len(plansResponse.PackageInfo))
			continue
		}

		// Plans return entitlement IDs (Z0177); the catalog needs content
		// provider IDs (200169), so map them via the provider directory.
		providerDirectory, directoryErr := fetchProviderDirectory(client, buildProviderConfigHeaders(credentials))
		if directoryErr != nil {
			utils.SafeLogf("Unable to fetch provider directory: %v", directoryErr)
		} else {
			resolveContentProviderIDs(plansBasedProviders, providerDirectory)
		}

		utils.SafeLogf("%s API returned %d package(s), %d premium provider(s): %s", attempt.name, len(plansResponse.PackageInfo), len(plansBasedProviders), describePremiumProviders(plansBasedProviders))
		return plansBasedProviders, nil
	}

	tokenBasedProviders := extractPremiumProvidersFromAccessToken(credentials.AccessToken)
	utils.SafeLogf("Falling back to access token plan IDs: %d premium provider(s)", len(tokenBasedProviders))
	return tokenBasedProviders, nil
}

func resolvePremiumProviderID(providerIdentifier string) string {
	normalizedIdentifier := strings.TrimSpace(providerIdentifier)
	if normalizedIdentifier == "" {
		return ""
	}

	upperIdentifier := strings.ToUpper(normalizedIdentifier)
	if _, exists := premiumProviderByID[upperIdentifier]; exists {
		return upperIdentifier
	}
	if providerLink, exists := premiumProviderByPlanID[upperIdentifier]; exists && providerLink.ProviderID != "" {
		return providerLink.ProviderID
	}

	lowerIdentifier := strings.ToLower(normalizedIdentifier)
	for providerID, providerLink := range premiumProviderByID {
		lowerDisplayName := strings.ToLower(providerLink.DisplayName)
		if strings.Contains(lowerDisplayName, lowerIdentifier) || strings.Contains(lowerIdentifier, lowerDisplayName) {
			return providerID
		}
	}
	for _, premiumProvider := range premiumProviderByKeyword {
		if strings.Contains(lowerIdentifier, premiumProvider.Keyword) && premiumProvider.Link.ProviderID != "" {
			return premiumProvider.Link.ProviderID
		}
	}

	return ""
}

// buildPlansAPIHeaders builds headers for the subscription packs API
// (v2.1/plans/get), which rejects requests without X-Platform.
func buildPlansAPIHeaders(credentials *utils.JIOTV_CREDENTIALS) map[string]string {
	plansHeaders := buildAuthenticatedHeaders(credentials)
	plansHeaders[headers.XPlatform] = headers.PlatformAndroidMobile
	return plansHeaders
}

// buildActivePlansHeaders builds headers for the active plans API
// (userservice/apis/v1/plans), which returns the plans actually subscribed on
// the account. This endpoint takes a smaller header set than the packs API.
func buildActivePlansHeaders(credentials *utils.JIOTV_CREDENTIALS) map[string]string {
	activePlansHeaders := map[string]string{
		headers.UserAgent:   headers.UserAgentOkHttp,
		headers.Accept:      headers.AcceptJSON,
		headers.DeviceType:  headers.DeviceTypePhone,
		headers.OS:          headers.OSAndroid,
		headers.VersionCode: headers.VersionCode413,
		"Connection":        "close",
	}
	if credentials == nil {
		return activePlansHeaders
	}
	if credentials.AccessToken != "" {
		activePlansHeaders[headers.AccessToken] = credentials.AccessToken
	}
	if credentials.UniqueID != "" {
		activePlansHeaders["uniqueId"] = credentials.UniqueID
	}
	return activePlansHeaders
}

// buildProviderConfigHeaders builds headers for the provider config API
// (/cnf/provider). This endpoint authenticates with X-AccessToken and expects
// X-VersionCode/X-Platform, unlike the rest of the provider APIs.
func buildProviderConfigHeaders(credentials *utils.JIOTV_CREDENTIALS) map[string]string {
	providerHeaders := buildAuthenticatedHeaders(credentials)
	if credentials != nil && credentials.AccessToken != "" {
		providerHeaders[headers.XAccessToken] = credentials.AccessToken
	}
	providerHeaders[headers.XVersionCode] = "4.0"
	providerHeaders[headers.XPlatform] = headers.OSAndroid
	return providerHeaders
}

// buildProviderCatalogHeaders builds headers for the provider metadata API
// (/apis/browse/provider/{id}). This endpoint authenticates with the plain
// accessToken header and must not receive the X-AccessToken variant.
func buildProviderCatalogHeaders(credentials *utils.JIOTV_CREDENTIALS) map[string]string {
	return buildAuthenticatedHeaders(credentials)
}

func fetchPremiumProviderFilterFromAPI(client *fasthttp.Client, providerID string, requestHeaders map[string]string) (PremiumProviderFilterResponse, error) {
	requestConfig := utils.HTTPRequestConfig{
		URL:     strings.TrimRight(PROVIDER_CONFIG_API_BASE_URL, "/") + "/cnf/provider?providerId=" + neturl.QueryEscape(providerID),
		Method:  "GET",
		Headers: requestHeaders,
	}

	resp, err := utils.MakeHTTPRequest(requestConfig, client)
	if err != nil {
		return PremiumProviderFilterResponse{}, err
	}
	defer fasthttp.ReleaseResponse(resp)

	var filterResponse PremiumProviderFilterResponse
	if err := utils.ParseJSONResponse(resp, &filterResponse); err != nil {
		return PremiumProviderFilterResponse{}, err
	}

	return filterResponse, nil
}

func extractProviderQueryNames(filterResponse PremiumProviderFilterResponse, providerID string) []string {
	queryNames := make([]string, 0)
	seenNames := make(map[string]struct{})
	normalizedProviderID := strings.ToUpper(strings.TrimSpace(providerID))

	for _, providerFilters := range filterResponse.Data.Provider {
		currentProviderID := strings.ToUpper(strings.TrimSpace(providerFilters.ProviderID))
		if currentProviderID != "" && normalizedProviderID != "" && currentProviderID != normalizedProviderID {
			continue
		}
		for _, filterItem := range providerFilters.Filters {
			filterName := strings.TrimSpace(filterItem.FilterName)
			if filterName == "" {
				continue
			}
			if _, alreadySeen := seenNames[filterName]; alreadySeen {
				continue
			}
			seenNames[filterName] = struct{}{}
			queryNames = append(queryNames, filterName)
		}
	}

	return queryNames
}

func buildProviderQueryMap(queryNames, selectedGenres []string, selectedContentType, selectedContentLanguage string) map[string]string {
	queryMap := make(map[string]string)
	genresCSV := strings.Join(selectedGenres, ",")

	for _, queryName := range queryNames {
		normalizedName := strings.ToLower(strings.TrimSpace(queryName))
		if normalizedName == "" {
			continue
		}

		value := ""
		switch {
		case strings.Contains(normalizedName, "genre") && genresCSV != "":
			value = genresCSV
		case strings.Contains(normalizedName, "type") && selectedContentType != "":
			value = strings.TrimSpace(selectedContentType)
		case strings.Contains(normalizedName, "language") && selectedContentLanguage != "":
			value = strings.TrimSpace(selectedContentLanguage)
		}

		if value != "" {
			queryMap[queryName] = value
		}
	}

	return queryMap
}

func buildProviderDefaultQueryMap(filterResponse PremiumProviderFilterResponse, providerID string) map[string]string {
	defaultQueryMap := make(map[string]string)
	normalizedProviderID := strings.ToUpper(strings.TrimSpace(providerID))

	for _, providerFilters := range filterResponse.Data.Provider {
		currentProviderID := strings.ToUpper(strings.TrimSpace(providerFilters.ProviderID))
		if currentProviderID != "" && normalizedProviderID != "" && currentProviderID != normalizedProviderID {
			continue
		}
		for _, filterItem := range providerFilters.Filters {
			filterName := strings.TrimSpace(filterItem.FilterName)
			if filterName == "" {
				continue
			}

			normalizedFilterName := strings.ToLower(filterName)
			if !strings.Contains(normalizedFilterName, "genre") && !strings.Contains(normalizedFilterName, "type") && !strings.Contains(normalizedFilterName, "language") {
				continue
			}

			// "All" is a UI-only sentinel; sending it makes the catalog API
			// reject the request with "Parameter Missing/Bad Request".
			selectedValue := ""
			for _, filterValue := range filterItem.Values {
				key := strings.TrimSpace(filterValue.Key)
				if key == "" || strings.EqualFold(key, "all") {
					continue
				}
				selectedValue = key
				break
			}

			if selectedValue != "" {
				defaultQueryMap[filterName] = selectedValue
			}
		}
	}

	return defaultQueryMap
}

func buildProviderCatalogURL(providerID string, page, limit int, queryMap map[string]string) string {
	baseURL := strings.TrimRight(PROVIDER_METADATA_API_BASE_URL, "/")
	catalogURL := fmt.Sprintf("%s/apis/browse/provider/%s?page=%d&limit=%d", baseURL, neturl.PathEscape(providerID), page, limit)
	if len(queryMap) == 0 {
		return catalogURL
	}

	queryParams := neturl.Values{}
	for key, value := range queryMap {
		trimmedKey := strings.TrimSpace(key)
		trimmedValue := strings.TrimSpace(value)
		if trimmedKey == "" || trimmedValue == "" {
			continue
		}
		queryParams.Set(trimmedKey, trimmedValue)
	}
	encodedQuery := queryParams.Encode()
	if encodedQuery == "" {
		return catalogURL
	}

	return catalogURL + "&" + encodedQuery
}

func fetchPremiumProviderCatalogFromAPI(client *fasthttp.Client, providerID string, page, limit int, queryMap map[string]string, requestHeaders map[string]string) (PremiumProviderCatalogEnvelope, error) {
	requestConfig := utils.HTTPRequestConfig{
		URL:     buildProviderCatalogURL(providerID, page, limit, queryMap),
		Method:  "GET",
		Headers: requestHeaders,
	}

	resp, err := utils.MakeHTTPRequest(requestConfig, client)
	if err != nil {
		return PremiumProviderCatalogEnvelope{}, err
	}
	defer fasthttp.ReleaseResponse(resp)

	var catalogResponse PremiumProviderCatalogEnvelope
	if err := utils.ParseJSONResponse(resp, &catalogResponse); err != nil {
		return PremiumProviderCatalogEnvelope{}, err
	}

	return catalogResponse, nil
}

func sameQueryMap(firstMap, secondMap map[string]string) bool {
	if len(firstMap) != len(secondMap) {
		return false
	}
	for key, value := range firstMap {
		if secondMap[key] != value {
			return false
		}
	}
	return true
}

func premiumCatalogCode(catalog PremiumProviderCatalogEnvelope) int {
	if catalog.Code != 0 {
		return catalog.Code
	}
	return catalog.CodeLower
}

func premiumCatalogMessage(catalog PremiumProviderCatalogEnvelope) string {
	if strings.TrimSpace(catalog.Message) != "" {
		return catalog.Message
	}
	return catalog.MessageLower
}

func premiumCatalogData(catalog PremiumProviderCatalogEnvelope) []map[string]interface{} {
	if len(catalog.Data) > 0 {
		return catalog.Data
	}
	return catalog.DataLower
}

func interfaceToString(value interface{}) string {
	switch typedValue := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(typedValue)
	case json.Number:
		return strings.TrimSpace(typedValue.String())
	case float64:
		if typedValue == float64(int64(typedValue)) {
			return strconv.FormatInt(int64(typedValue), 10)
		}
		return strconv.FormatFloat(typedValue, 'f', -1, 64)
	case float32:
		floatValue := float64(typedValue)
		if floatValue == float64(int64(floatValue)) {
			return strconv.FormatInt(int64(floatValue), 10)
		}
		return strconv.FormatFloat(floatValue, 'f', -1, 64)
	case int:
		return strconv.Itoa(typedValue)
	case int8:
		return strconv.FormatInt(int64(typedValue), 10)
	case int16:
		return strconv.FormatInt(int64(typedValue), 10)
	case int32:
		return strconv.FormatInt(int64(typedValue), 10)
	case int64:
		return strconv.FormatInt(typedValue, 10)
	case uint:
		return strconv.FormatUint(uint64(typedValue), 10)
	case uint8:
		return strconv.FormatUint(uint64(typedValue), 10)
	case uint16:
		return strconv.FormatUint(uint64(typedValue), 10)
	case uint32:
		return strconv.FormatUint(uint64(typedValue), 10)
	case uint64:
		return strconv.FormatUint(typedValue, 10)
	case bool:
		return strconv.FormatBool(typedValue)
	case map[string]interface{}, []interface{}:
		// Composite values have no useful string form; callers should address
		// the nested field directly via a dotted path.
		return ""
	default:
		stringValue := strings.TrimSpace(fmt.Sprint(typedValue))
		if stringValue == "<nil>" {
			return ""
		}
		return stringValue
	}
}

func stringByPath(rawItem map[string]interface{}, path string) string {
	pathParts := strings.Split(path, ".")
	var currentValue interface{} = rawItem

	for _, pathPart := range pathParts {
		currentMap, ok := currentValue.(map[string]interface{})
		if !ok {
			return ""
		}
		nextValue, exists := currentMap[pathPart]
		if !exists {
			return ""
		}
		currentValue = nextValue
	}

	return interfaceToString(currentValue)
}

func firstStringByPaths(rawItem map[string]interface{}, paths ...string) string {
	for _, path := range paths {
		pathValue := stringByPath(rawItem, path)
		if pathValue != "" {
			return pathValue
		}
	}
	return ""
}

// absoluteImageURL resolves catalog artwork paths, which the API returns
// relative to the JioTV image host.
func absoluteImageURL(imagePath string) string {
	trimmedPath := strings.TrimSpace(imagePath)
	if trimmedPath == "" {
		return ""
	}
	if !strings.HasPrefix(trimmedPath, "http://") && !strings.HasPrefix(trimmedPath, "https://") {
		trimmedPath = urls.ImageBaseURL + strings.TrimPrefix(trimmedPath, "/")
	}
	return upscaleCatalogArtwork(trimmedPath)
}

// catalogArtworkWidth is the width requested for premium poster art. The API
// hands back width=216 by default, which is visibly soft on a poster-sized
// card, let alone on a high-density display.
const catalogArtworkWidth = 480

var catalogArtworkWidthPattern = regexp.MustCompile(`([?&])width=(\d+)`)

// upscaleCatalogArtwork raises the width parameter on JioTV image URLs and
// drops the low-quality flag, leaving URLs without a width untouched.
func upscaleCatalogArtwork(imageURL string) string {
	upgraded := catalogArtworkWidthPattern.ReplaceAllStringFunc(imageURL, func(match string) string {
		submatches := catalogArtworkWidthPattern.FindStringSubmatch(match)
		if len(submatches) != 3 {
			return match
		}
		currentWidth, err := strconv.Atoi(submatches[2])
		if err != nil || currentWidth >= catalogArtworkWidth {
			return match
		}
		return submatches[1] + "width=" + strconv.Itoa(catalogArtworkWidth)
	})

	return strings.ReplaceAll(upgraded, "optimize=low", "optimize=high")
}

func mapPremiumProviderCatalogItems(rawItems []map[string]interface{}, providerID string) []PremiumProviderCatalogItem {
	catalogItems := make([]PremiumProviderCatalogItem, 0, len(rawItems))
	seenItems := make(map[string]struct{})
	normalizedProviderID := strings.ToUpper(strings.TrimSpace(providerID))

	for _, rawItem := range rawItems {
		channelID := firstStringByPaths(rawItem, "channel_id", "channelId", "metadata.channel.channel_id", "metadata.channel.channelId")
		contentID := firstStringByPaths(rawItem, "content_id", "contentId", "metadata.content.content_id", "metadata.content.contentId")
		subCategoryID := firstStringByPaths(rawItem, "sub_category_id", "subCategoryId", "metadata.content.sub_category_id", "metadata.content.subCategoryId")
		streamType := strings.ToLower(firstStringByPaths(rawItem, "stream_type", "streamType"))
		businessType := strings.ToLower(firstStringByPaths(rawItem, "business_type", "businessType"))

		// In provider catalogs metadata.channel.channel_id is the provider ID
		// itself and is identical for every item, so entries are keyed and
		// played by content_id instead. business_type is not always populated,
		// so also treat a channel ID equal to the provider ID as the sentinel.
		// A populated sub_category_id means a genuine VOD item, which keeps the
		// provider_id/sub_category_id/content_id playback path instead.
		isProviderScopedChannel := channelID != "" && normalizedProviderID != "" && strings.EqualFold(channelID, normalizedProviderID)
		if contentID != "" && subCategoryID == "" && (businessType == "svod" || businessType == "avod" || isProviderScopedChannel) {
			channelID = contentID
			streamType = "provider"
		}

		if streamType == "" {
			if channelID != "" {
				streamType = "provider"
			} else if contentID != "" {
				streamType = "vod"
			}
		}

		if channelID == "" && contentID == "" {
			continue
		}
		if streamType != "provider" && streamType != "vod" {
			continue
		}

		itemID := firstStringByPaths(rawItem, "id", "ID", "content_id", "contentId", "channel_id", "channelId", "metadata.content.content_id", "metadata.channel.channel_id")
		if itemID == "" {
			if streamType == "provider" {
				itemID = channelID
			} else {
				itemID = contentID
			}
		}
		if itemID == "" {
			continue
		}

		dedupeKey := streamType + ":" + itemID
		if _, exists := seenItems[dedupeKey]; exists {
			continue
		}
		seenItems[dedupeKey] = struct{}{}

		// content_title first: in provider catalogs channel_name is the
		// provider name and identical for every item.
		title := firstStringByPaths(rawItem,
			"content_title",
			"name",
			"title",
			"display_name",
			"show_name",
			"metadata.parent.show_name",
			"metadata.content.name",
			"metadata.content.title",
			"channel_name",
			"metadata.channel.channel_name",
		)
		if title == "" {
			title = "Premium stream"
		}

		catalogItems = append(catalogItems, PremiumProviderCatalogItem{
			ID:            itemID,
			Title:         title,
			Subtitle:      firstStringByPaths(rawItem, "genre", "category", "language", "type", "metadata.content.genre", "metadata.channel.language"),
			Description:   firstStringByPaths(rawItem, "content_description", "description", "short_description", "metadata.content.description"),
			ImageURL:      absoluteImageURL(firstStringByPaths(rawItem, "logoUrl", "image.portrait", "image.list", "image.square", "image.cover", "logo", "thumbnail", "poster", "poster_url", "metadata.channel.logo.portrait", "metadata.channel.logoUrl", "metadata.content.thumbnail", "metadata.content.poster")),
			StreamType:    streamType,
			ProviderID:    normalizedProviderID,
			ChannelID:     channelID,
			ContentID:     contentID,
			SubCategoryID: subCategoryID,
		})
	}

	return catalogItems
}

// PremiumProviderCatalog fetches provider catalog items available for a premium provider.
func PremiumProviderCatalog(providerIdentifier string, page, limit int) (PremiumProviderCatalogResult, error) {
	result := PremiumProviderCatalogResult{
		Result: make([]PremiumProviderCatalogItem, 0),
	}

	if page < 0 {
		page = 0
	}
	if limit <= 0 {
		limit = 30
	}

	if store.KVS == nil {
		return result, errors.New("not logged in")
	}

	credentials, err := utils.GetJIOTVCredentials()
	if err != nil {
		return result, err
	}
	if credentials == nil || credentials.AccessToken == "" {
		return result, errors.New("missing access token")
	}

	providerID := resolvePremiumProviderID(providerIdentifier)
	if providerID == "" {
		providerID = strings.ToUpper(strings.TrimSpace(providerIdentifier))
	}
	if providerID == "" {
		return result, errors.New("invalid provider identifier")
	}
	result.ProviderID = providerID

	client := utils.GetRequestClient()
	configHeaders := buildProviderConfigHeaders(credentials)
	catalogHeaders := buildProviderCatalogHeaders(credentials)

	filterResponse := PremiumProviderFilterResponse{}
	filterResponse, filterErr := fetchPremiumProviderFilterFromAPI(client, providerID, configHeaders)
	if filterErr != nil {
		utils.SafeLogf("Unable to fetch provider filters for %s: %v", providerID, filterErr)
	}

	queryNames := extractProviderQueryNames(filterResponse, providerID)
	attemptQueryMaps := []map[string]string{
		buildProviderQueryMap(queryNames, nil, "", ""),
	}
	fallbackQueryMap := buildProviderDefaultQueryMap(filterResponse, providerID)
	if len(fallbackQueryMap) > 0 && !sameQueryMap(fallbackQueryMap, attemptQueryMaps[0]) {
		attemptQueryMaps = append(attemptQueryMaps, fallbackQueryMap)
	}

	var lastCatalogResponse PremiumProviderCatalogEnvelope
	var hasCatalogResponse bool
	var firstFetchErr error

	for _, queryMap := range attemptQueryMaps {
		catalogResponse, fetchErr := fetchPremiumProviderCatalogFromAPI(client, providerID, page, limit, queryMap, catalogHeaders)
		if fetchErr != nil {
			// The unfiltered attempt comes first and gives the most accurate
			// diagnosis, so keep its error rather than a later one.
			if firstFetchErr == nil {
				firstFetchErr = fetchErr
			}
			continue
		}

		hasCatalogResponse = true
		lastCatalogResponse = catalogResponse

		catalogItems := mapPremiumProviderCatalogItems(premiumCatalogData(catalogResponse), providerID)
		if len(catalogItems) == 0 {
			continue
		}

		result.Code = premiumCatalogCode(catalogResponse)
		result.Message = premiumCatalogMessage(catalogResponse)
		result.Result = catalogItems
		return result, nil
	}

	if hasCatalogResponse {
		result.Code = premiumCatalogCode(lastCatalogResponse)
		result.Message = premiumCatalogMessage(lastCatalogResponse)
		result.Result = mapPremiumProviderCatalogItems(premiumCatalogData(lastCatalogResponse), providerID)
		return result, nil
	}
	if firstFetchErr != nil {
		// "Page out of bounds" means the provider simply has nothing to list for
		// this account, which is an empty catalog rather than a failure.
		if strings.Contains(strings.ToLower(firstFetchErr.Error()), "page out of bounds") {
			result.Code = 204
			result.Message = "No content available from this provider for your account"
			return result, nil
		}
		return result, firstFetchErr
	}

	return result, nil
}

func appendHdneaToPlaybackResult(result *LiveURLOutput) {
	if result == nil {
		return
	}

	extractHdneaFromURL := func(streamURL string) string {
		if streamURL == "" {
			return ""
		}
		index := strings.Index(streamURL, "hdnea=")
		if index == -1 {
			return ""
		}
		token := streamURL[index+len("hdnea="):]
		if separatorIndex := strings.IndexByte(token, '&'); separatorIndex != -1 {
			token = token[:separatorIndex]
		}
		return token
	}

	hdneaToken := extractHdneaFromURL(result.Bitrates.Auto)
	if hdneaToken == "" {
		hdneaToken = extractHdneaFromURL(result.Result)
	}
	if hdneaToken == "" {
		hdneaToken = extractHdneaFromURL(result.Mpd.Result)
	}
	result.Hdnea = hdneaToken

	if hdneaToken == "" {
		return
	}

	appendHdnea := func(streamURL string) string {
		if streamURL == "" || strings.Contains(streamURL, "hdnea=") {
			return streamURL
		}
		separator := "?"
		if strings.Contains(streamURL, "?") {
			separator = "&"
		}
		return streamURL + separator + "hdnea=" + hdneaToken
	}

	result.Bitrates.Auto = appendHdnea(result.Bitrates.Auto)
	result.Bitrates.High = appendHdnea(result.Bitrates.High)
	result.Bitrates.Medium = appendHdnea(result.Bitrates.Medium)
	result.Bitrates.Low = appendHdnea(result.Bitrates.Low)
	result.Result = appendHdnea(result.Result)
	result.Mpd.Result = appendHdnea(result.Mpd.Result)
	result.Mpd.Key = appendHdnea(result.Mpd.Key)
}

// ResolvePlaybackURL picks the best playable URL from a playback response.
func ResolvePlaybackURL(result *LiveURLOutput) string {
	if result == nil {
		return ""
	}
	if strings.TrimSpace(result.Bitrates.Auto) != "" {
		return strings.TrimSpace(result.Bitrates.Auto)
	}
	if strings.TrimSpace(result.Result) != "" {
		return strings.TrimSpace(result.Result)
	}
	// Premium provider VOD returns a plain HLS stream nested under "m3u8".
	if strings.TrimSpace(result.M3u8.Auto) != "" {
		return strings.TrimSpace(result.M3u8.Auto)
	}
	if strings.TrimSpace(result.Mpd.Result) != "" {
		return strings.TrimSpace(result.Mpd.Result)
	}
	return ""
}

// ErrPremiumNotSubscribed is returned when the account may browse a provider's
// catalog but is not entitled to play its content.
var ErrPremiumNotSubscribed = errors.New("your account is not subscribed to this premium provider")

// ErrPremiumUpstreamUnavailable is returned when JioTV's playback edge fails
// with a gateway error, which is usually transient.
var ErrPremiumUpstreamUnavailable = errors.New("JioTV's playback service is temporarily unavailable, please try again")

// PremiumProviderPlayback resolves playback URL for a premium provider item.
func PremiumProviderPlayback(providerIdentifier string, playRequest PremiumProviderPlayRequest) (*LiveURLOutput, error) {
	if store.KVS == nil {
		return nil, errors.New("not logged in")
	}

	credentials, err := utils.GetJIOTVCredentials()
	if err != nil {
		return nil, err
	}
	if credentials == nil || credentials.AccessToken == "" {
		return nil, errors.New("missing access token")
	}

	providerID := resolvePremiumProviderID(providerIdentifier)
	if providerID == "" {
		providerID = strings.ToUpper(strings.TrimSpace(providerIdentifier))
	}
	if providerID == "" {
		return nil, errors.New("invalid provider identifier")
	}

	streamType := strings.ToLower(strings.TrimSpace(playRequest.StreamType))
	channelID := strings.TrimSpace(playRequest.ChannelID)
	contentID := strings.TrimSpace(playRequest.ContentID)
	subCategoryID := strings.TrimSpace(playRequest.SubCategoryID)

	if streamType == "" {
		if channelID != "" {
			streamType = "provider"
		} else if contentID != "" {
			streamType = "vod"
		}
	}

	switch streamType {
	case "provider":
		if channelID == "" {
			return nil, errors.New("channelId is required for provider playback")
		}
	case "vod":
		if contentID == "" || subCategoryID == "" {
			return nil, errors.New("contentId and subCategoryId are required for vod playback")
		}
	default:
		return nil, fmt.Errorf("unsupported streamType: %s", streamType)
	}

	tv := New(credentials)
	formData := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(formData)

	formData.Add("stream_type", streamType)
	formData.Add("begin", utils.GenerateCurrentTime())
	formData.Add("srno", utils.GenerateDate())

	if streamType == "provider" {
		// The app posts only stream_type and channel_id for provider streams;
		// adding provider_id makes the request fail.
		formData.Add("channel_id", channelID)
	} else {
		formData.Add("provider_id", providerID)
		formData.Add("sub_category_id", subCategoryID)
		formData.Add("content_id", contentID)
	}

	request := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(request)
	for key, value := range tv.Headers {
		request.Header.Set(key, value)
	}
	request.Header.Set(headers.AccessToken, tv.AccessToken)
	request.SetRequestURI("https://" + JIOTV_API_DOMAIN + urls.PlaybackAPIPath)
	request.Header.SetMethod("POST")
	request.SetBody(formData.QueryString())

	response := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(response)

	// The playback edge sits behind Varnish, which intermittently answers 5xx
	// ("backend read error"). A single immediate retry clears it in practice.
	if err := tv.Client.Do(request, response); err != nil {
		return nil, err
	}
	if response.StatusCode() >= fasthttp.StatusInternalServerError {
		utils.SafeLogf("Premium playback got %d from upstream, retrying once", response.StatusCode())
		response.Reset()
		if err := tv.Client.Do(request, response); err != nil {
			return nil, err
		}
	}
	if response.StatusCode() != fasthttp.StatusOK {
		// Upstream gateway errors return an HTML error page; surface a short
		// message instead of dumping the markup at the caller.
		if response.StatusCode() >= fasthttp.StatusInternalServerError {
			utils.SafeLogf("Premium playback upstream error for provider %s (status %d): %s", providerID, response.StatusCode(), truncateForLog(response.Body(), 200))
			return nil, ErrPremiumUpstreamUnavailable
		}
		// The catalog is browsable for every provider, but playback is refused
		// unless the account is subscribed to that provider.
		if response.StatusCode() == fasthttp.StatusForbidden || response.StatusCode() == fasthttp.StatusUnauthorized {
			utils.SafeLogf("Premium playback refused for provider %s (status %d): %s", providerID, response.StatusCode(), truncateForLog(response.Body(), 200))
			return nil, ErrPremiumNotSubscribed
		}
		return nil, fmt.Errorf("request failed with status code: %d, body: %s", response.StatusCode(), response.Body())
	}

	var playbackResult LiveURLOutput
	if err := json.Unmarshal(response.Body(), &playbackResult); err != nil {
		return nil, err
	}

	appendHdneaToPlaybackResult(&playbackResult)
	return &playbackResult, nil
}

func mergeChannels(primary, secondary []Channel) []Channel {
	if len(primary) == 0 {
		return secondary
	}
	if len(secondary) == 0 {
		return primary
	}

	mergedChannels := make([]Channel, 0, len(primary)+len(secondary))
	seen := make(map[string]struct{}, len(primary)+len(secondary))

	getKey := func(channel Channel) string {
		if channel.ID != "" {
			return channel.ID
		}
		return strings.ToLower(strings.TrimSpace(channel.Name))
	}

	addChannel := func(channel Channel) {
		key := getKey(channel)
		if key == "" {
			mergedChannels = append(mergedChannels, channel)
			return
		}
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		mergedChannels = append(mergedChannels, channel)
	}

	for _, channel := range primary {
		addChannel(channel)
	}
	for _, channel := range secondary {
		addChannel(channel)
	}

	return mergedChannels
}

// Channels fetch channels from JioTV API and merge with custom channels
func Channels() (ChannelsResponse, error) {
	// Create a fasthttp.Client
	client := utils.GetRequestClient()

	// Set up request headers
	defaultHeaders := map[string]string{
		headers.UserAgent:  headers.UserAgentOkHttp,
		headers.Accept:     headers.AcceptJSON,
		headers.DeviceType: headers.DeviceTypePhone,
		headers.OS:         headers.OSAndroid,
		"appkey":           "NzNiMDhlYzQyNjJm",
		"lbcookie":         "1",
		"usertype":         "JIO",
	}

	var apiResponse ChannelsResponse
	hasAuthCredentials := false

	// Prefer account-aware channel listing when OTP credentials are available.
	if store.KVS != nil {
		credentials, err := utils.GetJIOTVCredentials()
		if err != nil {
			utils.SafeLogf("Unable to load credentials for authenticated channel list: %v", err)
		} else if credentials != nil && credentials.AccessToken != "" && credentials.SSOToken != "" && credentials.CRM != "" {
			hasAuthCredentials = true
			authHeaders := buildAuthenticatedHeaders(credentials)

			apiResponse, err = fetchChannelsFromAPI(client, CHANNELS_AUTH_API_URL, authHeaders)
			if err != nil {
				utils.Log.Printf("Error fetching authenticated channels, retrying default endpoint: %v", err)
			} else if len(apiResponse.Result) > 0 {
				// Merge default endpoint results to avoid missing channels exposed by only one endpoint.
				defaultResponse, fallbackErr := fetchChannelsFromAPI(client, CHANNELS_API_URL, defaultHeaders)
				if fallbackErr != nil {
					utils.Log.Printf("Error fetching fallback channels for merge: %v", fallbackErr)
				} else {
					apiResponse.Result = mergeChannels(apiResponse.Result, defaultResponse.Result)
				}
			}
		}
	}

	if len(apiResponse.Result) == 0 {
		fallbackResponse, err := fetchChannelsFromAPI(client, CHANNELS_API_URL, defaultHeaders)
		if err != nil {
			if hasAuthCredentials {
				utils.Log.Printf("Error fetching channels from both authenticated and default endpoints: %v", err)
			} else {
				utils.Log.Printf("Error fetching channels from JioTV API: %v", err)
			}
			return ChannelsResponse{}, err
		}
		apiResponse = fallbackResponse
	}

	// disable sony channels temporarily
	// apiResponse.Result = append(apiResponse.Result, SONY_CHANNELS_API...)

	// Load and append custom channels if configured
	if config.Cfg.CustomChannelsFile != "" {
		customChannels := getCustomChannels()
		apiResponse.Result = append(apiResponse.Result, customChannels...)
	}

	return apiResponse, nil
}

// FilterChannels Function is used to filter channels by language and category
func FilterChannels(channels []Channel, language, category int) []Channel {
	var filteredChannels []Channel
	for _, channel := range channels {
		// if both language and category is set, then use and operator
		if language != 0 && category != 0 {
			if channel.Language == language && channel.Category == category {
				filteredChannels = append(filteredChannels, channel)
			}
		} else if language != 0 {
			if channel.Language == language {
				filteredChannels = append(filteredChannels, channel)
			}
		} else if category != 0 {
			if channel.Category == category {
				filteredChannels = append(filteredChannels, channel)
			}
		} else {
			filteredChannels = append(filteredChannels, channel)
		}
	}
	return filteredChannels
}

// FilterChannelsByDefaults filters channels by arrays of default categories and languages
// If both arrays are provided, channels must match at least one category AND one language
// If only one array is provided, channels must match at least one item from that array
// If both arrays are empty, all channels are returned
func FilterChannelsByDefaults(channels []Channel, categories, languages []int) []Channel {
	// If both arrays are empty, return all channels
	if len(categories) == 0 && len(languages) == 0 {
		return channels
	}

	// Use maps for efficient O(1) lookups
	categorySet := make(map[int]struct{}, len(categories))
	for _, cat := range categories {
		categorySet[cat] = struct{}{}
	}

	languageSet := make(map[int]struct{}, len(languages))
	for _, lang := range languages {
		languageSet[lang] = struct{}{}
	}

	filteredChannels := make([]Channel, 0, len(channels))
	for _, channel := range channels {
		// If categories are specified, channel must match one of them
		categoryMatch := len(categories) == 0
		if !categoryMatch {
			_, categoryMatch = categorySet[channel.Category]
		}

		// If languages are specified, channel must match one of them
		languageMatch := len(languages) == 0
		if !languageMatch {
			_, languageMatch = languageSet[channel.Language]
		}

		// Include channel if it matches both criteria (or if a criterion is not specified)
		if categoryMatch && languageMatch {
			filteredChannels = append(filteredChannels, channel)
		}
	}
	return filteredChannels
}

func ReplaceM3U8(baseUrl, match []byte, params, channel_id string, quality string) []byte {
	config := EncryptedURLConfig{
		BaseURL:     string(baseUrl),
		Match:       string(match),
		Params:      params,
		ChannelID:   channel_id,
		EndpointURL: "/render.m3u8",
		Quality:     quality,
	}

	result, err := CreateEncryptedURL(config)
	if err != nil {
		return nil
	}
	return result
}

func ReplaceTS(baseUrl, match []byte, params, channelID string) []byte {
	if config.Cfg.DisableTSHandler {
		return []byte(string(baseUrl) + string(match) + "?" + params)
	}

	config := EncryptedURLConfig{
		BaseURL:     string(baseUrl),
		Match:       string(match),
		Params:      params,
		ChannelID:   channelID,
		EndpointURL: "/render.ts",
	}

	result, err := CreateEncryptedURL(config)
	if err != nil {
		return nil
	}
	return result
}

func ReplaceAAC(baseUrl, match []byte, params, channelID string) []byte {
	if config.Cfg.DisableTSHandler {
		return []byte(string(baseUrl) + string(match) + "?" + params)
	}

	config := EncryptedURLConfig{
		BaseURL:     string(baseUrl),
		Match:       string(match),
		Params:      params,
		ChannelID:   channelID,
		EndpointURL: "/render.ts",
	}

	result, err := CreateEncryptedURL(config)
	if err != nil {
		return nil
	}
	return result
}

func ReplaceKey(match []byte, params, channel_id string) []byte {
	config := EncryptedURLConfig{
		BaseURL:     "",
		Match:       string(match),
		Params:      params,
		ChannelID:   channel_id,
		EndpointURL: "/render.key",
	}

	result, err := CreateEncryptedURL(config)
	if err != nil {
		return nil
	}
	return result
}

func getSLChannel(channelID string) (*LiveURLOutput, error) {
	// Check if the channel is available in the SONY_CHANNELS map
	if val, ok := SONY_JIO_MAP[channelID]; ok {
		// If the channel is available in the SONY_CHANNELS map, then return the link
		result := new(LiveURLOutput)

		chu, err := base64.StdEncoding.DecodeString(SONY_CHANNELS[val])
		if err != nil {
			utils.Log.Println(err)
			return nil, err
		}

		channel_url := string(chu)

		// Make a get request to the channel url and store location header in actual_url
		req := fasthttp.AcquireRequest()
		defer fasthttp.ReleaseRequest(req)

		req.SetRequestURI(channel_url)
		req.Header.SetMethod("GET")

		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseResponse(resp)

		// Perform the HTTP GET request
		if err := utils.GetRequestClient().Do(req, resp); err != nil {
			utils.Log.Println(err)
			return nil, err
		}

		if resp.StatusCode() != fasthttp.StatusFound {
			err := fmt.Errorf("request failed with status code: %d, response: %s", resp.StatusCode(), string(resp.Body()))
			utils.Log.Println(err)
			return nil, err
		}

		// Store the location header in actual_url
		actual_url := string(resp.Header.Peek("Location"))

		result.Result = actual_url
		result.Bitrates.Auto = actual_url
		return result, nil
	} else {
		// If the channel is not available in the SONY_CHANNELS map, then return an error
		return nil, fmt.Errorf("Channel not found")
	}
}

func (tv *Television) GetCatchupURL(channelID, srno, start, end string) (*LiveURLOutput, error) {
	formData := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(formData)

	formData.Add("stream_type", "Catchup")
	formData.Add("channel_id", channelID)
	formData.Add("programId", srno)
	formData.Add("showtime", "000000")
	formData.Add("srno", srno)
	formData.Add("begin", start)
	formData.Add("end", end)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	for key, value := range tv.Headers {
		req.Header.Set(key, value)
	}

	url := "https://" + JIOTV_API_DOMAIN + urls.PlaybackAPIPath
	req.Header.Set(headers.AccessToken, tv.AccessToken)
	req.SetRequestURI(url)
	req.Header.SetMethod("POST")
	req.SetBody(formData.QueryString())
	req.Header.Set("channel_id", channelID)
	req.Header.Set("srno", srno)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		if err := tv.Client.Do(req, resp); err != nil {
			if strings.Contains(err.Error(), "server closed connection before returning the first response byte") {
				utils.Log.Printf("Retrying the catchup request (attempt %d/%d)...", i+1, maxRetries)
				continue
			}
			return nil, err
		}
		break
	}
	if resp.StatusCode() != fasthttp.StatusOK {
		response := string(resp.Body())
		utils.Log.Printf("Catchup request failed with status code: %d", resp.StatusCode())
		utils.Log.Println("Request headers:", req.Header.String())
		utils.Log.Println("Request data:", formData.String())
		utils.Log.Printf("API Response: %s", response)
		var errorResp map[string]interface{}
		if err := json.Unmarshal(resp.Body(), &errorResp); err == nil {
			if code, ok := errorResp["code"].(float64); ok {
				utils.Log.Printf("API Error Code: %.0f", code)
			}
			if message, ok := errorResp["message"].(string); ok {
				utils.Log.Printf("API Error Message: %s", message)
			}
		}
		return nil, fmt.Errorf("catchup request failed with status code: %d", resp.StatusCode())
	}

	var result LiveURLOutput
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	extractHdneaFromURL := func(u string) string {
		if u == "" {
			return ""
		}
		qIdx := strings.Index(u, "?")
		if qIdx == -1 {
			return ""
		}
		query := u[qIdx+1:]
		for _, p := range strings.Split(query, "&") {
			if strings.HasPrefix(p, "hdnea=") {
				return strings.TrimPrefix(p, "hdnea=")
			}
			if strings.HasPrefix(p, "__hdnea__=") {
				return strings.TrimPrefix(p, "__hdnea__=")
			}
		}
		return ""
	}
	hdnea := extractHdneaFromURL(result.Result)
	if hdnea == "" {
		hdnea = extractHdneaFromURL(result.Bitrates.Auto)
	}
	result.Hdnea = hdnea
	return &result, nil
}
