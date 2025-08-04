package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/jiotv-go/jiotv_go/v3/internal/config"
	"github.com/jiotv-go/jiotv_go/v3/internal/constants"
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

// Init initializes the TOML file, creates if not exist, otherwise reads and decodes to struct.
func Init() error {
	KVS = &TomlStore{}
	// store_vX.toml, where X is changed whenever new version requires re-login
	filename := filepath.Join(GetPathPrefix(), "store_v4.toml")

	KVS.mu.Lock()
	defer KVS.mu.Unlock()

	KVS.filename = filename
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// Create a new file with an empty configuration.
		KVS.config = Config{
			Data: make(map[string]string),
		}
		return saveConfig()
	}

	// Read and decode existing configuration from the file.
	_, err := toml.DecodeFile(filename, &KVS.config)
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
	file, err := os.Create(KVS.filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	return encoder.Encode(KVS.config)
}

// Errors
var (
	ErrKeyNotFound = errors.New("key not found")
)

const (
	// PATH_PREFIX is the default path prefix for all files managed by JioTV Go.
	PATH_PREFIX = constants.PathPrefix
)

// GetPathPrefix returns the path prefix for all files managed by JioTV Go.
func GetPathPrefix() string {
	pathPrefix := config.Cfg.PathPrefix
	if pathPrefix == "" {
		// add UserHomeDir to pathPrefix
		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(fmt.Errorf("GetPathPrefix: error getting user home directory: %v", err))
		}
		pathPrefix = filepath.Join(homeDir, PATH_PREFIX)
	}

	// if pathPrefix does not exist, create it
	if _, err := os.Stat(pathPrefix); os.IsNotExist(err) {
		if err := os.Mkdir(pathPrefix, 0755); err != nil {
			panic(fmt.Errorf("GetPathPrefix: error creating pathPrefix: %v", err))
		}
	}

	// if pathPrefix does not have a trailing slash, add it
	if pathPrefix[len(pathPrefix)-1] != '/' {
		pathPrefix += "/"
	}

	return pathPrefix
}

// SetupTestPathPrefix sets up a temporary directory for testing and configures
// the pathPrefix to use it. Returns a cleanup function that should be called
// when the test is complete to restore the original configuration and clean up
// the temporary directory.
func SetupTestPathPrefix() (cleanup func(), err error) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "jiotv_go_test_*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %v", err)
	}

	// Save the original pathPrefix
	originalPathPrefix := config.Cfg.PathPrefix

	// Set the pathPrefix to the temporary directory
	config.Cfg.PathPrefix = tempDir

	// Reset the global store to nil so Init() can be called again
	KVS = nil

	// Return cleanup function
	cleanup = func() {
		// Restore the original pathPrefix
		config.Cfg.PathPrefix = originalPathPrefix
		// Reset KVS to nil
		KVS = nil
		// Clean up the temporary directory
		os.RemoveAll(tempDir)
	}

	return cleanup, nil
}
