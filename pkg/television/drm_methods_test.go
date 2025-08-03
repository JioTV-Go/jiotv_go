package television

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jiotv-go/jiotv_go/v3/pkg/store"
	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
)

func TestTelevision_RequestDRMKey(t *testing.T) {
	// Initialize required dependencies
	store.Init()
	utils.Log = utils.GetLogger()

	tests := []struct {
		name       string
		url        string
		channelID  string
		serverResp string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "Successful DRM key request",
			url:        "http://mockserver/drm",
			channelID:  "123",
			serverResp: "mock-drm-key-data",
			statusCode: 200,
			wantErr:    false,
		},
		{
			name:       "DRM key request with 403 status",
			url:        "http://mockserver/drm",
			channelID:  "123",
			serverResp: "Forbidden",
			statusCode: 403,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify it's a POST request
				if r.Method != "POST" {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				
				// Verify headers are set correctly
				if r.Header.Get("channelid") != tt.channelID {
					t.Errorf("Expected channelid header %s, got %s", tt.channelID, r.Header.Get("channelid"))
				}
				
				if r.Header.Get("Content-Type") != "application/octet-stream" {
					t.Errorf("Expected Content-Type application/octet-stream, got %s", r.Header.Get("Content-Type"))
				}

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.serverResp))
			}))
			defer server.Close()

			// Create test TV instance
			tv := &Television{
				AccessToken: "test-token",
				SsoToken:    "test-sso",
				Crm:         "test-crm",
				UniqueID:    "test-unique",
				Client:      utils.GetRequestClient(),
			}

			// Make request
			body, statusCode, err := tv.RequestDRMKey(server.URL, tt.channelID)

			if (err != nil) != tt.wantErr {
				t.Errorf("RequestDRMKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if statusCode != tt.statusCode {
				t.Errorf("RequestDRMKey() statusCode = %v, want %v", statusCode, tt.statusCode)
			}

			if string(body) != tt.serverResp {
				t.Errorf("RequestDRMKey() body = %v, want %v", string(body), tt.serverResp)
			}
		})
	}
}

func TestTelevision_RequestMPD(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		serverResp string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "Successful MPD request",
			url:        "http://mockserver/mpd",
			serverResp: "<MPD><Period></Period></MPD>",
			statusCode: 200,
			wantErr:    false,
		},
		{
			name:       "MPD request with 403 status",
			url:        "http://mockserver/mpd",
			serverResp: "Forbidden",
			statusCode: 403,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify it's a GET request
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				
				// Verify User-Agent header
				expectedUA := "plaYtv/7.1.3 (Linux;Android 13) ExoPlayerLib/2.11.7"
				if r.Header.Get("User-Agent") != expectedUA {
					t.Errorf("Expected User-Agent %s, got %s", expectedUA, r.Header.Get("User-Agent"))
				}

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.serverResp))
			}))
			defer server.Close()

			// Create test TV instance
			tv := &Television{
				Client: utils.GetRequestClient(),
			}

			// Make request
			body, statusCode, err := tv.RequestMPD(server.URL)

			if (err != nil) != tt.wantErr {
				t.Errorf("RequestMPD() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if statusCode != tt.statusCode {
				t.Errorf("RequestMPD() statusCode = %v, want %v", statusCode, tt.statusCode)
			}

			if string(body) != tt.serverResp {
				t.Errorf("RequestMPD() body = %v, want %v", string(body), tt.serverResp)
			}
		})
	}
}

func TestTelevision_RequestDashSegment(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		serverResp string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "Successful DASH segment request",
			url:        "http://mockserver/segment.m4s",
			serverResp: "binary-segment-data",
			statusCode: 200,
			wantErr:    false,
		},
		{
			name:       "DASH segment request with 404 status",
			url:        "http://mockserver/segment.m4s",
			serverResp: "Not Found",
			statusCode: 404,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify it's a GET request
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				
				// Verify User-Agent header
				expectedUA := "plaYtv/7.1.3 (Linux;Android 13) ExoPlayerLib/2.11.7"
				if r.Header.Get("User-Agent") != expectedUA {
					t.Errorf("Expected User-Agent %s, got %s", expectedUA, r.Header.Get("User-Agent"))
				}

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.serverResp))
			}))
			defer server.Close()

			// Create test TV instance
			tv := &Television{
				Client: utils.GetRequestClient(),
			}

			// Make request
			body, statusCode, err := tv.RequestDashSegment(server.URL)

			if (err != nil) != tt.wantErr {
				t.Errorf("RequestDashSegment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if statusCode != tt.statusCode {
				t.Errorf("RequestDashSegment() statusCode = %v, want %v", statusCode, tt.statusCode)
			}

			if string(body) != tt.serverResp {
				t.Errorf("RequestDashSegment() body = %v, want %v", string(body), tt.serverResp)
			}
		})
	}
}

func Test_generateDateTime(t *testing.T) {
	t.Run("Generate datetime string", func(t *testing.T) {
		result1 := generateDateTime()
		result2 := generateDateTime()
		
		// Should be valid format (13 digits: YYMMDDHHMM + 3 digit milliseconds)
		if len(result1) != 13 {
			t.Errorf("generateDateTime() length = %d, want 13", len(result1))
		}
		
		// Should contain only digits
		for _, r := range result1 {
			if r < '0' || r > '9' {
				t.Errorf("generateDateTime() contains non-digit character: %c", r)
			}
		}
		
		// Two calls should be different (due to millisecond precision)
		if result1 == result2 {
			t.Logf("generateDateTime() returned same value twice: %s", result1)
		}
	})
}