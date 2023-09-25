package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
)

var (
	Log *log.Logger
)

func GetLogger() *log.Logger {
	var logger *log.Logger
	if os.Getenv("JIOTV_DEBUG") == "true" {
		logger = log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		// write logs to a file jiotv_go.log
		file, err := os.OpenFile("jiotv_go.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
		if err != nil {
			log.Println(err)
		}
		logger = log.New(file, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
	}
	return logger
}

func GetCredentialsPath() string {
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
		credentials_path += "jiotv_credentials_v2.json"
	} else {
		credentials_path = "jiotv_credentials_v2.json"
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
	payload := map[string]string{
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

		credentialsPath := GetCredentialsPath()
		file, err := os.Create(credentialsPath)
		if err != nil {
			return nil, err
		}
		defer file.Close() // skipcq: GO-S2307

		// Write result as credentials.json
		file.WriteString(`{"ssoToken":"` + ssotoken + `","crm":"` + crm + `","uniqueId":"` + uniqueId + `","accessToken":"` + accessToken + `","refreshToken":"` + refreshtoken + `","lastTokenRefreshTime":"` + strconv.FormatInt(time.Now().Unix(), 10) + `"}`)
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

func loadCredentialsFromFile(filename string) (*JIOTV_CREDENTIALS, error) {
	// check if given file exists, if not ask user username and password then call Login()
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		Log.Println("Credentials file not found, please login at the website or goto /login?username=xxx&password=xxx")
	} else {
		var credentials JIOTV_CREDENTIALS
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
		return &credentials, nil
	}
	return nil, err
}

func GetJIOTVCredentials() (*JIOTV_CREDENTIALS, error) {
	credentials_path := GetCredentialsPath()
	credentials, err := loadCredentialsFromFile(credentials_path)
	if err != nil {
		return nil, err
	}
	return credentials, nil
}

func WriteJIOTVCredentials(credentials *JIOTV_CREDENTIALS) error {
	credentialsPath := GetCredentialsPath()
	file, err := os.Create(credentialsPath)
	if err != nil {
		return err
	}
	// Write result as credentials.json
	file.WriteString(`{"ssoToken":"` + credentials.SSOToken + `","crm":"` + credentials.CRM + `","uniqueId":"` + credentials.UniqueID + `","accessToken":"` + credentials.AccessToken + `","refreshToken":"` + credentials.RefreshToken + `","lastTokenRefreshTime":"` + strconv.FormatInt(time.Now().Unix(), 10) + `"}`)
	return file.Close()
}

func CheckLoggedIn() bool {
	// Check if credentials.json exists
	_, err := GetJIOTVCredentials()
	if err != nil {
		return false
	} else {
		return true
	}
}

func ScheduleFunctionCall(fn func(), executeTime time.Time) {
	now := time.Now()
	if executeTime.After(now) {
		time.Sleep(executeTime.Sub(now))
	}
	fn()
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
