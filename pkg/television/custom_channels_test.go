package television

import (
	"encoding/json"
	"fmt"
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
		if channels[0].ID != "cc_test_channel_1" {
			t.Errorf("Expected channel ID 'cc_test_channel_1', got '%s'", channels[0].ID)
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
		if channels[1].ID != "cc_test_channel_2" {
			t.Errorf("Expected channel ID 'cc_test_channel_2', got '%s'", channels[1].ID)
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
					Category: 8, // Sports
					Language: 1, // Hindi
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
		if channels[0].ID != "cc_yaml_test_channel_1" {
			t.Errorf("Expected channel ID 'cc_yaml_test_channel_1', got '%s'", channels[0].ID)
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

	// Initialize custom channels now that config is set
	InitCustomChannels()

	// Test getCustomChannels function (which uses caching internally)
	channels1 := getCustomChannels()
	if len(channels1) != 1 {
		t.Fatalf("Expected 1 channel, got %d", len(channels1))
	}
	if channels1[0].ID != "cc_cache_test_channel" {
		t.Errorf("Expected channel ID 'cc_cache_test_channel', got '%s'", channels1[0].ID)
	}

	// Call again to test cache usage
	channels2 := getCustomChannels()
	if len(channels2) != 1 {
		t.Fatalf("Expected 1 channel from cache, got %d", len(channels2))
	}
	if channels2[0].ID != "cc_cache_test_channel" {
		t.Errorf("Expected channel ID 'cc_cache_test_channel' from cache, got '%s'", channels2[0].ID)
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
	if channels3[0].ID != "cc_cache_test_channel" {
		t.Errorf("Expected channel ID 'cc_cache_test_channel' after reload, got '%s'", channels3[0].ID)
	}

	// Clear cache
	ClearCustomChannelsCache()

	// After clearing cache, need to reinitialize to reload from file
	InitCustomChannels()
	channels4 := getCustomChannels()
	if len(channels4) != 1 {
		t.Fatalf("Expected 1 channel after cache clear, got %d", len(channels4))
	}
	if channels4[0].ID != "cc_cache_test_channel" {
		t.Errorf("Expected channel ID 'cc_cache_test_channel' after cache clear, got '%s'", channels4[0].ID)
	}
}

func TestCustomChannelEfficientLookup(t *testing.T) {
	// Save original config
	originalCustomChannelsFile := config.Cfg.CustomChannelsFile
	defer func() {
		config.Cfg.CustomChannelsFile = originalCustomChannelsFile
	}()

	// Clear cache before test
	ClearCustomChannelsCache()

	// Create temporary JSON file with multiple channels
	tempFile, err := os.CreateTemp("", "custom_channels_lookup_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	customConfig := CustomChannelsConfig{
		Channels: []CustomChannel{
			{
				ID:       "lookup_test_1",
				Name:     "Lookup Test Channel 1",
				URL:      "https://example.com/test1.m3u8",
				LogoURL:  "https://example.com/logo1.png",
				Category: 12,
				Language: 6,
				IsHD:     true,
			},
			{
				ID:       "lookup_test_2",
				Name:     "Lookup Test Channel 2",
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

	// Set config to use the temp file
	config.Cfg.CustomChannelsFile = tempFile.Name()

	// Initialize custom channels
	InitCustomChannels()

	// Test efficient lookup - channel exists
	channel, exists := getCustomChannelByID("cc_lookup_test_1")
	if !exists {
		t.Error("Expected channel 'cc_lookup_test_1' to exist")
	}
	if channel.ID != "cc_lookup_test_1" {
		t.Errorf("Expected channel ID 'cc_lookup_test_1', got '%s'", channel.ID)
	}
	if channel.Name != "Lookup Test Channel 1" {
		t.Errorf("Expected channel name 'Lookup Test Channel 1', got '%s'", channel.Name)
	}

	// Test efficient lookup - another channel exists
	channel2, exists2 := getCustomChannelByID("cc_lookup_test_2")
	if !exists2 {
		t.Error("Expected channel 'cc_lookup_test_2' to exist")
	}
	if channel2.ID != "cc_lookup_test_2" {
		t.Errorf("Expected channel ID 'cc_lookup_test_2', got '%s'", channel2.ID)
	}
	if channel2.IsHD {
		t.Error("Expected channel 'cc_lookup_test_2' to not be HD")
	}

	// Test efficient lookup - non-existent channel
	_, exists3 := getCustomChannelByID("non_existent_channel")
	if exists3 {
		t.Error("Expected channel 'non_existent_channel' to not exist")
	}

	// Test lookup with empty cache
	ClearCustomChannelsCache()
	_, exists4 := getCustomChannelByID("cc_lookup_test_1")
	if exists4 {
		t.Error("Expected no channel to exist after cache clear")
	}
}

func TestExcessiveChannelsWarning(t *testing.T) {
	// Create temporary JSON file with more than 1000 channels to test warning
	tempFile, err := os.CreateTemp("", "excessive_channels_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Generate 1500 channels to exceed the limit
	var channels []CustomChannel
	for i := 1; i <= 1500; i++ {
		channels = append(channels, CustomChannel{
			ID:       fmt.Sprintf("test_channel_%d", i),
			Name:     fmt.Sprintf("Test Channel %d", i),
			URL:      fmt.Sprintf("https://example.com/test%d.m3u8", i),
			LogoURL:  fmt.Sprintf("https://example.com/logo%d.png", i),
			Category: 5,
			Language: 1,
			IsHD:     i%2 == 0, // Half HD, half not
		})
	}

	customConfig := CustomChannelsConfig{
		Channels: channels,
	}

	jsonData, err := json.Marshal(customConfig)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	if _, err := tempFile.Write(jsonData); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Test LoadCustomChannels with excessive channels
	loadedChannels, err := LoadCustomChannels(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to load custom channels: %v", err)
	}

	// Verify we loaded the correct number of channels
	if len(loadedChannels) != 1500 {
		t.Errorf("Expected 1500 channels, got %d", len(loadedChannels))
	}

	// Test caching with excessive channels
	config.Cfg.CustomChannelsFile = tempFile.Name()
	ClearCustomChannelsCache()

	// This should trigger the warning in loadAndCacheCustomChannels
	InitCustomChannels()

	// Verify channels are cached
	cachedChannels := getCustomChannels()
	if len(cachedChannels) != 1500 {
		t.Errorf("Expected 1500 cached channels, got %d", len(cachedChannels))
	}

	// Test reload with excessive channels
	err = ReloadCustomChannels()
	if err != nil {
		t.Fatalf("Failed to reload custom channels: %v", err)
	}

	// Verify reload worked
	reloadedChannels := getCustomChannels()
	if len(reloadedChannels) != 1500 {
		t.Errorf("Expected 1500 reloaded channels, got %d", len(reloadedChannels))
	}
}

func TestCustomChannelPrefix(t *testing.T) {
	// Save original config
	originalCustomChannelsFile := config.Cfg.CustomChannelsFile
	defer func() {
		config.Cfg.CustomChannelsFile = originalCustomChannelsFile
	}()

	// Clear cache before test
	ClearCustomChannelsCache()

	// Create temporary JSON file
	tempFile, err := os.CreateTemp("", "custom_channels_prefix_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	customConfig := CustomChannelsConfig{
		Channels: []CustomChannel{
			{
				ID:       "test_prefix_channel",
				Name:     "Test Prefix Channel",
				URL:      "https://example.com/prefix_test.m3u8",
				LogoURL:  "https://example.com/prefix_logo.png",
				Category: 12,
				Language: 6,
				IsHD:     true,
			},
			{
				ID:       "cc_already_prefixed",
				Name:     "Already Prefixed Channel",
				URL:      "https://example.com/already_prefixed.m3u8", 
				LogoURL:  "https://example.com/already_logo.png",
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

	// Test loading custom channels with prefix logic
	channels, err := LoadCustomChannels(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to load custom channels: %v", err)
	}

	if len(channels) != 2 {
		t.Fatalf("Expected 2 channels, got %d", len(channels))
	}

	// First channel should be prefixed
	if channels[0].ID != "cc_test_prefix_channel" {
		t.Errorf("Expected channel ID 'cc_test_prefix_channel', got '%s'", channels[0].ID)
	}

	// Second channel should remain as-is since it's already prefixed
	if channels[1].ID != "cc_already_prefixed" {
		t.Errorf("Expected channel ID 'cc_already_prefixed', got '%s'", channels[1].ID)
	}

	// Set config to use the temp file and initialize cache
	config.Cfg.CustomChannelsFile = tempFile.Name()
	InitCustomChannels()

	// Test that both channels can be found
	channel1, exists1 := getCustomChannelByID("cc_test_prefix_channel")
	if !exists1 {
		t.Error("Expected channel 'cc_test_prefix_channel' to exist")
	}
	if channel1.Name != "Test Prefix Channel" {
		t.Errorf("Expected channel name 'Test Prefix Channel', got '%s'", channel1.Name)
	}

	channel2, exists2 := getCustomChannelByID("cc_already_prefixed")
	if !exists2 {
		t.Error("Expected channel 'cc_already_prefixed' to exist")
	}
	if channel2.Name != "Already Prefixed Channel" {
		t.Errorf("Expected channel name 'Already Prefixed Channel', got '%s'", channel2.Name)
	}
}
