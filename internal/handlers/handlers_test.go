package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"github.com/jiotv-go/jiotv_go/v3/internal/config"
	"github.com/jiotv-go/jiotv_go/v3/pkg/television"
	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
)

func TestErrorMessageHandler(t *testing.T) {
	app := fiber.New()
	app.Get("/error", func(c *fiber.Ctx) error {
		return ErrorMessageHandler(c, errors.New("test error occurred"))
	})

	req := httptest.NewRequest("GET", "/error", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	var body map[string]string
	err = json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, err)
	assert.Equal(t, "test error occurred", body["message"])
	resp.Body.Close()

	// Test with nil error (should not write to response or change status)
	app.Get("/no-error", func(c *fiber.Ctx) error {
		err := ErrorMessageHandler(c, nil) // Call with nil error
		if err != nil {                     // If ErrorMessageHandler returns an error itself
			return c.Status(fiber.StatusTeapot).SendString(err.Error())
		}
		// If ErrorMessageHandler returns nil, it means it didn't find an error to handle.
		// The response should be whatever the default is (e.g. 200 OK if nothing else is done).
		return c.SendStatus(fiber.StatusOK)
	})

	reqNoError := httptest.NewRequest("GET", "/no-error", nil)
	respNoError, errNoError := app.Test(reqNoError)
	assert.NoError(t, errNoError)
	assert.Equal(t, fiber.StatusOK, respNoError.StatusCode) // Default or explicitly set OK
	respNoError.Body.Close()
}

func TestCheckFieldExist(t *testing.T) {
	app := fiber.New()
	app.Get("/check", func(c *fiber.Ctx) error {
		fieldPresent := c.Query("fieldPresent")
		var check bool
		if fieldPresent == "true" {
			check = true
		} else {
			check = false
		}
		// checkFieldExist is not designed to be a terminal handler.
		// It returns an error if the field check fails, which the calling handler should then use.
		// Or, it writes to the context and the calling handler should not proceed.
		// The actual checkFieldExist writes to context and returns nil.
		// We need to see if it also halts execution or if Fiber allows multiple writes.
		// For testing, we will assume the calling handler stops if checkFieldExist writes.
		err := checkFieldExist("testField", check, c)
		if err != nil { // This implies checkFieldExist would return an error to stop.
			return err   // The current checkFieldExist in handlers.go doesn't return error, it writes directly.
		}
		// If checkFieldExist writes to the response and returns nil, we need to check if the response was written.
		// This is tricky. Let's assume the test setup means if checkFieldExist "handles" it, the response is set.
		if !check { // If field was missing, checkFieldExist should have sent the response.
			return nil // Stop further processing in this test handler
		}
		return c.SendString("Field was present")
	})

	t.Run("FieldPresent", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/check?fieldPresent=true", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode) // Or whatever the handler does after check
		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, "Field was present", string(body))
		resp.Body.Close()
	})

	t.Run("FieldMissing", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/check?fieldPresent=false", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		var body map[string]string
		json.NewDecoder(resp.Body).Decode(&body)
		assert.Equal(t, "testField not provided", body["message"])
		resp.Body.Close()
	})
}

func TestFaviconHandler(t *testing.T) {
	app := fiber.New()
	app.Get("/favicon.ico", FaviconHandler)

	req := httptest.NewRequest("GET", "/favicon.ico", nil)
	resp, err := app.Test(req, -1) // Use -1 to prevent following redirects

	assert.NoError(t, err)
	assert.Contains(t, []int{fiber.StatusMovedPermanently, fiber.StatusFound}, resp.StatusCode)
	assert.Equal(t, "/static/favicon.ico", resp.Header.Get("Location"))
	resp.Body.Close()
}

func TestPlaylistHandler(t *testing.T) {
	app := fiber.New()
	app.Get("/playlist.m3u", PlaylistHandler)

	// Test case 1: No query parameters
	t.Run("NoParams", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/playlist.m3u", nil)
		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		assert.Contains(t, []int{fiber.StatusMovedPermanently, fiber.StatusFound}, resp.StatusCode)
		assert.Equal(t, "/channels?type=m3u&q=&c=&l=&sg=", resp.Header.Get("Location"))
		resp.Body.Close()
	})

	// Test case 2: With query parameters
	t.Run("WithParams", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/playlist.m3u?q=high&c=Movies&l=English&sg=News", nil)
		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		assert.Contains(t, []int{fiber.StatusMovedPermanently, fiber.StatusFound}, resp.StatusCode)
		expectedLocation := "/channels?type=m3u&q=high&c=Movies&l=English&sg=News"
		assert.Equal(t, expectedLocation, resp.Header.Get("Location"))
		resp.Body.Close()
	})
}

func TestDASHTimeHandler(t *testing.T) {
	app := fiber.New()
	app.Get("/time", DASHTimeHandler)

	originalTimeNow := utils.TimeNow
	defer func() { utils.TimeNow = originalTimeNow }()

	fixedTime := time.Date(2023, 10, 26, 10, 30, 15, 123456789, time.UTC)
	utils.TimeNow = func() time.Time {
		return fixedTime
	}

	req := httptest.NewRequest("GET", "/time", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	// Format includes milliseconds (3 decimal places for nanoseconds)
	// 2006-01-02T15:04:05.000Z
	// The nanoseconds 123456789 should become 123 milliseconds
	expectedTimeString := "2023-10-26T10:30:15.123Z"
	assert.Equal(t, expectedTimeString, string(body))
}

func TestInitPartial(t *testing.T) {
	// Store original values
	originalCfgTitle := config.Cfg.Title
	originalCfgDisableTSHandler := config.Cfg.DisableTSHandler
	originalCfgDisableLogout := config.Cfg.DisableLogout
	originalCfgDRM := config.Cfg.DRM

	originalTitle := Title
	originalDisableTSHandler := DisableTSHandler
	originalIsLogoutDisabled := IsLogoutDisabled
	originalEnableDRM := EnableDRM

	originalGetDeviceID := utils.GetDeviceID
	originalGetJIOTVCredentials := utils.GetJIOTVCredentials
	originalRefreshTokenIfExpired := RefreshTokenIfExpired
	originalRefreshSSOTokenIfExpired := RefreshSSOTokenIfExpired
	originalTelevisionNew := television.New
	originalTV := TV

	defer func() {
		config.Cfg.Title = originalCfgTitle
		config.Cfg.DisableTSHandler = originalCfgDisableTSHandler
		config.Cfg.DisableLogout = originalCfgDisableLogout
		config.Cfg.DRM = originalCfgDRM

		Title = originalTitle
		DisableTSHandler = originalDisableTSHandler
		IsLogoutDisabled = originalIsLogoutDisabled
		EnableDRM = originalEnableDRM

		utils.GetDeviceID = originalGetDeviceID
		utils.GetJIOTVCredentials = originalGetJIOTVCredentials
		RefreshTokenIfExpired = originalRefreshTokenIfExpired
		RefreshSSOTokenIfExpired = originalRefreshSSOTokenIfExpired
		television.New = originalTelevisionNew
		TV = originalTV
	}()

	// Mock dependencies not under test for Init
	utils.GetDeviceID = func() string { return "test-device" }
	utils.GetJIOTVCredentials = func() (*utils.JIOTV_CREDENTIALS, error) {
		return &utils.JIOTV_CREDENTIALS{}, nil // Benign credentials
	}
	RefreshTokenIfExpired = func(credentials *utils.JIOTV_CREDENTIALS) error { return nil }
	RefreshSSOTokenIfExpired = func(credentials *utils.JIOTV_CREDENTIALS) error { return nil }
	television.New = func(credentials *utils.JIOTV_CREDENTIALS) *television.Television {
		return &television.Television{} // Dummy TV service
	}

	t.Run("ConfigValuesSet", func(t *testing.T) {
		config.Cfg.Title = "TestTitleFromConfig"
		config.Cfg.DisableTSHandler = true
		config.Cfg.DisableLogout = true // This will set IsLogoutDisabled to true
		config.Cfg.DRM = true

		Init() // Call the actual Init function

		assert.Equal(t, "TestTitleFromConfig", Title)
		assert.True(t, DisableTSHandler)
		assert.True(t, IsLogoutDisabled)
		assert.True(t, EnableDRM)
	})

	t.Run("ConfigValuesNotSet_Defaults", func(t *testing.T) {
		// Reset config values to empty/zero to test defaults
		config.Cfg.Title = ""
		config.Cfg.DisableTSHandler = false // Assuming false is the zero value/default for boolean
		config.Cfg.DisableLogout = false
		config.Cfg.DRM = false

		Init() // Call the actual Init function

		assert.Equal(t, "JioTV Go", Title) // Default title
		assert.False(t, DisableTSHandler)
		assert.False(t, IsLogoutDisabled)
		assert.False(t, EnableDRM)
	})
}

func TestMain(m *testing.M) {
	// Setup for all tests in this package
	if utils.Log == nil {
		utils.Log = utils.GetLogger() 
	}
	// Ensure any other package-level setup needed by handlers is done here
	// For example, if epg.EPG_URL or other consts/vars are used by handlers.
	
	// Example: Set a default for epg.EPG_URL if it's used by a handler being tested indirectly
	// This is more relevant if not all dependencies are perfectly mocked out in each test.
	// if epg.EPG_URL == "" {
	// 	epg.EPG_URL = "http://mock.epg.url/epg.xml.gz"
	// }


	// Run tests
	exitCode := m.Run()
	os.Exit(exitCode)
}
```
