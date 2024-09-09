package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rabilrbl/jiotv_go/v3/pkg/scheduler"
	"github.com/rabilrbl/jiotv_go/v3/pkg/television"
	"github.com/rabilrbl/jiotv_go/v3/pkg/utils"
	"github.com/valyala/fasthttp"
)

const (
	REFRESH_TOKEN_TASK_ID    = "jiotv_refresh_token"
	REFRESH_SSOTOKEN_TASK_ID = "jiotv_refresh_sso_token"
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
		utils.Log.Fatalln(err)
		return err
	}

	// Prepare the request
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(REFRESH_TOKEN_URL)
	req.Header.SetMethod("POST")
	req.Header.Set("devicetype", "phone")
	req.Header.Set("versionCode", "315")
	req.Header.Set("os", "android")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Host", "auth.media.jio.com")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("User-Agent", "okhttp/4.2.2")
	req.Header.Set("accessToken", tokenData.AccessToken)
	req.SetBody(requestBodyJSON)

	// Send the request
	resp := fasthttp.AcquireResponse()
	client := utils.GetRequestClient()
	if err := client.Do(req, resp); err != nil {
		utils.Log.Fatalln(err)
		return err
	}

	// Check the response
	if resp.StatusCode() != fasthttp.StatusOK {
		utils.Log.Fatalln("Request failed with status code:", resp.StatusCode())
		return fmt.Errorf("Request failed with status code: %d", resp.StatusCode())
	}

	// Parse the response body
	respBody, err := resp.BodyGunzip()
	if err != nil {
		utils.Log.Fatalln(err)
		return err
	}
	var response RefreshTokenResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		utils.Log.Fatalln(err)
		return err
	}

	// Update tokenData
	if response.AccessToken != "" {
		tokenData.AccessToken = response.AccessToken
		tokenData.LastTokenRefreshTime = strconv.FormatInt(time.Now().Unix(), 10)
		err := utils.WriteJIOTVCredentials(tokenData)
		if err != nil {
			utils.Log.Fatalln(err)
			return err
		}
		TV = television.New(tokenData)
		go RefreshTokenIfExpired(tokenData)
		return nil
	} else {
		return fmt.Errorf("AccessToken not found in response")
	}
}

// LoginRefreshSSOToken Function is used to refresh SSOToken
func LoginRefreshSSOToken() error {
	utils.Log.Println("Refreshing SsoToken...")
	tokenData, err := utils.GetJIOTVCredentials()
	if err != nil {
		return err
	}

	// Prepare the request
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(REFRESH_SSO_TOKEN_URL)
	req.Header.SetMethod("GET")
	req.Header.Set("devicetype", "phone")
	req.Header.Set("versionCode", "315")
	req.Header.Set("os", "android")
	req.Header.Set("Host", "tv.media.jio.com")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("User-Agent", "okhttp/4.2.2")
	req.Header.Set("ssoToken", tokenData.SSOToken)
	req.Header.Set("uniqueid", tokenData.UniqueID)
	req.Header.Set("deviceid",utils.GetDeviceID())

	// Send the request
	resp := fasthttp.AcquireResponse()
	client := utils.GetRequestClient()
	if err := client.Do(req, resp); err != nil {
		utils.Log.Fatalln(err)
		return err
	}

	// Check the response
	if resp.StatusCode() != fasthttp.StatusOK {
		utils.Log.Fatalln("Request failed with status code:", resp.StatusCode())
		return fmt.Errorf("Request failed with status code: %d", resp.StatusCode())
	}

	// Parse the response body
	respBody, err := resp.BodyGunzip()
	if err != nil {
		utils.Log.Fatalln(err)
		return err
	}
	var response RefreshSSOTokenResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		utils.Log.Fatalln(err)
		return err
	}

	// Update tokenData
	if response.SSOToken != "" {
		tokenData.SSOToken = response.SSOToken
		tokenData.LastSSOTokenRefreshTime = strconv.FormatInt(time.Now().Unix(), 10)
		err := utils.WriteJIOTVCredentials(tokenData)
		if err != nil {
			utils.Log.Fatalln(err)
			return err
		}
		TV = television.New(tokenData)
		go RefreshSSOTokenIfExpired(tokenData)
		return nil
	} else {
		return fmt.Errorf("SSOToken not found in response")
	}
}

// RefreshTokenIfExpired Function is used to handle AccessToken refresh
func RefreshTokenIfExpired(credentials *utils.JIOTV_CREDENTIALS) error {
	utils.Log.Println("Checking if AccessToken is expired...")
	lastTokenRefreshTime, err := strconv.ParseInt(credentials.LastTokenRefreshTime, 10, 64)
	if err != nil {
		utils.Log.Fatal(err)
		return err
	}
	lastTokenRefreshTimeUnix := time.Unix(lastTokenRefreshTime, 0)
	thresholdTime := lastTokenRefreshTimeUnix.Add(1*time.Hour + 50*time.Minute)

	if thresholdTime.Before(time.Now()) {
		LoginRefreshAccessToken()
	} else {
		utils.Log.Println("Refreshing AccessToken after", time.Until(thresholdTime).Truncate(time.Second))
		go scheduler.Add(REFRESH_TOKEN_TASK_ID, time.Until(thresholdTime), func() error {
			return RefreshTokenIfExpired(credentials)
		})
	}
	return nil
}

// RefreshSSOTokenIfExpired Function is used to handle SSOToken refresh
func RefreshSSOTokenIfExpired(credentials *utils.JIOTV_CREDENTIALS) error {
	utils.Log.Println("Checking if SSOToken is expired...")
	lastTokenRefreshTime, err := strconv.ParseInt(credentials.LastSSOTokenRefreshTime, 10, 64)
	if err != nil {
		utils.Log.Fatal(err)
		return err
	}
	lastTokenRefreshTimeUnix := time.Unix(lastTokenRefreshTime, 0)
	thresholdTime := lastTokenRefreshTimeUnix.Add(24 * time.Hour)

	if thresholdTime.Before(time.Now()) {
		LoginRefreshSSOToken()
	} else {
		utils.Log.Println("Refreshing SSOToken after", time.Until(thresholdTime).Truncate(time.Second))
		go scheduler.Add(REFRESH_SSOTOKEN_TASK_ID, time.Until(thresholdTime), func() error {
			return RefreshSSOTokenIfExpired(credentials)
		})
	}
	return nil
}
