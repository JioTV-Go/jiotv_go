package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jiotv-go/jiotv_go/v3/pkg/television"
	"github.com/valyala/fasthttp"
)

// istDateString formats the IST calendar date containing the given instant the
// way the getepg API stamps serverDate, e.g. "2026-08-13T00:00:00+05:30".
func istDateString(ms int64) string {
	ist := time.UnixMilli(ms).Add(5*time.Hour + 30*time.Minute).UTC()
	return fmt.Sprintf("%04d-%02d-%02dT00:00:00+05:30", ist.Year(), ist.Month(), ist.Day())
}

// epgJSONBody builds a minimal getepg-style response with one show entry.
func epgJSONBody(serverDate, showName string) []byte {
	now := time.Now().UnixMilli()
	b, _ := json.Marshal(map[string]any{
		"epg": []map[string]any{
			{
				"showname":   showName,
				"serverDate": serverDate,
				"startEpoch": now - 600000,
				"endEpoch":   now + 600000,
			},
		},
	})
	return b
}

func TestWebEPGDayOffset(t *testing.T) {
	now := time.Now().UnixMilli()
	tests := []struct {
		name string
		body []byte
		want int
	}{
		{
			name: "server a day behind returns 1",
			body: epgJSONBody(istDateString(now-24*time.Hour.Milliseconds()), "Yesterday Show"),
			want: 1,
		},
		{
			name: "server date today returns 0",
			body: epgJSONBody(istDateString(now), "Today Show"),
			want: 0,
		},
		{
			name: "server date in the future returns 0",
			body: epgJSONBody(istDateString(now+24*time.Hour.Milliseconds()), "Tomorrow Show"),
			want: 0,
		},
		{
			name: "empty epg array returns 0",
			body: []byte(`{"epg":[]}`),
			want: 0,
		},
		{
			name: "invalid json returns 0",
			body: []byte(`not-json`),
			want: 0,
		},
		{
			name: "missing serverDate returns 0",
			body: []byte(`{"epg":[{"showname":"No Date"}]}`),
			want: 0,
		},
		{
			name: "malformed serverDate returns 0",
			body: []byte(`{"epg":[{"serverDate":"not-a-date"}]}`),
			want: 0,
		},
		{
			name: "nil body returns 0",
			body: nil,
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := webEPGDayOffset(tt.body); got != tt.want {
				t.Errorf("webEPGDayOffset() = %d, want %d", got, tt.want)
			}
		})
	}
}

// setupWebEPGTest installs the global state WebEPGHandler needs (TV client and
// the upstream URL format) and returns a cleanup function.
func setupWebEPGTest(t *testing.T, upstreamURL string) func() {
	t.Helper()
	previousTV, previousURL := TV, webEPGURLFormat
	TV = &television.Television{
		Client: &fasthttp.Client{},
	}
	webEPGURLFormat = upstreamURL + "/getepg?offset=%d&channel_id=%d"
	return func() {
		TV = previousTV
		webEPGURLFormat = previousURL
	}
}

// epgUpstream returns a fake getepg server. offset 0 serves yesterday's
// schedule, any other offset serves today's schedule, so a corrected request
// (offset 1) is distinguishable from the original.
func epgUpstream(t *testing.T) (*httptest.Server, *[]int, *sync.Mutex) {
	t.Helper()
	var mu sync.Mutex
	var offsets []int

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		off, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		mu.Lock()
		offsets = append(offsets, off)
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		now := time.Now().UnixMilli()
		if off == 0 {
			_, _ = w.Write(epgJSONBody(istDateString(now-24*time.Hour.Milliseconds()), "Yesterday Show"))
			return
		}
		_, _ = w.Write(epgJSONBody(istDateString(now), "Today Show"))
	}))
	t.Cleanup(upstream.Close)
	return upstream, &offsets, &mu
}

func TestWebEPGHandlerCorrectsDayOffset(t *testing.T) {
	upstream, offsets, mu := epgUpstream(t)
	defer setupWebEPGTest(t, upstream.URL)()

	app := fiber.New()
	app.Get("/epg/:channelID/:offset", WebEPGHandler)

	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/epg/336/0", nil))
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
	if !strings.Contains(string(body), "Today Show") {
		t.Errorf("body serves yesterday's schedule, want corrected (today's) schedule: %s", body)
	}
	if strings.Contains(string(body), "Yesterday Show") {
		t.Errorf("body contains yesterday's schedule: %s", body)
	}
	if ct := response.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	mu.Lock()
	got := append([]int(nil), *offsets...)
	mu.Unlock()
	if len(got) != 2 || got[0] != 0 || got[1] != 1 {
		t.Errorf("upstream offsets = %v, want [0 1] (original request plus one corrected)", got)
	}
}

func TestWebEPGHandlerKeepsCorrectOffset(t *testing.T) {
	// When the API already serves today's schedule, no corrected re-fetch
	// should happen.
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(epgJSONBody(istDateString(time.Now().UnixMilli()), "Today Show"))
	}))
	defer upstream.Close()
	defer setupWebEPGTest(t, upstream.URL)()

	app := fiber.New()
	app.Get("/epg/:channelID/:offset", WebEPGHandler)

	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/epg/336/1", nil))
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
	if !strings.Contains(string(body), "Today Show") {
		t.Errorf("body = %s, want today's schedule unchanged", body)
	}
}

func TestWebEPGHandlerFallsBackWhenCorrectionFails(t *testing.T) {
	// If the corrected (offset 1) fetch fails, the handler must serve the
	// original response instead of erroring.
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		off, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		w.Header().Set("Content-Type", "application/json")
		if off == 0 {
			_, _ = w.Write(epgJSONBody(istDateString(time.Now().UnixMilli()-24*time.Hour.Milliseconds()), "Yesterday Show"))
			return
		}
		http.Error(w, "upstream failure", http.StatusInternalServerError)
	}))
	defer upstream.Close()
	defer setupWebEPGTest(t, upstream.URL)()

	app := fiber.New()
	app.Get("/epg/:channelID/:offset", WebEPGHandler)

	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/epg/336/0", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d (original response served on correction failure)", response.StatusCode, http.StatusOK)
	}
	if !strings.Contains(string(body), "Yesterday Show") {
		t.Errorf("body = %s, want the original (uncorrected) response", body)
	}
}

func TestWebEPGHandlerBadRequest(t *testing.T) {
	defer setupWebEPGTest(t, "http://127.0.0.1:1")()

	app := fiber.New()
	app.Get("/epg/:channelID/:offset", WebEPGHandler)

	for _, path := range []string{"/epg/abc/0", "/epg/336/abc"} {
		response, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want %d", path, response.StatusCode, http.StatusBadRequest)
		}
	}
}
