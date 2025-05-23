package television

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper function to create a temporary file with given content.
// Returns the path to the temporary file.
func createTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir() // Go 1.15+ creates a temporary directory cleaned up automatically
	tmpFile, err := os.CreateTemp(dir, "custom_channels_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	return tmpFile.Name()
}

func TestLoadCustomChannels(t *testing.T) {
	// Test Case 1: Valid JSON file
	t.Run("ValidJSONFile", func(t *testing.T) {
		validJSONContent := `{
			"channels": [
				{
					"ID": "test_001",
					"Name": "Test Channel 1",
					"LogoURL": "http://logo.url/tc1.png",
					"Category": "Entertainment",
					"Language": "English",
					"URL": "http://stream.url/tc1.m3u8",
					"EPGID": "tc1.epg"
				},
				{
					"ID": "test_002",
					"Name": "Test Channel 2",
					"LogoURL": "relative/logo.png",
					"Category": "News",
					"Language": "Hindi",
					"URL": "http://stream.url/tc2.m3u8"
				}
			]
		}`
		tmpFilePath := createTempFile(t, validJSONContent)
		// No explicit defer os.Remove(tmpFilePath) needed due to t.TempDir()

		channels, err := LoadCustomChannels(tmpFilePath)
		if err != nil {
			t.Errorf("LoadCustomChannels() with valid JSON returned error: %v", err)
		}
		if len(channels) != 2 {
			t.Errorf("Expected 2 channels, got %d", len(channels))
		}

		// Assertions for channel 1
		if len(channels) > 0 {
			ch1 := channels[0]
			if ch1.ID != "test_001" {
				t.Errorf("Ch1 ID: expected 'test_001', got '%s'", ch1.ID)
			}
			if ch1.Name != "Test Channel 1" {
				t.Errorf("Ch1 Name: expected 'Test Channel 1', got '%s'", ch1.Name)
			}
			if ch1.LogoURL != "http://logo.url/tc1.png" {
				t.Errorf("Ch1 LogoURL: expected 'http://logo.url/tc1.png', got '%s'", ch1.LogoURL)
			}
			if ch1.URL != "http://stream.url/tc1.m3u8" {
				t.Errorf("Ch1 URL: expected 'http://stream.url/tc1.m3u8', got '%s'", ch1.URL)
			}
			// IsHD removed from CustomChannel, Channel.IsHD will be false (zero-value)
			// Category "Entertainment" -> 5
			if ch1.Category != 5 {
				t.Errorf("Ch1 Category: expected 5 (Entertainment), got %d", ch1.Category)
			}
			// Language "English" -> 6
			if ch1.Language != 6 {
				t.Errorf("Ch1 Language: expected 6 (English), got %d", ch1.Language)
			}
			// EPGID is not directly mapped to Channel struct, so not tested here.
		}

		// Assertions for channel 2
		if len(channels) > 1 {
			ch2 := channels[1]
			if ch2.ID != "test_002" {
				t.Errorf("Ch2 ID: expected 'test_002', got '%s'", ch2.ID)
			}
			if ch2.Name != "Test Channel 2" {
				t.Errorf("Ch2 Name: expected 'Test Channel 2', got '%s'", ch2.Name)
			}
			if ch2.LogoURL != "relative/logo.png" {
				t.Errorf("Ch2 LogoURL: expected 'relative/logo.png', got '%s'", ch2.LogoURL)
			}
			if ch2.URL != "http://stream.url/tc2.m3u8" {
				t.Errorf("Ch2 URL: expected 'http://stream.url/tc2.m3u8', got '%s'", ch2.URL)
			}
			// IsHD removed from CustomChannel, Channel.IsHD will be false (zero-value)
			// Category "News" -> 12
			if ch2.Category != 12 {
				t.Errorf("Ch2 Category: expected 12 (News), got %d", ch2.Category)
			}
			// Language "Hindi" -> 1
			if ch2.Language != 1 {
				t.Errorf("Ch2 Language: expected 1 (Hindi), got %d", ch2.Language)
			}
		}
	})

	// Test Case 2: File Not Found
	t.Run("FileNotFound", func(t *testing.T) {
		nonExistentPath := filepath.Join(t.TempDir(), "non_existent_file.json")
		_, err := LoadCustomChannels(nonExistentPath)
		if err == nil {
			t.Errorf("LoadCustomChannels() with non-existent file path did not return an error")
		}
	})

	// Test Case 3: Invalid JSON
	t.Run("InvalidJSON", func(t *testing.T) {
		invalidJSONContent := `{
			"channels": [
				{
					"ID": "test_invalid",
					"Name": "Test Invalid Channel" 
					// Missing comma above
					"LogoURL": "http://logo.url/invalid.png",
					"Category": "Entertainment",
					"Language": "English",
					"URL": "http://stream.url/invalid.m3u8"
				}
			]
		}`
		tmpFilePath := createTempFile(t, invalidJSONContent)
		_, err := LoadCustomChannels(tmpFilePath)
		if err == nil {
			t.Errorf("LoadCustomChannels() with invalid JSON did not return an error")
		}
		// Check if the error is a JSON parsing error (optional, but good)
		// This depends on how specific the error returned by LoadCustomChannels is.
		// For now, just checking for any error is fine.
		if _, ok := err.(*json.SyntaxError); !ok && err != nil && !strings.Contains(err.Error(), "parse custom channels JSON") {
			 t.Logf("Warning: Error might not be a JSON syntax error, but: %v", err)
		}
	})

	// Test Case 4: Empty Channels Array
	t.Run("EmptyChannelsArray", func(t *testing.T) {
		emptyChannelsJSONContent := `{"channels": []}`
		tmpFilePath := createTempFile(t, emptyChannelsJSONContent)
		channels, err := LoadCustomChannels(tmpFilePath)
		if err != nil {
			t.Errorf("LoadCustomChannels() with empty channels array returned error: %v", err)
		}
		if len(channels) != 0 {
			t.Errorf("Expected 0 channels for empty array, got %d", len(channels))
		}
	})
	
	// Test Case 4b: Empty JSON file (should be handled by LoadCustomChannels)
	t.Run("EmptyJSONFile", func(t *testing.T) {
		tmpFilePath := createTempFile(t, "") // Empty content
		channels, err := LoadCustomChannels(tmpFilePath)
		// The function currently logs a warning and returns (nil, nil) for empty files
		if err != nil {
			t.Errorf("LoadCustomChannels() with empty JSON file returned error: %v", err)
		}
		if channels != nil { // Expect nil slice, not just empty
			t.Errorf("Expected nil channels for empty file, got %v (len %d)", channels, len(channels))
		}
	})


	// Test Case 5: Unmappable Category/Language
	t.Run("UnmappableCategoryLanguage", func(t *testing.T) {
		unmappableJSONContent := `{
			"channels": [
				{
					"ID": "test_003",
					"Name": "Unknown Lang/Cat Channel",
					"Category": "Unknown Category",
					"Language": "Klingon",
					"URL": "http://stream.url/tc3.m3u8"
				}
			]
		}`
		tmpFilePath := createTempFile(t, unmappableJSONContent)
		channels, err := LoadCustomChannels(tmpFilePath)
		if err != nil {
			t.Errorf("LoadCustomChannels() with unmappable category/language returned error: %v", err)
		}
		if len(channels) != 1 {
			t.Fatalf("Expected 1 channel, got %d", len(channels)) // Use Fatalf if subsequent checks depend on this
		}
		ch := channels[0]
		// Default Category ID: 0 ("All Categories")
		if ch.Category != 0 {
			t.Errorf("Category: expected 0 (All Categories) for unmappable, got %d", ch.Category)
		}
		// Default Language ID: 18 ("Other")
		if ch.Language != 18 {
			t.Errorf("Language: expected 18 (Other) for unmappable, got %d", ch.Language)
		}
	})

	// Test Case 6: Empty file path (should return nil, nil)
	t.Run("EmptyFilePath", func(t *testing.T) {
		channels, err := LoadCustomChannels("")
		if err != nil {
			t.Errorf("LoadCustomChannels() with empty file path returned error: %v", err)
		}
		if channels != nil {
			t.Errorf("Expected nil channels for empty file path, got %v", channels)
		}
	})
}
