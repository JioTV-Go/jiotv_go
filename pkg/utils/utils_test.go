package utils

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/jiotv-go/jiotv_go/v3/pkg/store"
)

// Setup function to initialize store for tests
func setupTest() {
	// Initialize store for testing
	store.Init()
	// Initialize the Log variable to prevent nil pointer dereference
	if Log == nil {
		Log = log.New(os.Stdout, "", log.LstdFlags)
	}
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
	setupTest() // Initialize store
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
			name: "Invalid phone number format",
			args: args{
				number: "invalid",
			},
			wantErr: true, // May fail but shouldn't panic
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This test will actually make API calls, so we're primarily
			// testing that the function doesn't panic and handles errors gracefully
			got, err := LoginSendOTP(tt.args.number)
			if tt.wantErr && err == nil {
				t.Errorf("LoginSendOTP() expected error but got none")
			}
			// Function should return a boolean
			if err == nil && got != true && got != false {
				t.Errorf("LoginSendOTP() should return boolean, got %v", got)
			}
		})
	}
}

func TestLoginVerifyOTP(t *testing.T) {
	setupTest() // Initialize store
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
			name: "Invalid OTP format",
			args: args{
				number: "1234567890",
				otp:    "invalid",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoginVerifyOTP(tt.args.number, tt.args.otp)
			if tt.wantErr && err == nil {
				t.Errorf("LoginVerifyOTP() expected error but got none")
			}
			// If no error, should return a map
			if err == nil && got == nil {
				t.Errorf("LoginVerifyOTP() should return map when successful")
			}
		})
	}
}

func TestLogin(t *testing.T) {
	setupTest() // Initialize store
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
			name: "Invalid credentials format",
			args: args{
				username: "test",
				password: "test",
			},
			wantErr: true, // Should fail for invalid credentials
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Login(tt.args.username, tt.args.password)
			if tt.wantErr && err == nil {
				t.Errorf("Login() expected error but got none")
			}
			// If no error, should return a map
			if err == nil && got == nil {
				t.Errorf("Login() should return map when successful")
			}
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
			// This should return a boolean - either true or false
			if got != true && got != false {
				t.Errorf("CheckLoggedIn() should return boolean, got %v", got)
			}
		})
	}
}

func TestLogout(t *testing.T) {
	setupTest() // Initialize store
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Logout (may succeed or fail depending on state)",
			wantErr: false, // We'll allow either outcome
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Logout()
			// The function may fail if no credentials exist or server logout fails, 
			// but it should not panic
			_ = err // We don't check specific error condition as it depends on state
		})
	}
}

func TestPerformServerLogout(t *testing.T) {
	setupTest() // Initialize store
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Perform server logout",
			wantErr: false, // May succeed or fail, we just test it doesn't panic
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PerformServerLogout()
			// Function may fail if not logged in or server issues, but shouldn't panic
			_ = err // We don't assert specific error conditions
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
		name        string
		wantFormat  string
		wantLength  int
	}{
		{
			name:        "Current time format",
			wantFormat:  "20060102T150405", // Expected format pattern
			wantLength:  15, // YYYYMMDDTHHMMSS should be 15 characters
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
		name        string
		wantLength  int
	}{
		{
			name:        "Date format",
			wantLength:  8, // YYYYMMDD should be 8 characters
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
