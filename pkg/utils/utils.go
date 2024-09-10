package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/rabilrbl/jiotv_go/v3/internal/config"
	"github.com/rabilrbl/jiotv_go/v3/pkg/store"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
)

var (
	// Log is a global logger
	// initialized in main.go
	// used to log debug messages and errors
	Log *log.Logger
)

// GetLogger creates a new logger instance with custom settings
func GetLogger() *log.Logger {
	logFilePath := GetPathPrefix() + "jiotv_go.log"
	var logger *log.Logger
	if config.Cfg.Debug {
		logger = log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		// write logs to a file jiotv_go.log
		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640) // skipcq: GSC-G302
		if err != nil {
			log.Println(err)
		}
		logger = log.New(file, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
		// rotate log file if it is larger than 10MB
		// neccessary to prevent filling up disk space with logs
		logger.SetOutput(&lumberjack.Logger{
			Filename:   logFilePath,
			MaxSize:    5, // megabytes
			MaxBackups: 3,
			MaxAge:     7, // days
		})
	}
	return logger
}

// LoginSendOTP sends OTP to the given number for login
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

// LoginVerifyOTP verifies OTP for login
func LoginVerifyOTP(number, otp string) (map[string]string, error) {
	// convert number string to base64
	encoded_number := base64.StdEncoding.EncodeToString([]byte(number))

	// Construct payload
	payload := LoginOTPPayload{
		Number: encoded_number,
		OTP:    otp,
		DeviceInfo: LoginPayloadDeviceInfo{
			ConsumptionDeviceName: "SM-G930F",
			Info: LoginPayloadDeviceInfoInfo{
				Type: "android",
				Platform: LoginPayloadDeviceInfoInfoPlatform{
					Name: "SM-G930F",
				},
				AndroidID: GetDeviceID(),
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

	var result LoginResponse

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	accessToken := result.AuthToken

	if accessToken != "" {
		refreshToken := result.RefreshToken
		ssoToken := result.SSOToken
		crm := result.SessionAttributes.User.SubscriberID
		uniqueId := result.SessionAttributes.User.Unique

		WriteJIOTVCredentials(&JIOTV_CREDENTIALS{
			SSOToken:             ssoToken,
			CRM:                  crm,
			UniqueID:             uniqueId,
			AccessToken:          accessToken,
			RefreshToken:         refreshToken,
			LastTokenRefreshTime: strconv.FormatInt(time.Now().Unix(), 10),
		})
		return map[string]string{
			"status":       "success",
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
			"ssoToken":     ssoToken,
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

// Login is used to login with username and password
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
	payload := LoginPasswordPayload{
		Identifier:           user,
		Password:             passw,
		RememberUser:         "T",
		UpgradeAuth:          "Y",
		ReturnSessionDetails: "T",
		DeviceInfo: LoginPayloadDeviceInfo{
			ConsumptionDeviceName: "Jio",
			Info: LoginPayloadDeviceInfoInfo{
				Type: "android",
				Platform: LoginPayloadDeviceInfoInfoPlatform{
					Name:    "vbox86p",
					Version: "8.0.0",
				},
				AndroidID: GetDeviceID(),
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

	var result LoginResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	ssoToken := result.SSOToken
	if ssoToken != "" {
		crm := result.SessionAttributes.User.SubscriberID
		uniqueId := result.SessionAttributes.User.Unique

		WriteJIOTVCredentials(&JIOTV_CREDENTIALS{
			SSOToken:    ssoToken,
			CRM:         crm,
			UniqueID:    uniqueId,
			AccessToken: "",
		})

		return map[string]string{
			"status":   "success",
			"ssoToken": ssoToken,
			"crm":      crm,
			"uniqueId": uniqueId,
		}, nil
	} else {
		return map[string]string{
			"status":  "failed",
			"message": "Invalid credentials",
		}, nil
	}
}

// GetPathPrefix alias for store.GetPathPrefix
func GetPathPrefix() string {
	return store.GetPathPrefix()
}

// GetDeviceID returns the device ID
func GetDeviceID() string {
	deviceID, err := store.Get("deviceId")
	if err != nil {
		Log.Println(err)
		err = GenerateRandomString()
		if err != nil {
			Log.Println(err)
			return ""
		}
		deviceID, err = store.Get("deviceId")
		if deviceID == "" {
			Log.Println("Device ID is empty")
			return ""
		} else if err != nil {
			Log.Println(err)
			return ""
		}
	}
	return deviceID
}

// GetJIOTVCredentials return credentials from environment variables or credentials file
// Important note: If credentials are provided from environment variables, they will be used instead of credentials file
func GetJIOTVCredentials() (*JIOTV_CREDENTIALS, error) {
	ssoToken, err := store.Get("ssoToken")
	if err != nil {
		return nil, err
	}

	crm, err := store.Get("crm")
	if err != nil {
		return nil, err
	}

	uniqueId, err := store.Get("uniqueId")
	if err != nil {
		return nil, err
	}

	// Empty for Password login
	accessToken, err := store.Get("accessToken")
	if err != nil {
		return nil, nil
	}

	// Empty for Password login
	refreshToken, err := store.Get("refreshToken")
	if err != nil {
		return nil, nil
	}

	// Empty for Password login
	lastTokenRefreshTime, err := store.Get("lastTokenRefreshTime")
	if err != nil {
		return nil, nil
	}

	lastSSOTokenRefreshTime, err := store.Get("lastSSOTokenRefreshTime")
	if err != nil {
		return nil, nil
	}

	return &JIOTV_CREDENTIALS{
		SSOToken:                ssoToken,
		CRM:                     crm,
		UniqueID:                uniqueId,
		AccessToken:             accessToken,
		RefreshToken:            refreshToken,
		LastTokenRefreshTime:    lastTokenRefreshTime,
		LastSSOTokenRefreshTime: lastSSOTokenRefreshTime,
	}, nil
}

// WriteJIOTVCredentials writes credentials data to file
func WriteJIOTVCredentials(credentials *JIOTV_CREDENTIALS) error {

	if err := store.Set("ssoToken", credentials.SSOToken); err != nil {
		return err
	}

	if err := store.Set("crm", credentials.CRM); err != nil {
		return err
	}

	if err := store.Set("uniqueId", credentials.UniqueID); err != nil {
		return err
	}

	if err := store.Set("accessToken", credentials.AccessToken); err != nil {
		return err
	}

	if err := store.Set("refreshToken", credentials.RefreshToken); err != nil {
		return err
	}

	if credentials.LastTokenRefreshTime != "" {
		if err := store.Set("lastTokenRefreshTime", credentials.LastTokenRefreshTime); err != nil {
			return err
		}
	} else {
		if err := store.Set("lastTokenRefreshTime", strconv.FormatInt(time.Now().Unix(), 10)); err != nil {
			return err
		}
	}

	if credentials.LastSSOTokenRefreshTime != "" {
		if err := store.Set("lastSSOTokenRefreshTime", credentials.LastSSOTokenRefreshTime); err != nil {
			return err
		}
	} else {
		if err := store.Set("lastSSOTokenRefreshTime", strconv.FormatInt(time.Now().Unix(), 10)); err != nil {
			return err
		}
	}

	return nil
}

// CheckLoggedIn function checks if user is logged in
func CheckLoggedIn() bool {
	// Check if credentials.json exists
	_, err := GetJIOTVCredentials()
	if err != nil {
		Log.Println(err)
		return false
	} else {
		return true
	}
}

// Logout function deletes credentials file
func Logout() error {
	// credentialsPath := GetCredentialsPath()
	// return os.Remove(credentialsPath)

	// Delete all key-value pairs from the store
	if err := store.Delete("ssoToken"); err != nil {
		return err
	}

	if err := store.Delete("crm"); err != nil {
		return err
	}

	if err := store.Delete("uniqueId"); err != nil {
		return err
	}

	if err := store.Delete("accessToken"); err != nil {
		return err
	}

	if err := store.Delete("refreshToken"); err != nil {
		return err
	}

	if err := store.Delete("lastTokenRefreshTime"); err != nil {
		return err
	}

	if err := store.Delete("lastSSOTokenRefreshTime"); err != nil {
		return err
	}

	return nil
}

// GetRequestClient create a HTTP client with proxy if given
// Otherwise create a HTTP client without proxy
// Returns a fasthttp.Client
func GetRequestClient() *fasthttp.Client {
	// The function shall return a fasthttp.client with proxy if given
	proxy := config.Cfg.Proxy

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

// FileExists function check if given file exists
func FileExists(filename string) bool {
	// check if given file exists
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

// GenerateCurrentTime generates current time in YYYYMMDDTHHMMSS format
func GenerateCurrentTime() string {
	currentTime := time.Now().UTC().Format("20060102T150405")
	return currentTime
}

// GenerateDate generates date in YYYYMMDD format
func GenerateDate() string {
	// 20231205
	currentTime := time.Now().UTC().Format("20060102")
	return currentTime
}

// ContainsString checks if item string is present in slice
func ContainsString(item string, slice []string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GenerateRandomString generates a random 16-character hexadecimal string.
func GenerateRandomString() error {
	bytes := make([]byte, 8) // 8 bytes will result in a 16-character hex string
	if _, err := rand.Read(bytes); err != nil {
		return err
	}
	if _, err := store.Get("deviceId"); err != nil {
		store.Set("deviceId", hex.EncodeToString(bytes))
	}
	return nil
}
