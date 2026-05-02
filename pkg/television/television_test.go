package television

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/jiotv-go/jiotv_go/v3/internal/config"
	"github.com/jiotv-go/jiotv_go/v3/pkg/secureurl"
	"github.com/jiotv-go/jiotv_go/v3/pkg/store"
	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
)

var (
	setupOnce sync.Once
)

// Setup function to initialize store for tests
func setupTest() {
	setupOnce.Do(func() {
		// Setup test environment with temporary pathPrefix
		_, err := store.SetupTestPathPrefix()
		if err != nil {
			panic(fmt.Sprintf("Failed to setup test environment: %v", err))
		}
		// Note: cleanup is handled by the temp directory system cleanup

		// Initialize store for testing
		store.Init()
		// Initialize secureurl for URL encryption/decryption
		secureurl.Init()
		// Initialize the Log variable to prevent nil pointer dereference
		if utils.Log == nil {
			utils.Log = log.New(os.Stdout, "", log.LstdFlags)
		}
	})
}

func TestFilterChannels(t *testing.T) {
	// Create test data
	testChannels := []Channel{
		{ID: "1", Name: "Hindi Entertainment", Language: 1, Category: 5}, // Hindi Entertainment
		{ID: "2", Name: "English Movies", Language: 6, Category: 6},      // English Movies
		{ID: "3", Name: "Hindi Movies", Language: 1, Category: 6},        // Hindi Movies
		{ID: "4", Name: "English Sports", Language: 6, Category: 8},      // English Sports
		{ID: "5", Name: "Tamil Entertainment", Language: 8, Category: 5}, // Tamil Entertainment
	}

	type args struct {
		channels []Channel
		language int
		category int
	}
	tests := []struct {
		name string
		args args
		want []Channel
	}{
		{
			name: "Filter by language only (Hindi)",
			args: args{
				channels: testChannels,
				language: 1, // Hindi
				category: 0, // No category filter
			},
			want: []Channel{
				{ID: "1", Name: "Hindi Entertainment", Language: 1, Category: 5},
				{ID: "3", Name: "Hindi Movies", Language: 1, Category: 6},
			},
		},
		{
			name: "Filter by category only (Movies)",
			args: args{
				channels: testChannels,
				language: 0, // No language filter
				category: 6, // Movies
			},
			want: []Channel{
				{ID: "2", Name: "English Movies", Language: 6, Category: 6},
				{ID: "3", Name: "Hindi Movies", Language: 1, Category: 6},
			},
		},
		{
			name: "Filter by both language and category (English Movies)",
			args: args{
				channels: testChannels,
				language: 6, // English
				category: 6, // Movies
			},
			want: []Channel{
				{ID: "2", Name: "English Movies", Language: 6, Category: 6},
			},
		},
		{
			name: "No filters (return all)",
			args: args{
				channels: testChannels,
				language: 0, // No filter
				category: 0, // No filter
			},
			want: testChannels,
		},
		{
			name: "Empty channels slice",
			args: args{
				channels: []Channel{},
				language: 1,
				category: 5,
			},
			want: []Channel{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterChannels(tt.args.channels, tt.args.language, tt.args.category)
			if len(got) != len(tt.want) {
				t.Errorf("FilterChannels() returned %d channels, want %d", len(got), len(tt.want))
				return
			}
			for i, channel := range got {
				if channel.ID != tt.want[i].ID {
					t.Errorf("FilterChannels() channel[%d].ID = %v, want %v", i, channel.ID, tt.want[i].ID)
				}
			}
		})
	}
}

func TestReplaceM3U8(t *testing.T) {
	setupTest() // Initialize necessary components
	type args struct {
		baseUrl    []byte
		match      []byte
		params     string
		channel_id string
		quality    string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Replace M3U8 URL with parameters and quality",
			args: args{
				baseUrl:    []byte("test.m3u8"),
				match:      []byte("test.m3u8"),
				params:     "param1=value1",
				channel_id: "123",
				quality:    "high",
			},
		},
		{
			name: "Replace M3U8 URL with empty params",
			args: args{
				baseUrl:    []byte("example.m3u8"),
				match:      []byte("example.m3u8"),
				params:     "",
				channel_id: "456",
				quality:    "auto",
			},
		},
		{
			name: "Replace M3U8 URL with empty quality",
			args: args{
				baseUrl:    []byte("original.m3u8"),
				match:      []byte("original.m3u8"),
				params:     "param1=value1",
				channel_id: "123",
				quality:    "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaceM3U8(tt.args.baseUrl, tt.args.match, tt.args.params, tt.args.channel_id, tt.args.quality)
			// The function encrypts URLs, so we check that it produces some output
			if len(got) == 0 {
				t.Errorf("ReplaceM3U8() returned empty result")
			}
			// Should contain render path
			if !strings.Contains(string(got), "/render") {
				t.Errorf("ReplaceM3U8() should contain /render path, got %s", string(got))
			}
			// If quality is provided, should contain quality parameter
			if tt.args.quality != "" && !strings.Contains(string(got), "q="+tt.args.quality) {
				t.Errorf("ReplaceM3U8() should contain quality parameter q=%s, got %s", tt.args.quality, string(got))
			}
		})
	}
}

func TestReplaceTS(t *testing.T) {
	setupTest() // Initialize necessary components
	type args struct {
		baseUrl []byte
		match   []byte
		params  string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Replace TS URL with parameters",
			args: args{
				baseUrl: []byte("segment.ts"),
				match:   []byte("segment.ts"),
				params:  "param1=value1",
			},
		},
		{
			name: "Replace TS URL with empty params",
			args: args{
				baseUrl: []byte("test.ts"),
				match:   []byte("test.ts"),
				params:  "",
			},
		},
		{
			name: "No match found",
			args: args{
				baseUrl: []byte("original content"),
				match:   []byte("not_found.ts"),
				params:  "param1=value1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaceTS(tt.args.baseUrl, tt.args.match, tt.args.params, "123")
			// The function encrypts URLs, so we check that it produces some output
			if len(got) == 0 {
				t.Errorf("ReplaceTS() returned empty result")
			}
			// Should contain render path
			if !strings.Contains(string(got), "/render") {
				t.Errorf("ReplaceTS() should contain /render path, got %s", string(got))
			}
		})
	}
}

func TestReplaceAAC(t *testing.T) {
	setupTest() // Initialize necessary components
	type args struct {
		baseUrl []byte
		match   []byte
		params  string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Replace AAC URL with parameters",
			args: args{
				baseUrl: []byte("audio.aac"),
				match:   []byte("audio.aac"),
				params:  "param1=value1",
			},
		},
		{
			name: "Replace AAC URL with empty params",
			args: args{
				baseUrl: []byte("test.aac"),
				match:   []byte("test.aac"),
				params:  "",
			},
		},
		{
			name: "No match found",
			args: args{
				baseUrl: []byte("original content"),
				match:   []byte("not_found.aac"),
				params:  "param1=value1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaceAAC(tt.args.baseUrl, tt.args.match, tt.args.params, "123")
			// The function encrypts URLs, so we check that it produces some output
			if len(got) == 0 {
				t.Errorf("ReplaceAAC() returned empty result")
			}
			// Should contain render path
			if !strings.Contains(string(got), "/render") {
				t.Errorf("ReplaceAAC() should contain /render path, got %s", string(got))
			}
		})
	}
}

func TestReplaceKey(t *testing.T) {
	setupTest() // Initialize necessary components
	type args struct {
		match      []byte
		params     string
		channel_id string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Replace key with parameters",
			args: args{
				match:      []byte("key.bin"),
				params:     "param1=value1",
				channel_id: "123",
			},
		},
		{
			name: "Replace key with empty params",
			args: args{
				match:      []byte("test.key"),
				params:     "",
				channel_id: "456",
			},
		},
		{
			name: "Replace key with empty channel_id",
			args: args{
				match:      []byte("test.key"),
				params:     "param1=value1",
				channel_id: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaceKey(tt.args.match, tt.args.params, tt.args.channel_id)
			// The function encrypts URLs, so we check that it produces some output
			if len(got) == 0 {
				t.Errorf("ReplaceKey() returned empty result")
			}
			// Should contain render path
			if !strings.Contains(string(got), "/render") {
				t.Errorf("ReplaceKey() should contain /render path, got %s", string(got))
			}
		})
	}
}

func TestInitCustomChannels(t *testing.T) {
	// Save original config
	originalCustomChannelsFile := config.Cfg.CustomChannelsFile
	defer func() {
		config.Cfg.CustomChannelsFile = originalCustomChannelsFile
	}()

	t.Run("No custom channels file configured", func(t *testing.T) {
		config.Cfg.CustomChannelsFile = ""
		// Should not panic
		InitCustomChannels()
	})

	t.Run("Custom channels file configured", func(t *testing.T) {
		// Create a temporary custom channels file
		tempDir := t.TempDir()
		customChannelsFile := filepath.Join(tempDir, "test_channels.json")

		customChannelsData := map[string]interface{}{
			"channels": []map[string]interface{}{
				{
					"id":       "test1",
					"name":     "Test Channel 1",
					"url":      "https://example.com/test1.m3u8",
					"logo_url": "https://example.com/logo1.png",
					"category": 1,
					"language": 6,
					"is_hd":    true,
				},
			},
		}

		jsonData, _ := json.Marshal(customChannelsData)
		err := os.WriteFile(customChannelsFile, jsonData, 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		config.Cfg.CustomChannelsFile = customChannelsFile
		// Should not panic and should load channels
		InitCustomChannels()

		// Test that channel is loaded
		channel, exists := GetCustomChannelByID("cc_test1")
		if !exists {
			t.Errorf("Expected custom channel to be loaded")
		}
		if channel.Name != "Test Channel 1" {
			t.Errorf("Expected channel name 'Test Channel 1', got '%s'", channel.Name)
		}
	})
}

func TestGetCustomChannelByID(t *testing.T) {
	setupTest()

	// Create test data
	tempDir := t.TempDir()
	customChannelsFile := filepath.Join(tempDir, "test_channels.json")

	customChannelsData := map[string]interface{}{
		"channels": []map[string]interface{}{
			{
				"id":       "test1",
				"name":     "Test Channel 1",
				"url":      "https://example.com/test1.m3u8",
				"logo_url": "https://example.com/logo1.png",
				"category": 1,
				"language": 6,
				"is_hd":    true,
			},
			{
				"id":       "test2",
				"name":     "Test Channel 2",
				"url":      "https://example.com/test2.m3u8",
				"logo_url": "https://example.com/logo2.png",
				"category": 2,
				"language": 1,
				"is_hd":    false,
			},
		},
	}

	jsonData, _ := json.Marshal(customChannelsData)
	err := os.WriteFile(customChannelsFile, jsonData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Save original config
	originalCustomChannelsFile := config.Cfg.CustomChannelsFile
	defer func() {
		config.Cfg.CustomChannelsFile = originalCustomChannelsFile
	}()

	config.Cfg.CustomChannelsFile = customChannelsFile
	InitCustomChannels()

	tests := []struct {
		name      string
		channelID string
		wantName  string
		wantURL   string
		wantFound bool
	}{
		{
			name:      "Existing custom channel with cc_ prefix",
			channelID: "cc_test1",
			wantName:  "Test Channel 1",
			wantURL:   "https://example.com/test1.m3u8",
			wantFound: true,
		},
		{
			name:      "Existing custom channel second entry",
			channelID: "cc_test2",
			wantName:  "Test Channel 2",
			wantURL:   "https://example.com/test2.m3u8",
			wantFound: true,
		},
		{
			name:      "Non-existing custom channel",
			channelID: "cc_nonexistent",
			wantFound: false,
		},
		{
			name:      "Empty channel ID",
			channelID: "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channel, found := GetCustomChannelByID(tt.channelID)

			if found != tt.wantFound {
				t.Errorf("GetCustomChannelByID() found = %v, want %v", found, tt.wantFound)
				return
			}

			if tt.wantFound {
				if channel.Name != tt.wantName {
					t.Errorf("GetCustomChannelByID() channel.Name = %v, want %v", channel.Name, tt.wantName)
				}
				if channel.URL != tt.wantURL {
					t.Errorf("GetCustomChannelByID() channel.URL = %v, want %v", channel.URL, tt.wantURL)
				}
				if channel.ID != tt.channelID {
					t.Errorf("GetCustomChannelByID() channel.ID = %v, want %v", channel.ID, tt.channelID)
				}
			}
		})
	}
}

func TestGetCustomChannelByID_NilCache(t *testing.T) {
	// Test when cache is nil
	customChannelsCacheMap = nil

	channel, found := GetCustomChannelByID("test")
	if found {
		t.Errorf("GetCustomChannelByID() with nil cache should return false, got true")
	}
	if channel.ID != "" {
		t.Errorf("GetCustomChannelByID() with nil cache should return empty channel, got %+v", channel)
	}
}

func TestLoadAndCacheCustomChannels(t *testing.T) {
	setupTest()

	// Save original config
	originalCustomChannelsFile := config.Cfg.CustomChannelsFile
	defer func() {
		config.Cfg.CustomChannelsFile = originalCustomChannelsFile
	}()

	t.Run("File does not exist", func(t *testing.T) {
		config.Cfg.CustomChannelsFile = "/nonexistent/path/channels.json"
		// Should not panic and should create empty cache
		loadAndCacheCustomChannels()

		// Check that cache is empty
		if customChannelsCacheMap == nil {
			t.Errorf("Expected cache to be initialized as empty map")
		}
		if len(customChannelsCacheMap) != 0 {
			t.Errorf("Expected empty cache, got %d items", len(customChannelsCacheMap))
		}
	})

	t.Run("Valid JSON file", func(t *testing.T) {
		tempDir := t.TempDir()
		customChannelsFile := filepath.Join(tempDir, "test_channels.json")

		customChannelsData := map[string]interface{}{
			"channels": []map[string]interface{}{
				{
					"id":       "json_test",
					"name":     "JSON Test Channel",
					"url":      "https://example.com/json.m3u8",
					"logo_url": "https://example.com/json_logo.png",
					"category": 5,
					"language": 1,
					"is_hd":    true,
				},
			},
		}

		jsonData, _ := json.Marshal(customChannelsData)
		err := os.WriteFile(customChannelsFile, jsonData, 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		config.Cfg.CustomChannelsFile = customChannelsFile
		loadAndCacheCustomChannels()

		// Check that channel is cached with cc_ prefix
		channel, exists := customChannelsCacheMap["cc_json_test"]
		if !exists {
			t.Errorf("Expected channel to be cached with cc_ prefix")
		}
		if channel.Name != "JSON Test Channel" {
			t.Errorf("Expected channel name 'JSON Test Channel', got '%s'", channel.Name)
		}
	})

	t.Run("Valid YAML file", func(t *testing.T) {
		tempDir := t.TempDir()
		customChannelsFile := filepath.Join(tempDir, "test_channels.yml")

		yamlData := `channels:
  - id: yaml_test
    name: YAML Test Channel
    url: https://example.com/yaml.m3u8
    logo_url: https://example.com/yaml_logo.png
    category: 6
    language: 6
    is_hd: false`

		err := os.WriteFile(customChannelsFile, []byte(yamlData), 0644)
		if err != nil {
			t.Fatalf("Failed to create test YAML file: %v", err)
		}

		config.Cfg.CustomChannelsFile = customChannelsFile
		loadAndCacheCustomChannels()

		// Check that channel is cached with cc_ prefix
		channel, exists := customChannelsCacheMap["cc_yaml_test"]
		if !exists {
			t.Errorf("Expected YAML channel to be cached with cc_ prefix")
		}
		if channel.Name != "YAML Test Channel" {
			t.Errorf("Expected channel name 'YAML Test Channel', got '%s'", channel.Name)
		}
	})

	t.Run("Channel with existing cc_ prefix", func(t *testing.T) {
		tempDir := t.TempDir()
		customChannelsFile := filepath.Join(tempDir, "test_channels.json")

		customChannelsData := map[string]interface{}{
			"channels": []map[string]interface{}{
				{
					"id":       "cc_already_prefixed",
					"name":     "Already Prefixed Channel",
					"url":      "https://example.com/prefixed.m3u8",
					"logo_url": "https://example.com/prefixed_logo.png",
					"category": 1,
					"language": 1,
					"is_hd":    true,
				},
			},
		}

		jsonData, _ := json.Marshal(customChannelsData)
		err := os.WriteFile(customChannelsFile, jsonData, 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		config.Cfg.CustomChannelsFile = customChannelsFile
		loadAndCacheCustomChannels()

		// Should not double-prefix
		channel, exists := customChannelsCacheMap["cc_already_prefixed"]
		if !exists {
			t.Errorf("Expected channel to be cached without double prefixing")
		}
		if channel.ID != "cc_already_prefixed" {
			t.Errorf("Expected channel ID to remain 'cc_already_prefixed', got '%s'", channel.ID)
		}
	})
}
