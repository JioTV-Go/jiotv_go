package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var Log *log.Logger

func GetLogger() *log.Logger {
	return log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
}

func getCredentialsPath() string {
	credentials_path := os.Getenv("JIOTV_CREDENTIALS_PATH")
	if credentials_path != "" {
		// if trailing slash is not present, add it
		if !strings.HasSuffix(credentials_path, "/") {
			credentials_path += "/"
		}
		// if folder path is not found, create the folder in current directory
		err := os.Mkdir(credentials_path, 0755)
		if err != nil {
			// if folder already exists, ignore the error
			if !os.IsExist(err) {
				Log.Println(err)
			}
		}
		credentials_path += "credentials.json"
	} else {
		credentials_path = "credentials.json"
	}
	return credentials_path
}

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

		credentials_path := getCredentialsPath()
		file, err := os.Create(credentials_path)
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
	// check if given file exists, if not ask user username and password then call Login()
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		Log.Println("Credentials file not found, please login at the website or goto /login?username=xxx&password=xxx")
	} else {
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
	return nil, err
}

func GetLoginCredentials() (map[string]string, error) {
	// Use credentials from environment variables if available
	jiotv_ssoToken := os.Getenv("JIOTV_SSO_TOKEN")
	jiotv_crm := os.Getenv("JIOTV_CRM")
	jiotv_uniqueId := os.Getenv("JIOTV_UNIQUE_ID")
	if jiotv_ssoToken != "" && jiotv_crm != "" && jiotv_uniqueId != "" {
		Log.Println("Using credentials from environment variables")
		return map[string]string{
			"ssoToken": jiotv_ssoToken,
			"crm":      jiotv_crm,
			"uniqueId": jiotv_uniqueId,
		}, nil
	}
	credentials_path := getCredentialsPath()
	credentials, err := loadCredentialsFromFile(credentials_path)
	if err != nil {
		return nil, err
	}
	return credentials, nil
}

func CheckLoggedIn() bool {
	// Check if credentials.json exists
	_, err := GetLoginCredentials()
	if err != nil {
		return false
	} else {
		return true
	}
}
