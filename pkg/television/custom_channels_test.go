package television

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/jiotv-go/jiotv_go/v3/internal/config"
	"gopkg.in/yaml.v3"
)

func TestLoadCustomChannels(t *testing.T) {
	// Test loading JSON format
	t.Run("LoadJSONCustomChannels", func(t *testing.T) {
		// Create temporary JSON file
		tempFile, err := os.CreateTemp("", "custom_channels_*.json")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		customConfig := CustomChannelsConfig{
			Channels: []CustomChannel{
				{
					ID:       "test_channel_1",
					Name:     "Test Channel 1",
					URL:      "https://example.com/test1.m3u8",
					LogoURL:  "https://example.com/logo1.png",
					Category: 12,
					Language: 6,
					IsHD:     true,
				},
				{
					ID:       "test_channel_2",
					Name:     "Test Channel 2",
					URL:      "https://example.com/test2.m3u8",
					LogoURL:  "https://example.com/logo2.png",
					Category: 5,
					Language: 1,
					IsHD:     false,
				},
			},
		}

		jsonData, err := json.Marshal(customConfig)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}

		if _, err := tempFile.Write(jsonData); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		tempFile.Close()

		// Test loading
		channels, err := LoadCustomChannels(tempFile.Name())
		if err != nil {
			t.Fatalf("Failed to load custom channels: %v", err)
		}

		if len(channels) != 2 {
			t.Errorf("Expected 2 channels, got %d", len(channels))
		}

		// Verify first channel
		if channels[0].ID != "test_channel_1" {
			t.Errorf("Expected channel ID 'test_channel_1', got '%s'", channels[0].ID)
		}
		if channels[0].Name != "Test Channel 1" {
			t.Errorf("Expected channel name 'Test Channel 1', got '%s'", channels[0].Name)
		}
		if channels[0].Category != 12 {
			t.Errorf("Expected category 12, got %d", channels[0].Category)
		}
		if channels[0].Language != 6 {
			t.Errorf("Expected language 6, got %d", channels[0].Language)
		}
		if !channels[0].IsHD {
			t.Error("Expected first channel to be HD")
		}

		// Verify second channel
		if channels[1].ID != "test_channel_2" {
			t.Errorf("Expected channel ID 'test_channel_2', got '%s'", channels[1].ID)
		}
		if channels[1].IsHD {
			t.Error("Expected second channel to not be HD")
		}
	})

	// Test loading YAML format
	t.Run("LoadYAMLCustomChannels", func(t *testing.T) {
		// Create temporary YAML file
		tempFile, err := os.CreateTemp("", "custom_channels_*.yml")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		customConfig := CustomChannelsConfig{
			Channels: []CustomChannel{
				{
					ID:       "yaml_test_channel_1",
					Name:     "YAML Test Channel 1",
					URL:      "https://example.com/yaml1.m3u8",
					LogoURL:  "https://example.com/yaml_logo1.png",
					Category: 8,  // Sports
					Language: 1,  // Hindi
					IsHD:     true,
				},
			},
		}

		yamlData, err := yaml.Marshal(customConfig)
		if err != nil {
			t.Fatalf("Failed to marshal YAML: %v", err)
		}

		if _, err := tempFile.Write(yamlData); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		tempFile.Close()

		// Test loading
		channels, err := LoadCustomChannels(tempFile.Name())
		if err != nil {
			t.Fatalf("Failed to load custom channels: %v", err)
		}

		if len(channels) != 1 {
			t.Errorf("Expected 1 channel, got %d", len(channels))
		}

		// Verify channel
		if channels[0].ID != "yaml_test_channel_1" {
			t.Errorf("Expected channel ID 'yaml_test_channel_1', got '%s'", channels[0].ID)
		}
		if channels[0].Name != "YAML Test Channel 1" {
			t.Errorf("Expected channel name 'YAML Test Channel 1', got '%s'", channels[0].Name)
		}
		if channels[0].Category != 8 {
			t.Errorf("Expected category 8, got %d", channels[0].Category)
		}
	})

	// Test empty file path
	t.Run("EmptyFilePath", func(t *testing.T) {
		channels, err := LoadCustomChannels("")
		if err != nil {
			t.Errorf("Expected no error for empty file path, got: %v", err)
		}
		if len(channels) != 0 {
			t.Errorf("Expected 0 channels for empty file path, got %d", len(channels))
		}
	})

	// Test non-existent file
	t.Run("NonExistentFile", func(t *testing.T) {
		channels, err := LoadCustomChannels("/path/that/does/not/exist.json")
		if err != nil {
			t.Errorf("Expected no error for non-existent file, got: %v", err)
		}
		if len(channels) != 0 {
			t.Errorf("Expected 0 channels for non-existent file, got %d", len(channels))
		}
	})

	// Test unsupported file format
	t.Run("UnsupportedFormat", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "custom_channels_*.txt")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		_, err = LoadCustomChannels(tempFile.Name())
		if err == nil {
			t.Error("Expected error for unsupported file format")
		}
	})
}

func TestChannelsWithCustomChannels(t *testing.T) {
	// Save original config
	originalCustomChannelsFile := config.Cfg.CustomChannelsFile
	defer func() {
		config.Cfg.CustomChannelsFile = originalCustomChannelsFile
	}()

	// Create temporary JSON file
	tempFile, err := os.CreateTemp("", "custom_channels_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	customConfig := CustomChannelsConfig{
		Channels: []CustomChannel{
			{
				ID:       "test_custom_channel",
				Name:     "Test Custom Channel",
				URL:      "https://example.com/test.m3u8",
				LogoURL:  "https://example.com/logo.png",
				Category: 12,
				Language: 6,
				IsHD:     true,
			},
		},
	}

	jsonData, err := json.Marshal(customConfig)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	if _, err := tempFile.Write(jsonData); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Set config to use the temp file
	config.Cfg.CustomChannelsFile = tempFile.Name()

	// Note: This test would require mocking the JioTV API call to fully test
	// For now, we just test that the function doesn't crash with custom channels configured
	t.Run("ChannelsDoesNotCrashWithCustomChannels", func(t *testing.T) {
		// This would make an actual HTTP request, so we skip it in tests
		// In a real test environment, you would mock the HTTP client
		t.Skip("Skipping actual API call test")
	})
}

func TestCustomChannelsCaching(t *testing.T) {
	// Save original config
	originalCustomChannelsFile := config.Cfg.CustomChannelsFile
	defer func() {
		config.Cfg.CustomChannelsFile = originalCustomChannelsFile
	}()

	// Clear cache before test
	ClearCustomChannelsCache()

	// Create temporary JSON file
	tempFile, err := os.CreateTemp("", "custom_channels_cache_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	customConfig := CustomChannelsConfig{
		Channels: []CustomChannel{
			{
				ID:       "cache_test_channel",
				Name:     "Cache Test Channel",
				URL:      "https://example.com/cache_test.m3u8",
				LogoURL:  "https://example.com/cache_logo.png",
				Category: 12,
				Language: 6,
				IsHD:     true,
			},
		},
	}

	jsonData, err := json.Marshal(customConfig)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	if _, err := tempFile.Write(jsonData); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Set config to use the temp file
	config.Cfg.CustomChannelsFile = tempFile.Name()

	// Test getCustomChannels function (which uses caching internally)
	channels1 := getCustomChannels()
	if len(channels1) != 1 {
		t.Fatalf("Expected 1 channel, got %d", len(channels1))
	}
	if channels1[0].ID != "cache_test_channel" {
		t.Errorf("Expected channel ID 'cache_test_channel', got '%s'", channels1[0].ID)
	}

	// Call again to test cache usage
	channels2 := getCustomChannels()
	if len(channels2) != 1 {
		t.Fatalf("Expected 1 channel from cache, got %d", len(channels2))
	}
	if channels2[0].ID != "cache_test_channel" {
		t.Errorf("Expected channel ID 'cache_test_channel' from cache, got '%s'", channels2[0].ID)
	}

	// Test cache reload
	err = ReloadCustomChannels()
	if err != nil {
		t.Fatalf("Failed to reload custom channels: %v", err)
	}

	// After reload, should still work
	channels3 := getCustomChannels()
	if len(channels3) != 1 {
		t.Fatalf("Expected 1 channel after reload, got %d", len(channels3))
	}
	if channels3[0].ID != "cache_test_channel" {
		t.Errorf("Expected channel ID 'cache_test_channel' after reload, got '%s'", channels3[0].ID)
	}

	// Clear cache
	ClearCustomChannelsCache()

	// After clearing cache, it should reload from file
	channels4 := getCustomChannels()
	if len(channels4) != 1 {
		t.Fatalf("Expected 1 channel after cache clear, got %d", len(channels4))
	}
	if channels4[0].ID != "cache_test_channel" {
		t.Errorf("Expected channel ID 'cache_test_channel' after cache clear, got '%s'", channels4[0].ID)
	}
}