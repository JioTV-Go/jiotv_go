package television

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
)

// MockTelevisionServer provides HTTP mocking functionality for television tests
type MockTelevisionServer struct {
	Server *httptest.Server
	URLs   map[string]string
}

// NewMockTelevisionServer creates a new mock server for television endpoints
func NewMockTelevisionServer() *MockTelevisionServer {
	mux := http.NewServeMux()
	
	// Mock JioTV API v1 (with access token)
	mux.HandleFunc("/playback/apis/v1/geturl", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		// Parse form data
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		
		channelID := r.FormValue("channel_id")
		if channelID == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error": "Missing channel_id"}`)
			return
		}
		
		// Mock live URL response
		response := LiveURLOutput{
			Result: fmt.Sprintf("https://mock.streaming.url/live_%s.m3u8", channelID),
			Bitrates: Bitrates{
				Auto: fmt.Sprintf("https://mock.streaming.url/live_%s.m3u8", channelID),
				High: fmt.Sprintf("https://mock.streaming.url/live_%s_high.m3u8", channelID),
				Low:  fmt.Sprintf("https://mock.streaming.url/live_%s_low.m3u8", channelID),
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	
	// Mock JioTV API v2.2 (with SSO token)
	mux.HandleFunc("/apis/v2.2/getchannelurl/getchannelurl", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		// Parse form data
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		
		channelID := r.FormValue("channel_id")
		if channelID == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error": "Missing channel_id"}`)
			return
		}
		
		// Mock live URL response
		response := LiveURLOutput{
			Result: fmt.Sprintf("https://mock.streaming.url/live_%s.m3u8", channelID),
			Bitrates: Bitrates{
				Auto: fmt.Sprintf("https://mock.streaming.url/live_%s.m3u8", channelID),
				High: fmt.Sprintf("https://mock.streaming.url/live_%s_high.m3u8", channelID),
				Low:  fmt.Sprintf("https://mock.streaming.url/live_%s_low.m3u8", channelID),
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	
	// Mock channels list endpoint
	mux.HandleFunc("/apis/v3.0/getMobileChannelList/get/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		// Mock channels response
		response := ChannelsResponse{
			Result: []Channel{
				{
					ID:       "1",
					Name:     "Mock Channel 1",
					Language: 1, // Hindi
					Category: 5, // Entertainment
				},
				{
					ID:       "2",
					Name:     "Mock Channel 2",
					Language: 6, // English
					Category: 6, // Movies
				},
				{
					ID:       "3",
					Name:     "Mock Channel 3",
					Language: 8, // Tamil
					Category: 5, // Entertainment
				},
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	
	// Mock render endpoint for testing arbitrary URL rendering
	mux.HandleFunc("/mock-content", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		fmt.Fprint(w, "#EXTM3U\n#EXT-X-VERSION:3\n#EXTINF:10.0,\nsegment1.ts\n#EXTINF:10.0,\nsegment2.ts\n")
	})
	
	server := httptest.NewServer(mux)
	
	// Create URL mappings
	serverURL, _ := url.Parse(server.URL)
	urls := map[string]string{
		JIOTV_API_DOMAIN: fmt.Sprintf("%s://%s", serverURL.Scheme, serverURL.Host),
		TV_MEDIA_DOMAIN:  fmt.Sprintf("%s://%s", serverURL.Scheme, serverURL.Host),
		JIOTV_CDN_DOMAIN: fmt.Sprintf("%s://%s", serverURL.Scheme, serverURL.Host),
	}
	
	return &MockTelevisionServer{
		Server: server,
		URLs:   urls,
	}
}

// Close closes the mock server
func (m *MockTelevisionServer) Close() {
	m.Server.Close()
}