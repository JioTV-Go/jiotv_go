package television

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"github.com/rabilrbl/jiotv_go/internals/utils"
)


type Television struct {
	ssoToken  string
	crm       string
	uniqueID  string
	headers   map[string][]string
	client    *http.Client
}

type Channel struct {
	ID   int    `json:"channel_id"`
	Name string `json:"channel_name"`
	URL  string `json:"channel_url"`
	LogoURL string `json:"logoUrl"`
	Category int `json:"channelCategoryId"`
	Language int `json:"channelLanguageId"` 
	IsHD bool `json:"isHD"`
}

type APIResponse struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Result  []Channel `json:"result"`
}

func NewTelevision(ssoToken, crm, uniqueID string) *Television {
	headers := http.Header{
		"Content-type":   {"application/x-www-form-urlencoded"},
		"appkey":         {"NzNiMDhlYzQyNjJm"},
		"channelId":      {""},
		"channel_id":     {""},
		"crmid":          {crm},
		"deviceId":       {"e4286d7b481d69b8"},
		"devicetype":     {"phone"},
		"isott":          {"true"},
		"languageId":     {"6"},
		"lbcookie":       {"1"},
		"os":             {"android"},
		"osVersion":      {"8.1.0"},
		"srno":           {"230203144000"},
		"ssotoken":       {ssoToken},
		"subscriberId":   {crm},
		"uniqueId":       {uniqueID},
		"User-Agent":     {"plaYtv/7.0.5 (Linux;Android 8.1.0) ExoPlayerLib/2.11.7"},
		"usergroup":      {"tvYR7NSNn7rymo3F"},
		"versionCode":    {"277"},
	}

	// Create a new cookie jar
	jar, _ := cookiejar.New(nil)

	// Create an http.Client using the cookie jar
	client := &http.Client{
		Jar: jar,
	}

	return &Television{
		ssoToken: ssoToken,
		crm:      crm,
		uniqueID: uniqueID,
		headers:  headers,
		client:   client,
	}
}

func (tv *Television) Live(channelID string) string {
	formData := url.Values{
		"channel_id":   []string{channelID},
		"channelId":    []string{channelID},
		"stream_type":  []string{"Seek"},
	}
	data := formData.Encode()

	url := "https://tv.media.jio.com/apis/v2.2/getchannelurl/getchannelurl"

	// remove old cookies
	tv.client.Jar, _ = cookiejar.New(nil)

	req, _ := http.NewRequest("POST", url, strings.NewReader(data))
	req.Header = tv.headers
	resp, err := tv.client.Do(req)
	if err != nil {
		utils.Log.Panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 400 {
		// store string response 
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		response := buf.String()
		// add headers and data from request
		utils.Log.Println("Request headers:", req.Header)
		utils.Log.Println("Request data:", data)
		utils.Log.Panicln("Response: ", response)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	return result["result"].(string)
}

func (tv *Television) Render(url string) []byte {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		utils.Log.Fatal(err)
	}
	req.Header = tv.headers

	// go http keeps adding more cookies to the request header, leading large request header size
	// so we reset the cookie header, so that only new cookies are present
	req.Header.Del("Cookie")

	resp, err := tv.client.Do(req)
	if err != nil {
		utils.Log.Panic(err)
	}

	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	return buf.Bytes()
}

func (tv *Television) RenderKey(url string, channelID string) ([]byte, int) {
	headers := tv.headers
	headers["channelId"] = []string{channelID}
	headers["channel_id"] = []string{channelID}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header = headers

	resp, err := tv.client.Do(req)
	if err != nil {
		utils.Log.Panic(err)
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	return buf.Bytes(), resp.StatusCode
}

func Channels() APIResponse {
	url := "https://jiotv.data.cdn.jio.com/apis/v1.3/getMobileChannelList/get/?os=android&devicetype=phone"
	resp, err := http.Get(url)
	if err != nil {
		utils.Log.Panic(err)
	}
	defer resp.Body.Close()

	var apiResponse APIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResponse)
	if err != nil {
		utils.Log.Panic(err)
	}
	return apiResponse

}

func FilterChannels(channels []Channel, language int, category int) []Channel {
	var filteredChannels []Channel
	for _, channel := range channels {
		if channel.Language == language || channel.Category == category {
			filteredChannels = append(filteredChannels, channel)
		}
	}
	return filteredChannels
}
