package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestCORS(t *testing.T) {
	app := fiber.New()
	app.Use(CORS())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	tests := []struct {
		name           string
		method         string
		path           string
		wantStatus     int
		wantAllowOrigin string
	}{
		{
			name:           "GET request sets CORS headers",
			method:         http.MethodGet,
			path:           "/test",
			wantStatus:     200,
			wantAllowOrigin: "*",
		},
		{
			name:           "OPTIONS preflight returns 204",
			method:         http.MethodOptions,
			path:           "/test",
			wantStatus:     204,
			wantAllowOrigin: "*",
		},
		{
			name:           "Whitelist path skips CORS headers",
			method:         http.MethodGet,
			path:           "/render.ts",
			wantStatus:     404, // No handler, so 404, but CORS header should be absent
			wantAllowOrigin: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			resp, _ := app.Test(req)
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("got status %d, want %d", resp.StatusCode, tt.wantStatus)
			}
			got := resp.Header.Get("Access-Control-Allow-Origin")
			if got != tt.wantAllowOrigin {
				t.Errorf("got Access-Control-Allow-Origin %q, want %q", got, tt.wantAllowOrigin)
			}
		})
	}
}
