package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/jiotv-go/jiotv_go/v3/internal/config"
)

// Config represents the structure of the TOML file.
type Config struct {
	Data map[string]string `toml:"data"`
}

// TomlStore represents the TOML storage.
type TomlStore struct {
	filename string
	config   Config
	mu       sync.Mutex
}

// KVS represents global key-value store.
var KVS *TomlStore

// osStat is a variable for os.Stat, allowing it to be mocked for testing.
var osStat = os.Stat
// osUserHomeDir is a variable for os.UserHomeDir, allowing it to be mocked for testing.
var osUserHomeDir = os.UserHomeDir
// osMkdir is a variable for os.Mkdir, allowing it to be mocked for testing.
var osMkdir = os.Mkdir
// osCreate is a variable for os.Create, allowing it to be mocked for testing.
var osCreate = os.Create

// tomlDecodeFile is a variable for toml.DecodeFile, for potential mocking (though direct file mocking is preferred).
var tomlDecodeFile = toml.DecodeFile
// tomlEncode is a helper variable for toml.NewEncoder().Encode(), for potential mocking.
var tomlEncode = func(w io.Writer, v interface{}) error {
	return toml.NewEncoder(w).Encode(v)
}


// Init initializes the TOML file, creates if not exist, otherwise reads and decodes to struct.
func Init() error {
	KVS = &TomlStore{}
	// store_vX.toml, where X is changed whenever new version requires re-login
	filename := filepath.Join(getPathPrefix(), "store_v4.toml")

	KVS.mu.Lock()
	defer KVS.mu.Unlock()

	KVS.filename = filename
	if _, err := osStat(filename); os.IsNotExist(err) {
		// Create a new file with an empty configuration.
		KVS.config = Config{
			Data: make(map[string]string),
		}
		return saveConfig()
	}

	// Read and decode existing configuration from the file.
	_, err := tomlDecodeFile(filename, &KVS.config)
	return err
}

// Get retrieves the value for the specified key from the TOML store.
func Get(key string) (string, error) {
	KVS.mu.Lock()
	defer KVS.mu.Unlock()

	value, ok := KVS.config.Data[key]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}
	return value, nil
}

// Set sets the value for the specified key in the TOML store.
func Set(key, value string) error {
	KVS.mu.Lock()
	defer KVS.mu.Unlock()

	KVS.config.Data[key] = value
	return saveConfig()
}

// Delete removes the entry for the specified key from the TOML store.
func Delete(key string) error {
	KVS.mu.Lock()
	defer KVS.mu.Unlock()

	delete(KVS.config.Data, key)
	return saveConfig()
}

// saveConfig saves the current configuration to the TOML file.
func saveConfig() error {
	file, err := osCreate(KVS.filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return tomlEncode(file, KVS.config)
}

// Errors
var (
	ErrKeyNotFound = errors.New("key not found")
)

const (
	// PATH_PREFIX is the default path prefix for all files managed by JioTV Go.
	PATH_PREFIX = ".jiotv_go"
)

// getPathPrefix returns the path prefix for all files managed by JioTV Go.
// Made into a variable for easier mocking in tests.
var getPathPrefix = func() string {
	pathPrefix := config.Cfg.PathPrefix
	if pathPrefix == "" {
		// add UserHomeDir to pathPrefix
		homeDir, err := osUserHomeDir()
		if err != nil {
			panic(fmt.Errorf("GetPathPrefix: error getting user home directory: %v", err))
		}
		pathPrefix = filepath.Join(homeDir, PATH_PREFIX)
	}

	// if pathPrefix does not exist, create it
	if _, err := osStat(pathPrefix); os.IsNotExist(err) {
		if err := osMkdir(pathPrefix, 0755); err != nil {
			panic(fmt.Errorf("GetPathPrefix: error creating pathPrefix: %v", err))
		}
	}

	// if pathPrefix does not have a trailing slash, add it
	// Use filepath.Separator for OS-agnostic path handling.
	// However, the original code explicitly adds '/', so we keep that for now.
	if pathPrefix[len(pathPrefix)-1] != '/' && pathPrefix[len(pathPrefix)-1] != filepath.Separator {
		pathPrefix += string(filepath.Separator) 
	}


	return pathPrefix
}

// GetPathPrefix is the exported version that uses the internal variable getPathPrefix.
// This allows users of the package to use it as before, while tests can mock the internal one.
func GetPathPrefix() string {
	return getPathPrefix()
}
