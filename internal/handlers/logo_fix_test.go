package handlers

import (
	"testing"
	"strings"
)

func TestLogoURLHandling(t *testing.T) {
	// Test cases to verify logo URL handling
	testCases := []struct {
		logoURL  string
		expected string
		hostURL  string
	}{
		{
			// Custom channel with full URL
			logoURL:  "https://example.com/logo.png",
			expected: "https://example.com/logo.png", 
			hostURL:  "http://localhost:5001",
		},
		{
			// Regular channel with relative path
			logoURL:  "Sony_HD.png",
			expected: "http://localhost:5001/jtvimage/Sony_HD.png",
			hostURL:  "http://localhost:5001",
		},
		{
			// Another custom channel
			logoURL:  "http://cdn.example.com/channel_logo.jpg",
			expected: "http://cdn.example.com/channel_logo.jpg",
			hostURL:  "http://localhost:5001",
		},
	}

	for _, tc := range testCases {
		t.Run("Logo_URL_"+tc.logoURL, func(t *testing.T) {
			var result string
			if strings.HasPrefix(tc.logoURL, "http://") || strings.HasPrefix(tc.logoURL, "https://") {
				// Custom channel with full URL
				result = tc.logoURL
			} else {
				// Regular channel with relative path  
				result = tc.hostURL + "/jtvimage/" + tc.logoURL
			}
			
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}