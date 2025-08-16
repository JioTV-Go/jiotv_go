package television

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/valyala/fasthttp"
	"gopkg.in/yaml.v3"

	"github.com/jiotv-go/jiotv_go/v3/internal/config"
	"github.com/jiotv-go/jiotv_go/v3/internal/constants"
	"github.com/jiotv-go/jiotv_go/v3/internal/constants/headers"
	"github.com/jiotv-go/jiotv_go/v3/internal/constants/urls"
	"github.com/jiotv-go/jiotv_go/v3/pkg/secureurl"
	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
)

const (
	// JioTV API domain constants
	JIOTV_API_DOMAIN = urls.JioTVAPIDomain
	TV_MEDIA_DOMAIN  = urls.TVMediaDomain
	JIOTV_CDN_DOMAIN = urls.JioTVCDNDomain

	// URL for fetching channels from JioTV API
	CHANNELS_API_URL = urls.ChannelsAPIURL
	// Error message for unsupported custom channels file formats
	errUnsupportedChannelsFormat = constants.ErrUnsupportedChannelsFormat
	// Maximum recommended number of custom channels before performance warnings
	maxRecommendedChannels = constants.MaxRecommendedChannels
)

// logExcessiveChannelsWarning logs a comprehensive warning when the number of custom channels exceeds the recommended limit
func logExcessiveChannelsWarning(channelCount int, context string) {
	if channelCount <= maxRecommendedChannels || utils.Log == nil {
		return
	}

	utils.Log.Printf("WARNING: %s %d custom channels, which exceeds the recommended limit of %d channels.", context, channelCount, maxRecommendedChannels)
	utils.Log.Printf("WARNING: Large numbers of custom channels may impact performance:")
	utils.Log.Printf("  - Slower channel listing and filtering operations")
	utils.Log.Printf("  - Increased memory usage")
	utils.Log.Printf("  - Longer startup times")
	utils.Log.Printf("  - Potential UI responsiveness issues")
	utils.Log.Printf("Consider splitting channels into multiple configuration files or reducing the total number.")
}

var (
	// customChannelsCache holds cached custom channels
	customChannelsCache []Channel
	// customChannelsCacheMap holds cached custom channels indexed by ID for efficient lookups
	customChannelsCacheMap map[string]Channel
	// customChannelsCacheMutex protects the cache from concurrent access
	customChannelsCacheMutex sync.RWMutex
	// customChannelsCacheLoaded indicates if cache has been loaded
	customChannelsCacheLoaded bool
)

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
		"versionCode":     "330",
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

// getCustomChannels returns cached custom channels
func getCustomChannels() []Channel {
	customChannelsCacheMutex.RLock()
	defer customChannelsCacheMutex.RUnlock()

	// Return a copy to prevent external modifications
	channels := make([]Channel, len(customChannelsCache))
	copy(channels, customChannelsCache)
	return channels
}

// getCustomChannelByID efficiently looks up a custom channel by ID
func getCustomChannelByID(channelID string) (Channel, bool) {
	customChannelsCacheMutex.RLock()
	defer customChannelsCacheMutex.RUnlock()

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
func loadAndCacheCustomChannels() []Channel {
	customChannelsCacheMutex.Lock()
	defer customChannelsCacheMutex.Unlock()

	// Load channels from file
	channels, err := LoadCustomChannels(config.Cfg.CustomChannelsFile)
	if err != nil {
		if utils.Log != nil {
			utils.Log.Printf("Error loading custom channels: %v", err)
		}
		// Cache empty result to avoid repeated file I/O errors
		customChannelsCache = []Channel{}
		customChannelsCacheMap = make(map[string]Channel)
	} else {
		customChannelsCache = channels
		// Populate the map for efficient lookups
		customChannelsCacheMap = make(map[string]Channel)
		for _, channel := range channels {
			customChannelsCacheMap[channel.ID] = channel
		}

		// Warn user about performance implications if too many channels
		logExcessiveChannelsWarning(len(channels), "Cached")
	}

	customChannelsCacheLoaded = true

	// Return a copy to prevent external modifications
	result := make([]Channel, len(customChannelsCache))
	copy(result, customChannelsCache)
	return result
}

// ReloadCustomChannels reloads custom channels from file and updates cache
func ReloadCustomChannels() error {
	customChannelsCacheMutex.Lock()
	defer customChannelsCacheMutex.Unlock()

	channels, err := LoadCustomChannels(config.Cfg.CustomChannelsFile)
	if err != nil {
		return err
	}

	customChannelsCache = channels
	// Update the map for efficient lookups
	customChannelsCacheMap = make(map[string]Channel)
	for _, channel := range channels {
		customChannelsCacheMap[channel.ID] = channel
	}
	customChannelsCacheLoaded = true

	if utils.Log != nil {
		utils.Log.Printf("Reloaded %d custom channels", len(channels))

		// Warn user about performance implications if too many channels
		logExcessiveChannelsWarning(len(channels), "Reloaded")
	}

	return nil
}

// ClearCustomChannelsCache clears the custom channels cache
func ClearCustomChannelsCache() {
	customChannelsCacheMutex.Lock()
	defer customChannelsCacheMutex.Unlock()

	customChannelsCache = nil
	customChannelsCacheMap = nil
	customChannelsCacheLoaded = false
}

// Live method generates m3u8 link from JioTV API with the provided channel ID
func (tv *Television) Live(channelID string) (*LiveURLOutput, error) {
	// Check if this is a custom channel by looking it up efficiently
	// Custom channels are identified by the "cc_" prefix or by checking the custom channels cache
	if config.Cfg.CustomChannelsFile != "" {
		if channel, exists := getCustomChannelByID(channelID); exists {
			// For custom channels, return the URL directly (always m3u8 HLS)
			result := &LiveURLOutput{
				Result: channel.URL,
				Bitrates: Bitrates{
					Auto:   channel.URL,
					High:   channel.URL,
					Medium: channel.URL,
					Low:    channel.URL,
				},
				Code:    200,
				Message: "success",
			}
			return result, nil
		}
		
		// Also check for custom channels without "cc_" prefix for backward compatibility
		if !strings.HasPrefix(channelID, "cc_") {
			if channel, exists := getCustomChannelByID("cc_" + channelID); exists {
				// For custom channels, return the URL directly (always m3u8 HLS)
				result := &LiveURLOutput{
					Result: channel.URL,
					Bitrates: Bitrates{
						Auto:   channel.URL,
						High:   channel.URL,
						Medium: channel.URL,
						Low:    channel.URL,
					},
					Code:    200,
					Message: "success",
				}
				return result, nil
			}
		}
	}

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

	var url string
	if tv.AccessToken != "" {
		url = "https://" + JIOTV_API_DOMAIN + "/playback/apis/v1/geturl?langId=6"
		req.Header.Set(headers.AccessToken, tv.AccessToken)
	} else {
		req.Header.Set("osVersion", "8.1.0")
		req.Header.Set("ssotoken", tv.SsoToken)
		req.Header.Set("versionCode", "277")
		url = "https://" + TV_MEDIA_DOMAIN + "/apis/v2.2/getchannelurl/getchannelurl"
		req.Header.SetUserAgent("plaYtv/7.0.5 (Linux;Android 8.1.0) ExoPlayerLib/2.11.7")
	}
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
		utils.Log.Panic(err)
		return nil, err
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		// Store the response body as a string
		response := string(resp.Body())

		// Log headers and request data
		utils.Log.Println("Request headers:", req.Header.String())
		utils.Log.Println("Request data:", formData.String())
		utils.Log.Panicln("Response: ", response)

		return nil, fmt.Errorf("Request failed with status code: %d\nresponse: %s", resp.StatusCode(), response)
	}

	var result LiveURLOutput
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		utils.Log.Panic(err)
		return nil, err
	}

	return &result, nil
}

// Render method does HTTP GET request to the provided URL and return the response body
func (tv *Television) Render(url string) ([]byte, int) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)
	req.Header.SetMethod("GET")

	// Copy headers from the Television headers map to the request
	for key, value := range tv.Headers {
		req.Header.Set(key, value)
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Perform the HTTP GET request
	if err := tv.Client.Do(req, resp); err != nil {
		utils.Log.Panic(err)
	}

	buf := resp.Body()

	return buf, resp.StatusCode()
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
		return customConfig, fmt.Errorf(errUnsupportedChannelsFormat)
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
			return customConfig, fmt.Errorf(errUnsupportedChannelsFormat)
		}
		return customConfig, nil
	}

	// Try YAML for other content
	err := yaml.Unmarshal(data, &customConfig)
	if err != nil {
		return customConfig, fmt.Errorf(errUnsupportedChannelsFormat)
	}
	return customConfig, nil
}

// LoadCustomChannels loads custom channels from configuration file
func LoadCustomChannels(filePath string) ([]Channel, error) {
	if filePath == "" {
		return []Channel{}, nil
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if utils.Log != nil {
			utils.Log.Printf("Custom channels file not found: %s", filePath)
		}
		return []Channel{}, nil
	}

	// Read file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read custom channels file: %w", err)
	}

	// Parse the file using format detection
	customConfig, err := detectAndParseFormat(data, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse custom channels file: %w", err)
	}

	// Convert CustomChannel to Channel
	var channels []Channel
	for _, customChannel := range customConfig.Channels {
		// Prefix custom channel ID with "cc_" if not already prefixed
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

	if utils.Log != nil {
		utils.Log.Printf("Loaded %d custom channels from %s", len(channels), filePath)

		// Warn user about performance implications if too many channels
		logExcessiveChannelsWarning(len(channels), "You have loaded")
	}
	return channels, nil
}

// Channels fetch channels from JioTV API and merge with custom channels
func Channels() ChannelsResponse {

	// Create a fasthttp.Client
	client := utils.GetRequestClient()

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(CHANNELS_API_URL)

	req.Header.SetMethod("GET")
	req.Header.Add(headers.UserAgent, headers.UserAgentOkHttp)
	req.Header.Add(headers.Accept, headers.AcceptJSON)
	req.Header.Add(headers.DeviceType, headers.DeviceTypePhone)
	req.Header.Add(headers.OS, headers.OSAndroid)
	req.Header.Add("appkey", "NzNiMDhlYzQyNjJm")
	req.Header.Add("lbcookie", "1")
	req.Header.Add("usertype", "JIO")

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Perform the HTTP GET request
	if err := client.Do(req, resp); err != nil {
		utils.Log.Panic(err)
	}

	var apiResponse ChannelsResponse

	// Check the response status code
	if resp.StatusCode() != fasthttp.StatusOK {
		utils.Log.Panicf("Request failed with status code: %d", resp.StatusCode())
	}

	resp_body := resp.Body()

	// Parse the JSON response
	if err := json.Unmarshal(resp_body, &apiResponse); err != nil {
		utils.Log.Panic(err)
	}

	// disable sony channels temporarily
	// apiResponse.Result = append(apiResponse.Result, SONY_CHANNELS_API...)

	// Load and append custom channels if configured
	if config.Cfg.CustomChannelsFile != "" {
		customChannels := getCustomChannels()
		apiResponse.Result = append(apiResponse.Result, customChannels...)
	}

	return apiResponse
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

func ReplaceM3U8(baseUrl, match []byte, params, channel_id string) []byte {
	coded_url, err := secureurl.EncryptURL(string(baseUrl) + string(match) + "?" + params)
	if err != nil {
		utils.Log.Println(err)
		return nil
	}
	return []byte("/render.m3u8?auth=" + coded_url + "&channel_key_id=" + channel_id)
}

func ReplaceTS(baseUrl, match []byte, params string) []byte {
	if config.Cfg.DisableTSHandler {
		return []byte(string(baseUrl) + string(match) + "?" + params)
	}
	coded_url, err := secureurl.EncryptURL(string(baseUrl) + string(match) + "?" + params)
	if err != nil {
		utils.Log.Println(err)
		return nil
	}
	return []byte("/render.ts?auth=" + coded_url)
}

func ReplaceAAC(baseUrl, match []byte, params string) []byte {
	if config.Cfg.DisableTSHandler {
		return []byte(string(baseUrl) + string(match) + "?" + params)
	}
	coded_url, err := secureurl.EncryptURL(string(baseUrl) + string(match) + "?" + params)
	if err != nil {
		utils.Log.Println(err)
		return nil
	}
	return []byte("/render.ts?auth=" + coded_url)
}

func ReplaceKey(match []byte, params, channel_id string) []byte {
	coded_url, err := secureurl.EncryptURL(string(match) + "?" + params)
	if err != nil {
		utils.Log.Println(err)
		return nil
	}
	return []byte("/render.key?auth=" + coded_url + "&channel_key_id=" + channel_id)
}

func getSLChannel(channelID string) (*LiveURLOutput, error) {
	// Check if the channel is available in the SONY_CHANNELS map
	if val, ok := SONY_JIO_MAP[channelID]; ok {
		// If the channel is available in the SONY_CHANNELS map, then return the link
		result := new(LiveURLOutput)

		chu, err := base64.StdEncoding.DecodeString(SONY_CHANNELS[val])
		if err != nil {
			utils.Log.Panic(err)
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
			utils.Log.Panic(err)
		}

		if resp.StatusCode() != fasthttp.StatusFound {
			utils.Log.Panicf("Request failed with status code: %d", resp.StatusCode())
			utils.Log.Panicln("Response: ", string(resp.Body()))
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
