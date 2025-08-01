package secureurl

import (
	"reflect"
	"testing"
)

func Test_generateKey(t *testing.T) {
	tests := []struct {
		name string
		want []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateKey(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateKey() = %v, want %v", got, tt.want)
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
