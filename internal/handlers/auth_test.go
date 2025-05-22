package handlers

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
)

func TestLoginSendOTPHandler(t *testing.T) {
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := LoginSendOTPHandler(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("LoginSendOTPHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoginVerifyOTPHandler(t *testing.T) {
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := LoginVerifyOTPHandler(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("LoginVerifyOTPHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoginPasswordHandler(t *testing.T) {
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := LoginPasswordHandler(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("LoginPasswordHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLogoutHandler(t *testing.T) {
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := LogoutHandler(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("LogoutHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoginRefreshAccessToken(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := LoginRefreshAccessToken(); (err != nil) != tt.wantErr {
				t.Errorf("LoginRefreshAccessToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoginRefreshSSOToken(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := LoginRefreshSSOToken(); (err != nil) != tt.wantErr {
				t.Errorf("LoginRefreshSSOToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRefreshTokenIfExpired(t *testing.T) {
	type args struct {
		credentials *utils.JIOTV_CREDENTIALS
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RefreshTokenIfExpired(tt.args.credentials); (err != nil) != tt.wantErr {
				t.Errorf("RefreshTokenIfExpired() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRefreshSSOTokenIfExpired(t *testing.T) {
	type args struct {
		credentials *utils.JIOTV_CREDENTIALS
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RefreshSSOTokenIfExpired(tt.args.credentials); (err != nil) != tt.wantErr {
				t.Errorf("RefreshSSOTokenIfExpired() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
