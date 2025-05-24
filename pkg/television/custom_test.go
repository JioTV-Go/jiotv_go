package television

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jiotv-go/jiotv_go/v3/internal/config"
)

func TestIsCustomChannel(t *testing.T) {
	tests := []struct {
		channelID string
		expected  bool
	}{
		{"custom_test1", true},
		{"custom_", true},
		{"test1", false},
		{"123", false},
		{"sl291", false},
		{"", false},
	}

	for _, test := range tests {
		result := IsCustomChannel(test.channelID)
		if result != test.expected {
			t.Errorf("IsCustomChannel(%s) = %v, expected %v", test.channelID, result, test.expected)
		}
	}
}

func TestGetCustomChannelURL(t *testing.T) {
	channels := []Channel{
		{ID: "custom_test1", URL: "https://example.com/test1.m3u8"},
		{ID: "custom_test2", URL: "https://example.com/test2.m3u8"},
		{ID: "regular_channel", URL: "https://jio.com/regular.m3u8"},
	}

	tests := []struct {
		channelID    string
		expectedURL  string
		expectError  bool
	}{
		{"custom_test1", "https://example.com/test1.m3u8", false},
		{"custom_test2", "https://example.com/test2.m3u8", false},
		{"custom_nonexistent", "", true},
		{"regular_channel", "https://jio.com/regular.m3u8", false},
	}

	for _, test := range tests {
		url, err := GetCustomChannelURL(test.channelID, channels)
		if test.expectError {
			if err == nil {
				t.Errorf("GetCustomChannelURL(%s) expected error but got none", test.channelID)
			}
		} else {
			if err != nil {
				t.Errorf("GetCustomChannelURL(%s) unexpected error: %v", test.channelID, err)
			}
			if url != test.expectedURL {
				t.Errorf("GetCustomChannelURL(%s) = %s, expected %s", test.channelID, url, test.expectedURL)
			}
		}
	}
}

func TestConvertCustomChannels(t *testing.T) {
	customChannels := []CustomChannel{
		{
			ID:       "test1",
			Name:     "Test Channel 1",
			URL:      "https://example.com/test1.m3u8",
			LogoURL:  "https://example.com/logo1.png",
			Category: 5,
			Language: 6,
			IsHD:     true,
		},
		{
			ID:       "custom_test2", // Already has custom_ prefix
			Name:     "Test Channel 2",
			URL:      "https://example.com/test2.m3u8",
			Category: 8,
			Language: 1,
			IsHD:     false,
		},
		{
			// Missing required fields (should be skipped)
			ID:   "",
			Name: "Invalid Channel",
		},
	}

	result := convertCustomChannels(customChannels, "test_source")

	// Should have 2 valid channels (the third is invalid)
	if len(result) != 2 {
		t.Errorf("Expected 2 channels, got %d", len(result))
	}

	// Check first channel
	if result[0].ID != "custom_test1" {
		t.Errorf("Expected ID custom_test1, got %s", result[0].ID)
	}
	if result[0].Name != "Test Channel 1" {
		t.Errorf("Expected name 'Test Channel 1', got %s", result[0].Name)
	}

	// Check second channel (should keep existing custom_ prefix)
	if result[1].ID != "custom_test2" {
		t.Errorf("Expected ID custom_test2, got %s", result[1].ID)
	}
}

func TestLoadFromJSON(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "jiotv_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create test JSON file
	testConfig := CustomChannelConfig{
		Channels: []CustomChannel{
			{
				ID:       "test1",
				Name:     "Test Channel 1",
				URL:      "https://example.com/test1.m3u8",
				Category: 5,
				Language: 6,
				IsHD:     true,
			},
		},
	}

	jsonData, err := json.Marshal(testConfig)
	if err != nil {
		t.Fatal(err)
	}

	testFile := filepath.Join(tempDir, "test-channels.json")
	err = os.WriteFile(testFile, jsonData, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test loading
	channels := loadFromJSON(testFile)
	if len(channels) != 1 {
		t.Errorf("Expected 1 channel, got %d", len(channels))
	}

	if channels[0].ID != "custom_test1" {
		t.Errorf("Expected ID custom_test1, got %s", channels[0].ID)
	}

	// Test loading non-existent file
	nonExistentChannels := loadFromJSON(filepath.Join(tempDir, "nonexistent.json"))
	if len(nonExistentChannels) != 0 {
		t.Errorf("Expected 0 channels for non-existent file, got %d", len(nonExistentChannels))
	}
}

func TestLoadCustomChannelsWithConfig(t *testing.T) {
	// Save original config
	originalConfig := config.Cfg
	defer func() { config.Cfg = originalConfig }()

	// Test with custom channels disabled
	config.Cfg.DisableCustomChannels = true
	channels := LoadCustomChannels()
	if len(channels) != 0 {
		t.Errorf("Expected 0 channels when disabled, got %d", len(channels))
	}

	// Test with custom channels enabled (no actual files)
	config.Cfg.DisableCustomChannels = false
	channels = LoadCustomChannels()
	// Should return empty slice since no files exist, but should not error
	if channels == nil {
		t.Error("Expected empty slice, got nil")
	}
}