package config

import (
	"os"
	"log"
	"reflect"

	"github.com/ilyakaznacheev/cleanenv"
)

type JioTVConfig struct {
	// Enable Or Disable EPG Generation. Default: false
	EPG bool `yaml:"epg" env:"JIOTV_EPG"`
	// Enable Or Disable Debug Mode. Default: false
	Debug bool `yaml:"debug" env:"JIOTV_DEBUG"`
	// Enable Or Disable TS Handler. While TS Handler is enabled, the server will serve the TS files directly from JioTV API. Default: false
	DisableTSHandler bool `yaml:"disable_ts_handler" env:"JIOTV_DISABLE_TS_HANDLER"`
	// Enable Or Disable Logout feature. Default: true
	DisableLogout bool `yaml:"disable_logout" env:"JIOTV_DISABLE_LOGOUT"`
	// Enable Or Disable DRM. As DRM is not supported by most of the players, it is disabled by default. Default: false
	DRM bool `yaml:"drm" env:"JIOTV_DRM"`
	// Title of the server. Default: JioTV Go
	Title string `yaml:"title" env:"JIOTV_TITLE"`
	// Enable Or Disable URL Encryption. URL Encryption prevents hackers from injecting URLs into the server. Default: true
	DisableURLEncryption bool `yaml:"disable_url_encryption" env:"JIOTV_DISABLE_URL_ENCRYPTION"`
	// Path to the credentials file. Default: credentials.json
	CredentialsPath string `yaml:"credentials_path" env:"JIOTV_CREDENTIALS_PATH"`
	// Proxy URL. Proxy is useful to bypass geo-restrictions and ip-restrictions for JioTV API. Default: ""
	Proxy string `yaml:"proxy" env:"JIOTV_PROXY"`
}

// Global config variable
var Cfg JioTVConfig

func (c *JioTVConfig) Load(filename string) error {
	if filename == "" {
		//  check if config.yml exists
		if fileExists("config.yml") {
			filename = "config.yml"
		} else {
			log.Println("INFO: No config file found, using environment variables")
			return cleanenv.ReadEnv(c)
		}
	}
	log.Println("INFO: Using config file:", filename)
	return cleanenv.ReadConfig(filename, c)
}

func (c *JioTVConfig) Get(key string) interface{} {
	r := reflect.ValueOf(Cfg)
	f := reflect.Indirect(r).FieldByName(key)
	if f.IsValid() {
		return f.Interface()
	}
	return nil
}

// FileExists function check if given file exists
func fileExists(filename string) bool {
	// check if given file exists
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}
