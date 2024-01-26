package config

import (
	"log"
	"os"
	"path/filepath"
	"reflect"

	"github.com/ilyakaznacheev/cleanenv"
)

// JioTVConfig defines the configuration options for the JioTV client.
// It includes options for enabling features like EPG, debug mode, DRM, etc.
// As well as configuration for credentials, proxies, file paths and more.
type JioTVConfig struct {
	// Enable Or Disable EPG Generation. Default: false
	EPG bool `yaml:"epg" env:"JIOTV_EPG" json:"epg" toml:"epg"`
	// Enable Or Disable Debug Mode. Default: false
	Debug bool `yaml:"debug" env:"JIOTV_DEBUG" json:"debug" toml:"debug"`
	// Enable Or Disable TS Handler. While TS Handler is enabled, the server will serve the TS files directly from JioTV API. Default: false
	DisableTSHandler bool `yaml:"disable_ts_handler" env:"JIOTV_DISABLE_TS_HANDLER" json:"disable_ts_handler" toml:"disable_ts_handler"`
	// Enable Or Disable Logout feature. Default: true
	DisableLogout bool `yaml:"disable_logout" env:"JIOTV_DISABLE_LOGOUT" json:"disable_logout" toml:"disable_logout"`
	// Enable Or Disable DRM. As DRM is not supported by most of the players, it is disabled by default. Default: false
	DRM bool `yaml:"drm" env:"JIOTV_DRM" json:"drm" toml:"drm"`
	// Title of the webpage. Default: JioTV Go
	Title string `yaml:"title" env:"JIOTV_TITLE" json:"title" toml:"title"`
	// Enable Or Disable URL Encryption. URL Encryption prevents hackers from injecting URLs into the server. Default: true
	DisableURLEncryption bool `yaml:"disable_url_encryption" env:"JIOTV_DISABLE_URL_ENCRYPTION" json:"disable_url_encryption" toml:"disable_url_encryption"`
	// Proxy URL. Proxy is useful to bypass geo-restrictions and ip-restrictions for JioTV API. Default: ""
	Proxy string `yaml:"proxy" env:"JIOTV_PROXY" json:"proxy" toml:"proxy"`
	// PathPrefix is the prefix for all file paths managed by JioTV Go. Default: "$HOME/.jiotv_go"
	PathPrefix string `yaml:"path_prefix" env:"JIOTV_PATH_PREFIX" json:"path_prefix" toml:"path_prefix"`
}

// Cfg is the global config variable
var Cfg JioTVConfig

// Load loads the JioTVConfig from a file.
// It first checks if a filename is provided, otherwise tries to find a common config file.
// If no file is found, it loads config from environment variables.
// It logs messages about which config source is being used.
func (c *JioTVConfig) Load(filename string) error {
	if filename == "" {
		filename = commonFileExists()
	}
	if filename == "" {
		log.Println("INFO: No config file found, using environment variables")
		return cleanenv.ReadEnv(c)
	}
	log.Println("INFO: Using config file:", filename)
	return cleanenv.ReadConfig(filename, c)
}

// Get retrieves the value of the config field specified by key.
// It uses reflection to get the field value from the global Cfg variable.
// Returns the field value as an interface{}, or nil if the field is invalid.
func (*JioTVConfig) Get(key string) interface{} {
	r := reflect.ValueOf(Cfg)
	f := reflect.Indirect(r).FieldByName(key)
	if f.IsValid() {
		return f.Interface()
	}
	return nil
}

// commonFileExists checks for the existence of common config
// file names and returns the first one found. It searches
// for config files in the following formats:
//   - jiotv_go.{yml,toml,json}
//   - config.{json,yml,toml}
//
// If no file is found, an empty string is returned.
func commonFileExists() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Println("ERROR: Unable to get user home directory:", err)
	}
	pathPrefix := filepath.Join(homeDir, ".jiotv_go")
	commonFiles := []string{"jiotv_go.yml", "jiotv_go.toml", "jiotv_go.json", "config.json", "config.yml", "config.toml", pathPrefix + "config.json", pathPrefix + "config.yml", pathPrefix + "config.toml"}
	for _, filename := range commonFiles {
		if _, err := os.Stat(filename); err == nil {
			return filename
		}
	}
	return ""
}
