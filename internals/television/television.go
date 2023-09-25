package television

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/rabilrbl/jiotv_go/internals/utils"
)

func NewTelevision(credentials *utils.JIOTV_CREDENTIALS) *Television {
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
		"versionCode":  "315",
	}

	client := utils.GetRequestClient()

	return &Television{
		accessToken: credentials.AccessToken,
		ssoToken:    credentials.SSOToken,
		crm:         credentials.CRM,
		uniqueID:    credentials.UniqueID,
		headers:     headers,
		client:      client,
	}
}

func (tv *Television) Live(channelID string) (string, error) {
	formData := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(formData)

	formData.Add("channel_id", channelID)
	formData.Add("stream_type", "Seek")

	url := "https://jiotvapi.media.jio.com/playback/apis/v1/geturl?langId=6"

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)
	req.Header.SetMethod("POST")

	// Encode the form data and set it as the request body
	req.SetBody(formData.QueryString())

	// Copy headers from the Television headers map to the request
	for key, value := range tv.headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("accesstoken", tv.accessToken)
	req.Header.Set("channel_id", channelID)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Perform the HTTP POST request
	if err := tv.client.Do(req, resp); err != nil {
		utils.Log.Panic(err)
		return "", err
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		// Store the response body as a string
		response := string(resp.Body())

		// Log headers and request data
		utils.Log.Println("Request headers:", req.Header.String())
		utils.Log.Println("Request data:", formData.String())
		utils.Log.Panicln("Response: ", response)

		return "", fmt.Errorf("Request failed with status code: %d\nresponse: %s", resp.StatusCode(), response)
	}

	var result LiveURLOutput
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		utils.Log.Panic(err)
		return "", err
	}

	return result.Bitrates.Auto, nil
}

func (tv *Television) Render(url string) []byte {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)
	req.Header.SetMethod("GET")

	// Copy headers from the Television headers map to the request
	for key, value := range tv.headers {
		req.Header.Set(key, value)
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Perform the HTTP GET request
	if err := tv.client.Do(req, resp); err != nil {
		utils.Log.Panic(err)
	}

	buf := resp.Body()

	return buf
}

func (tv *Television) RenderKey(url, channelID string) ([]byte, int) {
	// extract params from url
	params := strings.Split(url, "?")[1]

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)
	req.Header.SetMethod("GET")

	// set params as cookies as JioTV uses cookies to authenticate
	for _, param := range strings.Split(params, "&") {
		key := strings.Split(param, "=")[0]
		value := strings.Split(param, "=")[1]
		req.Header.SetCookie(key, value)
	}

	// Copy headers from the Television headers map to the request
	for key, value := range tv.headers {
		req.Header.Set(key, value) // Assuming only one value for each header
	}
	req.Header.Set("srno", "230203144000")
	req.Header.Set("ssotoken", tv.ssoToken)
	req.Header.Set("channelId", channelID)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Perform the HTTP GET request
	if err := tv.client.Do(req, resp); err != nil {
		utils.Log.Panic(err)
	}

	buf := resp.Body()

	return buf, resp.StatusCode()
}

func (tv *Television) RenderTS(url string) ([]byte, int, map[string]string) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)
	req.Header.SetMethod("GET")

	// Copy headers from the Television headers map to the request
	for key, value := range tv.headers {
		req.Header.Set(key, value)
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Perform the HTTP GET request
	if err := tv.client.Do(req, resp); err != nil {
		utils.Log.Panic(err)
	}

	headers := make(map[string]string)
	resp.Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = string(value)
	})

	return resp.Body(), resp.StatusCode(), headers
}

func Channels() APIResponse {
	url := "https://jiotvapi.cdn.jio.com/apis/v3.0/getMobileChannelList/get/?langId=6&os=android&devicetype=phone&usertype=JIO&version=315&langId=6"

	// Create a fasthttp.Client
	client := utils.GetRequestClient()

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)

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

	var apiResponse APIResponse

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
