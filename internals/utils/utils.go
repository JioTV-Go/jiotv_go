package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
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
		credentials_path += "jiotv_credentials.json"
	} else {
		credentials_path = "jiotv_credentials.json"
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

		credentialsPath := getCredentialsPath()
		file, err := os.Create(credentialsPath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		// Write result as credentials.json
		file.WriteString(`{"ssoToken":"` + ssoToken + `","crm":"` + crm + `","uniqueId":"` + uniqueId + `"}`)
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

	client := &fasthttp.Client{}

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
					"name":    "SM-G930F",
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
		defer file.Close()

		// Write result as credentials.json
		file.WriteString(`{"ssoToken":"` + ssotoken + `","crm":"` + crm + `","uniqueId":"` + uniqueId + `","accessToken":"` + accessToken + `","refreshToken":"` + refreshtoken + `"}`)
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

func GetCustomDialer() func(ctx context.Context, network string, addr string) (net.Conn, error) {
	USER_CUSTOM_DNS := os.Getenv("JIOTV_DNS")
	if USER_CUSTOM_DNS == "" {
		USER_CUSTOM_DNS = "1.1.1.1"
	}

	var (
		dnsResolverIP        = USER_CUSTOM_DNS + ":53" // Cloudflare DNS resolver.
		dnsResolverProto     = "udp"                   // Protocol to use for the DNS resolver
		dnsResolverTimeoutMs = 5000                    // Timeout (ms) for the DNS resolver (optional)
	)

	dialer := &net.Dialer{
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Duration(dnsResolverTimeoutMs) * time.Millisecond,
				}
				return d.DialContext(ctx, dnsResolverProto, dnsResolverIP)
			},
		},
	}

	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, addr)
	}

	return dialContext
}
