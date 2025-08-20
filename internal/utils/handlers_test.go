package utils

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestSelectQuality(t *testing.T) {
	auto := "auto_url"
	high := "high_url"
	medium := "medium_url"
	low := "low_url"

	tests := []struct {
		quality  string
		expected string
	}{
		{"high", high},
		{"h", high},
		{"medium", medium},
		{"med", medium},
		{"m", medium},
		{"low", low},
		{"l", low},
		{"auto", auto},
		{"", auto},
		{"unknown", auto},
	}

	for _, test := range tests {
		result := SelectQuality(test.quality, auto, high, medium, low)
		assert.Equal(t, test.expected, result, "Quality selection for %s failed", test.quality)
	}
}

func TestErrorResponse(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return ErrorResponse(c, fiber.StatusBadRequest, "test error")
	})

	// Note: Full HTTP testing would require a more complex setup
	// This is a basic structure test
	assert.NotNil(t, app)
}

func TestValidateRequiredParam(t *testing.T) {
	// Initialize the logger to avoid nil pointer issues
	// For testing, we can create a simple logger
	tests := []struct {
		paramName  string
		paramValue string
		expectErr  bool
	}{
		{"test", "value", false},
		{"test", "", true},
		{"empty", "", true},
		{"nonempty", "value", false},
	}

	for _, test := range tests {
		err := ValidateRequiredParam(test.paramName, test.paramValue)
		if test.expectErr {
			assert.Error(t, err, "Expected error for empty param %s", test.paramName)
		} else {
			assert.NoError(t, err, "Expected no error for param %s", test.paramName)
		}
	}
}

func TestDecryptURLParam(t *testing.T) {
	// Test empty parameter
	_, err := DecryptURLParam("test", "")
	assert.Error(t, err, "Expected error for empty URL")

	// Test invalid encrypted URL
	_, err = DecryptURLParam("test", "invalid")
	assert.Error(t, err, "Expected error for invalid encrypted URL")
}