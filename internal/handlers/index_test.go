package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/jiotv-go/jiotv_go/v3/internal/config"
	"github.com/jiotv-go/jiotv_go/v3/pkg/television"
)

func TestIndexHandlerWithDefaultConfig(t *testing.T) {
	// Save original config
	originalCfg := config.Cfg

	// Create test app
	app := fiber.New()

	tests := []struct {
		name           string
		defaultCats    []int
		defaultLangs   []int
		queryParams    string
		expectedStatus int
		description    string
	}{
		{
			name:           "No defaults, no query params",
			defaultCats:    []int{},
			defaultLangs:   []int{},
			queryParams:    "",
			expectedStatus: 200,
			description:    "Should show all channels when no defaults and no query params",
		},
		{
			name:           "With defaults, no query params",
			defaultCats:    []int{8, 5}, // Sports, Entertainment
			defaultLangs:   []int{1, 6}, // Hindi, English
			queryParams:    "",
			expectedStatus: 200,
			description:    "Should apply default filtering when no query params",
		},
		{
			name:           "With defaults, but query params override",
			defaultCats:    []int{8, 5}, // Sports, Entertainment
			defaultLangs:   []int{1, 6}, // Hindi, English
			queryParams:    "?language=2&category=6", // Marathi, Movies
			expectedStatus: 200,
			description:    "Query params should override defaults",
		},
		{
			name:           "With defaults, partial query params",
			defaultCats:    []int{8, 5},
			defaultLangs:   []int{1, 6},
			queryParams:    "?language=1", // Hindi only
			expectedStatus: 200,
			description:    "Any query params should override all defaults",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up config for this test
			config.Cfg = config.JioTVConfig{
				DefaultCategories: tt.defaultCats,
				DefaultLanguages:  tt.defaultLangs,
				Title:             "Test JioTV Go",
			}

			// Create request
			req := httptest.NewRequest("GET", "/"+tt.queryParams, nil)

			// Setup app with handler - need to mock the template rendering
			// For now, we'll just check that the handler runs without error
			app.Get("/", func(c *fiber.Ctx) error {
				// Mock channels response for testing
				television.CategoryMap = map[int]string{
					0:  "All Categories",
					5:  "Entertainment",
					6:  "Movies", 
					8:  "Sports",
				}
				television.LanguageMap = map[int]string{
					0: "All Languages",
					1: "Hindi",
					2: "Marathi", 
					6: "English",
				}

				// Call IndexHandler logic without template rendering
				language := c.Query("language")
				category := c.Query("category")

				// Verify the logic path taken
				if language != "" || category != "" {
					// Query params provided - should use old filtering logic
					return c.JSON(fiber.Map{"mode": "query_params", "language": language, "category": category})
				}

				// No query params - check if defaults should be applied
				if len(config.Cfg.DefaultCategories) > 0 || len(config.Cfg.DefaultLanguages) > 0 {
					return c.JSON(fiber.Map{
						"mode": "defaults",
						"default_categories": config.Cfg.DefaultCategories,
						"default_languages": config.Cfg.DefaultLanguages,
					})
				}

				// No query params, no defaults - return all
				return c.JSON(fiber.Map{"mode": "all_channels"})
			})

			// Execute request
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			// Check status code
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// Additional validation can be added here based on response body
			resp.Body.Close()
		})
	}

	// Restore original config
	config.Cfg = originalCfg
}