package television

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jiotv-go/jiotv_go/v3/internal/config"
	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
)

// CustomChannel represents a custom channel definition
type CustomChannel struct {
	ID       string `json:"id" yaml:"id"`
	Name     string `json:"name" yaml:"name"`
	URL      string `json:"url" yaml:"url"`
	LogoURL  string `json:"logo_url" yaml:"logo_url"`
	Category int    `json:"category" yaml:"category"`
	Language int    `json:"language" yaml:"language"`
	IsHD     bool   `json:"is_hd" yaml:"is_hd"`
}

// CustomChannelConfig represents the configuration file for custom channels
type CustomChannelConfig struct {
	Channels []CustomChannel `json:"channels" yaml:"channels"`
}

// LoadCustomChannels loads custom channels from configuration files
// It looks for custom-channels.json and custom-channels.yml in the path prefix directory
func LoadCustomChannels() []Channel {
	// Check if custom channels are disabled in config
	if config.Cfg.DisableCustomChannels {
		utils.Log.Println("Custom channels are disabled via configuration")
		return []Channel{}
	}
	
	pathPrefix := utils.GetPathPrefix()
	
	var customChannels []Channel
	
	// Try to load from JSON file first
	jsonFile := filepath.Join(pathPrefix, "custom-channels.json")
	if channels := loadFromJSON(jsonFile); len(channels) > 0 {
		customChannels = append(customChannels, channels...)
	}
	
	// Try to load from YAML file
	yamlFile := filepath.Join(pathPrefix, "custom-channels.yml")
	if channels := loadFromYAML(yamlFile); len(channels) > 0 {
		customChannels = append(customChannels, channels...)
	}
	
	// Alternative YAML extension
	yamlFile2 := filepath.Join(pathPrefix, "custom-channels.yaml")
	if channels := loadFromYAML(yamlFile2); len(channels) > 0 {
		customChannels = append(customChannels, channels...)
	}
	
	if len(customChannels) > 0 {
		utils.Log.Printf("Loaded %d custom channels", len(customChannels))
	}
	
	return customChannels
}

// loadFromJSON loads custom channels from a JSON file
func loadFromJSON(filename string) []Channel {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil
	}
	
	data, err := os.ReadFile(filename)
	if err != nil {
		utils.Log.Printf("Error reading custom channels JSON file %s: %v", filename, err)
		return nil
	}
	
	var config CustomChannelConfig
	if err := json.Unmarshal(data, &config); err != nil {
		utils.Log.Printf("Error parsing custom channels JSON file %s: %v", filename, err)
		return nil
	}
	
	return convertCustomChannels(config.Channels, filename)
}

// loadFromYAML loads custom channels from a YAML file
func loadFromYAML(filename string) []Channel {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil
	}
	
	// Note: For now, we'll skip YAML support since it requires additional dependencies
	// This can be implemented later if needed
	utils.Log.Printf("YAML support for custom channels not yet implemented: %s", filename)
	return nil
}

// convertCustomChannels converts CustomChannel structs to Channel structs
func convertCustomChannels(customChannels []CustomChannel, source string) []Channel {
	var channels []Channel
	
	for i, custom := range customChannels {
		// Validate required fields
		if custom.ID == "" {
			utils.Log.Printf("Warning: Custom channel at index %d in %s has no ID, skipping", i, source)
			continue
		}
		if custom.Name == "" {
			utils.Log.Printf("Warning: Custom channel %s in %s has no name, skipping", custom.ID, source)
			continue
		}
		if custom.URL == "" {
			utils.Log.Printf("Warning: Custom channel %s in %s has no URL, skipping", custom.ID, source)
			continue
		}
		
		// Ensure custom channel IDs have a prefix to distinguish them
		channelID := custom.ID
		if !strings.HasPrefix(channelID, "custom_") {
			channelID = "custom_" + channelID
		}
		
		// Validate category and language
		category := custom.Category
		if category < 0 || category > 19 {
			utils.Log.Printf("Warning: Custom channel %s has invalid category %d, setting to 0", channelID, category)
			category = 0
		}
		
		language := custom.Language
		if language < 0 || language > 18 {
			utils.Log.Printf("Warning: Custom channel %s has invalid language %d, setting to 0", channelID, language)
			language = 0
		}
		
		channel := Channel{
			ID:       channelID,
			Name:     custom.Name,
			URL:      custom.URL,
			LogoURL:  custom.LogoURL,
			Category: category,
			Language: language,
			IsHD:     custom.IsHD,
		}
		
		channels = append(channels, channel)
	}
	
	return channels
}

// IsCustomChannel checks if a channel ID belongs to a custom channel
func IsCustomChannel(channelID string) bool {
	return strings.HasPrefix(channelID, "custom_")
}

// GetCustomChannelURL returns the direct URL for a custom channel
func GetCustomChannelURL(channelID string, customChannels []Channel) (string, error) {
	for _, channel := range customChannels {
		if channel.ID == channelID {
			return channel.URL, nil
		}
	}
	return "", fmt.Errorf("custom channel not found: %s", channelID)
}