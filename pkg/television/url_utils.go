package television

import (
	"fmt"
	"strings"

	"github.com/jiotv-go/jiotv_go/v3/pkg/secureurl"
	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
)

// EncryptedURLConfig holds configuration for URL encryption with auth parameters
type EncryptedURLConfig struct {
	BaseURL     string
	Match       string
	Params      string
	ChannelID   string
	EndpointURL string // The endpoint URL pattern (e.g., "/render.m3u8", "/render.ts")
	Quality     string // Quality parameter for live streams
	Hdnea       string // Akamai token value to be appended as query param hdnea
}

// CreateEncryptedURL creates an encrypted URL with auth parameters for various endpoints
func CreateEncryptedURL(config EncryptedURLConfig) ([]byte, error) {
	fullURL := config.BaseURL + config.Match
	if config.Params != "" {
		sep := "?"
		if strings.Contains(fullURL, "?") {
			sep = "&"
		}
		fullURL += sep + config.Params
	}

	encryptedURL, err := secureurl.EncryptURL(fullURL)
	if err != nil {
		utils.Log.Println(err)
		return nil, err
	}

	var result string
	if config.ChannelID != "" {
		result = fmt.Sprintf("%s?auth=%s&channel_key_id=%s", config.EndpointURL, encryptedURL, config.ChannelID)
	} else {
		result = fmt.Sprintf("%s?auth=%s", config.EndpointURL, encryptedURL)
	}

	if config.Quality != "" {
		result += "&q=" + config.Quality
	}

	if config.Hdnea != "" {
		result += "&hdnea=" + config.Hdnea
	}

	return []byte(result), nil
}
