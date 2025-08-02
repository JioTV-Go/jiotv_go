package epg

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
)

// convertToUint16WithBounds safely converts a string to uint16 with bounds checking
// This helper function encapsulates the repeated pattern of int conversion with bounds checking
func convertToUint16WithBounds(s string) uint16 {
	intVal, err := strconv.Atoi(s)
	if err == nil && intVal >= 0 && intVal <= int(math.MaxUint16) {
		return uint16(intVal)
	}
	return 0 // Default value for invalid input
}

// MockEPGServer provides HTTP mocking functionality for EPG tests
type MockEPGServer struct {
	Server *httptest.Server
	URLs   map[string]string
}

// NewMockEPGServer creates a new mock server for EPG endpoints
func NewMockEPGServer() *MockEPGServer {
	mux := http.NewServeMux()
	
	// Mock channels list endpoint
	mux.HandleFunc("/apis/v3.0/getMobileChannelList/get/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		// Mock channels response
		response := ChannelsResponse{
			Code: 200,
			Channels: []ChannelObject{
				{
					ChannelID:   1,
					ChannelName: "Mock Channel 1",
					LogoURL:     "https://example.com/logo1.png",
				},
				{
					ChannelID:   2,
					ChannelName: "Mock Channel 2",
					LogoURL:     "https://example.com/logo2.png",
				},
				{
					ChannelID:   3,
					ChannelName: "Mock Channel 3",
					LogoURL:     "https://example.com/logo3.png",
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
		
		// Parse query parameters
		offset := r.URL.Query().Get("offset")
		channelID := r.URL.Query().Get("channel_id")
		
		offsetInt, err := strconv.Atoi(offset)
		if err != nil {
			offsetInt = 0 // or some other default value
		}
		channelIDUint16 := convertToUint16WithBounds(channelID)
		channelIDInt := int(channelIDUint16) // For display purposes
		
		// Mock EPG response with different data based on offset and channel
		var epgData []EPGObject
		
		// Base timestamp: 2022-01-01 00:00:00 UTC
		baseTimestamp := int64(1640995200000)
		
		// Generate mock programmes based on offset and channel
		for i := 0; i < 3; i++ {
			programIndex := offsetInt*3 + i
			startTime := baseTimestamp + int64(programIndex*3600*1000) // Each program 1 hour apart
			endTime := startTime + 3600*1000 // 1 hour duration
			
			epgData = append(epgData, EPGObject{
				StartEpoch:   startTime,
				EndEpoch:     endTime,
				ChannelID:    channelIDUint16,
				ChannelName:  fmt.Sprintf("Mock Channel %d", channelIDInt),
				ShowCategory: "Entertainment",
				Description:  fmt.Sprintf("Mock Program %d Description for Channel %d", programIndex+1, channelIDInt),
				Title:        fmt.Sprintf("Mock Program %d", programIndex+1),
				Thumbnail:    fmt.Sprintf("mock_thumb_%d.jpg", programIndex+1),
				Poster:       fmt.Sprintf("mock_poster_%d.jpg", programIndex+1),
			})
		}
		
		response := EPGResponse{
			EPG: epgData,
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	
	server := httptest.NewServer(mux)
	
	// Create URL mappings
	serverURL, _ := url.Parse(server.URL)
	urls := map[string]string{
		"jiotv.data.cdn.jio.com": fmt.Sprintf("%s://%s", serverURL.Scheme, serverURL.Host),
	}
	
	return &MockEPGServer{
		Server: server,
		URLs:   urls,
	}
}

// Close closes the mock server
func (m *MockEPGServer) Close() {
	m.Server.Close()
}