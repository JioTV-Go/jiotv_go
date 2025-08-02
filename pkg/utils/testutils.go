package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

// Test-friendly API functions that can use custom base URLs

// LoginSendOTPWithBaseURL sends OTP to the given number for login using custom base URL
func LoginSendOTPWithBaseURL(number string, baseURL string) (bool, error) {
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
	url := baseURL + "/userservice/apis/v1/loginotp/send"

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

// LoginVerifyOTPWithBaseURL verifies OTP for login using custom base URL
func LoginVerifyOTPWithBaseURL(number, otp string, baseURL string) (map[string]string, error) {
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
	url := baseURL + "/userservice/apis/v1/loginotp/verify"

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

// LoginWithBaseURL is used to login with username and password using custom base URL
func LoginWithBaseURL(username, password string, baseURL string) (map[string]string, error) {
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
	url := baseURL + "/v3/dip/user/unpw/verify"
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

// PerformServerLogoutWithBaseURL attempts to log out the user from the JioTV servers using custom base URL
func PerformServerLogoutWithBaseURL(baseURL string) error {
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

	req.SetRequestURI(baseURL + "/tokenservice/apis/v1/logout?langId=6")
	req.Header.SetMethod("POST")
	req.Header.SetUserAgent("okhttp/4.9.3")
	req.Header.Set("Accept-Encoding", "gzip")
	if creds.AccessToken != "" {
		req.Header.Set("accesstoken", creds.AccessToken)
	} else {
		Log.Println("AccessToken is missing, proceeding without it for server logout.")
	}
	req.Header.Set("devicetype", "phone")
	req.Header.Set("versioncode", "371") // As per new requirement
	req.Header.Set("os", "android")
	if creds.UniqueID != "" {
		req.Header.Set("uniqueid", creds.UniqueID)
	} else {
		Log.Println("UniqueID is missing, proceeding without it for server logout.")
	}
	req.Header.Set("content-type", "application/json; charset=utf-8")
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