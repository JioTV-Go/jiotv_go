package secureurl

import (
	"reflect"
	"testing"
)

func Test_generateKey(t *testing.T) {
	tests := []struct {
		name        string
		wantLength  int
		checkUnique bool // Flag to check if subsequent calls generate unique keys
	}{
		{
			name:       "Check key length",
			wantLength: 32,
		},
		{
			name:        "Check key uniqueness",
			wantLength:  32, // Still check length for this test
			checkUnique: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateKey()
			if len(got) != tt.wantLength {
				t.Errorf("generateKey() length = %v, want %v", len(got), tt.wantLength)
			}

			if tt.checkUnique {
				got2 := generateKey()
				if reflect.DeepEqual(got, got2) {
					t.Errorf("generateKey() called twice, expected different keys but got the same")
				}
			}
		})
	}
}

func TestEncryptURL(t *testing.T) {
	originalDisableConfig := config.Cfg.DisableURLEncryption
	originalKey := key 
	defer func() {
		config.Cfg.DisableURLEncryption = originalDisableConfig
		key = originalKey 
		Init() 
	}()

	testURL := "http://example.com/test?query=123&another=param"
	emptyURL := ""

	t.Run("Encryption Enabled", func(t *testing.T) {
		config.Cfg.DisableURLEncryption = false
		Init() 

		if key == nil {
			t.Fatal("Key not initialized after Init() with encryption enabled")
		}

		encrypted1, err1 := EncryptURL(testURL)
		if err1 != nil {
			t.Fatalf("EncryptURL() with valid URL error = %v, wantErr false", err1)
		}
		if encrypted1 == "" {
			t.Error("EncryptURL() with valid URL returned empty string")
		}
		_, err := base64.URLEncoding.DecodeString(encrypted1)
		if err != nil {
			t.Errorf("EncryptURL() output for valid URL is not valid Base64: %v", err)
		}

		encryptedEmpty, errEmpty := EncryptURL(emptyURL)
		if errEmpty != nil {
			t.Fatalf("EncryptURL() with empty URL error = %v, wantErr false", errEmpty)
		}
		if encryptedEmpty == "" {
			t.Error("EncryptURL() with empty URL returned empty string, expected IV block")
		}
		decodedEmpty, err := base64.URLEncoding.DecodeString(encryptedEmpty)
		if err != nil {
			t.Errorf("EncryptURL() output for empty URL is not valid Base64: %v", err)
		}
		if len(decodedEmpty) < aes.BlockSize {
			t.Errorf("EncryptURL() output for empty URL is too short, got %d, want at least %d", len(decodedEmpty), aes.BlockSize)
		}


		encrypted2, err2 := EncryptURL(testURL)
		if err2 != nil {
			t.Fatalf("EncryptURL() second call error = %v, wantErr false", err2)
		}
		if encrypted1 == encrypted2 {
			t.Error("EncryptURL() called twice on same URL, got same ciphertext, want different (due to random IV)")
		}
	})

	t.Run("Encryption Disabled", func(t *testing.T) {
		config.Cfg.DisableURLEncryption = true
		Init() 

		escapedTestURL := url.QueryEscape(testURL)
		got, err := EncryptURL(testURL)
		if err != nil {
			t.Errorf("EncryptURL() with encryption disabled error = %v, wantErr false", err)
		}
		if got != escapedTestURL {
			t.Errorf("EncryptURL() with encryption disabled = %q, want %q", got, escapedTestURL)
		}

		escapedEmptyURL := url.QueryEscape(emptyURL)
		gotEmpty, errEmpty := EncryptURL(emptyURL)
		if errEmpty != nil {
			t.Errorf("EncryptURL() with empty URL and encryption disabled error = %v, wantErr false", errEmpty)
		}
		if gotEmpty != escapedEmptyURL {
			t.Errorf("EncryptURL() with empty URL and encryption disabled = %q, want %q", gotEmpty, escapedEmptyURL)
		}
	})
}

func TestDecryptURL(t *testing.T) {
	originalDisableConfig := config.Cfg.DisableURLEncryption
	originalKey := key
	defer func() {
		config.Cfg.DisableURLEncryption = originalDisableConfig
		key = originalKey
		Init()
	}()

	testURL := "http://example.com/test?query=123&another=param with spaces"
	
	t.Run("Encryption Enabled - Success", func(t *testing.T) {
		config.Cfg.DisableURLEncryption = false
		Init() // Generate a key
		if key == nil {t.Fatal("Key not initialized for encryption enabled success test")}

		encrypted, err := EncryptURL(testURL)
		if err != nil {
			t.Fatalf("EncryptURL failed: %v", err)
		}

		decrypted, err := DecryptURL(encrypted)
		if err != nil {
			t.Errorf("DecryptURL() error = %v, wantErr false. Encrypted: %s", err, encrypted)
		}
		if decrypted != testURL {
			t.Errorf("DecryptURL() = %q, want %q", decrypted, testURL)
		}
	})

	t.Run("Encryption Enabled - Error Cases", func(t *testing.T) {
		config.Cfg.DisableURLEncryption = false
		Init() // Ensure key is set for this sub-test block
		if key == nil {t.Fatal("Key not initialized for encryption enabled error cases test")}


		// Case: Encrypted with a different key
		t.Run("Encrypted with different key", func(t *testing.T) {
			config.Cfg.DisableURLEncryption = false // Ensure it's still false
			Init() // Key K1 generated and stored in package var `key`
			dataToEncrypt := "some secret data"
			encryptedWithK1, _ := EncryptURL(dataToEncrypt)

			Init() // Key K2 generated and replaces K1 in package var `key`
			
			decryptedWithK2, err := DecryptURL(encryptedWithK1)
			if err != nil {
				// This is also acceptable, CFB might error if padding is wrong or due to other reasons
				// t.Logf("DecryptURL() with different key returned error as expected: %v", err)
			} else if decryptedWithK2 == dataToEncrypt {
				// This is the main failure condition for this test if no error occurred.
				t.Errorf("DecryptURL() with different key unexpectedly decrypted correctly to %q", decryptedWithK2)
			}
			// If err is not nil, and decryptedWithK2 is not dataToEncrypt, it's a pass.
		})

		// Reset key for subsequent error cases that might rely on a consistent (but not necessarily specific) key
		Init()


		tests := []struct {
			name         string
			encryptedURL string
			wantErrMsg   string 
		}{
			{"Malformed Base64", "not-base64-%", "illegal base64 data"},
			{"Ciphertext too short", base64.URLEncoding.EncodeToString([]byte("short")), "ciphertext too short"},
			{"Empty encrypted string", "", "illegal base64 data"}, 
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := DecryptURL(tt.encryptedURL)
				if err == nil {
					t.Errorf("DecryptURL() expected an error for %q, but got nil", tt.name)
				} else if tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("DecryptURL() for %q error = %q, want error containing %q", tt.name, err.Error(), tt.wantErrMsg)
				}
			})
		}
	})

	t.Run("Encryption Disabled", func(t *testing.T) {
		config.Cfg.DisableURLEncryption = true
		Init()

		urlToTest := "http://example.com/test?query=123&another=param with spaces"
		escapedURL := url.QueryEscape(urlToTest)
		
		decrypted, err := DecryptURL(escapedURL)
		if err != nil {
			t.Errorf("DecryptURL() with encryption disabled error = %v, wantErr false", err)
		}
		if decrypted != urlToTest {
			t.Errorf("DecryptURL() with encryption disabled = %q, want %q", decrypted, urlToTest)
		}

		decryptedEmpty, errEmpty := DecryptURL("")
		if errEmpty != nil {
			t.Errorf("DecryptURL() with empty string and encryption disabled error = %v, wantErr false", errEmpty)
		}
		if decryptedEmpty != "" {
			t.Errorf("DecryptURL() with empty string and encryption disabled = %q, want \"\"", decryptedEmpty)
		}
	})
}

func TestInit(t *testing.T) {
	// Store original config and key values to restore after test
	originalDisableConfig := config.Cfg.DisableURLEncryption
	originalKey := key
	defer func() {
		config.Cfg.DisableURLEncryption = originalDisableConfig
		key = originalKey
		// Re-run Init() to restore original package state if necessary,
		// or ensure other tests call Init() themselves if they depend on it.
		// For simplicity here, we assume tests are independent or will re-Init.
		Init()
	}()

	tests := []struct {
		name                   string
		disableEncryption      bool
		expectKeyInitialization bool
		keyShouldChange        bool // Only relevant if expectKeyInitialization is true
	}{
		{
			name:                   "Encryption enabled - key initialized",
			disableEncryption:      false,
			expectKeyInitialization: true,
			keyShouldChange:        true, // Key should be new after Init()
		},
		{
			name:                   "Encryption disabled - key not initialized",
			disableEncryption:      true,
			expectKeyInitialization: false,
		},
		{
			name:                   "Encryption enabled - Init called twice - key changes",
			disableEncryption:      false,
			expectKeyInitialization: true,
			keyShouldChange:        true, // Key should change on second Init() too
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set config for the current test case
			config.Cfg.DisableURLEncryption = tt.disableEncryption
			key = nil // Reset key before each Init call for this test structure

			Init()

			if tt.expectKeyInitialization {
				if key == nil {
					t.Errorf("Init() with encryption enabled: key is nil, want non-nil")
				}
				if len(key) != 32 {
					t.Errorf("Init() with encryption enabled: key length = %d, want 32", len(key))
				}

				if tt.keyShouldChange {
					keyBefore := make([]byte, len(key))
					copy(keyBefore, key)
					
					Init() // Call Init again
					
					if key == nil { // Should still be initialized
						t.Errorf("Second Init() with encryption enabled: key is nil, want non-nil")
					}
					if len(key) != 32 { // Should still be 32 bytes
						t.Errorf("Second Init() with encryption enabled: key length = %d, want 32", len(key))
					}
					if reflect.DeepEqual(key, keyBefore) {
						t.Errorf("Second Init() with encryption enabled: key did not change, want new key")
					}
				}
			} else { // Encryption disabled
				if key != nil {
					t.Errorf("Init() with encryption disabled: key is not nil, want nil. Key: %x", key)
				}
			}
		})
	}
}
