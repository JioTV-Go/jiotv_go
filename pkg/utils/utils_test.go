package utils

import (
	"log"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/jiotv-go/jiotv_go/v3/pkg/store"
)

var (
	setupOnce sync.Once
)

// Setup function to initialize store for tests
func setupTest() {
	setupOnce.Do(func() {
		// Initialize store for testing
		store.Init()
		// Initialize the Log variable to prevent nil pointer dereference
		if Log == nil {
			Log = log.New(os.Stdout, "", log.LstdFlags)
		}
	})
}

func TestGetLogger(t *testing.T) {
	setupTest() // Initialize store
	tests := []struct {
		name string
	}{
		{
			name: "Get logger instance",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetLogger()
			if got == nil {
				t.Errorf("GetLogger() returned nil")
			}
		})
	}
}

func TestLoginSendOTP(t *testing.T) {

	type args struct {
		number string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Empty phone number",
			args: args{
				number: "",
			},
			wantErr: true, // Should handle empty input gracefully
		},
		{
			name: "Valid phone number",
			args: args{
				number: "1234567890",
			},
			wantErr: false, // Should succeed with mock server
		},
		{
			name: "Invalid phone number format",
			args: args{
				number: "invalid",
			},
			wantErr: false, // Mock server accepts any non-empty number
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO
		})
	}
}

func TestLoginVerifyOTP(t *testing.T) {

	type args struct {
		number string
		otp    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Empty inputs",
			args: args{
				number: "",
				otp:    "",
			},
			wantErr: true,
		},
		{
			name: "Valid credentials",
			args: args{
				number: "1234567890",
				otp:    "123456",
			},
			wantErr: false,
		},
		{
			name: "Valid number, empty OTP",
			args: args{
				number: "1234567890",
				otp:    "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO
		})
	}
}

func TestLogin(t *testing.T) {

	type args struct {
		username string
		password string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Empty credentials",
			args: args{
				username: "",
				password: "",
			},
			wantErr: true,
		},
		{
			name: "Valid credentials",
			args: args{
				username: "test@example.com",
				password: "testpassword",
			},
			wantErr: false,
		},
		{
			name: "Valid phone number credentials",
			args: args{
				username: "1234567890",
				password: "testpassword",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO
		})
	}
}

func TestGetPathPrefix(t *testing.T) {
	setupTest() // Initialize store
	tests := []struct {
		name string
	}{
		{
			name: "Get path prefix",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPathPrefix()
			// Should return a non-empty string
			if got == "" {
				t.Errorf("GetPathPrefix() returned empty string")
			}
			// Should end with a separator
			if !strings.HasSuffix(got, string(os.PathSeparator)) {
				t.Errorf("GetPathPrefix() = %v, should end with path separator", got)
			}
		})
	}
}

func TestGetDeviceID(t *testing.T) {
	setupTest() // Initialize store
	tests := []struct {
		name string
	}{
		{
			name: "Get device ID",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDeviceID()
			// Should return a non-empty string (either existing or newly generated)
			if got == "" {
				t.Errorf("GetDeviceID() returned empty string")
			}
			// Should be a hex string (16 characters if newly generated)
			if len(got) > 0 {
				// Check if it's a valid hex string
				for _, c := range got {
					if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
						t.Errorf("GetDeviceID() returned invalid hex character: %c", c)
					}
				}
			}
		})
	}
}

func TestGetJIOTVCredentials(t *testing.T) {
	setupTest() // Initialize store
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Get credentials (may exist or not)",
			wantErr: false, // Error is acceptable if credentials don't exist
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetJIOTVCredentials()
			// The function may return error if no credentials exist, which is valid
			if err != nil && got != nil {
				t.Errorf("GetJIOTVCredentials() should return nil when error occurs, got %v", got)
			}
			// If no error, should return a valid credential struct
			if err == nil && got == nil {
				t.Errorf("GetJIOTVCredentials() should return credentials when no error occurs")
			}
		})
	}
}

func TestWriteJIOTVCredentials(t *testing.T) {
	setupTest() // Initialize store
	type args struct {
		credentials *JIOTV_CREDENTIALS
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Write valid credentials",
			args: args{
				credentials: &JIOTV_CREDENTIALS{
					SSOToken:     "test_sso_token",
					UniqueID:     "test_unique_id",
					CRM:          "test_crm",
					AccessToken:  "test_access_token",
					RefreshToken: "test_refresh_token",
				},
			},
			wantErr: false,
		},
		{
			name: "Write credentials with empty fields",
			args: args{
				credentials: &JIOTV_CREDENTIALS{
					SSOToken: "test_sso_token_2",
					UniqueID: "",
					CRM:      "",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := WriteJIOTVCredentials(tt.args.credentials); (err != nil) != tt.wantErr {
				t.Errorf("WriteJIOTVCredentials() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckLoggedIn(t *testing.T) {
	setupTest() // Initialize store
	tests := []struct {
		name string
	}{
		{
			name: "Check if user is logged in",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckLoggedIn()
			// This should return a boolean - got is already declared as bool type,
			// so no need to check if it's true or false
			_ = got // Use the variable to avoid "unused variable" error
		})
	}
}

func TestLogout(t *testing.T) {

	tests := []struct {
		name    string
		setup   func() // Function to set up test conditions
		wantErr bool
	}{
		{
			name: "Logout with no credentials",
			setup: func() {
				// Clear all credentials to simulate no login state
				store.Delete("ssoToken")
				store.Delete("refreshToken")
				store.Delete("accessToken")
			},
			wantErr: false, // Logout should not fail even if no credentials exist
		},
		{
			name: "Logout with valid credentials",
			setup: func() {
				// Set up valid credentials
				WriteJIOTVCredentials(&JIOTV_CREDENTIALS{
					SSOToken:     "test_sso",
					AccessToken:  "test_access",
					RefreshToken: "test_refresh",
					UniqueID:     "test_unique",
				})
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO
		})
	}
}

func TestPerformServerLogout(t *testing.T) {

	tests := []struct {
		name    string
		setup   func() // Function to set up test conditions
		wantErr bool
	}{
		{
			name: "No credentials available",
			setup: func() {
				// Clear all credentials to simulate no login state
				store.Delete("ssoToken")
				store.Delete("refreshToken")
				store.Delete("accessToken")
			},
			wantErr: true,
		},
		{
			name: "Missing refresh token",
			setup: func() {
				// Set up credentials but without refresh token
				WriteJIOTVCredentials(&JIOTV_CREDENTIALS{
					SSOToken:     "test_sso",
					AccessToken:  "test_access",
					RefreshToken: "", // Missing refresh token
				})
			},
			wantErr: true,
		},
		{
			name: "Valid credentials",
			setup: func() {
				// Set up valid credentials including refresh token
				WriteJIOTVCredentials(&JIOTV_CREDENTIALS{
					SSOToken:     "test_sso",
					AccessToken:  "test_access",
					RefreshToken: "test_refresh",
					UniqueID:     "test_unique",
				})
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO
		})
	}
}

func TestGetRequestClient(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Get HTTP client",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetRequestClient()
			if got == nil {
				t.Errorf("GetRequestClient() returned nil")
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Existing file",
			args: args{filename: "utils.go"}, // This file should exist
			want: true,
		},
		{
			name: "Non-existing file",
			args: args{filename: "nonexistent_file.txt"},
			want: false,
		},
		{
			name: "Empty filename",
			args: args{filename: ""},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FileExists(tt.args.filename); got != tt.want {
				t.Errorf("FileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateCurrentTime(t *testing.T) {
	tests := []struct {
		name       string
		wantFormat string
		wantLength int
	}{
		{
			name:       "Current time format",
			wantFormat: "20060102T150405", // Expected format pattern
			wantLength: 15,                // YYYYMMDDTHHMMSS should be 15 characters
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateCurrentTime()

			// Check length
			if len(got) != tt.wantLength {
				t.Errorf("GenerateCurrentTime() length = %v, want %v", len(got), tt.wantLength)
			}

			// Check format by trying to parse it
			if len(got) == 15 {
				// Should have T at position 8
				if got[8] != 'T' {
					t.Errorf("GenerateCurrentTime() should have 'T' at position 8, got %c", got[8])
				}

				// All other characters should be digits
				for i, c := range got {
					if i == 8 { // Skip the 'T'
						continue
					}
					if c < '0' || c > '9' {
						t.Errorf("GenerateCurrentTime() character at position %d should be digit, got %c", i, c)
					}
				}
			}
		})
	}
}

func TestGenerateDate(t *testing.T) {
	tests := []struct {
		name       string
		wantLength int
	}{
		{
			name:       "Date format",
			wantLength: 8, // YYYYMMDD should be 8 characters
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateDate()

			// Check length
			if len(got) != tt.wantLength {
				t.Errorf("GenerateDate() length = %v, want %v", len(got), tt.wantLength)
			}

			// All characters should be digits
			for i, c := range got {
				if c < '0' || c > '9' {
					t.Errorf("GenerateDate() character at position %d should be digit, got %c", i, c)
				}
			}
		})
	}
}

func TestContainsString(t *testing.T) {
	type args struct {
		item  string
		slice []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Item exists in slice",
			args: args{
				item:  "apple",
				slice: []string{"apple", "banana", "cherry"},
			},
			want: true,
		},
		{
			name: "Item does not exist in slice",
			args: args{
				item:  "grape",
				slice: []string{"apple", "banana", "cherry"},
			},
			want: false,
		},
		{
			name: "Empty slice",
			args: args{
				item:  "apple",
				slice: []string{},
			},
			want: false,
		},
		{
			name: "Empty item in slice with empty string",
			args: args{
				item:  "",
				slice: []string{"", "apple", "banana"},
			},
			want: true,
		},
		{
			name: "Empty item not in slice",
			args: args{
				item:  "",
				slice: []string{"apple", "banana"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsString(tt.args.item, tt.args.slice); got != tt.want {
				t.Errorf("ContainsString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateRandomString(t *testing.T) {
	setupTest() // Initialize store
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Generate random string successfully",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := GenerateRandomString(); (err != nil) != tt.wantErr {
				t.Errorf("GenerateRandomString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
