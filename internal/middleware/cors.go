package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// CORS middleware to enable CORS
// https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
func CORS() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// ignore direct pass-through routes (proxy requests through server)
		whitelist := []string{"/render.ts", "/jtvimage"}
		for _, path := range whitelist {
			// if path is in whitelist, skip CORS
			if strings.Contains(c.Path(), path) {
				return c.Next()
			}
		}
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET, POST, HEAD, OPTIONS")

		// handle preflight requests
		if c.Method() == "OPTIONS" {
			return c.SendStatus(204)
		}

		// continue request handler chain
		return c.Next()
	}
}
