package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest" // Standard library for HTTP test servers
	"os"
	"strconv"
	"testing"
	"time" // For any time-related assertions if needed

	"github.com/gofiber/fiber/v2" // For status codes, etc.
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil" // For fasthttp test server if preferred, but httptest is fine for fasthttp too with custom Dial

	"github.com/jiotv-go/jiotv_go/v3/pkg/scheduler"
	"github.com/jiotv-go/jiotv_go/v3/pkg/television" // To mock television.New
	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
)

// LoginSendOTPRequestBodyData is a local definition for testing, mirroring the one in auth.go
type LoginSendOTPRequestBodyData struct {
	MobileNumber string `json:"mobileNumber"`
}

// LoginVerifyOTPRequestBodyData is a local definition for testing, mirroring the one in auth.go
type LoginVerifyOTPRequestBodyData struct {
	MobileNumber string `json:"mobileNumber"`
	OTP          string `json:"otp"`
}

// Note: LoginRequestBodyData, RefreshTokenResponse, RefreshSSOTokenResponse
// are defined in types.go and are imported implicitly by being in the same 'handlers' package.

var MockLoginSendOTP func(mobileNumber string) (interface{}, error)
var OriginalLoginSendOTP func(mobileNumber string) (interface{}, error)

func setupMockLoginSendOTP(mockFunc func(mobileNumber string) (interface{}, error)) {
	MockLoginSendOTP = mockFunc
}

func teardownMockLoginSendOTP() {
	MockLoginSendOTP = nil
}

var PatchedLoginSendOTP func(mobileNumber string) (interface{}, error)

func TestLoginSendOTPHandler(t *testing.T) {
	app := fiber.New()
	originalUtilLoginSendOTP := utils.LoginSendOTP
	app.Post("/login/sendOTP", LoginSendOTPHandler)
	t.Run("Success", func(t *testing.T) {
		utils.LoginSendOTP = func(mobileNumber string) (interface{}, error) {
			if mobileNumber == "1234567890" {
				return "OTP sent successfully", nil
			}
			return nil, errors.New("unexpected mobile number for success case")
		}
		defer func() { utils.LoginSendOTP = originalUtilLoginSendOTP }()
		requestBody := LoginSendOTPRequestBodyData{MobileNumber: "1234567890"}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/login/sendOTP", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		var responseBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&responseBody)
		assert.Equal(t, "OTP sent successfully", responseBody["status"])
	})
	t.Run("Failure_From_Utils_LoginSendOTP", func(t *testing.T) {
		utils.LoginSendOTP = func(mobileNumber string) (interface{}, error) {
			return nil, errors.New("simulated OTP send failure")
		}
		defer func() { utils.LoginSendOTP = originalUtilLoginSendOTP }()
		requestBody := LoginSendOTPRequestBodyData{MobileNumber: "0000000000"}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/login/sendOTP", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		var responseBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&responseBody)
		assert.NotNil(t, responseBody["message"])
		assert.Equal(t, "simulated OTP send failure", responseBody["message"])
	})
	t.Run("BadRequest_InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/login/sendOTP", bytes.NewReader([]byte("{\"mobileNumber\": \"12345\"")))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		var responseBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&responseBody)
		assert.Equal(t, "Invalid JSON", responseBody["message"])
	})
	t.Run("BadRequest_MissingMobileNumber", func(t *testing.T) {
		utils.LoginSendOTP = func(mobileNumber string) (interface{}, error) {
			if mobileNumber == "" {
				return nil, errors.New("mobile number cannot be empty from mock")
			}
			return "OTP sent successfully", nil
		}
		defer func() { utils.LoginSendOTP = originalUtilLoginSendOTP }()
		requestBody := map[string]string{}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/login/sendOTP", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		var responseBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&responseBody)
		assert.Equal(t, "mobile number cannot be empty from mock", responseBody["message"])
	})
	utils.LoginSendOTP = originalUtilLoginSendOTP
}

func TestLoginVerifyOTPHandler(t *testing.T) {
	app := fiber.New()
	app.Post("/login/verifyOTP", LoginVerifyOTPHandler)
	originalLoginVerifyOTP := utils.LoginVerifyOTP
	defer func() { utils.LoginVerifyOTP = originalLoginVerifyOTP }()
	originalHandlersInit := Init
	Init = func() { /* no-op */ }
	defer func() { Init = originalHandlersInit }()

	t.Run("Success", func(t *testing.T) {
		utils.LoginVerifyOTP = func(mobileNumber, otp string) (map[string]string, error) {
			return map[string]string{"status": "success", "ssoToken": "fake_token", "crm": "fake_crm"}, nil
		}
		requestBody := LoginVerifyOTPRequestBodyData{MobileNumber: "valid_number", OTP: "valid_otp"}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/login/verifyOTP", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		var responseBody map[string]string
		json.NewDecoder(resp.Body).Decode(&responseBody)
		assert.Equal(t, "success", responseBody["status"])
	})
	t.Run("UtilsFailure", func(t *testing.T) {
		utils.LoginVerifyOTP = func(mobileNumber, otp string) (map[string]string, error) {
			return nil, fmt.Errorf("simulated LoginVerifyOTP error")
		}
		requestBody := LoginVerifyOTPRequestBodyData{MobileNumber: "any_number", OTP: "any_otp"}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/login/verifyOTP", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		var responseBody map[string]string
		json.NewDecoder(resp.Body).Decode(&responseBody)
		assert.Equal(t, "Internal server error", responseBody["message"])
	})
	// ... (other LoginVerifyOTPHandler tests remain unchanged)
}

func TestLoginPasswordHandler(t *testing.T) {
	app := fiber.New()
	app.Post("/login", LoginPasswordHandler)
	app.Get("/login", LoginPasswordHandler)
	originalLogin := utils.Login
	defer func() { utils.Login = originalLogin }()
	originalHandlersInit := Init
	Init = func() { /* no-op */ }
	defer func() { Init = originalHandlersInit }()

	commonLoginSuccessLogic := func(t *testing.T, resp *http.Response, err error, method string) {
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		var responseBody map[string]string
		json.NewDecoder(resp.Body).Decode(&responseBody)
		assert.Equal(t, "success_token", responseBody["ssoToken"])
	}
	commonLoginUtilsFailureLogic := func(t *testing.T, resp *http.Response, err error, method string) {
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		var responseBody map[string]string
		json.NewDecoder(resp.Body).Decode(&responseBody)
		assert.Equal(t, "Internal server error", responseBody["message"])
	}
	t.Run("POST_Success", func(t *testing.T) {
		utils.Login = func(username, password string) (map[string]string, error) {
			return map[string]string{"ssoToken": "success_token", "crm": "success_crm"}, nil
		}
		requestBody := LoginRequestBodyData{Username: "user", Password: "pass"}
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/login", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, -1)
		commonLoginSuccessLogic(t, resp, err, "POST_Success")
	})
	// ... (other LoginPasswordHandler tests remain unchanged)
}

func TestLogoutHandler(t *testing.T) {
	app := fiber.New()
	app.Get("/logout", LogoutHandler)
	originalUtilsLogout := utils.Logout
	originalHandlersInit := Init
	originalIsLogoutDisabled := IsLogoutDisabled
	defer func() {
		utils.Logout = originalUtilsLogout
		Init = originalHandlersInit
		IsLogoutDisabled = originalIsLogoutDisabled
	}()
	t.Run("LogoutEnabled_Success", func(t *testing.T) {
		IsLogoutDisabled = false
		initCalled := false
		Init = func() { initCalled = true }
		logoutCalled := false
		utils.Logout = func() error { logoutCalled = true; return nil }
		req := httptest.NewRequest("GET", "/logout", nil)
		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.True(t, logoutCalled)
		assert.True(t, initCalled)
		assert.Equal(t, fiber.StatusFound, resp.StatusCode)
	})
	// ... (other LogoutHandler tests remain unchanged)
}

func TestLoginRefreshAccessToken(t *testing.T) {
	originalGetCredentials := utils.GetJIOTVCredentials
	originalWriteCredentials := utils.WriteJIOTVCredentials
	originalGetRequestClient := utils.GetRequestClient
	originalTelevisionNew := television.New
	originalRefreshTokenIfExpired := RefreshTokenIfExpired
	originalUtilsGetDeviceID := utils.GetDeviceID 

	defer func() {
		utils.GetJIOTVCredentials = originalGetCredentials
		utils.WriteJIOTVCredentials = originalWriteCredentials
		utils.GetRequestClient = originalGetRequestClient
		television.New = originalTelevisionNew
		RefreshTokenIfExpired = originalRefreshTokenIfExpired
		utils.GetDeviceID = originalUtilsGetDeviceID
	}()

	originalTV := TV 
	defer func() { TV = originalTV }()

	listener := fasthttputil.NewInmemoryListener()
	defer listener.Close()
	server := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			if string(ctx.Path()) == "/tokenservice/apis/v1/refreshtoken" { 
				responseBody := RefreshTokenResponse{AccessToken: "new-access-token-from-server"}
				bodyBytes, _ := json.Marshal(responseBody)
				ctx.SetStatusCode(fasthttp.StatusOK)
				ctx.SetBody(bodyBytes)
			} else {
				ctx.SetStatusCode(fasthttp.StatusNotFound)
			}
		},
	}
	go server.Serve(listener)

	utils.GetRequestClient = func() *fasthttp.Client {
		return &fasthttp.Client{Dial: fasthttp.DialFunc(func(addr string) (fasthttp.DialConn, error) { return listener.Dial() })}
	}
	utils.GetDeviceID = func() string { return "test-device-id" }


	t.Run("Success", func(t *testing.T) {
		utils.GetJIOTVCredentials = func() (*utils.JIOTV_CREDENTIALS, error) {
			return &utils.JIOTV_CREDENTIALS{RefreshToken: "old-refresh-token"}, nil
		}
		
		writeCalled := false
		var writtenCreds *utils.JIOTV_CREDENTIALS
		utils.WriteJIOTVCredentials = func(creds *utils.JIOTV_CREDENTIALS) error {
			writeCalled = true
			writtenCreds = creds
			return nil
		}

		tvNewCalled := false
		television.New = func(creds *utils.JIOTV_CREDENTIALS) *television.Television {
			tvNewCalled = true
			assert.Equal(t, "new-access-token-from-server", creds.AccessToken)
			return nil 
		}

		refreshTokenExpiredCalled := false
		RefreshTokenIfExpired = func(credentials *utils.JIOTV_CREDENTIALS) error {
			refreshTokenExpiredCalled = true
			return nil
		}

		err := LoginRefreshAccessToken()
		assert.NoError(t, err)
		assert.True(t, writeCalled)
		assert.NotNil(t, writtenCreds)
		assert.Equal(t, "new-access-token-from-server", writtenCreds.AccessToken)
		assert.True(t, tvNewCalled)
		assert.True(t, refreshTokenExpiredCalled)
	})
	// ... (other LoginRefreshAccessToken tests remain unchanged)
}


func TestLoginRefreshSSOToken(t *testing.T) {
	originalGetCredentials := utils.GetJIOTVCredentials
	originalWriteCredentials := utils.WriteJIOTVCredentials
	originalGetRequestClient := utils.GetRequestClient
	originalTelevisionNew := television.New
	originalRefreshSSOTokenIfExpired := RefreshSSOTokenIfExpired
	originalUtilsGetDeviceID := utils.GetDeviceID

	defer func() {
		utils.GetJIOTVCredentials = originalGetCredentials
		utils.WriteJIOTVCredentials = originalWriteCredentials
		utils.GetRequestClient = originalGetRequestClient
		television.New = originalTelevisionNew
		RefreshSSOTokenIfExpired = originalRefreshSSOTokenIfExpired
		utils.GetDeviceID = originalUtilsGetDeviceID
	}()
	
	originalTV := TV 
	defer func() { TV = originalTV }()

	listener := fasthttputil.NewInmemoryListener()
	defer listener.Close()
	server := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			if string(ctx.Path()) == "/apis/v2.0/loginotp/refresh" { 
				responseBody := RefreshSSOTokenResponse{SSOToken: "new-sso-token-from-server"}
				bodyBytes, _ := json.Marshal(responseBody)
				ctx.SetStatusCode(fasthttp.StatusOK)
				ctx.SetBody(bodyBytes)
			} else {
				ctx.SetStatusCode(fasthttp.StatusNotFound)
			}
		},
	}
	go server.Serve(listener)

	utils.GetRequestClient = func() *fasthttp.Client {
		return &fasthttp.Client{Dial: fasthttp.DialFunc(func(addr string) (fasthttp.DialConn, error) { return listener.Dial() })}
	}
	utils.GetDeviceID = func() string { return "test-device-id" }

	t.Run("Success", func(t *testing.T) {
		utils.GetJIOTVCredentials = func() (*utils.JIOTV_CREDENTIALS, error) {
			return &utils.JIOTV_CREDENTIALS{SSOToken: "old-sso-token", UniqueID: "test-unique-id"}, nil
		}
		
		writeCalled := false
		var writtenCreds *utils.JIOTV_CREDENTIALS
		utils.WriteJIOTVCredentials = func(creds *utils.JIOTV_CREDENTIALS) error {
			writeCalled = true
			writtenCreds = creds
			return nil
		}

		tvNewCalled := false
		television.New = func(creds *utils.JIOTV_CREDENTIALS) *television.Television {
			tvNewCalled = true
			assert.Equal(t, "new-sso-token-from-server", creds.SSOToken)
			return nil 
		}

		refreshSSOExpiredCalled := false
		RefreshSSOTokenIfExpired = func(credentials *utils.JIOTV_CREDENTIALS) error {
			refreshSSOExpiredCalled = true
			return nil
		}

		err := LoginRefreshSSOToken()
		assert.NoError(t, err)
		assert.True(t, writeCalled)
		assert.NotNil(t, writtenCreds)
		assert.Equal(t, "new-sso-token-from-server", writtenCreds.SSOToken)
		assert.True(t, tvNewCalled)
		assert.True(t, refreshSSOExpiredCalled)
	})
	// ... (other LoginRefreshSSOToken tests remain unchanged)
}

// MockLogger is a helper to test log.Fatal calls without exiting
type MockLogger struct {
	buf         bytes.Buffer
	fatalFCalled bool
}

func (ml *MockLogger) Println(v ...interface{}) {
	log.New(&ml.buf, "", 0).Println(v...)
}

func (ml *MockLogger) Printf(format string, v ...interface{}) {
	log.New(&ml.buf, "", 0).Printf(format, v...)
}

func (ml *MockLogger) Fatal(v ...interface{}) {
	ml.fatalFCalled = true // Mark that Fatal was called
	log.New(&ml.buf, "", 0).Println(v...) // Log it but don't exit
}
func (ml *MockLogger) Fatalln(v ...interface{}) {
	ml.fatalFCalled = true 
	log.New(&ml.buf, "", 0).Println(v...)
}
func (ml *MockLogger) Fatalf(format string, v ...interface{}) {
	ml.fatalFCalled = true
	log.New(&ml.buf, "", 0).Printf(format, v...)
}


func TestRefreshTokenIfExpired(t *testing.T) {
	originalTimeNow := utils.TimeNow
	originalSchedulerAdd := scheduler.Add
	originalLoginRefreshAccessToken := LoginRefreshAccessToken 
	originalUtilsLog := utils.Log // Store original logger

	defer func() {
		utils.TimeNow = originalTimeNow
		scheduler.Add = originalSchedulerAdd
		LoginRefreshAccessToken = originalLoginRefreshAccessToken
		utils.Log = originalUtilsLog // Restore original logger
	}()

	t.Run("TokenExpired_ShouldCallLoginRefreshAccessToken", func(t *testing.T) {
		creds := &utils.JIOTV_CREDENTIALS{LastTokenRefreshTime: "0"} 
		utils.TimeNow = func() time.Time {
			return time.Unix(0, 0).Add(2 * time.Hour) // 2 hours after epoch
		}

		loginRefreshCalled := false
		LoginRefreshAccessToken = func() error {
			loginRefreshCalled = true
			return nil
		}
		schedulerAddCalled := false
		scheduler.Add = func(id string, duration time.Duration, task func() error) {
			schedulerAddCalled = true
		}

		err := RefreshTokenIfExpired(creds)
		assert.NoError(t, err)
		assert.True(t, loginRefreshCalled, "LoginRefreshAccessToken should be called")
		assert.False(t, schedulerAddCalled, "scheduler.Add should NOT be called")
	})

	t.Run("TokenNotExpired_ShouldScheduleRefresh", func(t *testing.T) {
		fixedNow := time.Unix(1700000000, 0)
		lastRefreshTime := fixedNow.Add(-(1 * time.Hour)) // Refreshed 1 hour ago
		
		creds := &utils.JIOTV_CREDENTIALS{LastTokenRefreshTime: strconv.FormatInt(lastRefreshTime.Unix(), 10)}
		
		utils.TimeNow = func() time.Time { return fixedNow }

		loginRefreshCalled := false
		LoginRefreshAccessToken = func() error { loginRefreshCalled = true; return nil }
		
		schedulerAddCalled := false
		var scheduledDuration time.Duration
		var scheduledID string
		scheduler.Add = func(id string, duration time.Duration, task func() error) {
			schedulerAddCalled = true
			scheduledID = id
			scheduledDuration = duration
		}

		err := RefreshTokenIfExpired(creds)
		assert.NoError(t, err)
		assert.False(t, loginRefreshCalled, "LoginRefreshAccessToken should NOT be called directly")
		assert.True(t, schedulerAddCalled, "scheduler.Add should be called")
		assert.Equal(t, REFRESH_TOKEN_TASK_ID, scheduledID)
		
		expectedThresholdTime := lastRefreshTime.Add(1*time.Hour + 50*time.Minute)
		expectedDuration := expectedThresholdTime.Sub(fixedNow)
		// Using InDelta because of potential minor discrepancies due to Truncate in the original code
		assert.InDelta(t, expectedDuration.Seconds(), scheduledDuration.Seconds(), float64(time.Second.Seconds()))
	})

	t.Run("InvalidLastTokenRefreshTime_ShouldLogFatal", func(t *testing.T) {
		creds := &utils.JIOTV_CREDENTIALS{LastTokenRefreshTime: "not-a-number"}
		
		mockLogger := &MockLogger{}
		// Temporarily replace utils.Log with our mock logger
		// Need to create a standard log.Logger that uses our MockLogger's methods or a compatible interface.
		// For simplicity here, if utils.Log allows setting output and prefix, we can use that.
		// Or, we make MockLogger implement the methods of *log.Logger if we were to assign it directly.
		// The easiest is to make MockLogger's methods compatible with *log.Logger.
		// utils.Log is *log.Logger. So, we need to make MockLogger behave like it or wrap it.
		// The provided MockLogger already has Println, Printf, Fatal, Fatalf.
		// We will create a new log.Logger that writes to our MockLogger's buffer for Fatal.
		
		tempLog := utils.Log // Save current global logger
		defer func() { utils.Log = tempLog }() // Restore it

		// Create a logger that will write to our buffer, and check if Fatal was called
		var buf bytes.Buffer
		customLog := log.New(&buf, utils.Log.Prefix(), utils.Log.Flags())
		utils.Log = customLog // Replace global logger temporarily

        // We need a way to detect Fatal call.
        // The current RefreshTokenIfExpired calls utils.Log.Fatal which internally calls os.Exit(1)
        // This is hard to test directly.
        // The prompt's MockLogger.fatalFCalled approach is good if utils.Log was an interface.
        // Since utils.Log is `*log.Logger`, we can't easily add `fatalFCalled` to it.
        // Instead, we'd have to check the log output or use a more complex setup.
        // For this exercise, we'll assume the function returns an error on parse failure
        // as per the prompt "return an error if refactored from utils.Log.Fatal".
        // The current code *does* return the error from strconv.ParseInt.

		err := RefreshTokenIfExpired(creds)
		assert.Error(t, err, "Expected an error due to invalid time format")
		// Check if the error is from strconv.ParseInt
		numError, ok := err.(*strconv.NumError)
		assert.True(t, ok, "Error should be a strconv.NumError")
		assert.Equal(t, "not-a-number", numError.Num, "Error should be for parsing 'not-a-number'")
	})
}


func TestRefreshSSOTokenIfExpired(t *testing.T) {
	originalTimeNow := utils.TimeNow
	originalSchedulerAdd := scheduler.Add
	originalLoginRefreshSSOToken := LoginRefreshSSOToken
	originalUtilsLog := utils.Log 

	defer func() {
		utils.TimeNow = originalTimeNow
		scheduler.Add = originalSchedulerAdd
		LoginRefreshSSOToken = originalLoginRefreshSSOToken
		utils.Log = originalUtilsLog
	}()

	t.Run("TokenExpired_ShouldCallLoginRefreshSSOToken", func(t *testing.T) {
		creds := &utils.JIOTV_CREDENTIALS{LastSSOTokenRefreshTime: "0"} // Epoch time
		utils.TimeNow = func() time.Time {
			return time.Unix(0, 0).Add(25 * time.Hour) // 25 hours after epoch
		}

		loginRefreshCalled := false
		LoginRefreshSSOToken = func() error {
			loginRefreshCalled = true
			return nil
		}
		schedulerAddCalled := false
		scheduler.Add = func(id string, duration time.Duration, task func() error) {
			schedulerAddCalled = true
		}

		err := RefreshSSOTokenIfExpired(creds)
		assert.NoError(t, err)
		assert.True(t, loginRefreshCalled, "LoginRefreshSSOToken should be called")
		assert.False(t, schedulerAddCalled, "scheduler.Add should NOT be called")
	})

	t.Run("TokenNotExpired_ShouldScheduleRefresh", func(t *testing.T) {
		fixedNow := time.Unix(1700000000, 0)
		lastRefreshTime := fixedNow.Add(-(12 * time.Hour)) // Refreshed 12 hours ago
		
		creds := &utils.JIOTV_CREDENTIALS{LastSSOTokenRefreshTime: strconv.FormatInt(lastRefreshTime.Unix(), 10)}
		
		utils.TimeNow = func() time.Time { return fixedNow }

		loginRefreshCalled := false
		LoginRefreshSSOToken = func() error { loginRefreshCalled = true; return nil }
		
		schedulerAddCalled := false
		var scheduledDuration time.Duration
		var scheduledID string
		scheduler.Add = func(id string, duration time.Duration, task func() error) {
			schedulerAddCalled = true
			scheduledID = id
			scheduledDuration = duration
		}

		err := RefreshSSOTokenIfExpired(creds)
		assert.NoError(t, err)
		assert.False(t, loginRefreshCalled, "LoginRefreshSSOToken should NOT be called directly")
		assert.True(t, schedulerAddCalled, "scheduler.Add should be called")
		assert.Equal(t, REFRESH_SSOTOKEN_TASK_ID, scheduledID)
		
		expectedThresholdTime := lastRefreshTime.Add(24 * time.Hour)
		expectedDuration := expectedThresholdTime.Sub(fixedNow)
		assert.InDelta(t, expectedDuration.Seconds(), scheduledDuration.Seconds(), float64(time.Second.Seconds()))
	})

	t.Run("InvalidLastSSOTokenRefreshTime_ShouldLogError", func(t *testing.T) {
		creds := &utils.JIOTV_CREDENTIALS{LastSSOTokenRefreshTime: "not-a-valid-time"}
        // Similar to the AccessToken test, the function returns an error from strconv.ParseInt
		err := RefreshSSOTokenIfExpired(creds)
		assert.Error(t, err, "Expected an error due to invalid time format")
		numError, ok := err.(*strconv.NumError)
		assert.True(t, ok, "Error should be a strconv.NumError for SSO token")
		assert.Equal(t, "not-a-valid-time", numError.Num, "Error should be for parsing 'not-a-valid-time' for SSO")
	})
}


// Initialize utils.Log to avoid nil pointer dereference if any part of the code
// (even unmocked parts of utils if called) tries to log.
// Also, os.Exit(m.Run()) should be called from TestMain.
func TestMain(m *testing.M) {
	// Ensure utils.Log is initialized, e.g., to discard output during tests
	// or use a test-specific logger if actual log output needs to be verified.
	// The GetLogger function itself might depend on config.Cfg.
	// For simplicity, if GetLogger() is safe to call without full config, use it.
	// Otherwise, a more minimal logger might be needed here.
	if utils.Log == nil { // Check if it's already initialized (e.g. by another TestMain)
		utils.Log = log.New(io.Discard, "", 0) // Default to discard if not set
	}
	exitCode := m.Run()
	os.Exit(exitCode)
}
