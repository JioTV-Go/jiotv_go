package television

import (
	"encoding/json"
	"fmt"

	"github.com/valyala/fasthttp"

	"github.com/rabilrbl/jiotv_go/v2/pkg/utils"
)

const (
	// URL for fetching channels from JioTV API
	CHANNELS_API_URL = "https://jiotvapi.cdn.jio.com/apis/v3.0/getMobileChannelList/get/?langId=6&os=android&devicetype=phone&usertype=JIO&version=315&langId=6"
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
		"Content-type": "application/x-www-form-urlencoded",
		"appkey":       "NzNiMDhlYzQyNjJm",
		"channel_id":   "",
		"crmid":        credentials.CRM,
		"userId":       credentials.CRM,
		"deviceId":     "e4286d7b481d69b8",
		"devicetype":   "phone",
		"isott":        "false",
		"languageId":   "6",
		"lbcookie":     "1",
		"os":           "android",
		"osVersion":    "13",
		"subscriberId": credentials.CRM,
		"uniqueId":     credentials.UniqueID,
		"User-Agent":   "okhttp/4.2.2",
		"usergroup":    "tvYR7NSNn7rymo3F",
		"versionCode":  "330",
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

// Live method generates m3u8 link from JioTV API with the provided channel ID
func (tv *Television) Live(channelID string) (*LiveURLOutput, error) {
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
		url = "https://jiotvapi.media.jio.com/playback/apis/v1/geturl?langId=6"
		req.Header.Set("accesstoken", tv.AccessToken)
	} else {
		req.Header.Set("osVersion", "8.1.0")
		req.Header.Set("ssotoken", tv.SsoToken)
		req.Header.Set("versionCode", "277")
		url = "https://tv.media.jio.com/apis/v2.2/getchannelurl/getchannelurl"
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
func (tv *Television) Render(url string) []byte {
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

	return buf
}

// Channels fetch channels from JioTV API
func Channels() ChannelsResponse {

	// Create a fasthttp.Client
	client := utils.GetRequestClient()

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(CHANNELS_API_URL)

	req.Header.SetMethod("GET")
	req.Header.Add("User-Agent", "okhttp/4.2.2")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("devicetype", "phone")
	req.Header.Add("os", "android")
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

	resp_body, err := resp.BodyGunzip()
	if err != nil {
		utils.Log.Panic(err)
	}

	// Parse the JSON response
	if err := json.Unmarshal(resp_body, &apiResponse); err != nil {
		utils.Log.Panic(err)
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
