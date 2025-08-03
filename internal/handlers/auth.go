package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jiotv-go/jiotv_go/v3/pkg/scheduler"
	"github.com/jiotv-go/jiotv_go/v3/pkg/television"
	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
	"github.com/valyala/fasthttp"
)

const (
	REFRESH_TOKEN_TASK_ID    = "jiotv_refresh_token"
	REFRESH_SSOTOKEN_TASK_ID = "jiotv_refresh_sso_token"
	HEALTH_CHECK_TASK_ID     = "jiotv_token_health_check"
)

// LoginSendOTPHandler sends OTP for login
func LoginSendOTPHandler(c *fiber.Ctx) error {
	// get mobile number from post request
	formBody := new(LoginSendOTPRequestBodyData)
	err := c.BodyParser(&formBody)
	if err != nil {
		utils.Log.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid JSON",
		})
	}
	mobileNumber := formBody.MobileNumber
	checkFieldExist("Mobile Number", mobileNumber != "", c)

	result, err := utils.LoginSendOTP(mobileNumber)
	if err != nil {
		utils.Log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err,
		})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid JSON",
		})
	}
	mobileNumber := formBody.MobileNumber
	checkFieldExist("Mobile Number", mobileNumber != "", c)
	otp := formBody.OTP
	checkFieldExist("OTP", otp != "", c)

	result, err := utils.LoginVerifyOTP(mobileNumber, otp)
	if err != nil {
		utils.Log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}
	Init()
	return c.JSON(result)
}

// LoginPasswordHandler is used to login with password
func LoginPasswordHandler(c *fiber.Ctx) error {
	var username, password string
	if c.Method() == "GET" {
		username = c.Query("username")
		checkFieldExist("Username", username != "", c)
		password = c.Query("password")
		checkFieldExist("Password", password != "", c)
	} else if c.Method() == "POST" {
		formBody := new(LoginRequestBodyData)
		err := c.BodyParser(&formBody)
		if err != nil {
			utils.Log.Println(err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid JSON",
			})
		}
		username = formBody.Username
		checkFieldExist("Username", username != "", c)
		password = formBody.Password
		checkFieldExist("Password", password != "", c)
	}

	result, err := utils.Login(username, password)
	if err != nil {
		utils.Log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
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
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Internal server error",
			})
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
	req.Header.Set("devicetype", "phone")
	req.Header.Set("versionCode", "315")
	req.Header.Set("os", "android")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Host", "auth.media.jio.com")
	req.Header.Set("User-Agent", "okhttp/4.2.2")
	req.Header.Set("accessToken", tokenData.AccessToken)
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
		
		// Schedule next refresh based on the new refresh time
		go RefreshTokenIfExpired(tokenData)
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
	req.Header.Set("devicetype", "phone")
	req.Header.Set("versionCode", "315")
	req.Header.Set("os", "android")
	req.Header.Set("Host", "tv.media.jio.com")
	req.Header.Set("User-Agent", "okhttp/4.2.2")
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
		
		// Schedule next refresh based on the new refresh time
		go RefreshSSOTokenIfExpired(tokenData)
		return nil
	} else {
		err := fmt.Errorf("SSOToken not found in response")
		utils.Log.Printf("Error: %v", err)
		return err
	}
}

// RefreshTokenIfExpired Function is used to handle AccessToken refresh
func RefreshTokenIfExpired(credentials *utils.JIOTV_CREDENTIALS) error {
	utils.Log.Println("Checking if AccessToken is expired...")
	lastTokenRefreshTime, err := strconv.ParseInt(credentials.LastTokenRefreshTime, 10, 64)
	if err != nil {
		utils.Log.Printf("Error parsing LastTokenRefreshTime: %v. Scheduling refresh in 10 minutes.", err)
		// Schedule refresh in 10 minutes if we can't parse the time
		go scheduler.Add(REFRESH_TOKEN_TASK_ID, 10*time.Minute, func() error {
			freshCreds, err := utils.GetJIOTVCredentials()
			if err != nil {
				utils.Log.Printf("Error getting fresh credentials for scheduled refresh: %v", err)
				return err
			}
			return RefreshTokenIfExpired(freshCreds)
		})
		return err
	}
	lastTokenRefreshTimeUnix := time.Unix(lastTokenRefreshTime, 0)
	thresholdTime := lastTokenRefreshTimeUnix.Add(1*time.Hour + 50*time.Minute)

	if thresholdTime.Before(time.Now()) {
		err := LoginRefreshAccessToken()
		if err != nil {
			utils.Log.Printf("AccessToken refresh failed: %v. Retrying in 5 minutes.", err)
			// Retry in 5 minutes if refresh failed
			go scheduler.Add(REFRESH_TOKEN_TASK_ID, 5*time.Minute, func() error {
				// Get fresh credentials in case they were updated
				freshCreds, err := utils.GetJIOTVCredentials()
				if err != nil {
					utils.Log.Printf("Error getting fresh credentials for scheduled retry: %v", err)
					return err
				}
				return RefreshTokenIfExpired(freshCreds)
			})
		}
	} else {
		utils.Log.Println("Refreshing AccessToken after", time.Until(thresholdTime).Truncate(time.Second))
		go scheduler.Add(REFRESH_TOKEN_TASK_ID, time.Until(thresholdTime), func() error {
			// Get fresh credentials in case they were updated
			freshCreds, err := utils.GetJIOTVCredentials()
			if err != nil {
				utils.Log.Printf("Error getting fresh credentials for scheduled refresh: %v", err)
				return err
			}
			return RefreshTokenIfExpired(freshCreds)
		})
	}
	return nil
}

// RefreshSSOTokenIfExpired Function is used to handle SSOToken refresh
func RefreshSSOTokenIfExpired(credentials *utils.JIOTV_CREDENTIALS) error {
	utils.Log.Println("Checking if SSOToken is expired...")
	lastTokenRefreshTime, err := strconv.ParseInt(credentials.LastSSOTokenRefreshTime, 10, 64)
	if err != nil {
		utils.Log.Printf("Error parsing LastSSOTokenRefreshTime: %v. Scheduling refresh in 1 hour.", err)
		// Schedule refresh in 1 hour if we can't parse the time
		go scheduler.Add(REFRESH_SSOTOKEN_TASK_ID, 1*time.Hour, func() error {
			// Get fresh credentials in case they were updated
			freshCreds, err := utils.GetJIOTVCredentials()
			if err != nil {
				utils.Log.Printf("Error getting fresh credentials for scheduled refresh: %v", err)
				return err
			}
			return RefreshSSOTokenIfExpired(freshCreds)
		})
		return err
	}
	lastTokenRefreshTimeUnix := time.Unix(lastTokenRefreshTime, 0)
	thresholdTime := lastTokenRefreshTimeUnix.Add(24 * time.Hour)

	if thresholdTime.Before(time.Now()) {
		err := LoginRefreshSSOToken()
		if err != nil {
			utils.Log.Printf("SSOToken refresh failed: %v. Retrying in 30 minutes.", err)
			// Retry in 30 minutes if refresh failed
			go scheduler.Add(REFRESH_SSOTOKEN_TASK_ID, 30*time.Minute, func() error {
				// Get fresh credentials in case they were updated
				freshCreds, err := utils.GetJIOTVCredentials()
				if err != nil {
					utils.Log.Printf("Error getting fresh credentials for scheduled refresh: %v", err)
					return err
				}
				return RefreshSSOTokenIfExpired(freshCreds)
			})
		}
	} else {
		utils.Log.Println("Refreshing SSOToken after", time.Until(thresholdTime).Truncate(time.Second))
		go scheduler.Add(REFRESH_SSOTOKEN_TASK_ID, time.Until(thresholdTime), func() error {
			// Get fresh credentials in case they were updated
			freshCreds, err := utils.GetJIOTVCredentials()
			if err != nil {
				utils.Log.Printf("Error getting fresh credentials for scheduled refresh: %v", err)
				return err
			}
			return RefreshSSOTokenIfExpired(freshCreds)
		})
	}
	return nil
}

// LoginRefreshAccessTokenSync synchronously refreshes AccessToken without scheduling new tasks
func LoginRefreshAccessTokenSync() error {
	utils.Log.Println("Synchronously refreshing AccessToken...")
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
	req.Header.Set("devicetype", "phone")
	req.Header.Set("versionCode", "315")
	req.Header.Set("os", "android")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Host", "auth.media.jio.com")
	req.Header.Set("User-Agent", "okhttp/4.2.2")
	req.Header.Set("accessToken", tokenData.AccessToken)
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

// LoginRefreshSSOTokenSync synchronously refreshes SSOToken without scheduling new tasks
func LoginRefreshSSOTokenSync() error {
	utils.Log.Println("Synchronously refreshing SsoToken...")
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
	req.Header.Set("devicetype", "phone")
	req.Header.Set("versionCode", "315")
	req.Header.Set("os", "android")
	req.Header.Set("Host", "tv.media.jio.com")
	req.Header.Set("User-Agent", "okhttp/4.2.2")
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

// TokenHealthCheck verifies that token refresh tasks are properly scheduled and tokens are valid
func TokenHealthCheck() error {
	utils.Log.Println("Running token health check...")
	
	credentials, err := utils.GetJIOTVCredentials()
	if err != nil {
		utils.Log.Printf("Health check: No credentials found: %v", err)
		// Schedule next health check
		go scheduler.Add(HEALTH_CHECK_TASK_ID, 1*time.Hour, TokenHealthCheck)
		return nil // Don't return error when no credentials are found - this is expected behavior
	}

	var refreshedTokens bool

	// Check AccessToken health and refresh synchronously if needed
	if credentials.AccessToken != "" && credentials.RefreshToken != "" {
		if credentials.LastTokenRefreshTime != "" {
			lastTime, err := strconv.ParseInt(credentials.LastTokenRefreshTime, 10, 64)
			if err == nil {
				lastRefresh := time.Unix(lastTime, 0)
				// If token was last refreshed more than 3 hours ago, it might be stale
				if time.Since(lastRefresh) > 3*time.Hour {
					utils.Log.Printf("Health check: AccessToken may be stale (last refresh: %v). Refreshing synchronously.", lastRefresh)
					if err := LoginRefreshAccessTokenSync(); err != nil {
						utils.Log.Printf("Health check: Failed to refresh AccessToken: %v", err)
					} else {
						refreshedTokens = true
						// Re-fetch credentials after refresh
						credentials, _ = utils.GetJIOTVCredentials()
					}
				}
			}
		}
	}

	// Check SSOToken health and refresh synchronously if needed
	if credentials.SSOToken != "" && credentials.UniqueID != "" {
		if credentials.LastSSOTokenRefreshTime != "" {
			lastTime, err := strconv.ParseInt(credentials.LastSSOTokenRefreshTime, 10, 64)
			if err == nil {
				lastRefresh := time.Unix(lastTime, 0)
				// If token was last refreshed more than 26 hours ago, it might be stale
				if time.Since(lastRefresh) > 26*time.Hour {
					utils.Log.Printf("Health check: SSOToken may be stale (last refresh: %v). Refreshing synchronously.", lastRefresh)
					if err := LoginRefreshSSOTokenSync(); err != nil {
						utils.Log.Printf("Health check: Failed to refresh SSOToken: %v", err)
					} else {
						refreshedTokens = true
					}
				}
			}
		}
	}

	if refreshedTokens {
		utils.Log.Println("Health check: Tokens refreshed synchronously")
	} else {
		utils.Log.Println("Health check: All tokens appear healthy")
	}

	// Schedule next health check in 2 hours
	go scheduler.Add(HEALTH_CHECK_TASK_ID, 2*time.Hour, TokenHealthCheck)
	return nil
}
