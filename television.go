package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
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

func (tv *Television) live(channelID string) string {
	formData := url.Values{
		"channel_id":   []string{channelID},
		"channelId":    []string{channelID},
		"stream_type":  []string{"Seek"},
	}
	data := formData.Encode()

	url := "https://tv.media.jio.com/apis/v2.2/getchannelurl/getchannelurl"
	req, _ := http.NewRequest("POST", url, strings.NewReader(data))
	req.Header = tv.headers
	resp, err := tv.client.Do(req)
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	return result["result"].(string)
}

func (tv *Television) render(url string) []byte {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		Log.Fatal(err)
	}
	req.Header = tv.headers

	resp, err := tv.client.Do(req)
	if err != nil {
		log.Panic(err)
	}

	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	return buf.Bytes()
}

func (tv *Television) renderKey(url string, channelID string) ([]byte, int) {
	headers := tv.headers
	headers["channelId"] = []string{channelID}
	headers["channel_id"] = []string{channelID}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header = headers

	resp, err := tv.client.Do(req)
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	return buf.Bytes(), resp.StatusCode
}

func (tv *Television) getRequest(url string) *http.Request {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header = tv.headers
	return req
}

func (tv *Television) channels() APIResponse {
	url := "https://jiotv.data.cdn.jio.com/apis/v1.3/getMobileChannelList/get/?os=android&devicetype=phone"
	req := tv.getRequest(url)
	resp, err := tv.client.Do(req)
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()

	var apiResponse APIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResponse)
	if err != nil {
		Log.Panic(err)
	}
	return apiResponse

} 
