package main

import (
	"os"
	"io"
	"encoding/json"
	"bytes"
	"net/http"
	"strings"
)

func Login(username, password string) (map[string]string, error) {
	postData := map[string]string{
		"username": username,
		"password": password,
	}

	// Process the username
	u := postData["username"]
	var user string
	if strings.Contains(u, "@") {
		user = u
	} else {
		user = "+91" + u
	}

	passw := postData["password"]

	// Set headers
	headers := map[string]string{
		"x-api-key":    "l7xx75e822925f184370b2e25170c5d5820a",
		"Content-Type": "application/json",
	}

	// Construct payload
	payload := map[string]interface{}{
		"identifier":          user,
		"password":            passw,
		"rememberUser":        "T",
		"upgradeAuth":         "Y",
		"returnSessionDetails": "T",
		"deviceInfo": map[string]interface{}{
			"consumptionDeviceName": "Jio",
			"info": map[string]interface{}{
				"type": "android",
				"platform": map[string]string{
					"name":    "vbox86p",
					"version": "8.0.0",
				},
				"androidId": "6fcadeb7b4b10d77",
			},
		},
	}

	// Convert payload to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Make the request
	url := "https://api.jio.com/v3/dip/user/unpw/verify"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	ssoToken := result["ssoToken"].(string)
	if ssoToken != "" {
		crm := result["sessionAttributes"].(map[string]interface{})["user"].(map[string]interface{})["subscriberId"].(string)
		uniqueId := result["sessionAttributes"].(map[string]interface{})["user"].(map[string]interface{})["unique"].(string)

		// write result as credentials.json
		file, err := os.Create("credentials.json")
		if err != nil {
			return nil, err
		}
		defer file.Close()

		// write result as credentials.json
		file.WriteString(`{"ssoToken":"` + ssoToken + `","crm":"` + crm + `","uniqueId":"` + uniqueId + `"}`)
		return map[string]string{
			"status":    "success",
			"ssoToken":  ssoToken,
			"crm":       result["sessionAttributes"].(map[string]interface{})["user"].(map[string]interface{})["subscriberId"].(string),
			"uniqueId":  result["sessionAttributes"].(map[string]interface{})["user"].(map[string]interface{})["unique"].(string),
		}, nil
	} else {
		return map[string]string{
			"status":  "failed",
			"message": "Invalid credentials",
		}, nil
	}
}

func loadCredentialsFromFile(filename string) (map[string]string, error) {
	credentials := make(map[string]string)
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &credentials)
	if err != nil {
		return nil, err
	}

	return credentials, nil
}

func getLoginCredentials() (map[string]string, error) {
	// Check if credentials.json exists
	if _, err := os.Stat("credentials.json"); os.IsNotExist(err) {
		Log.Fatal("credentials.json not found, please login first")
		return nil, err
	}
	Log.Output(2,"Accessing credentials")
	credentials, err := loadCredentialsFromFile("credentials.json")
	if err != nil {
		return nil, err
	}
	return credentials, nil
}

func getTV() *Television {
	credentials, err := getLoginCredentials()
	if err != nil {
		Log.Fatal(err)
	}
	tv := NewTelevision(credentials["ssoToken"], credentials["crm"], credentials["uniqueId"])
	return tv
}
