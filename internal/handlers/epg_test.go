package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"

	"github.com/jiotv-go/jiotv_go/v3/pkg/epg" // For epg.EPG_URL
	"github.com/jiotv-go/jiotv_go/v3/pkg/television"
	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
)

func TestWebEPGHandler(t *testing.T) {
	app := fiber.New()
	app.Get("/epg/:channelID/:offset", WebEPGHandler)

	originalTelevisionNew := television.New
	originalInit := Init
	originalGetCredentials := utils.GetJIOTVCredentials
	originalRefreshTokenIfExpired := RefreshTokenIfExpired
	originalRefreshSSOTokenIfExpired := RefreshSSOTokenIfExpired
	originalTV := TV

	defer func() {
		television.New = originalTelevisionNew
		Init = originalInit
		utils.GetJIOTVCredentials = originalGetCredentials
		RefreshTokenIfExpired = originalRefreshTokenIfExpired
		RefreshSSOTokenIfExpired = originalRefreshSSOTokenIfExpired
		TV = originalTV
	}()

	proxyTargetListenerEPG := fasthttputil.NewInmemoryListener()
	defer proxyTargetListenerEPG.Close()

	mockTvClientEPG := &fasthttp.Client{Dial: proxyTargetListenerEPG.Dial}
	mockTvServiceEPG := &television.Television{Client: mockTvClientEPG}

	// Setup Init to use our mocked TV service
	utils.GetJIOTVCredentials = func() (*utils.JIOTV_CREDENTIALS, error) {
		return &utils.JIOTV_CREDENTIALS{AccessToken: "dummy_token_for_epg"}, nil
	}
	RefreshTokenIfExpired = func(credentials *utils.JIOTV_CREDENTIALS) error { return nil }
	RefreshSSOTokenIfExpired = func(credentials *utils.JIOTV_CREDENTIALS) error { return nil }
	television.New = func(creds *utils.JIOTV_CREDENTIALS) *television.Television {
		// Ensure the client within the TV service is the one dialing our mock server
		mockTvServiceEPG.AccessToken = creds.AccessToken // Keep creds if needed by other parts of TV service
		return mockTvServiceEPG
	}
	originalInit() // Call the original Init, it will use the mocked television.New

	proxyTargetServerEPG := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			// Default EPG server handler, can be overridden in subtests
			assert.True(t, strings.HasPrefix(string(ctx.Path()), "/apis/v1.3/getepg/get"), "Request path for EPG is wrong")
			// Example: /apis/v1.3/getepg/get?offset=0&channel_id=100
			expectedOffset := string(ctx.QueryArgs().Peek("offset"))
			expectedChannelID := string(ctx.QueryArgs().Peek("channel_id"))

			// These assertions can be made more specific in sub-tests
			assert.NotEmpty(t, expectedOffset, "Offset should be present in EPG request")
			assert.NotEmpty(t, expectedChannelID, "Channel ID should be present in EPG request")

			ctx.SetStatusCode(fasthttp.StatusOK)
			ctx.SetBodyString(`{"epg_data": "mock EPG for ` + expectedChannelID + ` at offset ` + expectedOffset + `"}`)
		},
	}
	go proxyTargetServerEPG.Serve(proxyTargetListenerEPG) //nolint:errcheck
	defer proxyTargetServerEPG.Shutdown()


	t.Run("Success_NumericChannelID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/epg/123/1", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fasthttp.StatusOK, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		assert.Equal(t, `{"epg_data": "mock EPG for 123 at offset 1"}`, string(body))
		assert.Empty(t, resp.Header.Get(fiber.HeaderServer))
	})

	t.Run("Success_SLPrefixedChannelID", func(t *testing.T) {
		// Specific handler for this test to check "sl" stripping
		proxyTargetServerEPG.Handler = func(ctx *fasthttp.RequestCtx) {
			assert.Equal(t, "/apis/v1.3/getepg/get", string(ctx.Path()))
			assert.Equal(t, "2", string(ctx.QueryArgs().Peek("offset")))
			assert.Equal(t, "456", string(ctx.QueryArgs().Peek("channel_id"))) // "sl" should be stripped
			ctx.SetStatusCode(fasthttp.StatusOK)
			ctx.SetBodyString(`{"epg_data": "mock EPG for 456 at offset 2"}`)
		}
		defer func() { // Restore default handler
			proxyTargetServerEPG.Handler = func(ctx *fasthttp.RequestCtx) {
				ctx.SetStatusCode(fasthttp.StatusOK); ctx.SetBodyString(`{"epg_data":"default"}`)
			}
		}()


		req := httptest.NewRequest("GET", "/epg/sl456/2", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fasthttp.StatusOK, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		assert.Equal(t, `{"epg_data": "mock EPG for 456 at offset 2"}`, string(body))
		assert.Empty(t, resp.Header.Get(fiber.HeaderServer))
	})

	t.Run("InvalidChannelID_NonNumeric", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/epg/abc/1", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fasthttp.StatusBadRequest, resp.StatusCode)
		// Further check error response if desired
	})
	
	t.Run("InvalidChannelID_SLNonNumeric", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/epg/slabc/1", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fasthttp.StatusBadRequest, resp.StatusCode)
	})

	t.Run("InvalidOffset_NonNumeric", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/epg/123/abc", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fasthttp.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Failure_ProxyDoFails", func(t *testing.T) {
		proxyTargetServerEPG.Handler = func(ctx *fasthttp.RequestCtx) {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.SetBodyString("mock EPG server error")
		}
		defer func() { // Restore default handler
			proxyTargetServerEPG.Handler = func(ctx *fasthttp.RequestCtx) {
				ctx.SetStatusCode(fasthttp.StatusOK); ctx.SetBodyString(`{"epg_data":"default"}`)
			}
		}()

		req := httptest.NewRequest("GET", "/epg/789/3", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fasthttp.StatusInternalServerError, resp.StatusCode)
		// Body might be empty or contain error from proxy if Fiber handles it that way
	})
}

func TestPosterHandler(t *testing.T) {
	app := fiber.New()
	app.Get("/jtvposter/:date/:file", PosterHandler)

	originalTelevisionNew := television.New
	originalInit := Init
	originalGetCredentials := utils.GetJIOTVCredentials
	originalRefreshTokenIfExpired := RefreshTokenIfExpired
	originalRefreshSSOTokenIfExpired := RefreshSSOTokenIfExpired
	originalTV := TV

	defer func() {
		television.New = originalTelevisionNew
		Init = originalInit
		utils.GetJIOTVCredentials = originalGetCredentials
		RefreshTokenIfExpired = originalRefreshTokenIfExpired
		RefreshSSOTokenIfExpired = originalRefreshSSOTokenIfExpired
		TV = originalTV
	}()

	proxyTargetListenerPoster := fasthttputil.NewInmemoryListener()
	defer proxyTargetListenerPoster.Close()

	mockTvClientPoster := &fasthttp.Client{Dial: proxyTargetListenerPoster.Dial}
	mockTvServicePoster := &television.Television{Client: mockTvClientPoster}

	// Setup Init to use our mocked TV service
	utils.GetJIOTVCredentials = func() (*utils.JIOTV_CREDENTIALS, error) {
		return &utils.JIOTV_CREDENTIALS{AccessToken: "dummy_token_for_poster"}, nil
	}
	RefreshTokenIfExpired = func(credentials *utils.JIOTV_CREDENTIALS) error { return nil }
	RefreshSSOTokenIfExpired = func(credentials *utils.JIOTV_CREDENTIALS) error { return nil }
	television.New = func(creds *utils.JIOTV_CREDENTIALS) *television.Television {
		mockTvServicePoster.AccessToken = creds.AccessToken
		return mockTvServicePoster
	}
	originalInit() // Call original Init

	proxyTargetServerPoster := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			// Default poster server handler
			assert.True(t, strings.HasPrefix(string(ctx.Path()), "/dare_images/shows/"), "Request path for Poster is wrong")
			// Example: /dare_images/shows/20231026/poster_file.jpg
			// We can assert path components here based on date and file from the test case
			ctx.SetStatusCode(fasthttp.StatusOK)
			ctx.SetBodyString("mock_poster_image_data")
		},
	}
	go proxyTargetServerPoster.Serve(proxyTargetListenerPoster) //nolint:errcheck
	defer proxyTargetServerPoster.Shutdown()

	t.Run("Success_Poster", func(t *testing.T) {
		date := "20231026"
		file := "test_poster.jpg"

		// Customize handler for this specific test for more detailed assertions
		proxyTargetServerPoster.Handler = func(ctx *fasthttp.RequestCtx) {
			expectedPath := fmt.Sprintf("/dare_images/shows/%s/%s", date, file)
			assert.Equal(t, expectedPath, string(ctx.Path()))
			ctx.SetStatusCode(fasthttp.StatusOK)
			ctx.SetBodyString("specific_poster_data")
		}
		defer func() { // Restore default handler
			proxyTargetServerPoster.Handler = func(ctx *fasthttp.RequestCtx) {
				ctx.SetStatusCode(fasthttp.StatusOK); ctx.SetBodyString(`default_poster_data`)
			}
		}()


		req := httptest.NewRequest("GET", fmt.Sprintf("/jtvposter/%s/%s", date, file), nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fasthttp.StatusOK, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		assert.Equal(t, "specific_poster_data", string(body))
		assert.Empty(t, resp.Header.Get(fiber.HeaderServer))
	})

	t.Run("Failure_ProxyDoFails_Poster", func(t *testing.T) {
		proxyTargetServerPoster.Handler = func(ctx *fasthttp.RequestCtx) {
			ctx.SetStatusCode(fasthttp.StatusNotFound) // Simulate target error
			ctx.SetBodyString("mock poster server error")
		}
		defer func() { // Restore default handler
			proxyTargetServerPoster.Handler = func(ctx *fasthttp.RequestCtx) {
				ctx.SetStatusCode(fasthttp.StatusOK); ctx.SetBodyString(`default_poster_data`)
			}
		}()

		req := httptest.NewRequest("GET", "/jtvposter/20231101/error_poster.jpg", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fasthttp.StatusNotFound, resp.StatusCode)
		// Body might be empty or contain error from proxy if Fiber handles it that way
	})
}

// TestMain for utils.Log initialization if needed by any underlying functions
// (though not directly by the handlers under test here, good practice)
func TestMain(m *testing.M) {
	if utils.Log == nil {
		utils.Log = utils.GetLogger() // Ensure logger is initialized
	}
	// EPG URL is a const, so no need to mock for these specific handlers
	// if epg.EPG_URL == "" { epg.EPG_URL = "http://default.epg.url/for/test" } 
	// if EPG_POSTER_URL == "" { EPG_POSTER_URL = "http://default.poster.url/for/test" }
	os.Exit(m.Run())
}
