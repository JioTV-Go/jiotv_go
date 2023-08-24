package main

import (
	"bytes"
	"net/http"
)

type CustomRequest struct {
	BaseURL  string
	Method   string
	Data     []byte
	Headers  map[string]string
}

func NewRequest(url, method string) *CustomRequest {
	credentials, err := getLoginCredentials()
	if err != nil {
		Log.Fatal(err)
	}
	Log.Output(2, "Creating new request")
	headers := map[string]string{
		"Content-type": "application/x-www-form-urlencoded",
        "appkey": "NzNiMDhlYzQyNjJm",
        "channelId": "",
        "channel_id": "",
        "crmid": credentials["crm"],
        "deviceId": "e4286d7b481d69b8",
        "devicetype": "phone",
        "isott": "true",
        "languageId": "6",
        "lbcookie": "1",
        "os": "android",
        "osVersion": "8.1.0",
        "srno": "230203144000",
        "ssotoken": credentials["ssoToken"],
        "subscriberId": credentials["crm"],
        "uniqueId": credentials["uniqueId"],
        "User-Agent": "plaYtv/7.0.5 (Linux;Android 8.1.0) ExoPlayerLib/2.11.7",
        "usergroup": "tvYR7NSNn7rymo3F",
        "versionCode": "277",
	}

	return &CustomRequest{
		BaseURL:  url,
		Method:   method,
		Data:     nil,
		Headers:  headers,
	}
}

func (cr *CustomRequest) SetData(data []byte) *CustomRequest {
	cr.Data = data
	return cr
}

func (cr *CustomRequest) SetHeader(key, value string) *CustomRequest {
	cr.Headers[key] = value
	return cr
}

func (cr *CustomRequest) MakeRequest() (*http.Response, error) {
	client := &http.Client{}

	Log.Output(2, "Making request to "+cr.BaseURL)
	req, err := http.NewRequest(cr.Method, cr.BaseURL, bytes.NewBuffer(cr.Data))

	if err != nil {
		Log.Fatal(err)
		return nil, err
	}

	for key, value := range cr.Headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	return resp, nil
}