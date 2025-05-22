package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httptest" // Standard library for HTTP test servers
	"os"
	// "strings" // Not used yet, but might be for more complex assertions
	"testing"
	"time"

	"github.com/gofiber/fiber/v2" // For status codes, etc.
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil" // For fasthttp test server

	"github.com/jiotv-go/jiotv_go/v3/pkg/secureurl"
	"github.com/jiotv-go/jiotv_go/v3/pkg/television"
	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
)

// Note: DrmMpdOutput is defined in types.go and is accessible.
// television.LiveURLOutput, television.Mpd, television.Bitrates are from the television package.

func TestGetDrmMpd(t *testing.T) {
	originalSecureURLEncrypt := secureurl.EncryptURL
	originalGetRequestClient := utils.GetRequestClient
	originalTelevisionNew := television.New // To control TV instance
	originalHandlersInit := Init           // Store original Init for handlers
	originalTV := TV                       // Store original global TV

	defer func() {
		secureurl.EncryptURL = originalSecureURLEncrypt
		utils.GetRequestClient = originalGetRequestClient
		television.New = originalTelevisionNew
		Init = originalHandlersInit
		TV = originalTV
	}()

	// Mock server for TV.Live() HTTP calls
	listener := fasthttputil.NewInmemoryListener()
	defer listener.Close()

	// Patch GetRequestClient to use the mock server's dialer
	utils.GetRequestClient = func() *fasthttp.Client {
		return &fasthttp.Client{Dial: fasthttp.DialFunc(func(addr string) (fasthttp.DialConn, error) { return listener.Dial() })}
	}

	// Patch television.New to use the client that dials our mock server.
	television.New = func(creds *utils.JIOTV_CREDENTIALS) *television.Television {
		return originalTelevisionNew(creds) 
	}
	Init() 

	mockHTTPServer := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) { /* Default - overridden per test */ },
	}
	go mockHTTPServer.Serve(listener) 

	testCases := []struct {
		name               string
		channelID          string
		quality            string
		setupMockServer    func(ctx *fasthttp.RequestCtx) 
		mockEncryptFunc    func(data string) (string, error)
		expectedPlayURL    string
		expectedLicenseURL string
		expectedHost       string
		expectedPath       string
		expectError        bool
		expectedErrorMsg   string
	}{
		{
			name:      "Success_HighQuality",
			channelID: "ch1",
			quality:   "high",
			setupMockServer: func(ctx *fasthttp.RequestCtx) {
				liveResult := television.LiveURLOutput{
					Mpd: television.Mpd{Key: "key_url_high", Bitrates: television.Bitrates{High: "http://jio.com/high.mpd"}}}
				body, _ := json.Marshal(liveResult)
				ctx.SetStatusCode(fasthttp.StatusOK)
				ctx.SetBody(body)
			},
			mockEncryptFunc: func(data string) (string, error) { return "enc_" + data, nil },
			expectedPlayURL: "/render.mpd?auth=enc_http://jio.com/high.mpd",
			expectedLicenseURL: "/drm?auth=enc_key_url_high&channel_id=ch1&channel=enc_http://jio.com/high.mpd",
			expectedHost:    "enc_jio.com",
			expectedPath:    "enc_/",
		},
		{
			name:      "Success_AutoQuality",
			channelID: "ch2",
			quality:   "auto", 
			setupMockServer: func(ctx *fasthttp.RequestCtx) {
				liveResult := television.LiveURLOutput{
					Mpd: television.Mpd{Key: "key_url_auto", Bitrates: television.Bitrates{Auto: "http://jio.com/auto.mpd"}}}
				body, _ := json.Marshal(liveResult)
				ctx.SetStatusCode(fasthttp.StatusOK)
				ctx.SetBody(body)
			},
			mockEncryptFunc: func(data string) (string, error) { return "enc_auto_" + data, nil },
			expectedPlayURL: "/render.mpd?auth=enc_auto_http://jio.com/auto.mpd",
			expectedLicenseURL: "/drm?auth=enc_auto_key_url_auto&channel_id=ch2&channel=enc_auto_http://jio.com/auto.mpd",
			expectedHost:    "enc_auto_jio.com",
			expectedPath:    "enc_auto_/",
		},
		{
			name:      "Failure_TV.LiveError",
			channelID: "ch_err",
			quality:   "auto",
			setupMockServer: func(ctx *fasthttp.RequestCtx) {
				ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			},
			mockEncryptFunc:  func(data string) (string, error) { return "enc_" + data, nil },
			expectError:      true,
			expectedErrorMsg: "Request failed with status code: 500",
		},
		{
			name:      "Failure_EncryptMPDKeyError",
			channelID: "ch_enc_key_err",
			quality:   "auto",
			setupMockServer: func(ctx *fasthttp.RequestCtx) { 
				liveResult := television.LiveURLOutput{Mpd: television.Mpd{Key: "key_to_fail_encrypt", Bitrates: television.Bitrates{Auto: "http://jio.com/auto.mpd"}}}
				body, _ := json.Marshal(liveResult)
				ctx.SetStatusCode(fasthttp.StatusOK)
				ctx.SetBody(body)
			},
			mockEncryptFunc: func(data string) (string, error) {
				if data == "key_to_fail_encrypt" {
					return "", errors.New("encrypt key error")
				}
				return "enc_" + data, nil
			},
			expectError:      true,
			expectedErrorMsg: "failed to encrypt Mpd.Key: encrypt key error",
		},
		{
			name:      "Failure_EncryptTVURLError",
			channelID: "ch_enc_tvurl_err",
			quality:   "auto",
			setupMockServer: func(ctx *fasthttp.RequestCtx) { 
				liveResult := television.LiveURLOutput{Mpd: television.Mpd{Key: "key_ok", Bitrates: television.Bitrates{Auto: "http://jio.com/tv_to_fail_encrypt.mpd"}}}
				body, _ := json.Marshal(liveResult)
				ctx.SetStatusCode(fasthttp.StatusOK)
				ctx.SetBody(body)
			},
			mockEncryptFunc: func(data string) (string, error) {
				if data == "http://jio.com/tv_to_fail_encrypt.mpd" {
					return "", errors.New("encrypt tv_url error")
				}
				return "enc_" + data, nil
			},
			expectError:      true,
			expectedErrorMsg: "failed to encrypt tv_url: encrypt tv_url error",
		},
		{
			name:      "Failure_ParseTVURLError", 
			channelID: "ch_parse_err",
			quality:   "auto",
			setupMockServer: func(ctx *fasthttp.RequestCtx) {
				liveResult := television.LiveURLOutput{Mpd: television.Mpd{Key: "key_ok", Bitrates: television.Bitrates{Auto: "http://[::1]:namedport"}}} 
				body, _ := json.Marshal(liveResult)
				ctx.SetStatusCode(fasthttp.StatusOK)
				ctx.SetBody(body)
			},
			mockEncryptFunc:  func(data string) (string, error) { return "enc_" + data, nil },
			expectError:      true,
			expectedErrorMsg: "failed to parse tv_url:", 
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHTTPServer.Handler = tc.setupMockServer
			secureurl.EncryptURL = tc.mockEncryptFunc
			output, err := internalGetDrmMpd(tc.channelID, tc.quality)

			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), tc.expectedErrorMsg)
				}
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				assert.Equal(t, tc.expectedPlayURL, output.PlayUrl)
				assert.Equal(t, tc.expectedLicenseURL, output.LicenseUrl)
				assert.Equal(t, tc.expectedHost, output.Tv_url_host)
				assert.Equal(t, tc.expectedPath, output.Tv_url_path)
			}
		})
	}
}

func TestLiveMpdHandler(t *testing.T) {
	app := fiber.New(fiber.Config{Views: nil})
	app.Get("/mpd/:channelID", LiveMpdHandler)

	originalInternalGetDrmMpd := internalGetDrmMpd 
	defer func() { internalGetDrmMpd = originalInternalGetDrmMpd }() 

	originalInit := Init
	Init = func() {} 
	defer func() { Init = originalInit }()

	t.Run("Success", func(t *testing.T) {
		expectedOutput := &DrmMpdOutput{ 
			PlayUrl: "mock_play_url", LicenseUrl: "mock_license_url",
			Tv_url_host: "mock_host", Tv_url_path: "mock_path",
		}
		internalGetDrmMpd = func(channelID, quality string) (*DrmMpdOutput, error) {
			assert.Equal(t, "ch1", channelID)
			assert.Equal(t, "q_auto", quality)
			return expectedOutput, nil
		}

		req := httptest.NewRequest("GET", "/mpd/ch1?q=q_auto", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("Failure_getDrmMpdError", func(t *testing.T) {
		internalGetDrmMpd = func(channelID, quality string) (*DrmMpdOutput, error) {
			return nil, fmt.Errorf("simulated internalGetDrmMpd error")
		}
		req := httptest.NewRequest("GET", "/mpd/ch1", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		var jsonResponse map[string]string
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		json.Unmarshal(bodyBytes, &jsonResponse)
		assert.Contains(t, jsonResponse["message"], "Failed to get DRM MPD details: simulated internalGetDrmMpd error")
	})
}

func TestDRMKeyHandler(t *testing.T) {
	app := fiber.New()
	app.All("/drm", DRMKeyHandler) 

	originalDecryptURL := secureurl.DecryptURL
	originalGetRequestClient := utils.GetRequestClient
	originalGetDeviceID := utils.GetDeviceID
	originalTimeNow := utils.TimeNow
	originalTelevisionNew := television.New
	originalInit := Init
	originalTV := TV 
	originalGetCreds := utils.GetJIOTVCredentials
	originalRefreshTokenIfExpired := RefreshTokenIfExpired
	originalRefreshSSOTokenIfExpired := RefreshSSOTokenIfExpired


	defer func() {
		secureurl.DecryptURL = originalDecryptURL
		utils.GetRequestClient = originalGetRequestClient
		utils.GetDeviceID = originalGetDeviceID
		utils.TimeNow = originalTimeNow
		television.New = originalTelevisionNew
		Init = originalInit
		TV = originalTV
		utils.GetJIOTVCredentials = originalGetCreds
		RefreshTokenIfExpired = originalRefreshTokenIfExpired
		RefreshSSOTokenIfExpired = originalRefreshSSOTokenIfExpired
	}()
	
	headRequestListener := fasthttputil.NewInmemoryListener()
	defer headRequestListener.Close()
	headRequestServer := &fasthttp.Server{ Handler: func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Set("Set-Cookie", "default_head_cookie=val")
		ctx.SetStatusCode(fasthttp.StatusOK)
	}}
	go headRequestServer.Serve(headRequestListener)

	proxyTargetListener := fasthttputil.NewInmemoryListener()
	defer proxyTargetListener.Close()
	proxyTargetServer := &fasthttp.Server{ Handler: func(ctx *fasthttp.RequestCtx) {
		ctx.SetBodyString("default_license_data")
		ctx.SetStatusCode(fasthttp.StatusOK)
	}}
	go proxyTargetServer.Serve(proxyTargetListener)

	mockUserCredsForTV := &utils.JIOTV_CREDENTIALS{
		AccessToken: "tv_access_token", Crm: "tv_crm_id",
		SsoToken: "tv_sso_token", UniqueID: "tv_unique_id",
	}
	
	testTvService := &television.Television{
		AccessToken: mockUserCredsForTV.AccessToken, Crm: mockUserCredsForTV.Crm,
		SsoToken: mockUserCredsForTV.SsoToken, UniqueID: mockUserCredsForTV.UniqueID,
		Headers: map[string]string{"X-TV-Header": "GlobalTV"}, 
		Client:  &fasthttp.Client{Dial: proxyTargetListener.Dial}, 
	}
	
	television.New = func(creds *utils.JIOTV_CREDENTIALS) *television.Television {
		testTvService.AccessToken = creds.AccessToken
		testTvService.Crm = creds.CRM
		testTvService.SsoToken = creds.SSOToken
		testTvService.UniqueID = creds.UniqueID
		testTvService.Client = &fasthttp.Client{Dial: proxyTargetListener.Dial}
		return testTvService
	}
	utils.GetJIOTVCredentials = func() (*utils.JIOTV_CREDENTIALS, error) {
		return mockUserCredsForTV, nil 
	}
	RefreshTokenIfExpired = func(_ *utils.JIOTV_CREDENTIALS) error { return nil } 
	RefreshSSOTokenIfExpired = func(_ *utils.JIOTV_CREDENTIALS) error { return nil } 
	utils.GetDeviceID = func() string { return "init_device_id" } 

	originalInit() 


	t.Run("Success", func(t *testing.T) {
		decryptedLicenseURL := "http://" + proxyTargetListener.Addr().String() + "/getkey_success_target"
		decryptedChannelURL := "http://" + headRequestListener.Addr().String() + "/channel_stream_for_head"

		secureurl.DecryptURL = func(data string) (string, error) {
			if data == "valid_auth_token" { return decryptedLicenseURL, nil }
			if data == "valid_channel_token" { return decryptedChannelURL, nil }
			return "", fmt.Errorf("Success: unexpected DecryptURL call with %s", data)
		}
		currentGetRequestClient := utils.GetRequestClient 
		utils.GetRequestClient = func() *fasthttp.Client { 
			return &fasthttp.Client{Dial: headRequestListener.Dial}
		}
		defer func() { utils.GetRequestClient = currentGetRequestClient }()
		
		currentGetDeviceID := utils.GetDeviceID
		utils.GetDeviceID = func() string { return "drm_device_id_success" }
		defer func() { utils.GetDeviceID = currentGetDeviceID }()

		currentTimeNow := utils.TimeNow
		utils.TimeNow = func() time.Time { return time.Date(2023, 10, 26, 12, 0, 0, 0, time.UTC) }
		defer func() { utils.TimeNow = currentTimeNow }()
		
		headRequestHit := false
		headRequestServer.Handler = func(ctx *fasthttp.RequestCtx) {
			headRequestHit = true
			assert.Equal(t, "HEAD", string(ctx.Method()))
			assert.Equal(t, "/channel_stream_for_head", string(ctx.Path()))
			ctx.Response.Header.Set("Set-Cookie", "cookie_from_head=test_value")
			ctx.SetStatusCode(fasthttp.StatusOK)
		}
		proxyTargetHit := false
		proxyTargetServer.Handler = func(ctx *fasthttp.RequestCtx) {
			proxyTargetHit = true
			assert.Equal(t, decryptedLicenseURL, string(ctx.Request.URI().FullURI()))
			assert.Equal(t, testTvService.AccessToken, string(ctx.Request.Header.Peek("accesstoken")))
			assert.Equal(t, testTvService.Crm, string(ctx.Request.Header.Peek("subscriberId")))
			assert.Equal(t, "cookie_from_head=test_value", string(ctx.Request.Header.Peek("Cookie")))
			assert.Equal(t, "drm_device_id_success", string(ctx.Request.Header.Peek("deviceId")))
			assert.Equal(t, "ch_drm_success", string(ctx.Request.Header.Peek("channelid")))
			assert.Equal(t, "231026120000000", string(ctx.Request.Header.Peek("srno"))) 
			assert.Equal(t, PLAYER_USER_AGENT, string(ctx.Request.Header.Peek("User-Agent")))
			assert.Equal(t, "application/octet-stream", string(ctx.Request.Header.Peek("Content-Type")))
			ctx.SetStatusCode(fasthttp.StatusOK)
			ctx.SetBodyString("license_data_success_case")
		}

		req := httptest.NewRequest("GET", "/drm?auth=valid_auth_token&channel=valid_channel_token&channel_id=ch_drm_success", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.True(t, headRequestHit, "HEAD request server should have been hit")
		assert.True(t, proxyTargetHit, "Proxy target server should have been hit")
		assert.Equal(t, fasthttp.StatusOK, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		assert.Equal(t, "license_data_success_case", string(bodyBytes))
	})

	t.Run("Failure_DecryptAuthFails", func(t *testing.T) {
		secureurl.DecryptURL = func(data string) (string, error) {
			if data == "auth_fail_token" { return "", errors.New("decrypt auth error") } 
			return "channel_url_ok_for_auth_fail", nil 
		}
		req := httptest.NewRequest("GET", "/drm?auth=auth_fail_token&channel=channel_ok_for_auth_fail&channel_id=ch1", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
		var jsonResp map[string]string
		json.NewDecoder(resp.Body).Decode(&jsonResp); resp.Body.Close()
		assert.Contains(t, jsonResp["message"], "Failed to decrypt auth URL: decrypt auth error")
	})

	t.Run("Failure_DecryptChannelFails", func(t *testing.T) {
		secureurl.DecryptURL = func(data string) (string, error) {
			if data == "auth_ok_for_channel_fail" { return "license_url_ok", nil } 
			if data == "channel_fail_token" { return "", errors.New("decrypt channel error") } 
			return "", fmt.Errorf("unexpected DecryptURL call")
		}
		req := httptest.NewRequest("GET", "/drm?auth=auth_ok_for_channel_fail&channel=channel_fail_token&channel_id=ch1", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
		var jsonResp map[string]string
		json.NewDecoder(resp.Body).Decode(&jsonResp); resp.Body.Close()
		assert.Contains(t, jsonResp["message"], "Failed to decrypt channel URL: decrypt channel error")
	})
	
	t.Run("Failure_HeadRequestClientDoError", func(t *testing.T) {
		decryptedChannelURLForHeadClientFail := "http://" + headRequestListener.Addr().String() + "/head_client_fail_path"
		secureurl.DecryptURL = func(data string) (string, error) {
			if data == "auth_ok_head_client_fail" { return "license_url_ok", nil }
			if data == "channel_ok_head_client_fail" { return decryptedChannelURLForHeadClientFail, nil }
			return "", fmt.Errorf("unexpected DecryptURL")
		}
		currentGetRequestClient := utils.GetRequestClient
		utils.GetRequestClient = func() *fasthttp.Client { 
			return &fasthttp.Client{Dial: func(addr string) (fasthttp.DialConn, error) {
				return nil, errors.New("simulated dial error for HEAD") 
			}}
		}
		defer func() { utils.GetRequestClient = currentGetRequestClient }()

		req := httptest.NewRequest("GET", "/drm?auth=auth_ok_head_client_fail&channel=channel_ok_head_client_fail&channel_id=ch1", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		var jsonResp map[string]string
		json.NewDecoder(resp.Body).Decode(&jsonResp); resp.Body.Close()
		assert.Contains(t, jsonResp["message"], "Failed to perform HEAD request: simulated dial error for HEAD")
	})
	
	t.Run("Failure_ProxyDoFails", func(t *testing.T) {
		decryptedLicenseURLForProxyFail := "http://"+proxyTargetListener.Addr().String()+"/license_proxy_fail"
		decryptedChannelURLForProxyFail := "http://" + headRequestListener.Addr().String() + "/channel_proxy_fail"

		secureurl.DecryptURL = func(data string) (string, error) {
			if data == "auth_proxy_fail" { return decryptedLicenseURLForProxyFail, nil }
			if data == "channel_proxy_fail" { return decryptedChannelURLForProxyFail, nil }
			return "", fmt.Errorf("unexpected DecryptURL")
		}
		currentGetRequestClient := utils.GetRequestClient
        utils.GetRequestClient = func() *fasthttp.Client { return &fasthttp.Client{Dial: headRequestListener.Dial} } 
		defer func() { utils.GetRequestClient = currentGetRequestClient }()

		headRequestServer.Handler = func(ctx *fasthttp.RequestCtx) { 
			ctx.Response.Header.Set("Set-Cookie", "dummy_cookie_from_head=123"); ctx.SetStatusCode(fasthttp.StatusOK)
		}
		proxyTargetServer.Handler = func(ctx *fasthttp.RequestCtx) { 
			ctx.Error("simulated proxy.Do target error", fasthttp.StatusServiceUnavailable)
		}
		
		req := httptest.NewRequest("GET", "/drm?auth=auth_proxy_fail&channel=channel_proxy_fail&channel_id=ch1", nil)
		resp, err := app.Test(req) 

		assert.NoError(t, err) 
		assert.Equal(t, fasthttp.StatusServiceUnavailable, resp.StatusCode) 
	})

	// Restore default handlers for listeners
	headRequestServer.Handler = func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Set("Set-Cookie", "default_head_cookie=val"); ctx.SetStatusCode(fasthttp.StatusOK)
	}
	proxyTargetServer.Handler = func(ctx *fasthttp.RequestCtx) {
		ctx.SetBodyString("default_license_data"); ctx.SetStatusCode(fasthttp.StatusOK)
	}
}

func TestMpdHandler(t *testing.T) {
	app := fiber.New()
	app.Get("/render.mpd", MpdHandler)

	originalDecryptURL := secureurl.DecryptURL
	originalTelevisionNew := television.New
	originalInit := Init
	originalTV := TV
	originalGetCreds := utils.GetJIOTVCredentials
	originalRefreshTokenIfExpired := RefreshTokenIfExpired
	originalRefreshSSOTokenIfExpired := RefreshSSOTokenIfExpired

	defer func() {
		secureurl.DecryptURL = originalDecryptURL
		television.New = originalTelevisionNew
		Init = originalInit
		TV = originalTV
		utils.GetJIOTVCredentials = originalGetCreds
		RefreshTokenIfExpired = originalRefreshTokenIfExpired
		RefreshSSOTokenIfExpired = originalRefreshSSOTokenIfExpired
	}()

	proxyTargetListener := fasthttputil.NewInmemoryListener()
	defer proxyTargetListener.Close()
	proxyTargetServer := &fasthttp.Server{Handler: func(ctx *fasthttp.RequestCtx) {
		// Default handler, override in tests
		ctx.SetBodyString("<MPD><Period><BaseURL>http://original/base/</BaseURL></Period></MPD>")
		ctx.Response.Header.Set("Set-Cookie", "original_cookie=val; Domain=target-host.com; path=/")
		ctx.SetStatusCode(fasthttp.StatusOK)
	}}
	go proxyTargetServer.Serve(proxyTargetListener)

	mockUserCredsForTV_Mpd := &utils.JIOTV_CREDENTIALS{} 
	mockTvService_Mpd := &television.Television{
		Client: &fasthttp.Client{Dial: proxyTargetListener.Dial},
	}
	television.New = func(creds *utils.JIOTV_CREDENTIALS) *television.Television {
		mockTvService_Mpd.Client = &fasthttp.Client{Dial: proxyTargetListener.Dial} 
		return mockTvService_Mpd
	}
	utils.GetJIOTVCredentials = func() (*utils.JIOTV_CREDENTIALS, error) { return mockUserCredsForTV_Mpd, nil }
	RefreshTokenIfExpired = func(_ *utils.JIOTV_CREDENTIALS) error { return nil }
	RefreshSSOTokenIfExpired = func(_ *utils.JIOTV_CREDENTIALS) error { return nil }
	originalInit()

	t.Run("Success", func(t *testing.T) {
		decryptedURL := "http://target-host.com/manifest.mpd"
		secureurl.DecryptURL = func(data string) (string, error) {
			if data == "enc_auth_val_mpd" { return decryptedURL, nil }
			return "", fmt.Errorf("MpdHandler DecryptURL unexpected: %s", data)
		}

		proxyTargetServer.Handler = func(ctx *fasthttp.RequestCtx) {
			assert.Equal(t, decryptedURL, string(ctx.Request.URI().FullURI()))
			assert.Equal(t, "target-host.com", string(ctx.Request.Header.Host()))
			assert.Equal(t, PLAYER_USER_AGENT, string(ctx.Request.Header.UserAgent()))
			ctx.Response.Header.Set("Set-Cookie", "original_cookie=val; Domain=target-host.com; path=/")
			ctx.SetBodyString("<MPD><Period><BaseURL>http://original/base/</BaseURL></Period></MPD>")
			ctx.SetStatusCode(fasthttp.StatusOK)
		}

		req := httptest.NewRequest("GET", "/render.mpd?auth=enc_auth_val_mpd", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fasthttp.StatusOK, resp.StatusCode)
		
		newCookie := resp.Header.Get("Set-Cookie")
		assert.NotContains(t, newCookie, "Domain=target-host.com")
		assert.Contains(t, newCookie, "path=/render.dash") 

		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		assert.Contains(t, string(bodyBytes), "<BaseURL>/render.dash/dash/</BaseURL>") 
	})

	t.Run("Success_NoBaseURLInResponse", func(t *testing.T) {
		decryptedURL := "http://target-host.com/no_base_url.mpd"
		secureurl.DecryptURL = func(data string) (string, error) {
			if data == "enc_no_base" { return decryptedURL, nil }
			return "", fmt.Errorf("MpdHandler DecryptURL unexpected: %s", data)
		}
		proxyTargetServer.Handler = func(ctx *fasthttp.RequestCtx) {
			ctx.SetBodyString("<MPD><Period id=\"1\"></Period></MPD>") // No BaseURL
			ctx.SetStatusCode(fasthttp.StatusOK)
		}
		req := httptest.NewRequest("GET", "/render.mpd?auth=enc_no_base", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fasthttp.StatusOK, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		assert.Contains(t, string(bodyBytes), "<BaseURL>/render.dash/</BaseURL>") 
	})

	t.Run("Failure_AuthQueryMissing", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/render.mpd", nil) 
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode) 
	})

	t.Run("Failure_DecryptURLFails", func(t *testing.T) {
		secureurl.DecryptURL = func(data string) (string, error) {
			return "", errors.New("decrypt error mpd")
		}
		req := httptest.NewRequest("GET", "/render.mpd?auth=fail_decrypt_mpd", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
		var jsonResp map[string]string
		json.NewDecoder(resp.Body).Decode(&jsonResp); resp.Body.Close()
		assert.Contains(t, jsonResp["message"], "Failed to decrypt proxy URL: decrypt error mpd")
	})
	
	t.Run("Failure_ParseURLFails", func(t *testing.T) {
		secureurl.DecryptURL = func(data string) (string, error) {
			return "http://[::1]:this_is_bad", nil // Malformed URL for url.Parse
		}
		req := httptest.NewRequest("GET", "/render.mpd?auth=malformed_url_auth", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		var jsonResp map[string]string
		json.NewDecoder(resp.Body).Decode(&jsonResp); resp.Body.Close()
		assert.Contains(t, jsonResp["message"], "Failed to parse decrypted URL:")
	})
}

func TestDashHandler(t *testing.T) {
	app := fiber.New()
	app.Get("/render.dash/*", DashHandler) 

	originalDecryptURL := secureurl.DecryptURL
	originalTelevisionNew := television.New
	originalInit := Init
	originalTV := TV
	originalGetCreds := utils.GetJIOTVCredentials
	originalRefreshTokenIfExpired := RefreshTokenIfExpired
	originalRefreshSSOTokenIfExpired := RefreshSSOTokenIfExpired

	defer func() {
		secureurl.DecryptURL = originalDecryptURL
		television.New = originalTelevisionNew
		Init = originalInit
		TV = originalTV
		utils.GetJIOTVCredentials = originalGetCreds
		RefreshTokenIfExpired = originalRefreshTokenIfExpired
		RefreshSSOTokenIfExpired = originalRefreshSSOTokenIfExpired
	}()

	proxyTargetListener := fasthttputil.NewInmemoryListener()
	defer proxyTargetListener.Close()
	proxyTargetServer := &fasthttp.Server{Handler: func(ctx *fasthttp.RequestCtx) {
		ctx.SetBodyString("dash_segment_data")
		ctx.SetStatusCode(fasthttp.StatusOK)
	}}
	go proxyTargetServer.Serve(proxyTargetListener)

	mockUserCredsForTV_Dash := &utils.JIOTV_CREDENTIALS{}
	mockTvService_Dash := &television.Television{
		Client: &fasthttp.Client{Dial: proxyTargetListener.Dial},
	}
	television.New = func(creds *utils.JIOTV_CREDENTIALS) *television.Television {
		mockTvService_Dash.Client = &fasthttp.Client{Dial: proxyTargetListener.Dial}
		return mockTvService_Dash
	}
	utils.GetJIOTVCredentials = func() (*utils.JIOTV_CREDENTIALS, error) { return mockUserCredsForTV_Dash, nil }
	RefreshTokenIfExpired = func(_ *utils.JIOTV_CREDENTIALS) error { return nil }
	RefreshSSOTokenIfExpired = func(_ *utils.JIOTV_CREDENTIALS) error { return nil }
	originalInit()

	t.Run("Success", func(t *testing.T) {
		decryptedHost := "target-dash-host.com"
		decryptedPath := "/dash/path_prefix/" 
		requestedSegment := "segment1.m4s"

		secureurl.DecryptURL = func(data string) (string, error) {
			if data == "enc_dash_host" { return decryptedHost, nil }
			if data == "enc_dash_path" { return decryptedPath, nil }
			return "", fmt.Errorf("DashHandler DecryptURL unexpected: %s", data)
		}
		proxyTargetServer.Handler = func(ctx *fasthttp.RequestCtx) {
			expectedURI := fmt.Sprintf("https://%s%s%s", decryptedHost, decryptedPath, requestedSegment)
			assert.Equal(t, expectedURI, string(ctx.Request.URI().FullURI()))
			assert.Equal(t, PLAYER_USER_AGENT, string(ctx.Request.Header.UserAgent()))
			ctx.SetBodyString("segment_data_for_" + requestedSegment)
			ctx.SetStatusCode(fasthttp.StatusOK)
		}

		req := httptest.NewRequest("GET", "/render.dash/"+requestedSegment+"?host=enc_dash_host&path=enc_dash_path", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fasthttp.StatusOK, resp.StatusCode)
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		assert.Equal(t, "segment_data_for_"+requestedSegment, string(bodyBytes))
	})

	t.Run("Failure_HostQueryMissing", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/render.dash/segment.m4s?path=enc_path", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Failure_PathQueryMissing", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/render.dash/segment.m4s?host=enc_host", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Failure_DecryptHostFails", func(t *testing.T) {
		secureurl.DecryptURL = func(data string) (string, error) {
			if data == "enc_host_fail" { return "", errors.New("decrypt host error") }
			return "ok_path", nil
		}
		req := httptest.NewRequest("GET", "/render.dash/segment.m4s?host=enc_host_fail&path=ok_path_enc", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
		var jsonResp map[string]string
		json.NewDecoder(resp.Body).Decode(&jsonResp); resp.Body.Close()
		assert.Contains(t, jsonResp["message"], "Failed to decrypt proxy host: decrypt host error")
	})

	t.Run("Failure_DecryptPathFails", func(t *testing.T) {
		secureurl.DecryptURL = func(data string) (string, error) {
			if data == "enc_host_ok" { return "ok_host", nil }
			if data == "enc_path_fail" { return "", errors.New("decrypt path error") }
			return "", fmt.Errorf("DashHandler DecryptURL unexpected: %s", data)
		}
		req := httptest.NewRequest("GET", "/render.dash/segment.m4s?host=enc_host_ok&path=enc_path_fail", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
		var jsonResp map[string]string
		json.NewDecoder(resp.Body).Decode(&jsonResp); resp.Body.Close()
		assert.Contains(t, jsonResp["message"], "Failed to decrypt proxy path: decrypt path error")
	})
}


func TestMain(m *testing.M) {
	if utils.Log == nil {
		utils.Log = utils.GetLogger() 
	}
	os.Exit(m.Run())
}
