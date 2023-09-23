package television

import (
	"encoding/json"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/rabilrbl/jiotv_go/internals/utils"
)

type Television struct {
	accessToken string
	ssoToken    string
	crm         string
	uniqueID    string
	headers     map[string]string
	client      *fasthttp.Client
}

type Channel struct {
	ID       int    `json:"channel_id"`
	Name     string `json:"channel_name"`
	URL      string `json:"channel_url"`
	LogoURL  string `json:"logoUrl"`
	Category int    `json:"channelCategoryId"`
	Language int    `json:"channelLanguageId"`
	IsHD     bool   `json:"isHD"`
}

type APIResponse struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Result  []Channel `json:"result"`
}

var CategoryMap = map[int]string{
	5:  "Entertainment",
	6:  "Movies",
	7:  "Kids",
	8:  "Sports",
	9:  "Lifestyle",
	10: "Infotainment",
	12: "News",
	13: "Music",
	15: "Devotional",
	16: "Business",
	17: "Educational",
	18: "Shopping",
	19: "JioDarshan",
}

var LanguageMap = map[int]string{
	1:  "Hindi",
	2:  "Marathi",
	3:  "Punjabi",
	4:  "Urdu",
	5:  "Bengali",
	6:  "English",
	7:  "Malayalam",
	8:  "Tamil",
	9:  "Gujarati",
	10: "Odia",
	11: "Telugu",
	12: "Bhojpuri",
	13: "Kannada",
	14: "Assamese",
	15: "Nepali",
	16: "French",
	18: "Other",
}

func NewTelevision(accessToken, ssoToken, crm, uniqueID string) *Television {
	headers := map[string]string{
		"Content-type": "application/x-www-form-urlencoded",
		"appkey":       "NzNiMDhlYzQyNjJm",
		"channel_id":   "",
		"crmid":        crm,
		"userId":       crm,
		"deviceId":     "e4286d7b481d69b8",
		"devicetype":   "phone",
		"isott":        "false",
		"languageId":   "6",
		"lbcookie":     "1",
		"os":           "android",
		"osVersion":    "13",
		"subscriberId": crm,
		"uniqueId":     uniqueID,
		"User-Agent":   "okhttp/4.2.2",
		"usergroup":    "tvYR7NSNn7rymo3F",
		"versionCode":  "315",
	}

	client := utils.GetRequestClient()

	return &Television{
		accessToken: accessToken,
		ssoToken:    ssoToken,
		crm:         crm,
		uniqueID:    uniqueID,
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

	if resp.StatusCode() == fasthttp.StatusBadRequest {
		// Store the response body as a string
		response := string(resp.Body())

		// Log headers and request data
		utils.Log.Println("Request headers:", req.Header.String())
		utils.Log.Println("Request data:", formData.String())
		utils.Log.Panicln("Response: ", response)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		utils.Log.Panic(err)
		return "", err
	}

	return result["result"].(string), nil
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
