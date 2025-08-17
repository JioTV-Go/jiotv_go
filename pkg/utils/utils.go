package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io" // Ensure io is imported
	"log"
	"net"
	"os"
	"path/filepath" // Ensure path/filepath is imported
	"strconv"
	"strings"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/jiotv-go/jiotv_go/v3/internal/config"
	"github.com/jiotv-go/jiotv_go/v3/internal/constants/headers"
	"github.com/jiotv-go/jiotv_go/v3/internal/constants/urls"
	"github.com/jiotv-go/jiotv_go/v3/pkg/store"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
)

const (
	// JioTV API domain constants
	JIOTV_API_DOMAIN  = urls.JioTVAPIDomain
	API_JIO_DOMAIN    = urls.APIJioDomain
	AUTH_MEDIA_DOMAIN = urls.AuthMediaDomain
)

// Log is a global logger
// initialized in main.go
// used to log debug messages and errors
var Log *log.Logger

// GetLogger creates a new logger instance with custom settings
func GetLogger() *log.Logger {
	// Step 1: Determine Log File Path
	logFilePath := "" // Initialize logFilePath
	if config.Cfg.LogPath != "" {
		logFilePath = filepath.Join(config.Cfg.LogPath, "jiotv_go.log")
		// Ensure the directory config.Cfg.LogPath exists.
		if _, err := os.Stat(config.Cfg.LogPath); os.IsNotExist(err) {
			if err := os.MkdirAll(config.Cfg.LogPath, 0755); err != nil {
				// Log error if directory creation fails. Lumberjack will handle actual file I/O errors.
				log.Printf("Error creating custom log directory %s: %v. File logging by lumberjack might fail.", config.Cfg.LogPath, err)
			}
		}
	} else {
		// If LogPath is empty, use the default path.
		logFilePath = filepath.Join(GetPathPrefix(), "jiotv_go.log")
		// Ensure the default log directory exists.
		defaultLogDir := filepath.Dir(logFilePath) // Get directory from path
		if _, err := os.Stat(defaultLogDir); os.IsNotExist(err) {
			if err := os.MkdirAll(defaultLogDir, 0755); err != nil {
				// Log error if default directory creation fails. Lumberjack will handle actual file I/O errors.
				log.Printf("Error creating default log directory %s: %v. File logging by lumberjack might fail.", defaultLogDir, err)
			}
		}
	}

	// Step 2: Initialize Writers
	outputWriters := []io.Writer{}
	if config.Cfg.LogToStdout {
		outputWriters = append(outputWriters, os.Stdout)
	}

	fileLogger := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    5, // megabytes
		MaxBackups: 3,
		MaxAge:     7, // days
	}
	outputWriters = append(outputWriters, fileLogger)

	// Step 3: Create Logger
	if len(outputWriters) == 0 {
		// This case means LogToStdout was false and file logging was somehow skipped (e.g. logFilePath became empty).
		// Default to os.Stdout with a warning.
		log.Println("Warning: No logging output explicitly configured (e.g., LogToStdout is false and file path is invalid or empty). Defaulting to Stdout.")
		outputWriters = append(outputWriters, os.Stdout)
	}

	multiWriter := io.MultiWriter(outputWriters...)
	logger := log.New(multiWriter, "", 0) // Initial prefix and flags are set to zero values.

	// Step 4: Set Logger Flags and Prefix
	if config.Cfg.Debug {
		logger.SetPrefix("[DEBUG] ")
		logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	} else {
		logger.SetPrefix("[INFO] ")
		logger.SetFlags(log.Ldate | log.Ltime)
	}

	return logger // Step 5: Return the configured logger
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
	url := "https://" + JIOTV_API_DOMAIN + "/userservice/apis/v1/loginotp/send"

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)

	req.Header.SetContentType(headers.ContentTypeJSON)
	req.Header.SetMethod("POST")
	req.Header.SetUserAgent(headers.UserAgentOkHttp)
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
	url := "https://" + JIOTV_API_DOMAIN + "/userservice/apis/v1/loginotp/verify"

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)

	req.Header.SetContentType(headers.ContentTypeJSON)
	req.Header.SetMethod("POST")
	req.Header.SetUserAgent(headers.UserAgentOkHttp)
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
	headerMap := map[string]string{
		headers.XAPIKey:     headers.APIKeyJio,
		headers.ContentType: headers.ContentTypeJSON,
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
	url := "https://" + API_JIO_DOMAIN + "/v3/dip/user/unpw/verify"
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(url)
	req.Header.SetContentType(headers.ContentTypeJSON)
	req.Header.SetMethod("POST")
	for key, value := range headerMap {
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
	// Perform server-side logout first
	if err := PerformServerLogout(); err != nil {
		// Log the error but continue with local logout
		Log.Printf("PerformServerLogout failed: %v", err)
	}

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

// PerformServerLogout attempts to log out the user from the JioTV servers.
func PerformServerLogout() error {
	Log.Println("Attempting server-side logout...")

	creds, err := GetJIOTVCredentials()
	if err != nil {
		Log.Printf("Error getting credentials for server logout: %v\n", err)
		// Depending on the error, we might still proceed if critical info like refreshToken is available
		// For now, we'll attempt to proceed if creds is not nil, or return if it is.
		if creds == nil {
			return fmt.Errorf("failed to get credentials: %w", err)
		}
	}

	deviceID := GetDeviceID()
	if deviceID == "" {
		Log.Println("Device ID is empty, cannot perform server logout.")
		return fmt.Errorf("deviceId is empty")
	}

	// refreshToken is crucial for logout
	if creds.RefreshToken == "" {
		Log.Println("RefreshToken is missing, cannot perform server logout.")
		return fmt.Errorf("refreshToken is missing")
	}

	// Construct the request body
	requestBodyMap := map[string]string{
		"appName":      "RJIL_JioTV",
		"deviceId":     deviceID,
		"refreshToken": creds.RefreshToken,
	}

	requestBodyJSON, err := json.Marshal(requestBodyMap)
	if err != nil {
		Log.Printf("Error marshalling server logout request body: %v", err)
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Construct the request
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI("https://" + AUTH_MEDIA_DOMAIN + "/tokenservice/apis/v1/logout?langId=6")
	req.Header.SetMethod("POST")
	req.Header.SetUserAgent(headers.UserAgentOkHttp)
	req.Header.Set(headers.AcceptEncoding, headers.AcceptEncodingGzip)
	if creds.AccessToken != "" {
		req.Header.Set(headers.AccessToken, creds.AccessToken)
	} else {
		Log.Println("AccessToken is missing, proceeding without it for server logout.")
	}
	req.Header.Set(headers.DeviceType, headers.DeviceTypePhone)
	req.Header.Set("versioncode", "371") // As per new requirement
	req.Header.Set(headers.OS, headers.OSAndroid)
	if creds.UniqueID != "" {
		req.Header.Set("uniqueid", creds.UniqueID)
	} else {
		Log.Println("UniqueID is missing, proceeding without it for server logout.")
	}
	req.Header.Set(headers.ContentType, headers.ContentTypeJSONCharsetUTF8)
	req.SetBody(requestBodyJSON)

	// Get the HTTP client
	client := GetRequestClient()

	// Perform the HTTP POST request
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	if err := client.Do(req, resp); err != nil {
		Log.Printf("Error performing server logout request: %v\n", err)
		return fmt.Errorf("http request failed: %w", err)
	}

	// Log the response status code
	Log.Printf("Server logout API response status code: %d", resp.StatusCode())

	if resp.StatusCode() >= 200 && resp.StatusCode() < 300 {
		Log.Println("Server-side logout successful.")
		return nil
	}

	Log.Printf("Server-side logout failed with status code: %d, body: %s\n", resp.StatusCode(), string(resp.Body()))
	return fmt.Errorf("server logout API request failed with status code: %d", resp.StatusCode())
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
				Dial: fasthttpproxy.FasthttpSocksDialerDualStack(proxy),
			}
		} else {
			// http proxy
			return &fasthttp.Client{
				Dial: fasthttpproxy.FasthttpHTTPDialerDualStackTimeout(proxy, 10*time.Second),
			}
		}
	}
	return &fasthttp.Client{
		Dial: fasthttp.DialFunc(func(addr string) (netConn net.Conn, err error) {
			return fasthttp.DialDualStackTimeout(addr, 5*time.Second)
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

func BuildHLSPlayURL(quality, channelID string) string {
    if quality != "" {
        return fmt.Sprintf("/live/%s/%s.m3u8", quality, channelID)
    }
    return fmt.Sprintf("/live/%s.m3u8", channelID)
}
