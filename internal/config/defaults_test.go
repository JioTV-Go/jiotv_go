package config

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDefaultCategoriesAndLanguagesConfig(t *testing.T) {
	tests := []struct {
		name       string
		configData interface{}
		configType string
		expected   JioTVConfig
	}{
		{
			name: "JSON config with default categories and languages",
			configData: map[string]interface{}{
				"default_categories": []int{1, 2, 3},
				"default_languages":  []int{6, 1},
				"debug":              true,
			},
			configType: "json",
			expected: JioTVConfig{
				DefaultCategories: []int{1, 2, 3},
				DefaultLanguages:  []int{6, 1},
				Debug:             true,
			},
		},
		{
			name: "YAML config with default categories and languages",
			configData: map[string]interface{}{
				"default_categories": []int{8, 5},
				"default_languages":  []int{1},
				"epg":                false,
			},
			configType: "yaml",
			expected: JioTVConfig{
				DefaultCategories: []int{8, 5},
				DefaultLanguages:  []int{1},
				EPG:               false,
			},
		},
		{
			name: "Empty arrays should work",
			configData: map[string]interface{}{
				"default_categories": []int{},
				"default_languages":  []int{},
				"title":              "Test App",
			},
			configType: "json",
			expected: JioTVConfig{
				DefaultCategories: []int{},
				DefaultLanguages:  []int{},
				Title:             "Test App",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tmpFile, err := os.CreateTemp("", "test-config-*."+tt.configType)
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			// Write config data to file
			var data []byte
			switch tt.configType {
			case "json":
				data, err = json.Marshal(tt.configData)
			case "yaml":
				data, err = yaml.Marshal(tt.configData)
			default:
				t.Fatalf("unsupported config type: %s", tt.configType)
			}
			if err != nil {
				t.Fatalf("Failed to marshal config data: %v", err)
			}

			if _, err := tmpFile.Write(data); err != nil {
				t.Fatalf("Failed to write config file: %v", err)
			}
			tmpFile.Close()

			// Load config
			var config JioTVConfig
			err = config.Load(tmpFile.Name())
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			// Check default categories
			if !reflect.DeepEqual(config.DefaultCategories, tt.expected.DefaultCategories) {
				t.Errorf("DefaultCategories mismatch. Got %v, expected %v", config.DefaultCategories, tt.expected.DefaultCategories)
			}

			// Check default languages
			if !reflect.DeepEqual(config.DefaultLanguages, tt.expected.DefaultLanguages) {
				t.Errorf("DefaultLanguages mismatch. Got %v, expected %v", config.DefaultLanguages, tt.expected.DefaultLanguages)
			}

			// Check other fields
			if config.Debug != tt.expected.Debug {
				t.Errorf("Debug = %v, expected %v", config.Debug, tt.expected.Debug)
			}
			if config.EPG != tt.expected.EPG {
				t.Errorf("EPG = %v, expected %v", config.EPG, tt.expected.EPG)
			}
			if config.Title != tt.expected.Title {
				t.Errorf("Title = %v, expected %v", config.Title, tt.expected.Title)
			}
		})
	}
}