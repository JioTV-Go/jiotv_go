package store

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	KVS   KVSInst
	mutex sync.Mutex
)

// KVSInst represents the key-value store.
type KVSInst struct {
	Data map[string]string
}

// Init initializes the key-value store.
func Init() error {
	mutex.Lock()
	defer mutex.Unlock()

	storeFilePath := getStoreFilePath()

	// Check if the folder exists, create it if not
	storeFolder := filepath.Dir(storeFilePath)
	if _, err := os.Stat(storeFolder); os.IsNotExist(err) {
		if err := os.MkdirAll(storeFolder, 0755); err != nil {
			return fmt.Errorf("error creating .jiotv_go folder: %v", err)
		}
	}

	// Create the store file
	storeFile, err := os.Create(storeFilePath)
	if err != nil {
		return fmt.Errorf("error creating store file: %v", err)
	}
	defer storeFile.Close()

	KVS = KVSInst{
		Data: make(map[string]string),
	}

	return saveStore()
}

// Get retrieves the value associated with the given key.
func Get(key string) (string, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if err := loadStore(); err != nil {
		return "", fmt.Errorf("error loading store: %v", err)
	}

	val, exists := KVS.Data[key]
	if !exists {
		return "", fmt.Errorf("key not found: %s", key)
	}

	return val, nil
}

// Set sets the value for the given key.
func Set(key, val string) error {
	mutex.Lock()
	defer mutex.Unlock()

	if err := loadStore(); err != nil {
		return fmt.Errorf("error loading store: %v", err)
	}

	KVS.Data[key] = val

	return saveStore()
}

// Delete removes the entry associated with the given key.
func Delete(key string) error {
	mutex.Lock()
	defer mutex.Unlock()

	if err := loadStore(); err != nil {
		return fmt.Errorf("error loading store: %v", err)
	}

	delete(KVS.Data, key)

	return saveStore()
}

// getStoreFilePath returns the full path to the store file.
func getStoreFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("error getting user home directory: %v", err))
	}

	return filepath.Join(homeDir, ".jiotv_go", ".store")
}

// loadStore loads the key-value store from the file.
func loadStore() error {
	storeFile, err := os.Open(getStoreFilePath())
	if err != nil {
		return fmt.Errorf("error opening store file: %v", err)
	}
	defer storeFile.Close()

	decoder := gob.NewDecoder(storeFile)
	if err := decoder.Decode(&KVS); err != nil {
		return fmt.Errorf("error decoding store data: %v", err)
	}

	return nil
}

// saveStore saves the key-value store to the file.
func saveStore() error {
	storeFile, err := os.Create(getStoreFilePath())
	if err != nil {
		return fmt.Errorf("error creating store file: %v", err)
	}
	defer storeFile.Close()

	encoder := gob.NewEncoder(storeFile)
	if err := encoder.Encode(KVS); err != nil {
		return fmt.Errorf("error encoding store data: %v", err)
	}

	return nil
}
