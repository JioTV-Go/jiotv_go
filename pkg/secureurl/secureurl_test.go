package secureurl

import (
	"testing"
)

func TestGenerateKey(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Generate encryption key",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateKey()
			// Key should be 32 bytes (256 bits) for AES-256
			if len(got) != 32 {
				t.Errorf("generateKey() returned key of length %d, want 32", len(got))
			}
			// Should not be all zeros
			allZeros := true
			for _, b := range got {
				if b != 0 {
					allZeros = false
					break
				}
			}
			if allZeros {
				t.Errorf("generateKey() returned all zeros")
			}
		})
	}
}

func TestEncryptURL(t *testing.T) {
	// Initialize the package first
	Init()

	type args struct {
		inputURL string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Basic URL encryption",
			args:    args{inputURL: "https://example.com/test"},
			wantErr: false,
		},
		{
			name:    "URL with parameters",
			args:    args{inputURL: "https://example.com/test?param1=value1&param2=value2"},
			wantErr: false,
		},
		{
			name:    "Empty URL",
			args:    args{inputURL: ""},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncryptURL(tt.args.inputURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncryptURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Should return a non-empty string for valid inputs
				if len(got) == 0 && len(tt.args.inputURL) > 0 {
					t.Errorf("EncryptURL() returned empty string for non-empty input")
				}
			}
		})
	}
}

func TestDecryptURL(t *testing.T) {
	// Initialize the package first
	Init()

	// Test round-trip encryption/decryption
	testURLs := []string{
		"https://example.com/test",
		"https://example.com/test?param1=value1&param2=value2",
		"",
	}

	for _, testURL := range testURLs {
		t.Run("Round-trip test for: "+testURL, func(t *testing.T) {
			// First encrypt
			encrypted, err := EncryptURL(testURL)
			if err != nil {
				t.Errorf("EncryptURL() error = %v", err)
				return
			}

			// Then decrypt
			decrypted, err := DecryptURL(encrypted)
			if err != nil {
				t.Errorf("DecryptURL() error = %v", err)
				return
			}

			// Should match original
			if decrypted != testURL {
				t.Errorf("DecryptURL() = %v, want %v", decrypted, testURL)
			}
		})
	}

	t.Run("Invalid encrypted URL", func(t *testing.T) {
		_, err := DecryptURL("invalid-base64!")
		if err == nil {
			t.Errorf("DecryptURL() should return error for invalid input")
		}
	})
}

func TestInit(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Initialize secureurl package",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init()
			// Test that the package is initialized by trying to encrypt something
			_, err := EncryptURL("test")
			if err != nil {
				t.Errorf("Init() failed - EncryptURL returned error: %v", err)
			}
		})
	}
}
