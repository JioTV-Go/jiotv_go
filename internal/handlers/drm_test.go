package handlers

import (
	"reflect"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func Test_getDrmMpd(t *testing.T) {
	type args struct {
		channelID string
		quality   string
	}
	tests := []struct {
		name    string
		args    args
		want    *DrmMpdOutput
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getDrmMpd(tt.args.channelID, tt.args.quality)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDrmMpd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDrmMpd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiveMpdHandler(t *testing.T) {
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
			if err := LiveMpdHandler(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("LiveMpdHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_generateDateTime(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateDateTime(); got != tt.want {
				t.Errorf("generateDateTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDRMKeyHandler(t *testing.T) {
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
			if err := DRMKeyHandler(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("DRMKeyHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMpdHandler(t *testing.T) {
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
			if err := MpdHandler(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("MpdHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDashHandler(t *testing.T) {
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
			if err := DashHandler(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("DashHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
