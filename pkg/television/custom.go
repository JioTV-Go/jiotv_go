package television

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/jiotv-go/jiotv_go/v3/internal/config"
	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
)

// CustomChannel represents a custom channel configuration
type CustomChannel struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	LogoURL  string `json:"logo_url,omitempty"`
	Category int    `json:"category"`
	Language int    `json:"language"`
	IsHD     bool   `json:"is_hd"`
}

// CustomChannelsConfig represents the structure of the custom channels JSON file
type CustomChannelsConfig struct {
	Channels []CustomChannel `json:"channels"`
}

// customChannels holds the loaded custom channels
var customChannels []Channel

// LoadCustomChannels loads custom channels from the configured JSON file
func LoadCustomChannels() error {
	// Reset custom channels
	customChannels = nil

	// Check if custom channels path is configured
	if config.Cfg.CustomChannelsPath == "" {
		utils.Log.Println("INFO: Custom channels disabled (no path configured)")
		return nil
	}

	// Check if file exists
	if _, err := os.Stat(config.Cfg.CustomChannelsPath); os.IsNotExist(err) {
		utils.Log.Printf("INFO: Custom channels file not found at %s, skipping custom channels\n", config.Cfg.CustomChannelsPath)
		return nil
	}

	// Read the JSON file
	data, err := ioutil.ReadFile(config.Cfg.CustomChannelsPath)
	if err != nil {
		utils.Log.Printf("ERROR: Failed to read custom channels file: %v\n", err)
		return err
	}

	// Parse the JSON
	var customConfig CustomChannelsConfig
	if err := json.Unmarshal(data, &customConfig); err != nil {
		utils.Log.Printf("ERROR: Failed to parse custom channels JSON: %v\n", err)
		return err
	}

	// Convert custom channels to standard Channel format
	for _, customCh := range customConfig.Channels {
		if customCh.ID == "" || customCh.Name == "" || customCh.URL == "" {
			utils.Log.Printf("WARNING: Skipping invalid custom channel (missing required fields): %+v\n", customCh)
			continue
		}

		// Validate category and language
		if customCh.Category < 0 || customCh.Language < 0 {
			utils.Log.Printf("WARNING: Invalid category or language for custom channel %s, using defaults\n", customCh.ID)
			if customCh.Category < 0 {
				customCh.Category = 18 // Other category
			}
			if customCh.Language < 0 {
				customCh.Language = 18 // Other language
			}
		}

		// Add custom_ prefix to ID to avoid conflicts with JioTV channels
		channelID := "custom_" + customCh.ID

		channel := Channel{
			ID:       channelID,
			Name:     customCh.Name,
			URL:      customCh.URL,
			LogoURL:  customCh.LogoURL,
			Category: customCh.Category,
			Language: customCh.Language,
			IsHD:     customCh.IsHD,
		}

		customChannels = append(customChannels, channel)
	}

	utils.Log.Printf("INFO: Loaded %d custom channels from %s\n", len(customChannels), config.Cfg.CustomChannelsPath)
	return nil
}

// GetCustomChannels returns the loaded custom channels
func GetCustomChannels() []Channel {
	return customChannels
}

// IsCustomChannel checks if a channel ID is a custom channel
func IsCustomChannel(channelID string) bool {
	return strings.HasPrefix(channelID, "custom_")
}

// GetCustomChannelURL returns the direct URL for a custom channel
func GetCustomChannelURL(channelID string) (string, error) {
	if !IsCustomChannel(channelID) {
		return "", nil
	}

	for _, channel := range customChannels {
		if channel.ID == channelID {
			return channel.URL, nil
		}
	}

	return "", nil
}

// GetChannelsWithCustom returns all channels (JioTV + custom) combined
func GetChannelsWithCustom() ChannelsResponse {
	// Get JioTV channels
	jioTVChannels := Channels()
	
	// Get custom channels
	customChannelsList := GetCustomChannels()
	
	// Combine both
	allChannels := append(jioTVChannels.Result, customChannelsList...)
	
	return ChannelsResponse{
		Code:    jioTVChannels.Code,
		Message: jioTVChannels.Message,
		Result:  allChannels,
	}
}