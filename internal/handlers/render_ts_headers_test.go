package handlers

import (
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/jiotv-go/jiotv_go/v3/pkg/secureurl"
	"github.com/jiotv-go/jiotv_go/v3/pkg/store"
	"github.com/jiotv-go/jiotv_go/v3/pkg/television"
	pkgUtils "github.com/jiotv-go/jiotv_go/v3/pkg/utils"
)

func TestRenderTSHandlerStripsBrowserHeaders(t *testing.T) {
	// Browsers add Referer (same-origin) or Origin (cross-origin) to media
	// segment requests. The CDN answers 403 to /render.ts when either is
	// present, so they must not be forwarded upstream. Native players send
	// neither, which is why only browser playback failed.
	const channelID = "291"

	var gotReferer, gotOrigin, gotAccept, gotUserAgent string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReferer = r.Header.Get("Referer")
		gotOrigin = r.Header.Get("Origin")
		gotAccept = r.Header.Get("Accept")
		gotUserAgent = r.Header.Get("User-Agent")
		_, _ = w.Write([]byte("segment"))
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
	pkgUtils.Log = log.New(os.Stderr, "", 0)
	TV = television.New(nil)
	secureurl.Init()
	defer func() {
		TV = previousTV
		pkgUtils.Log = previousLog
		renderHDNEACache.Delete(channelID)
	}()

	streamURL := upstream.URL + "/Fallback/segment.ts"
	auth, err := secureurl.EncryptURL(streamURL)
	if err != nil {
		t.Fatal(err)
	}

	app := fiber.New()
	app.Get("/render.ts", RenderTSHandler)

	req := httptest.NewRequest(http.MethodGet, "/render.ts?auth="+url.QueryEscape(auth)+"&channel_key_id="+channelID, nil)
	req.Header.Set("Referer", "https://jiotv.example/player/291?q=auto")
	req.Header.Set("Origin", "https://jiotv.example")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Browser)")

	response, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if gotReferer != "" {
		t.Errorf("upstream Referer = %q, want empty", gotReferer)
	}
	if gotOrigin != "" {
		t.Errorf("upstream Origin = %q, want empty", gotOrigin)
	}
	if gotAccept != "" {
		t.Errorf("upstream Accept = %q, want empty", gotAccept)
	}
	if gotUserAgent != PLAYER_USER_AGENT {
		t.Errorf("upstream User-Agent = %q, want %q", gotUserAgent, PLAYER_USER_AGENT)
	}
}
