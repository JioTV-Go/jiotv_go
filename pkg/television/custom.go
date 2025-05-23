package television

import (
	"encoding/json"
	"fmt"
	"os" // Using os.ReadFile

	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
)

// LoadCustomChannels reads a JSON file containing custom channel definitions,
// converts them to the standard Channel struct, and returns them.
//
// Parameters:
//   - customChannelsPath: The file path to the JSON configuration for custom channels.
//
// Returns:
//   - A slice of Channel structs populated from the custom configuration.
//   - An error if the file cannot be read or parsed, or if there's a critical issue
//     with the data. Returns (nil, nil) if customChannelsPath is empty.
func LoadCustomChannels(customChannelsPath string) ([]Channel, error) {
	if customChannelsPath == "" {
		return nil, nil // No path provided, so no custom channels to load.
	}

	data, err := os.ReadFile(customChannelsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read custom channels file '%s': %w", customChannelsPath, err)
	}

	if len(data) == 0 {
		// File is empty, treat as no custom channels
		utils.Log.Printf("Custom channels file '%s' is empty.", customChannelsPath)
		return nil, nil
	}
	
	var config CustomChannelsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse custom channels JSON from '%s': %w", customChannelsPath, err)
	}

	var channels []Channel
	for _, customCh := range config.Channels {
		// Basic mapping
		ch := Channel{
			ID:       customCh.ID,
			Name:     customCh.Name,
			LogoURL:  customCh.LogoURL,
			URL:      customCh.URL, // Mapping custom URL to Channel.URL
			// IsHD is removed from CustomChannel, so Channel.IsHD will be its zero-value (false)
			// TODO: Handle EPGID for custom channels (customCh.EPGID)
		}

		// Map Category (string to int)
		categoryID := -1 // Use -1 to indicate not found initially
		for id, name := range CategoryMap {
			if name == customCh.Category {
				categoryID = id
				break
			}
		}
		if categoryID != -1 {
			ch.Category = categoryID
		} else {
			utils.Log.Printf("Custom channel '%s': Category '%s' not found in CategoryMap. Defaulting to 'All Categories' (ID 0).", customCh.Name, customCh.Category)
			ch.Category = 0 // Default to "All Categories"
		}

		// Map Language (string to int)
		languageID := -1 // Use -1 to indicate not found initially
		for id, name := range LanguageMap {
			if name == customCh.Language {
				languageID = id
				break
			}
		}
		if languageID != -1 {
			ch.Language = languageID
		} else {
			utils.Log.Printf("Custom channel '%s': Language '%s' not found in LanguageMap. Defaulting to 'Other' (ID 18).", customCh.Name, customCh.Language)
			ch.Language = 18 // Default to "Other"
		}

		channels = append(channels, ch)
	}

	return channels, nil
}
