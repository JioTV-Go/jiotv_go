package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestWebEPGHandlerShortChannelID covers channel IDs shorter than the "sl"
// prefix, which used to panic on a two-byte slice of a one-byte ID.
func TestWebEPGHandlerShortChannelID(t *testing.T) {
	app := fiber.New()
	app.Get("/epg/:channelID/:offset", WebEPGHandler)

	// Each ID is shorter than or equal to the "sl" prefix and non-numeric, so
	// the handler rejects it before making any upstream request.
	tests := []struct {
		name      string
		channelID string
		offset    string
		wantCode  int
	}{
		{"single character ID", "s", "0", fiber.StatusBadRequest},
		{"bare sl prefix", "sl", "0", fiber.StatusBadRequest},
		{"non-numeric ID", "x", "0", fiber.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/epg/"+tt.channelID+"/"+tt.offset, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test() error = %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantCode {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantCode)
			}
		})
	}
}

func TestWebEPGHandler(t *testing.T) {
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// No test cases - EPG handler function
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
		// No test cases - EPG handler function
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := PosterHandler(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("PosterHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
