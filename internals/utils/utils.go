package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
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
		err := os.Mkdir(credentials_path, 0640)
		if err != nil {
			// if folder already exists, ignore the error
			if !os.IsExist(err) {
				Log.Println(err)
			}
		}
		credentials_path += "jiotv_credentials.json"
	} else {
		credentials_path = "jiotv_credentials.json"
	}
	return credentials_path
}

func LoginSendOTP(number string) (bool, error) {
	postData := map[string]string{
		"number": number,
	}

	// convert number string to base64
	postData["number"] = base64.StdEncoding.EncodeToString([]byte(postData["number"]))

	// Construct payload
	payload := map[string]interface{}{
		"number": postData["number"],
	}

	// Convert payload to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return false, err
	}

	// Make the request
	url := "https://jiotvapi.media.jio.com/userservice/apis/v1/loginotp/send"

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)

	req.Header.SetContentType("application/json")
	req.Header.SetMethod("POST")
	req.Header.SetUserAgent("okhttp/4.2.2")
	// Set headers
	req.Header.Add("appname", "RJIL_JioTV")
	req.Header.Add("os", "android")
	req.Header.Add("devicetype", "phone")

	req.SetBody(payloadJSON)

	client := GetRequestClient()

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Perform the HTTP POST request
	if err := client.Do(req, resp); err != nil {
		return false, err
	}

	// Check the response status code
	if resp.StatusCode() != fasthttp.StatusNoContent {
		return false, fmt.Errorf("request failed with status code: %d body: %s", resp.StatusCode(), resp.Body())
	} else {
		return true, nil
	}
}

func LoginVerifyOTP(number, otp string) (map[string]string, error) {
	postData := map[string]string{
		"number": number,
		"otp":    otp,
	}

	// convert number string to base64
	postData["number"] = base64.StdEncoding.EncodeToString([]byte(postData["number"]))

	// Construct payload
	payload := map[string]interface{}{
		"number": postData["number"],
		"otp":    postData["otp"],
		"deviceInfo": map[string]interface{}{
			"consumptionDeviceName": "SM-G930F",
			"info": map[string]interface{}{
				"type": "android",
				"platform": map[string]string{
					"name": "SM-G930F",
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
	url := "https://jiotvapi.media.jio.com/userservice/apis/v1/loginotp/verify"

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)

	req.Header.SetContentType("application/json")
	req.Header.SetMethod("POST")
	req.Header.SetUserAgent("okhttp/4.2.2")
	// Set headers
	req.Header.Add("appname", "RJIL_JioTV")
	req.Header.Add("os", "android")
	req.Header.Add("devicetype", "phone")

	req.SetBody(payloadJSON)

	client := GetRequestClient()

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Perform the HTTP POST request
	if err := client.Do(req, resp); err != nil {
		return nil, err
	}

	// Check the response status code
	if resp.StatusCode() != fasthttp.StatusOK {
		return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode())
	}

	// Read response body
	body := resp.Body()

	var result map[string]interface{}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	accessToken := result["authToken"].(string)

	if accessToken != "" {
		refreshtoken := result["refreshToken"].(string)
		ssotoken := result["ssoToken"].(string)
		crm := result["sessionAttributes"].(map[string]interface{})["user"].(map[string]interface{})["subscriberId"].(string)
		uniqueId := result["sessionAttributes"].(map[string]interface{})["user"].(map[string]interface{})["unique"].(string)

		credentialsPath := getCredentialsPath()
		file, err := os.Create(credentialsPath)
		if err != nil {
			return nil, err
		}
		defer file.Close() // skipcq: GO-S2307

		// Write result as credentials.json
		file.WriteString(`{"ssoToken":"` + ssotoken + `","crm":"` + crm + `","uniqueId":"` + uniqueId + `","accessToken":"` + accessToken + `","refreshToken":"` + refreshtoken + `"}`)
		return map[string]string{
			"status":       "success",
			"accessToken":  accessToken,
			"refreshToken": refreshtoken,
			"ssoToken":     ssotoken,
			"crm":          crm,
			"uniqueId":     uniqueId,
		}, file.Sync()
	} else {
		return map[string]string{
			"status":  "failed",
			"message": "Invalid OTP",
		}, nil
	}
}

func LoginRefreshAccessToken() map[string]interface{} {
	Log.Println("Refreshing AccessToken...")
	tokenData, err := GetLoginCredentials()
	if err != nil {
		Log.Fatalln(err)
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}

	// Prepare the request body
	requestBody := map[string]interface{}{
		"appName":      "RJIL_JioTV",
		"deviceId":     "6fcadeb7b4b10d77",
		"refreshToken": tokenData["refreshToken"],
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		Log.Fatalln(err)
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}

	// Prepare the request
	req := fasthttp.AcquireRequest()
	req.SetRequestURI("https://auth.media.jio.com/tokenservice/apis/v1/refreshtoken?langId=6")
	req.Header.SetMethod("POST")
	req.Header.Set("devicetype", "phone")
	req.Header.Set("versionCode", "315")
	req.Header.Set("os", "android")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Host", "auth.media.jio.com")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("User-Agent", "okhttp/4.2.2")
	req.Header.Set("accessToken", tokenData["accessToken"])
	req.SetBody(requestBodyJSON)

	// Send the request
	resp := fasthttp.AcquireResponse()
	if err := fasthttp.Do(req, resp); err != nil {
		Log.Fatalln(err)
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}

	// Check the response
	if resp.StatusCode() != fasthttp.StatusOK {
		Log.Fatalln("Request failed with status code:", resp.StatusCode())
		return map[string]interface{}{
			"success": false,
			"message": "Token expired, please log in again.",
		}
	}

	// Parse the response body
	respBody, err := resp.BodyGunzip()
	if err != nil {
		Log.Fatalln(err)
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}
	var res map[string]interface{}
	if err := json.Unmarshal(respBody, &res); err != nil {
		Log.Fatalln(err)
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
	}

	// Update tokenData
	if authToken, ok := res["authToken"].(string); ok {
		tokenData["accessToken"] = authToken
		err := os.WriteFile(getCredentialsPath(), []byte(`{"ssoToken":"`+tokenData["ssoToken"]+`","crm":"`+tokenData["crm"]+`","uniqueId":"`+tokenData["uniqueId"]+`","accessToken":"`+tokenData["accessToken"]+`","refreshToken":"`+tokenData["refreshToken"]+`"}`), 0640)
		if err != nil {
			Log.Fatalln(err)
			return map[string]interface{}{
				"success": false,
				"message": err.Error(),
			}
		}
		return map[string]interface{}{
			"success": true,
			"message": "AccessToken Generated",
		}
	} else {
		return map[string]interface{}{
			"success": false,
			"message": "AccessToken not found in response",
		}
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
		defer file.Close() // skipcq: GO-S2307

		data, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(data, &credentials)
		if err != nil {
			return nil, err
		}
		return credentials, file.Sync()
	}
	return nil, err
}

func GetLoginCredentials() (map[string]string, error) {
	// Use credentials from environment variables if available
	jiotv_accessToken := os.Getenv("JIOTV_ACCESS_TOKEN")
	jiotv_ssoToken := os.Getenv("JIOTV_SSO_TOKEN")
	jiotv_crm := os.Getenv("JIOTV_CRM")
	jiotv_uniqueId := os.Getenv("JIOTV_UNIQUE_ID")
	if jiotv_accessToken != "" && jiotv_ssoToken != "" && jiotv_crm != "" && jiotv_uniqueId != "" {
		Log.Println("Using credentials from environment variables")
		return map[string]string{
			"accessToken": jiotv_accessToken,
			"ssoToken":    jiotv_ssoToken,
			"crm":         jiotv_crm,
			"uniqueId":    jiotv_uniqueId,
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

func GetRequestClient() *fasthttp.Client {
	// The function shall return a fasthttp.client with proxy if given
	proxy := os.Getenv("JIOTV_PROXY")

	if proxy != "" {
		Log.Println("Using proxy: " + proxy)
		// check if given proxy is socks5 or http
		if strings.HasPrefix(proxy, "socks5://") {
			// socks5 proxy
			return &fasthttp.Client{
				Dial: fasthttpproxy.FasthttpSocksDialer(proxy),
			}
		} else {
			// http proxy
			return &fasthttp.Client{
				Dial: fasthttpproxy.FasthttpHTTPDialer(proxy),
			}
		}
	}
	return &fasthttp.Client{}
}
