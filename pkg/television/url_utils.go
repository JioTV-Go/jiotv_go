package television

import (
	"fmt"

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
}

// CreateEncryptedURL creates an encrypted URL with auth parameters for various endpoints
func CreateEncryptedURL(config EncryptedURLConfig) ([]byte, error) {
	fullURL := config.BaseURL + config.Match + "?" + config.Params
	
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

	return []byte(result), nil
}