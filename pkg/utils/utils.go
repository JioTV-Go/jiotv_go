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

	"github.com/rabilrbl/jiotv_go/v2/internal/config"
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
	var logger *log.Logger
	if config.Cfg.Debug {
		logger = log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		// write logs to a file jiotv_go.log
		file, err := os.OpenFile("jiotv_go.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640) // skipcq: GSC-G302
		if err != nil {
			log.Println(err)
		}
		logger = log.New(file, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
		// rotate log file if it is larger than 10MB
		// neccessary to prevent filling up disk space with logs
		logger.SetOutput(&lumberjack.Logger{
			Filename:   "jiotv_go.log",
			MaxSize:    5, // megabytes
			MaxBackups: 3,
			MaxAge:     7, // days
		})
	}
	return logger
}

// GetCredentialsPath returns the file path to credentials file
func GetCredentialsPath() string {
	filename := "jiotv_credentials_v2.json"
	credentials_path := config.Cfg.CredentialsPath
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
				AndroidID: "6fcadeb7b4b10d77",
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
				AndroidID: "6fcadeb7b4b10d77",
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

// loadCredentialsFromFile loads credentials from file if available
// Returns JIOTV_CREDENTIALS struct
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

// GetJIOTVCredentials return credentials from environment variables or credentials file
// Important note: If credentials are provided from environment variables, they will be used instead of credentials file
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

// WriteJIOTVCredentials writes credentials data to file
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

// CheckLoggedIn function checks if user is logged in
func CheckLoggedIn() bool {
	// Check if credentials.json exists
	_, err := GetJIOTVCredentials()
	if err != nil {
		return false
	} else {
		return true
	}
}

// Logout function deletes credentials file
func Logout() error {
	credentialsPath := GetCredentialsPath()
	err := os.Remove(credentialsPath)
	if err != nil {
		return err
	}
	return nil
}

// ScheduleFunctionCall schedules a function call at a given time
func ScheduleFunctionCall(fn func(), executeTime time.Time) {
	now := time.Now()
	if executeTime.After(now) {
		time.Sleep(executeTime.Sub(now))
	}
	fn()
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
