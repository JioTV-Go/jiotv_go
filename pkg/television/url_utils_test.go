package television

import (
	"strings"
	"testing"

	"github.com/jiotv-go/jiotv_go/v3/pkg/secureurl"
	"github.com/jiotv-go/jiotv_go/v3/pkg/store"
)

func setupURLUtilsTest() {
	// Setup test environment with temporary pathPrefix
	_, err := store.SetupTestPathPrefix()
	if err != nil {
		panic("Failed to setup test environment")
	}
	// Initialize store for testing
	store.Init()
	// Initialize secureurl for URL encryption/decryption
	secureurl.Init()
}

func TestCreateEncryptedURL(t *testing.T) {
	setupURLUtilsTest()

	tests := []struct {
		name            string
		config          EncryptedURLConfig
		wantErr         bool
		wantContains    []string
		wantNotContains []string
	}{
		{
			name: "Basic URL with channel ID",
			config: EncryptedURLConfig{
				BaseURL:     "https://example.com",
				Match:       "/segment.m3u8",
				Params:      "token=abc123",
				ChannelID:   "123",
				EndpointURL: "/render.m3u8",
			},
			wantErr:      false,
			wantContains: []string{"/render.m3u8", "auth=", "channel_key_id=123"},
		},
		{
			name: "URL with quality parameter",
			config: EncryptedURLConfig{
				BaseURL:     "https://example.com",
				Match:       "/segment.ts",
				Params:      "token=def456",
				ChannelID:   "456",
				EndpointURL: "/render.ts",
				Quality:     "high",
			},
			wantErr:      false,
			wantContains: []string{"/render.ts", "auth=", "channel_key_id=456", "q=high"},
		},
		{
			name: "URL without channel ID",
			config: EncryptedURLConfig{
				BaseURL:     "https://example.com",
				Match:       "/file.key",
				Params:      "token=ghi789",
				EndpointURL: "/render.key",
			},
			wantErr:         false,
			wantContains:    []string{"/render.key", "auth="},
			wantNotContains: []string{"channel_key_id="},
		},
		{
			name: "URL without quality",
			config: EncryptedURLConfig{
				BaseURL:     "https://example.com",
				Match:       "/segment.aac",
				Params:      "token=jkl012",
				ChannelID:   "789",
				EndpointURL: "/render.ts",
			},
			wantErr:         false,
			wantContains:    []string{"/render.ts", "auth=", "channel_key_id=789"},
			wantNotContains: []string{"q="},
		},
		{
			name: "Empty base URL",
			config: EncryptedURLConfig{
				BaseURL:     "",
				Match:       "/segment.m3u8",
				Params:      "token=mno345",
				ChannelID:   "101",
				EndpointURL: "/render.m3u8",
			},
			wantErr:      false,
			wantContains: []string{"/render.m3u8", "auth=", "channel_key_id=101"},
		},
		{
			name: "Empty params",
			config: EncryptedURLConfig{
				BaseURL:     "https://example.com",
				Match:       "/segment.m3u8",
				Params:      "",
				ChannelID:   "202",
				EndpointURL: "/render.m3u8",
			},
			wantErr:      false,
			wantContains: []string{"/render.m3u8", "auth=", "channel_key_id=202"},
		},
		{
			name: "All parameters provided",
			config: EncryptedURLConfig{
				BaseURL:     "https://cdn.example.com",
				Match:       "/playlist.m3u8",
				Params:      "hdntl=exp=1234567890~acl=/*~data=hdntl~hmac=abc123",
				ChannelID:   "999",
				EndpointURL: "/render.m3u8",
				Quality:     "medium",
			},
			wantErr:      false,
			wantContains: []string{"/render.m3u8", "auth=", "channel_key_id=999", "q=medium"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateEncryptedURL(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateEncryptedURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return // Skip further checks if error was expected
			}

			gotStr := string(got)

			// Check that result contains expected strings
			for _, want := range tt.wantContains {
				if !strings.Contains(gotStr, want) {
					t.Errorf("CreateEncryptedURL() result should contain %q, got %q", want, gotStr)
				}
			}

			// Check that result does not contain unwanted strings
			for _, unwanted := range tt.wantNotContains {
				if strings.Contains(gotStr, unwanted) {
					t.Errorf("CreateEncryptedURL() result should not contain %q, got %q", unwanted, gotStr)
				}
			}

			// Check that auth parameter is not empty
			if strings.Contains(gotStr, "auth=&") || strings.Contains(gotStr, "auth=") && !strings.Contains(gotStr, "auth=") {
				// This check ensures auth parameter has a value
				authIndex := strings.Index(gotStr, "auth=")
				if authIndex != -1 {
					remaining := gotStr[authIndex+5:] // Skip "auth="
					if len(remaining) == 0 || remaining[0] == '&' {
						t.Errorf("CreateEncryptedURL() auth parameter should not be empty")
					}
				}
			}
		})
	}
}

func TestEncryptedURLConfig_Structure(t *testing.T) {
	// Test that the EncryptedURLConfig struct has all expected fields
	config := EncryptedURLConfig{
		BaseURL:     "test_base",
		Match:       "test_match",
		Params:      "test_params",
		ChannelID:   "test_channel",
		EndpointURL: "test_endpoint",
		Quality:     "test_quality",
	}

	// Verify all fields are accessible and have correct values
	if config.BaseURL != "test_base" {
		t.Errorf("Expected BaseURL to be 'test_base', got %s", config.BaseURL)
	}
	if config.Match != "test_match" {
		t.Errorf("Expected Match to be 'test_match', got %s", config.Match)
	}
	if config.Params != "test_params" {
		t.Errorf("Expected Params to be 'test_params', got %s", config.Params)
	}
	if config.ChannelID != "test_channel" {
		t.Errorf("Expected ChannelID to be 'test_channel', got %s", config.ChannelID)
	}
	if config.EndpointURL != "test_endpoint" {
		t.Errorf("Expected EndpointURL to be 'test_endpoint', got %s", config.EndpointURL)
	}
	if config.Quality != "test_quality" {
		t.Errorf("Expected Quality to be 'test_quality', got %s", config.Quality)
	}
}

func TestCreateEncryptedURL_Integration(t *testing.T) {
	setupURLUtilsTest()

	// Test the integration with the Replace functions
	t.Run("Integration with ReplaceM3U8 pattern", func(t *testing.T) {
		config := EncryptedURLConfig{
			BaseURL:     "https://jiotvapi.media.jio.com/v1",
			Match:       "/playlist.m3u8",
			Params:      "hdntl=exp=1640995200~acl=/*~data=hdntl~hmac=test",
			ChannelID:   "123",
			EndpointURL: "/render.m3u8",
			Quality:     "auto",
		}

		result, err := CreateEncryptedURL(config)
		if err != nil {
			t.Fatalf("CreateEncryptedURL() failed: %v", err)
		}

		resultStr := string(result)

		// Check the format matches what's expected by the handlers
		if !strings.HasPrefix(resultStr, "/render.m3u8?auth=") {
			t.Errorf("Result should start with '/render.m3u8?auth=', got %s", resultStr)
		}

		// Check that all required parameters are present
		requiredParams := []string{"auth=", "channel_key_id=123", "q=auto"}
		for _, param := range requiredParams {
			if !strings.Contains(resultStr, param) {
				t.Errorf("Result should contain %s, got %s", param, resultStr)
			}
		}
	})

	t.Run("Integration with ReplaceTS pattern", func(t *testing.T) {
		config := EncryptedURLConfig{
			BaseURL:     "https://jiotvapi.media.jio.com/v1",
			Match:       "/segment001.ts",
			Params:      "hdntl=exp=1640995200~acl=/*~data=hdntl~hmac=test",
			EndpointURL: "/render.ts",
		}

		result, err := CreateEncryptedURL(config)
		if err != nil {
			t.Fatalf("CreateEncryptedURL() failed: %v", err)
		}

		resultStr := string(result)

		// Check the format matches what's expected by the handlers
		if !strings.HasPrefix(resultStr, "/render.ts?auth=") {
			t.Errorf("Result should start with '/render.ts?auth=', got %s", resultStr)
		}

		// Should not contain channel_key_id or quality for TS files
		if strings.Contains(resultStr, "channel_key_id=") {
			t.Errorf("TS result should not contain channel_key_id, got %s", resultStr)
		}
		if strings.Contains(resultStr, "q=") {
			t.Errorf("TS result should not contain quality parameter, got %s", resultStr)
		}
	})

	t.Run("Integration with ReplaceKey pattern", func(t *testing.T) {
		config := EncryptedURLConfig{
			BaseURL:     "",
			Match:       "https://example.com/key.pkey",
			Params:      "token=abc123",
			ChannelID:   "456",
			EndpointURL: "/render.key",
		}

		result, err := CreateEncryptedURL(config)
		if err != nil {
			t.Fatalf("CreateEncryptedURL() failed: %v", err)
		}

		resultStr := string(result)

		// Check the format matches what's expected by the handlers
		if !strings.HasPrefix(resultStr, "/render.key?auth=") {
			t.Errorf("Result should start with '/render.key?auth=', got %s", resultStr)
		}

		// Should contain channel_key_id for key files
		if !strings.Contains(resultStr, "channel_key_id=456") {
			t.Errorf("Key result should contain channel_key_id=456, got %s", resultStr)
		}
	})
}

func TestCreateEncryptedURL_QueryJoinBehavior(t *testing.T) {
	setupURLUtilsTest()

	tests := []struct {
		name          string
		config        EncryptedURLConfig
		wantPlainAuth string
	}{
		{
			name: "No trailing question mark when params empty",
			config: EncryptedURLConfig{
				BaseURL:     "https://example.com",
				Match:       "/segment.m3u8",
				Params:      "",
				EndpointURL: "/render.m3u8",
			},
			wantPlainAuth: "https://example.com/segment.m3u8",
		},
		{
			name: "Append parent params when match already has query",
			config: EncryptedURLConfig{
				BaseURL:     "https://example.com",
				Match:       "/segment.m3u8?foo=1",
				Params:      "bar=2",
				EndpointURL: "/render.m3u8",
			},
			wantPlainAuth: "https://example.com/segment.m3u8?foo=1&bar=2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateEncryptedURL(tt.config)
			if err != nil {
				t.Fatalf("CreateEncryptedURL() failed: %v", err)
			}

			gotStr := string(got)
			authStart := strings.Index(gotStr, "auth=")
			if authStart == -1 {
				t.Fatalf("auth param missing in result: %s", gotStr)
			}

			authValue := gotStr[authStart+5:]
			if amp := strings.Index(authValue, "&"); amp != -1 {
				authValue = authValue[:amp]
			}

			plain, err := secureurl.DecryptURL(authValue)
			if err != nil {
				t.Fatalf("DecryptURL() failed: %v", err)
			}

			if plain != tt.wantPlainAuth {
				t.Fatalf("decrypted auth URL mismatch: got %q want %q", plain, tt.wantPlainAuth)
			}
		})
	}
}
