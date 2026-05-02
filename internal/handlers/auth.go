package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jiotv-go/jiotv_go/v3/internal/constants/headers"
	"github.com/jiotv-go/jiotv_go/v3/internal/constants/urls"
	internalUtils "github.com/jiotv-go/jiotv_go/v3/internal/utils"
	"github.com/jiotv-go/jiotv_go/v3/pkg/television"
	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
	"github.com/valyala/fasthttp"
)

var (
	// tokenRefreshMutex prevents concurrent token refreshes
	tokenRefreshMutex sync.Mutex
)

// IsAccessTokenExpired checks if the AccessToken needs refreshing
// Returns true if the token is expired or will expire within the next 10 minutes
func IsAccessTokenExpired(credentials *utils.JIOTV_CREDENTIALS) bool {
	if credentials == nil || credentials.AccessToken == "" {
		return true
	}

	return shouldRefreshToken(
		credentials.AccessToken,
		credentials.LastTokenRefreshTime,
		jwtTokenRefreshLeadTime,
		accessTokenFallbackTTL,
		accessTokenFallbackLeadTime,
		time.Now(),
	)
}

// IsSSOTokenExpired checks if the SSOToken needs refreshing
// Returns true if the token is expired or will expire within the next hour
func IsSSOTokenExpired(credentials *utils.JIOTV_CREDENTIALS) bool {
	if credentials == nil || credentials.SSOToken == "" {
		return true
	}

	return shouldRefreshToken(
		credentials.SSOToken,
		credentials.LastSSOTokenRefreshTime,
		jwtTokenRefreshLeadTime,
		ssoTokenFallbackTTL,
		ssoTokenFallbackLeadTime,
		time.Now(),
	)
}

// EnsureFreshTokens checks and refreshes tokens if needed
// This is the main function that should be called before making API requests
func EnsureFreshTokens() error {
	tokenRefreshMutex.Lock()
	defer tokenRefreshMutex.Unlock()

	credentials, err := utils.GetJIOTVCredentials()
	if err != nil {
		return fmt.Errorf("failed to get credentials: %v", err)
	}
	if credentials == nil {
		return fmt.Errorf("failed to get credentials: credentials are empty")
	}

	var refreshed bool

	// Check and refresh AccessToken if needed
	if credentials.AccessToken != "" && credentials.RefreshToken != "" {
		if IsAccessTokenExpired(credentials) {
			utils.Log.Println("AccessToken is expired, refreshing...")
			err := LoginRefreshAccessToken()
			if err != nil {
				utils.Log.Printf("AccessToken refresh failed: %v", err)
				return err
			}
			refreshed = true
		}
	}

	// Check and refresh SSOToken if needed
	if credentials.SSOToken != "" && credentials.UniqueID != "" {
		if IsSSOTokenExpired(credentials) {
			utils.Log.Println("SSOToken is expired, refreshing...")
			err := LoginRefreshSSOToken()
			if err != nil {
				utils.Log.Printf("SSOToken refresh failed: %v", err)
				return err
			}
			refreshed = true
		}
	}

	if refreshed {
		// Update the TV object with fresh credentials
		freshCreds, err := utils.GetJIOTVCredentials()
		if err != nil {
			return fmt.Errorf("failed to get fresh credentials: %v", err)
		}
		TV = television.New(freshCreds)
	}

	return nil
}

// LoginSendOTPHandler sends OTP for login
func LoginSendOTPHandler(c *fiber.Ctx) error {
	// get mobile number from post request
	formBody := new(LoginSendOTPRequestBodyData)
	err := c.BodyParser(&formBody)
	if err != nil {
		utils.Log.Println(err)
		return internalUtils.BadRequestError(c, "Invalid JSON")
	}
	mobileNumber := formBody.MobileNumber
	if err := internalUtils.CheckFieldExist(c, "Mobile Number", mobileNumber != ""); err != nil {
		return err
	}

	result, err := utils.LoginSendOTP(mobileNumber)
	if err != nil {
		utils.Log.Println(err)
		return internalUtils.InternalServerError(c, err)
	}
	return c.JSON(fiber.Map{
		"status": result,
	})
}

// LoginVerifyOTPHandler verifies OTP and login
func LoginVerifyOTPHandler(c *fiber.Ctx) error {
	// get mobile number and otp from post request
	formBody := new(LoginVerifyOTPRequestBodyData)
	err := c.BodyParser(&formBody)
	if err != nil {
		utils.Log.Println(err)
		return internalUtils.BadRequestError(c, "Invalid JSON")
	}
	mobileNumber := formBody.MobileNumber
	if err := internalUtils.CheckFieldExist(c, "Mobile Number", mobileNumber != ""); err != nil {
		return err
	}
	otp := formBody.OTP
	if err := internalUtils.CheckFieldExist(c, "OTP", otp != ""); err != nil {
		return err
	}

	result, err := utils.LoginVerifyOTP(mobileNumber, otp)
	if err != nil {
		utils.Log.Println(err)
		return internalUtils.InternalServerError(c, "Internal server error")
	}
	Init()
	return c.JSON(result)
}

// LogoutHandler is used to logout
func LogoutHandler(c *fiber.Ctx) error {
	if !isLogoutDisabled {
		err := utils.Logout()
		if err != nil {
			utils.Log.Println(err)
			return internalUtils.InternalServerError(c, "Internal server error")
		}
		Init()
	}
	return c.Redirect("/", fiber.StatusFound)
}

// LoginRefreshAccessToken Function is used to refresh AccessToken
func LoginRefreshAccessToken() error {
	utils.Log.Println("Refreshing AccessToken...")
	tokenData, err := utils.GetJIOTVCredentials()
	if err != nil {
		utils.Log.Printf("Error getting credentials for AccessToken refresh: %v", err)
		return err
	}

	// Validate that we have the required refresh token
	if tokenData.RefreshToken == "" {
		err := fmt.Errorf("RefreshToken is empty, cannot refresh AccessToken")
		utils.Log.Printf("Error: %v", err)
		return err
	}

	// Prepare the request body
	requestBody := map[string]string{
		"appName":      "RJIL_JioTV",
		"deviceId":     utils.GetDeviceID(),
		"refreshToken": tokenData.RefreshToken,
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		utils.Log.Printf("Error marshaling request body for AccessToken refresh: %v", err)
		return err
	}

	// Prepare the request
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(REFRESH_TOKEN_URL)
	req.Header.SetMethod("POST")
	req.Header.Set(headers.DeviceType, headers.DeviceTypePhone)
	req.Header.Set(headers.VersionCode, headers.VersionCode389)
	req.Header.Set(headers.OS, headers.OSAndroid)
	req.Header.Set(headers.ContentType, headers.ContentTypeJSONCharsetUTF8)
	req.Header.Set(headers.Host, urls.AuthMediaDomain)
	req.Header.Set(headers.UserAgent, headers.UserAgentOkHttp)
	req.Header.Set(headers.AccessToken, tokenData.AccessToken)
	req.SetBody(requestBodyJSON)

	// Send the request
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	client := utils.GetRequestClient()
	if err := client.Do(req, resp); err != nil {
		utils.Log.Printf("HTTP request failed for AccessToken refresh: %v", err)
		return err
	}

	// Check the response
	if resp.StatusCode() != fasthttp.StatusOK {
		err := fmt.Errorf("AccessToken refresh failed with status code: %d, body: %s", resp.StatusCode(), string(resp.Body()))
		utils.Log.Printf("Error: %v", err)
		return err
	}

	// Parse the response body
	respBody := resp.Body()

	var response RefreshTokenResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		utils.Log.Printf("Error unmarshaling AccessToken refresh response: %v", err)
		return err
	}

	// Update tokenData
	if response.AccessToken != "" {
		tokenData.AccessToken = response.AccessToken
		tokenData.LastTokenRefreshTime = strconv.FormatInt(time.Now().Unix(), 10)
		err := utils.WriteJIOTVCredentials(tokenData)
		if err != nil {
			utils.Log.Printf("Error saving refreshed credentials: %v", err)
			return err
		}
		TV = television.New(tokenData)
		utils.Log.Println("AccessToken refreshed successfully")
		return nil
	} else {
		err := fmt.Errorf("AccessToken not found in response")
		utils.Log.Printf("Error: %v", err)
		return err
	}
}

// LoginRefreshSSOToken Function is used to refresh SSOToken
func LoginRefreshSSOToken() error {
	utils.Log.Println("Refreshing SsoToken...")
	tokenData, err := utils.GetJIOTVCredentials()
	if err != nil {
		utils.Log.Printf("Error getting credentials for SSOToken refresh: %v", err)
		return err
	}

	// Validate that we have the required tokens
	if tokenData.SSOToken == "" {
		err := fmt.Errorf("SSOToken is empty, cannot refresh SSOToken")
		utils.Log.Printf("Error: %v", err)
		return err
	}
	if tokenData.UniqueID == "" {
		err := fmt.Errorf("UniqueID is empty, cannot refresh SSOToken")
		utils.Log.Printf("Error: %v", err)
		return err
	}

	deviceID := utils.GetDeviceID()
	if deviceID == "" {
		err := fmt.Errorf("DeviceID is empty, cannot refresh SSOToken")
		utils.Log.Printf("Error: %v", err)
		return err
	}

	// Prepare the request
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(REFRESH_SSO_TOKEN_URL)
	req.Header.SetMethod("GET")
	req.Header.Set(headers.DeviceType, headers.DeviceTypePhone)
	req.Header.Set(headers.VersionCode, headers.VersionCode389)
	req.Header.Set(headers.OS, headers.OSAndroid)
	req.Header.Set(headers.Host, urls.TVMediaDomain)
	req.Header.Set(headers.UserAgent, headers.UserAgentOkHttp)
	req.Header.Set("ssoToken", tokenData.SSOToken)
	req.Header.Set("uniqueid", tokenData.UniqueID)
	req.Header.Set("deviceid", deviceID)

	// Send the request
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	client := utils.GetRequestClient()
	if err := client.Do(req, resp); err != nil {
		utils.Log.Printf("HTTP request failed for SSOToken refresh: %v", err)
		return err
	}

	// Check the response
	if resp.StatusCode() != fasthttp.StatusOK {
		err := fmt.Errorf("SSOToken refresh failed with status code: %d, body: %s", resp.StatusCode(), string(resp.Body()))
		utils.Log.Printf("Error: %v", err)
		return err
	}

	// Parse the response body
	respBody := resp.Body()

	var response RefreshSSOTokenResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		utils.Log.Printf("Error unmarshaling SSOToken refresh response: %v", err)
		return err
	}

	// Update tokenData
	if response.SSOToken != "" {
		tokenData.SSOToken = response.SSOToken
		tokenData.LastSSOTokenRefreshTime = strconv.FormatInt(time.Now().Unix(), 10)
		err := utils.WriteJIOTVCredentials(tokenData)
		if err != nil {
			utils.Log.Printf("Error saving refreshed SSOToken credentials: %v", err)
			return err
		}
		TV = television.New(tokenData)
		utils.Log.Println("SSOToken refreshed successfully")
		return nil
	} else {
		err := fmt.Errorf("SSOToken not found in response")
		utils.Log.Printf("Error: %v", err)
		return err
	}
}

// RefreshTokenIfExpired Function is used to handle AccessToken refresh
// This function is now simplified for on-demand use only
func RefreshTokenIfExpired(credentials *utils.JIOTV_CREDENTIALS) error {
	utils.Log.Println("Checking if AccessToken is expired...")

	if IsAccessTokenExpired(credentials) {
		return LoginRefreshAccessToken()
	}

	utils.Log.Println("AccessToken is still valid")
	return nil
}

// RefreshSSOTokenIfExpired Function is used to handle SSOToken refresh
// This function is now simplified for on-demand use only
func RefreshSSOTokenIfExpired(credentials *utils.JIOTV_CREDENTIALS) error {
	utils.Log.Println("Checking if SSOToken is expired...")

	if IsSSOTokenExpired(credentials) {
		return LoginRefreshSSOToken()
	}

	utils.Log.Println("SSOToken is still valid")
	return nil
}
