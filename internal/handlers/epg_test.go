package handlers

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestWebEPGHandler(t *testing.T) {
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
			if err := WebEPGHandler(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("WebEPGHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPosterHandler(t *testing.T) {
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
			if err := PosterHandler(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("PosterHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
