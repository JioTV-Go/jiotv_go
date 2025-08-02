package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

// MockServer provides HTTP mocking functionality for tests
type MockServer struct {
	Server *httptest.Server
	URLs   map[string]string
}

// NewMockServer creates a new mock server with all JioTV API endpoints
func NewMockServer() *MockServer {
	mux := http.NewServeMux()
	
	// Mock LoginSendOTP endpoint
	mux.HandleFunc("/userservice/apis/v1/loginotp/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		// Check headers
		if r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		
		// If number is empty or invalid, return error
		if payload["number"] == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error": "Invalid number"}`)
			return
		}
		
		w.WriteHeader(http.StatusNoContent)
	})
	
	// Mock LoginVerifyOTP endpoint
	mux.HandleFunc("/userservice/apis/v1/loginotp/verify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		var payload LoginOTPPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		
		// If number or OTP is empty, return error
		if payload.Number == "" || payload.OTP == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error": "Invalid credentials"}`)
			return
		}
		
		// Mock successful response
		response := LoginResponse{
			AuthToken:    "mock_auth_token",
			RefreshToken: "mock_refresh_token",
			SSOToken:     "mock_sso_token",
		}
		response.SessionAttributes.User.SubscriberID = "mock_crm"
		response.SessionAttributes.User.Unique = "mock_unique_id"
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	
	// Mock Login (password) endpoint
	mux.HandleFunc("/v3/dip/user/unpw/verify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		var payload LoginPasswordPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		
		// If credentials are empty, return error
		if payload.Identifier == "" || payload.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error": "Invalid credentials"}`)
			return
		}
		
		// Mock successful response
		response := LoginResponse{
			SSOToken: "mock_sso_token",
		}
		response.SessionAttributes.User.SubscriberID = "mock_crm"
		response.SessionAttributes.User.Unique = "mock_unique_id"
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	
	// Mock server logout endpoint
	mux.HandleFunc("/tokenservice/apis/v1/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		
		// If refreshToken is missing, return error
		if payload["refreshToken"] == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error": "RefreshToken missing"}`)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status": "success"}`)
	})
	
	// Mock channels list endpoint (for EPG and television)
	mux.HandleFunc("/apis/v3.0/getMobileChannelList/get/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		// Mock channels response
		response := map[string]interface{}{
			"code": 200,
			"result": []map[string]interface{}{
				{
					"channel_id":      1,
					"channel_name":    "Mock Channel 1",
					"channel_url":     "mock_url_1",
					"language_id":     1,
					"category_id":     5,
					"isHD":            true,
				},
				{
					"channel_id":      2,
					"channel_name":    "Mock Channel 2", 
					"channel_url":     "mock_url_2",
					"language_id":     6,
					"category_id":     6,
					"isHD":            false,
				},
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	
	// Mock EPG endpoint
	mux.HandleFunc("/apis/v1.3/getepg/get/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		// Mock EPG response
		response := map[string]interface{}{
			"code": 200,
			"epg": []map[string]interface{}{
				{
					"startEpoch":      1640995200000, // 2022-01-01 00:00:00 UTC
					"endEpoch":        1640998800000, // 2022-01-01 01:00:00 UTC
					"title":           "Mock Program 1",
					"description":     "Mock Program Description 1",
					"showCategory":    "Entertainment",
					"poster":          "mock_poster_1.jpg",
				},
				{
					"startEpoch":      1640998800000, // 2022-01-01 01:00:00 UTC
					"endEpoch":        1641002400000, // 2022-01-01 02:00:00 UTC
					"title":           "Mock Program 2",
					"description":     "Mock Program Description 2",
					"showCategory":    "News",
					"poster":          "mock_poster_2.jpg",
				},
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	
	// Mock Live TV URL endpoint (JioTV API v1)
	mux.HandleFunc("/playback/apis/v1/geturl", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		// Mock live URL response
		response := map[string]interface{}{
			"code": 200,
			"result": "https://mock.streaming.url/live.m3u8",
			"bitrates": map[string]string{
				"auto": "https://mock.streaming.url/live.m3u8",
				"high": "https://mock.streaming.url/live_high.m3u8",
				"low":  "https://mock.streaming.url/live_low.m3u8",
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	
	// Mock Live TV URL endpoint (JioTV API v2.2)  
	mux.HandleFunc("/apis/v2.2/getchannelurl/getchannelurl", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		// Mock live URL response
		response := map[string]interface{}{
			"code": 200,
			"result": "https://mock.streaming.url/live.m3u8",
			"bitrates": map[string]string{
				"auto": "https://mock.streaming.url/live.m3u8",
				"high": "https://mock.streaming.url/live_high.m3u8",
				"low":  "https://mock.streaming.url/live_low.m3u8",
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	
	server := httptest.NewServer(mux)
	
	// Create URL mappings for each domain
	serverURL, _ := url.Parse(server.URL)
	urls := map[string]string{
		API_JIO_DOMAIN:   fmt.Sprintf("%s://%s", serverURL.Scheme, serverURL.Host),
		"jiotv.data.cdn.jio.com": fmt.Sprintf("%s://%s", serverURL.Scheme, serverURL.Host),
		JIOTV_API_DOMAIN: fmt.Sprintf("%s://%s", serverURL.Scheme, serverURL.Host),
		"tv.media.jio.com":       fmt.Sprintf("%s://%s", serverURL.Scheme, serverURL.Host),
		AUTH_MEDIA_DOMAIN:        fmt.Sprintf("%s://%s", serverURL.Scheme, serverURL.Host),
		"jiotvapi.cdn.jio.com":   fmt.Sprintf("%s://%s", serverURL.Scheme, serverURL.Host),
	}
	
	return &MockServer{
		Server: server,
		URLs:   urls,
	}
}

// Close closes the mock server
func (m *MockServer) Close() {
	m.Server.Close()
}

// ReplaceURLs replaces API URLs in the given text with mock server URLs
func (m *MockServer) ReplaceURLs(text string) string {
	for domain, mockURL := range m.URLs {
		text = strings.ReplaceAll(text, "https://"+domain, mockURL)
		text = strings.ReplaceAll(text, "http://"+domain, mockURL)
	}
	return text
}