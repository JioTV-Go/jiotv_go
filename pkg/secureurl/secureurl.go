package secureurl

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/url"

	"github.com/rabilrbl/jiotv_go/v3/internal/config"
	"github.com/rabilrbl/jiotv_go/v3/pkg/utils"
)

var (
	key                  []byte
	disableUrlEncryption bool
)

func generateKey() []byte {
	key := make([]byte, 32) // 32 bytes for AES-256
	_, err := rand.Read(key)
	if err != nil {
		utils.Log.Panicln("Error generating random key: ", err)
	}
	return key
}

func EncryptURL(inputURL string) (string, error) {
	if disableUrlEncryption {
		return url.QueryEscape(inputURL), nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(inputURL))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(inputURL))

	encryptedURL := base64.URLEncoding.EncodeToString(ciphertext)
	return encryptedURL, nil
}

func DecryptURL(encryptedURL string) (string, error) {
	if disableUrlEncryption {
		decoded_url, err := url.QueryUnescape(encryptedURL)
		return decoded_url, err
	}

	ciphertext, err := base64.URLEncoding.DecodeString(encryptedURL)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	decryptedURL := string(ciphertext)

	return decryptedURL, nil
}

func Init() {
	disableUrlEncryption = config.Cfg.DisableURLEncryption
	if disableUrlEncryption {
		fmt.Println("Warning! URL encryption is disabled. Anyone can pass modified URLs to your server.")
		return
	}
	key = generateKey()
}
