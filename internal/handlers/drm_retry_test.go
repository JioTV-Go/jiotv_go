package handlers

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jiotv-go/jiotv_go/v3/pkg/secureurl"
	"github.com/jiotv-go/jiotv_go/v3/pkg/store"
	"github.com/jiotv-go/jiotv_go/v3/pkg/television"
	pkgUtils "github.com/jiotv-go/jiotv_go/v3/pkg/utils"
	"github.com/valyala/fasthttp"
)

// TestDashHandlerRecoversFrom403WithoutForcedRefresh verifies the per-segment
// 403 retry contract. Segment requests carry only the CDN's __hdnea__ cookie
// (never the JioTV session tokens), so a 403 means CDN auth expired: the
// handler must clear the stale cookie and retry immediately, without the
// blocking ForceRefreshCredentials() that used to run on every expired
// segment. That blocking refresh is what turned a handful of expiring HDNEA
// tokens into hundreds of slow segment requests during DRM channel loading.
func TestDashHandlerRecoversFrom403WithoutForcedRefresh(t *testing.T) {
	const staleToken = "stale-hdnea-token"

	var mu sync.Mutex
	upstreamHits := 0
	firstRequestHadCookie := false
	retryCarriedCookie := false

	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		upstreamHits++
		hit := upstreamHits
		mu.Unlock()

		if hit == 1 {
			// Simulate an expired CDN token: the first request carries the
			// stale __hdnea__ cookie and is rejected with 403.
			cookie, err := r.Cookie("__hdnea__")
			if err == nil && cookie.Value == staleToken {
				mu.Lock()
				firstRequestHadCookie = true
				mu.Unlock()
			}
			http.Error(w, "expired token", http.StatusForbidden)
			return
		}
		// The retry must have had the stale cookie cleared.
		if cookie, err := r.Cookie("__hdnea__"); err == nil && cookie.Value != "" {
			mu.Lock()
			retryCarriedCookie = true
			mu.Unlock()
		}
		_, _ = w.Write([]byte("segment-data"))
	}))
	defer upstream.Close()

	cleanupStore, err := store.SetupTestPathPrefix()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanupStore()
	if err := store.Init(); err != nil {
		t.Fatal(err)
	}

	var logBuf bytes.Buffer
	previousTV, previousLog := TV, pkgUtils.Log
	pkgUtils.Log = log.New(&logBuf, "", 0)
	TV = &television.Television{
		Client: &fasthttp.Client{TLSConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	secureurl.Init()
	defer func() {
		TV = previousTV
		pkgUtils.Log = previousLog
		nextCredentialValidationTime = time.Time{}
	}()

	t.Setenv("JIOTV_DEBUG", "true")

	upstreamURL, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	encHost, err := secureurl.EncryptURL(upstreamURL.Host)
	if err != nil {
		t.Fatal(err)
	}
	encPath, err := secureurl.EncryptURL("/cdn/path/")
	if err != nil {
		t.Fatal(err)
	}
	encHdnea, err := secureurl.EncryptURL("__hdnea__=" + staleToken)
	if err != nil {
		t.Fatal(err)
	}

	app := fiber.New()
	app.Use("/render.dash", DashHandler)

	requestURL := fmt.Sprintf(
		"/render.dash/host/%s/path/%s/hdnea/%s/segment.m4s",
		encHost, encPath, encHdnea,
	)
	response, err := app.Test(httptest.NewRequest(http.MethodGet, requestURL, nil))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d (body: %s)", response.StatusCode, http.StatusOK, body)
	}
	if string(body) != "segment-data" {
		t.Errorf("body = %q, want %q", body, "segment-data")
	}

	mu.Lock()
	hits, firstCookie, cookieSeen := upstreamHits, firstRequestHadCookie, retryCarriedCookie
	mu.Unlock()
	if hits != 2 {
		t.Errorf("upstream hits = %d, want 2 (original request + one retry)", hits)
	}
	if !firstCookie {
		t.Errorf("first upstream request did not carry the stale __hdnea__ cookie")
	}
	if cookieSeen {
		t.Errorf("retry request still carried the stale __hdnea__ cookie")
	}
	if strings.Contains(logBuf.String(), "FORCED token refresh") {
		t.Errorf("retry triggered a blocking credential refresh:\n%s", logBuf.String())
	}
}

// TestMpdHandlerProducesStableSegmentURLs verifies that repeated manifest
// fetches rewrite <BaseURL> to byte-identical /render.dash URLs. Segment URLs
// used to be re-encrypted with a random IV on every MpdHandler call, so each
// live manifest refresh changed the URL identity of every buffered segment and
// Shaka re-downloaded them - multiplying the requests needed before playback
// could start.
func TestMpdHandlerProducesStableSegmentURLs(t *testing.T) {
	const hdneaToken = "hdnea-token-1"
	const mpdBody = `<MPD type="dynamic"><Period><BaseURL>https://cdn.example.com/cdn/</BaseURL></Period></MPD>`

	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/dash+xml")
		http.SetCookie(w, &http.Cookie{Name: "__hdnea__", Value: hdneaToken, Path: "/"})
		_, _ = w.Write([]byte(mpdBody))
	}))
	defer upstream.Close()

	cleanupStore, err := store.SetupTestPathPrefix()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanupStore()
	if err := store.Init(); err != nil {
		t.Fatal(err)
	}

	previousTV, previousLog := TV, pkgUtils.Log
	pkgUtils.Log = log.New(io.Discard, "", 0)
	TV = &television.Television{
		Client: &fasthttp.Client{TLSConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	secureurl.Init()
	defer func() {
		TV = previousTV
		pkgUtils.Log = previousLog
		nextCredentialValidationTime = time.Time{}
	}()

	encURL, err := secureurl.EncryptURL(upstream.URL + "/cdn/stream.mpd")
	if err != nil {
		t.Fatal(err)
	}

	app := fiber.New()
	app.Use("/render.mpd", MpdHandler)

	baseURLPattern := regexp.MustCompile(`<BaseURL>(.*?)</BaseURL>`)
	fetchBaseURL := func() string {
		response, err := app.Test(httptest.NewRequest(http.MethodGet, "/render.mpd?auth="+encURL, nil))
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want %d (body: %s)", response.StatusCode, http.StatusOK, body)
		}
		match := baseURLPattern.FindSubmatch(body)
		if len(match) != 2 {
			t.Fatalf("no <BaseURL> found in rewritten MPD: %s", body)
		}
		return string(match[1])
	}

	first := fetchBaseURL()
	second := fetchBaseURL()

	if first == "" {
		t.Fatal("first BaseURL is empty")
	}
	if !strings.HasPrefix(first, "/render.dash/host/") {
		t.Errorf("first BaseURL = %q, want /render.dash/host/ prefix", first)
	}
	if first != second {
		t.Errorf("BaseURL changed between manifest refreshes:\nfirst:  %s\nsecond: %s\n\nShaka treats a changed segment URL as a new segment and re-downloads it.", first, second)
	}
}

// TestMpdHandlerInjectsUTCTiming verifies the server adds a UTCTiming element
// to live manifests. JioTV MPDs have no UTCTiming and use an epoch
// availabilityStartTime, so without a clock source clients compute the live
// edge from their own (possibly skewed) clock and pre-roll a large chunk of
// the DVR window before playback starts.
func TestMpdHandlerInjectsUTCTiming(t *testing.T) {
	const mpdWithoutTiming = `<MPD type="dynamic" availabilityStartTime="1970-01-01T00:00:00Z"><Period><BaseURL>https://cdn.example.com/cdn/</BaseURL></Period></MPD>`
	const mpdWithTiming = `<MPD type="dynamic"><UTCTiming schemeIdUri="urn:mpeg:dash:utc:direct:2014" value="https://cdn.example.com/time"/><Period><BaseURL>https://cdn.example.com/cdn/</BaseURL></Period></MPD>`

	tests := []struct {
		name         string
		upstreamBody string
		wantInjected bool
		wantNotDupl  bool
	}{
		{
			name:         "inject when manifest has no UTCTiming",
			upstreamBody: mpdWithoutTiming,
			wantInjected: true,
		},
		{
			name:         "do not duplicate an existing UTCTiming",
			upstreamBody: mpdWithTiming,
			wantNotDupl:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/dash+xml")
				_, _ = w.Write([]byte(tt.upstreamBody))
			}))
			defer upstream.Close()

			cleanupStore, err := store.SetupTestPathPrefix()
			if err != nil {
				t.Fatal(err)
			}
			defer cleanupStore()
			if err := store.Init(); err != nil {
				t.Fatal(err)
			}

			previousTV, previousLog := TV, pkgUtils.Log
			pkgUtils.Log = log.New(io.Discard, "", 0)
			TV = &television.Television{
				Client: &fasthttp.Client{TLSConfig: &tls.Config{InsecureSkipVerify: true}},
			}
			secureurl.Init()
			defer func() {
				TV = previousTV
				pkgUtils.Log = previousLog
				nextCredentialValidationTime = time.Time{}
			}()

			encURL, err := secureurl.EncryptURL(upstream.URL + "/cdn/stream.mpd")
			if err != nil {
				t.Fatal(err)
			}

			app := fiber.New()
			app.Use("/render.mpd", MpdHandler)

			response, err := app.Test(httptest.NewRequest(http.MethodGet, "/render.mpd?auth="+encURL, nil))
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, want %d (body: %s)", response.StatusCode, http.StatusOK, body)
			}

			if tt.wantInjected {
				if !bytes.Contains(body, []byte(`urn:mpeg:dash:utc:http-xsdate:2014`)) {
					t.Errorf("rewritten MPD has no injected UTCTiming element:\n%s", body)
				}
				if !bytes.Contains(body, []byte(`/dashtime`)) {
					t.Errorf("injected UTCTiming does not point at /dashtime:\n%s", body)
				}
			}
			if tt.wantNotDupl {
				if count := bytes.Count(body, []byte(`<UTCTiming`)); count != 1 {
					t.Errorf("rewritten MPD contains %d UTCTiming elements, want 1 (existing preserved):\n%s", count, body)
				}
			}
		})
	}
}

// TestMpdHandlerRecordsCDNClock verifies the wiring between MpdHandler and the
// /dashtime clock source: proxying a live MPD must cache its publishTime so
// DASHTimeHandler can serve the CDN's clock instead of the machine's.
func TestMpdHandlerRecordsCDNClock(t *testing.T) {
	const publishTime = "2026-08-12T20:15:49.537686Z"
	const mpdBody = `<MPD type="dynamic" publishTime="` + publishTime + `"><Period><BaseURL>https://cdn.example.com/cdn/</BaseURL></Period></MPD>`

	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/dash+xml")
		_, _ = w.Write([]byte(mpdBody))
	}))
	defer upstream.Close()

	cleanupStore, err := store.SetupTestPathPrefix()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanupStore()
	if err := store.Init(); err != nil {
		t.Fatal(err)
	}

	previousTV, previousLog := TV, pkgUtils.Log
	pkgUtils.Log = log.New(io.Discard, "", 0)
	TV = &television.Television{
		Client: &fasthttp.Client{TLSConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	secureurl.Init()
	defer func() {
		TV = previousTV
		pkgUtils.Log = previousLog
		nextCredentialValidationTime = time.Time{}
	}()

	// Isolate the global CDN clock cache.
	origPT, origFA := cdnPublishTime, cdnPublishFetchedAt
	defer func() {
		cdnClockMu.Lock()
		defer cdnClockMu.Unlock()
		cdnPublishTime, cdnPublishFetchedAt = origPT, origFA
	}()

	encURL, err := secureurl.EncryptURL(upstream.URL + "/cdn/stream.mpd")
	if err != nil {
		t.Fatal(err)
	}

	app := fiber.New()
	app.Use("/render.mpd", MpdHandler)
	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/render.mpd?auth="+encURL, nil))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if _, err := io.ReadAll(response.Body); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}

	cdn, ok := cdnNow()
	if !ok {
		t.Fatal("cdnNow() reports no recorded CDN clock after a successful MPD proxy")
	}
	want := time.Date(2026, 8, 12, 20, 15, 49, 537686000, time.UTC)
	if cdn.Before(want) {
		t.Errorf("cached CDN clock %v is before MPD publishTime %v", cdn, want)
	}
	if cdn.Sub(want) > 2*time.Second {
		t.Errorf("cached CDN clock %v extrapolated too far past publishTime %v", cdn, want)
	}
}
