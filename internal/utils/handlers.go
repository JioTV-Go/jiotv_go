package utils

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/jiotv-go/jiotv_go/v3/pkg/secureurl"
	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
	"github.com/valyala/fasthttp"
)

// ErrorResponse sends a standardized error response
func ErrorResponse(c *fiber.Ctx, statusCode int, message interface{}) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"message": message,
	})
}

// InternalServerError sends a 500 error response
func InternalServerError(c *fiber.Ctx, err interface{}) error {
	return ErrorResponse(c, fiber.StatusInternalServerError, err)
}

// BadRequestError sends a 400 error response
func BadRequestError(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, fiber.StatusBadRequest, message)
}

// NotFoundError sends a 404 error response
func NotFoundError(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, fiber.StatusNotFound, message)
}

// ForbiddenError sends a 403 error response
func ForbiddenError(c *fiber.Ctx, err interface{}) error {
	return ErrorResponse(c, fiber.StatusForbidden, err)
}

// SetCommonHeaders sets common headers for proxy responses
func SetCommonHeaders(c *fiber.Ctx, userAgent string) {
	c.Request().Header.Set("User-Agent", userAgent)
	c.Response().Header.Del(fiber.HeaderServer)
}

// SetPlayerHeaders sets headers commonly used for player requests
func SetPlayerHeaders(c *fiber.Ctx, userAgent string) {
	c.Request().Header.Set("User-Agent", userAgent)
	c.Request().Header.Del("Accept")
	c.Request().Header.Del("Accept-Encoding")
	c.Request().Header.Del("Accept-Language")
	c.Request().Header.Del("Origin")
	c.Request().Header.Del("Referer")
}

// DecryptURLParam decrypts a URL parameter and handles errors
func DecryptURLParam(paramName, encryptedURL string) (string, error) {
	if encryptedURL == "" {
		return "", fmt.Errorf("%s not provided", paramName)
	}
	
	decoded, err := secureurl.DecryptURL(encryptedURL)
	if err != nil {
		if utils.Log != nil {
			utils.Log.Printf("Error decrypting %s: %v", paramName, err)
		}
		return "", err
	}
	
	return decoded, nil
}

// ProxyRequest performs a proxy request with common setup
func ProxyRequest(c *fiber.Ctx, url string, client *fasthttp.Client, userAgent string) error {
	if userAgent != "" {
		SetCommonHeaders(c, userAgent)
	}
	
	if err := proxy.Do(c, url, client); err != nil {
		return err
	}
	
	c.Response().Header.Del(fiber.HeaderServer)
	return nil
}

// ValidateRequiredParam checks if a required parameter is provided
func ValidateRequiredParam(paramName, paramValue string) error {
	if paramValue == "" {
		if utils.Log != nil {
			utils.Log.Printf("%s not provided", paramName)
		}
		return fmt.Errorf("%s not provided", paramName)
	}
	return nil
}

// CheckFieldExist validates field existence and sends error response if missing
func CheckFieldExist(c *fiber.Ctx, field string, condition bool) error {
	if !condition {
		if utils.Log != nil {
			utils.Log.Printf("%s not provided", field)
		}
		return BadRequestError(c, field+" not provided")
	}
	return nil
}

// SelectQuality returns the appropriate quality URL based on quality parameter
func SelectQuality(quality string, auto, high, medium, low string) string {
	switch quality {
	case "high", "h":
		return high
	case "medium", "med", "m":
		return medium
	case "low", "l":
		return low
	default:
		return auto
	}
}

// SetCacheHeader sets a cache control header with the specified max-age
func SetCacheHeader(c *fiber.Ctx, maxAge int) {
	c.Response().Header.Set("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
}

// SetMustRevalidateHeader sets a cache header that must revalidate
func SetMustRevalidateHeader(c *fiber.Ctx, maxAge int) {
	c.Response().Header.Set("Cache-Control", fmt.Sprintf("public, must-revalidate, max-age=%d", maxAge))
}