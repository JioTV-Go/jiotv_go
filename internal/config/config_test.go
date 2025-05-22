package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Helper function to create a temporary config file
func createTempConfigFile(t *testing.T, content string, extension string) string {
	t.Helper()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_config"+extension)
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}
	return filePath
}

func TestConfig_Load(t *testing.T) {
	// Define a sample config to compare against
	expectedConfig := JioTVConfig{
		EPG:                  true,
		Debug:                true,
		DisableTSHandler:     false,
		DisableLogout:        false,
		DRM:                  true,
		Title:                "Test Title",
		DisableURLEncryption: false,
		Proxy:                "http://localhost:8080",
		PathPrefix:           "/tmp/jiotv_go",
		LogPath:              "/tmp/logs",
		LogToStdout:          true,
	}

	t.Run("JSON_Valid", func(t *testing.T) {
		jsonContent := `{
			"epg": true,
			"debug": true,
			"disable_ts_handler": false,
			"disable_logout": false,
			"drm": true,
			"title": "Test Title",
			"disable_url_encryption": false,
			"proxy": "http://localhost:8080",
			"path_prefix": "/tmp/jiotv_go",
			"log_path": "/tmp/logs",
			"log_to_stdout": true
		}`
		jsonPath := createTempConfigFile(t, jsonContent, ".json")

		var cfg JioTVConfig
		err := cfg.Load(jsonPath)
		if err != nil {
			t.Fatalf("Load() with JSON failed: %v", err)
		}
		if !reflect.DeepEqual(cfg, expectedConfig) {
			t.Errorf("Loaded JSON config does not match expected. Got %+v, expected %+v", cfg, expectedConfig)
		}
	})

	t.Run("TOML_Valid", func(t *testing.T) {
		tomlContent := `
			epg = true
			debug = true
			disable_ts_handler = false
			disable_logout = false
			drm = true
			title = "Test Title"
			disable_url_encryption = false
			proxy = "http://localhost:8080"
			path_prefix = "/tmp/jiotv_go"
			log_path = "/tmp/logs"
			log_to_stdout = true
`
		tomlPath := createTempConfigFile(t, tomlContent, ".toml")
		var cfg JioTVConfig
		err := cfg.Load(tomlPath)
		if err != nil {
			t.Fatalf("Load() with TOML failed: %v", err)
		}
		if !reflect.DeepEqual(cfg, expectedConfig) {
			t.Errorf("Loaded TOML config does not match expected. Got %+v, expected %+v", cfg, expectedConfig)
		}
	})

	t.Run("YAML_Valid", func(t *testing.T) {
		yamlContent := `
epg: true
debug: true
disable_ts_handler: false
disable_logout: false
drm: true
title: "Test Title"
disable_url_encryption: false
proxy: "http://localhost:8080"
path_prefix: "/tmp/jiotv_go"
log_path: "/tmp/logs"
log_to_stdout: true
`
		yamlPath := createTempConfigFile(t, yamlContent, ".yaml")
		var cfg JioTVConfig
		err := cfg.Load(yamlPath)
		if err != nil {
			t.Fatalf("Load() with YAML failed: %v", err)
		}
		if !reflect.DeepEqual(cfg, expectedConfig) {
			t.Errorf("Loaded YAML config does not match expected. Got %+v, expected %+v", cfg, expectedConfig)
		}
	})

	t.Run("NonExistentFile", func(t *testing.T) {
		var cfg JioTVConfig
		err := cfg.Load("non_existent_config.json")
		if err == nil {
			t.Errorf("Expected error when loading non-existent file, but got nil")
		}
	})

	t.Run("InvalidFileFormat", func(t *testing.T) {
		invalidContent := `this is not a valid config format`
		// The library cleanenv attempts to detect format by extension,
		// and then by parsing. If we provide a .json extension with invalid content,
		// it will fail parsing. If we provide an unknown extension, it might also fail.
		invalidPath := createTempConfigFile(t, invalidContent, ".json")

		var cfg JioTVConfig
		err := cfg.Load(invalidPath)
		if err == nil {
			t.Errorf("Expected error when loading invalid config file (invalid JSON content), but got nil")
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		jsonContent := `{"debug": true, "title": "Test Title",}` // Extra comma makes it invalid
		jsonPath := createTempConfigFile(t, jsonContent, ".json")

		var cfg JioTVConfig
		err := cfg.Load(jsonPath)
		if err == nil {
			t.Errorf("Expected error when loading invalid JSON (extra comma), but got nil")
		}
	})

	t.Run("InvalidTOML", func(t *testing.T) {
		tomlContent := `
debug = true
title = 
` // Missing value for title
		tomlPath := createTempConfigFile(t, tomlContent, ".toml")
		var cfg JioTVConfig
		err := cfg.Load(tomlPath)
		if err == nil {
			t.Errorf("Expected error when loading invalid TOML (missing value), but got nil")
		}
	})

	t.Run("InvalidYAML", func(t *testing.T) {
		yamlContent := `
debug: true
title: "Test Title"
  path_prefix: "/tmp/jiotv_go" 
` // Inconsistent indentation
		yamlPath := createTempConfigFile(t, yamlContent, ".yaml")
		var cfg JioTVConfig
		err := cfg.Load(yamlPath)
		if err == nil {
			t.Errorf("Expected error when loading invalid YAML (inconsistent indentation), but got nil")
		}
	})

	t.Run("UnsupportedExtension", func(t *testing.T) {
		content := `data = "value"`
		path := createTempConfigFile(t, content, ".txt") // .txt is not a supported extension by cleanenv by default
		var cfg JioTVConfig
		err := cfg.Load(path)
		// cleanenv.ReadConfig returns an error if the file extension is not recognized.
		if err == nil {
			t.Errorf("Expected error when loading file with unsupported extension, but got nil")
		}
	})

	t.Run("LoadFromEnvironment", func(t *testing.T) {
		var cfg JioTVConfig
		// Set environment variables
		t.Setenv("JIOTV_EPG", "true")
		t.Setenv("JIOTV_DEBUG", "true")
		t.Setenv("JIOTV_TITLE", "Env Test Title")
		// Add other env vars as needed

		err := cfg.Load("") // Pass empty filename to trigger env loading
		if err != nil {
			t.Fatalf("Load() from environment failed: %v", err)
		}

		if !cfg.EPG {
			t.Errorf("Expected EPG to be true from env, got %v", cfg.EPG)
		}
		if !cfg.Debug {
			t.Errorf("Expected Debug to be true from env, got %v", cfg.Debug)
		}
		if cfg.Title != "Env Test Title" {
			t.Errorf("Expected Title to be 'Env Test Title' from env, got %s", cfg.Title)
		}
		// Unset environment variables after test if necessary, though t.Setenv handles cleanup.
	})
}
