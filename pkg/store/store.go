package store

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/rabilrbl/jiotv_go/v3/internal/config"
)

const (
	// PATH_PREFIX is the prefix for all file paths managed by JioTV Go.
	PATH_PREFIX = ".jiotv_go"
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

	return filepath.Join(homeDir, GetPathPrefix(), ".store")
}

// getKeyFilePath returns the full path to the key file.
func getKeyFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("error getting user home directory: %v", err))
	}

	return filepath.Join(homeDir, GetPathPrefix(), ".store_pass")
}

// loadStore loads the key-value store from the file.
func loadStore() error {
	storeFile, err := os.Open(getStoreFilePath())
	if err != nil {
		return fmt.Errorf("error opening store file: %v", err)
	}
	defer storeFile.Close()

	// Read the encrypted data
	encryptedData, err := io.ReadAll(storeFile)
	if err != nil {
		return fmt.Errorf("error reading store file: %v", err)
	}

	// Decrypt the data
	key, err := getKey()
	if err != nil {
		return fmt.Errorf("error getting encryption key: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("error creating cipher block: %v", err)
	}

	if len(encryptedData) < aes.BlockSize {
		return errors.New("ciphertext too short")
	}

	iv := encryptedData[:aes.BlockSize]
	encryptedData = encryptedData[aes.BlockSize:]

	// Use CBC mode with PKCS7 padding
	if len(encryptedData)%aes.BlockSize != 0 {
		return errors.New("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(encryptedData, encryptedData)

	// Unmarshal the decrypted data
	decoder := gob.NewDecoder(bytes.NewReader(encryptedData))
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

	// Marshal the data
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(KVS); err != nil {
		return fmt.Errorf("error encoding store data: %v", err)
	}

	// Pad the data using PKCS7
	padding := aes.BlockSize - buf.Len()%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	buf.Write(padtext)

	// Encrypt the data
	key, err := getKey()
	if err != nil {
		return fmt.Errorf("error getting encryption key: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("error creating cipher block: %v", err)
	}

	ciphertext := make([]byte, aes.BlockSize+buf.Len())
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return fmt.Errorf("error generating IV: %v", err)
	}

	// Use CBC mode with PKCS7 padding
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], buf.Bytes())

	// Write the encrypted data to the file
	if _, err := storeFile.Write(ciphertext); err != nil {
		return fmt.Errorf("error writing to store file: %v", err)
	}

	return nil
}

// getKey reads the encryption key from the .key file. If the file does not exist,
// it generates a new key and stores it in the file.
func getKey() ([]byte, error) {
	keyFilePath := getKeyFilePath()
	key, err := os.ReadFile(keyFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Generate a new key
			key = make([]byte, 32)
			if _, err := rand.Read(key); err != nil {
				return nil, fmt.Errorf("error generating key: %v", err)
			}

			// Store the key in the file
			if err := os.WriteFile(keyFilePath, key, 0600); err != nil {
				return nil, fmt.Errorf("error writing key file: %v", err)
			}
		} else {
			return nil, fmt.Errorf("error reading key file: %v", err)
		}
	}

	// Ensure the key is 32 bytes long
	if len(key) != 32 {
		return nil, errors.New("invalid key length, must be 32 bytes")
	}

	return key, nil
}

// GetPathPrefix returns the path prefix for all files managed by JioTV Go.
func GetPathPrefix() string {
	if config.Cfg.PathPrefix == "" {
		return PATH_PREFIX
	}
	return config.Cfg.PathPrefix
}
