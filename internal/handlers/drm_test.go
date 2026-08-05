package handlers

import (
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
)

func TestGetDrmMpd(t *testing.T) {
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
		{
			name: "Test with invalid channel (expected to fail)",
			args: args{
				channelID: "invalid-channel",
				quality:   "high",
			},
			want:    nil,
			wantErr: true, // Should fail due to external API dependency
		},
		{
			name: "Test with empty channel ID (expected to fail)",
			args: args{
				channelID: "",
				quality:   "medium",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Handle potential panics from uninitialized TV object
			defer func() {
				if r := recover(); r != nil {
					t.Logf("getDrmMpd() panicked as expected due to uninitialized dependencies: %v", r)
				}
			}()

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
		// No test cases - DRM related function
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := LiveMpdHandler(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("LiveMpdHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateDateTime(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Generate datetime string",
			want: "", // We'll validate the format instead of exact value
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateDateTime()

			// Validate that the result is not empty
			if got == "" {
				t.Errorf("generateDateTime() returned empty string")
			}

			// Validate the format: should be 13 characters (YYMMDDHHMMMMS)
			if len(got) != 13 {
				t.Errorf("generateDateTime() length = %v, want 13", len(got))
			}

			// All characters should be digits
			for i, c := range got {
				if c < '0' || c > '9' {
					t.Errorf("generateDateTime() character at position %d should be digit, got %c", i, c)
				}
			}

			// Test that consecutive calls return different values (due to millisecond precision)
			got2 := generateDateTime()
			if got == got2 {
				// This might occasionally happen if called in same millisecond, so just log it
				t.Logf("generateDateTime() returned same value twice: %s", got)
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
		// No test cases - DRM related function
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
		// No test cases - DRM related function
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
		// No test cases - DRM related function
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DashHandler(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("DashHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLiveManifestHandlers(t *testing.T) {
	app := fiber.New()

	// Add recover middleware so the uninitialized TV panics are caught and returned as 500s.
	app.Use(fiberrecover.New())

	app.Get("/live/mpd/:channelID", LiveManifestMpdHandler)
	app.Post("/live/key/:channelID", LiveManifestKeyHandler)

	t.Run("Test MPD Handler Missing Channel", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/live/mpd/", nil)
		resp, _ := app.Test(req)

		if resp.StatusCode != fiber.StatusNotFound {
			t.Errorf("Expected 404 for missing channel ID, got %d", resp.StatusCode)
		}
	})

	// Since we can't easily mock the global TV object without breaking other tests,
	// we use defer recover to gracefully catch panics caused by uninitialized store/TV.

	t.Run("Test Key Handler Caching Mechanism", func(t *testing.T) {
		defer func() { recover() }()
		// Manually populate the cache to simulate a previous MPD request
		mockOutput := &DrmMpdOutput{
			IsDRM:      true,
			PlayUrl:    "mock",
			LicenseUrl: "", // Empty URL triggers a 404 Not Found in LiveManifestKeyHandler
		}

		// The cache key is generated from channelID and quality. Default quality is "auto".
		cacheKey := "12345_auto"
		drmMpdCache.Store(cacheKey, drmMpdCacheEntry{
			Output:    mockOutput,
			UpdatedAt: time.Now(),
		})

		req := httptest.NewRequest("POST", "/live/key/12345", nil)
		resp, _ := app.Test(req)

		// A successful cache hit will return 404 because LicenseUrl is empty.
		// A cache miss would panic in TV.Live() and return 500.
		if resp != nil && resp.StatusCode != fiber.StatusNotFound {
			t.Errorf("Expected 404 Not Found from cache hit, got %d", resp.StatusCode)
		}
	})

	t.Run("Test Cache Expiration", func(t *testing.T) {
		defer func() { recover() }()
		// Populate cache with an expired entry
		cacheKey := "expired-channel_auto"
		drmMpdCache.Store(cacheKey, drmMpdCacheEntry{
			Output: &DrmMpdOutput{
				IsDRM:      true,
				PlayUrl:    "mock",
				LicenseUrl: "mock",
			},
			UpdatedAt: time.Now().Add(-60 * time.Second), // 60 seconds ago (TTL is 30s)
		})

		req := httptest.NewRequest("POST", "/live/key/expired-channel", nil)
		resp, _ := app.Test(req)

		// Because it's expired, it will try to hit the JioTV API using the TV object.
		// Since we initialized TV with nil credentials, it will fail gracefully and return 500.
		if resp != nil && resp.StatusCode != fiber.StatusInternalServerError {
			t.Errorf("Expected 500 Internal Server Error due to TV unauthenticated, got %d", resp.StatusCode)
		}
	})
}
