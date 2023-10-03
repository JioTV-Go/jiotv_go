package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

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
		file, err := os.OpenFile("jiotv_go.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640) // skipcq: GSC-G302
		if err != nil {
			log.Println(err)
		}
		logger = log.New(file, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
		// rotate log file if it is larger than 10MB
		logger.SetOutput(&lumberjack.Logger{
			Filename:   "jiotv_go.log",
			MaxSize:    5, // megabytes
			MaxBackups: 3,
			MaxAge:     7, // days
		})
	}
	return logger
}

func GetCredentialsPath() string {
	filename := "jiotv_credentials_v2.json"
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
		credentials_path += filename
	} else {
		credentials_path = filename
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

		WriteJIOTVCredentials(&JIOTV_CREDENTIALS{
			SSOToken:             ssotoken,
			CRM:                  crm,
			UniqueID:             uniqueId,
			AccessToken:          accessToken,
			RefreshToken:         refreshtoken,
			LastTokenRefreshTime: strconv.FormatInt(time.Now().Unix(), 10),
		})
		return map[string]string{
			"status":       "success",
			"accessToken":  accessToken,
			"refreshToken": refreshtoken,
			"ssoToken":     ssotoken,
			"crm":          crm,
			"uniqueId":     uniqueId,
		}, nil
	} else {
		return map[string]string{
			"status":  "failed",
			"message": "Invalid OTP",
		}, nil
	}
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
		"identifier":           user,
		"password":             passw,
		"rememberUser":         "T",
		"upgradeAuth":          "Y",
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
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)
	req.Header.SetContentType("application/json")
	req.Header.SetMethod("POST")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	req.SetBody(payloadJSON)

	client := &fasthttp.Client{}
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

	ssoToken := result["ssoToken"].(string)
	if ssoToken != "" {
		crm := result["sessionAttributes"].(map[string]interface{})["user"].(map[string]interface{})["subscriberId"].(string)
		uniqueId := result["sessionAttributes"].(map[string]interface{})["user"].(map[string]interface{})["unique"].(string)

		WriteJIOTVCredentials(&JIOTV_CREDENTIALS{
			SSOToken:    ssoToken,
			CRM:         crm,
			UniqueID:    uniqueId,
			AccessToken: "",
		})

		return map[string]string{
			"status":   "success",
			"ssoToken": ssoToken,
			"crm":      result["sessionAttributes"].(map[string]interface{})["user"].(map[string]interface{})["subscriberId"].(string),
			"uniqueId": result["sessionAttributes"].(map[string]interface{})["user"].(map[string]interface{})["unique"].(string),
		}, nil
	} else {
		return map[string]string{
			"status":  "failed",
			"message": "Invalid credentials",
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
	// Use credentials from environment variables if available
	jiotv_ssoToken := os.Getenv("JIOTV_SSO_TOKEN")
	jiotv_crm := os.Getenv("JIOTV_CRM")
	jiotv_uniqueId := os.Getenv("JIOTV_UNIQUE_ID")
	if jiotv_ssoToken != "" && jiotv_crm != "" && jiotv_uniqueId != "" {
		Log.Println("Using credentials from environment variables")
		return &JIOTV_CREDENTIALS{
			SSOToken:    jiotv_ssoToken,
			CRM:         jiotv_crm,
			UniqueID:    jiotv_uniqueId,
			AccessToken: "",
		}, nil
	}
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
				Dial: fasthttpproxy.FasthttpHTTPDialerTimeout(proxy, 10*time.Second),
			}
		}
	}
	return &fasthttp.Client{
		Dial: fasthttp.DialFunc(func(addr string) (netConn net.Conn, err error) {
			return fasthttp.DialTimeout(addr, 5*time.Second)
		}),
	}
}

func FileExists(filename string) bool {
	// check if given file exists
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}
