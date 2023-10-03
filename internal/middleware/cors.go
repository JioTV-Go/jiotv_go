package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func CORS() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// ignore direct pass-through routes (proxy requests through server)
		whitelist := []string{"/render.ts", "/jtvimage"}
		for _, path := range whitelist {
			if strings.Contains(c.Path(), path) {
				return c.Next()
			}
		}
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET")

		if c.Method() == "OPTIONS" {
			return c.SendStatus(204)
		}

		return c.Next()
	}
}
