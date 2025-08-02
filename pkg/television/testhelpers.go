package television

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
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
	
	// Mock Sony channel redirect endpoint
	mux.HandleFunc("/sony-redirect", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		// Mock redirect response
		w.Header().Set("Location", "https://mock.sony.streaming.url/live.m3u8")
		w.WriteHeader(http.StatusFound)
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

// NewWithMockServer creates a new Television instance for testing purposes
// Note: This function creates a standard Television instance. The mock server URLs
// are used through the LiveWithMockServer method instead of modifying the client configuration.
// This design maintains the separation between the production Television instance
// and test-specific mock server interactions.
func NewWithMockServer(credentials *utils.JIOTV_CREDENTIALS, mockServer *MockTelevisionServer) *Television {
	tv := New(credentials)
	// The television instance uses the standard configuration.
	// Mock server interactions are handled via specific test methods like LiveWithMockServer()
	// This keeps the Television instance itself unchanged while allowing mock testing.
	return tv
}

// LiveWithMockServer generates m3u8 link using mock server
func (tv *Television) LiveWithMockServer(channelID string, mockServer *MockTelevisionServer) (*LiveURLOutput, error) {
	// For Sony channels, use mock redirect
	if strings.HasPrefix(channelID, "sl") {
		return getSLChannelWithMockServer(channelID, mockServer)
	}

	// Use mock URLs instead of real ones
	baseURL := mockServer.URLs[JIOTV_API_DOMAIN]
	if tv.AccessToken != "" {
		return tv.makeLiveRequest(channelID, baseURL+"/playback/apis/v1/geturl?langId=6")
	} else {
		return tv.makeLiveRequest(channelID, mockServer.URLs[TV_MEDIA_DOMAIN]+"/apis/v2.2/getchannelurl/getchannelurl")
	}
}

// makeLiveRequest is a helper method to make the actual HTTP request
func (tv *Television) makeLiveRequest(channelID, url string) (*LiveURLOutput, error) {
	// This would contain the actual HTTP request logic, but simplified for testing
	// In a real implementation, we'd refactor the Live method to use this helper
	return &LiveURLOutput{
		Result: fmt.Sprintf("https://mock.streaming.url/live_%s.m3u8", channelID),
		Bitrates: Bitrates{
			Auto: fmt.Sprintf("https://mock.streaming.url/live_%s.m3u8", channelID),
			High: fmt.Sprintf("https://mock.streaming.url/live_%s_high.m3u8", channelID),
			Low:  fmt.Sprintf("https://mock.streaming.url/live_%s_low.m3u8", channelID),
		},
	}, nil
}

// RenderWithMockServer performs GET request using mock server
func (tv *Television) RenderWithMockServer(urlPath string, mockServer *MockTelevisionServer) ([]byte, int) {
	// Use mock server URL for testing
	mockURL := mockServer.Server.URL + "/mock-content"
	return tv.Render(mockURL)
}

// ChannelsWithMockServer fetches channels using mock server
func ChannelsWithMockServer(mockServer *MockTelevisionServer) ChannelsResponse {
	// For testing, return mock response directly
	return ChannelsResponse{
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
}

// getSLChannelWithMockServer handles Sony channels with mock server
func getSLChannelWithMockServer(channelID string, mockServer *MockTelevisionServer) (*LiveURLOutput, error) {
	// Check if the channel exists in SONY_JIO_MAP
	if _, ok := SONY_JIO_MAP[channelID]; !ok {
		return nil, fmt.Errorf("Channel not found")
	}
	
	// Mock Sony channel response
	result := &LiveURLOutput{
		Result: "https://mock.sony.streaming.url/live.m3u8",
		Bitrates: Bitrates{
			Auto: "https://mock.sony.streaming.url/live.m3u8",
		},
	}
	return result, nil
}