package middleware

import (
	"github.com/gofiber/fiber/v2"
)

func CORS() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// if path is /render.ts then return c.Next()
		if c.Path() == "/render.ts" {
			return c.Next()
		}

		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET")

		if c.Method() == "OPTIONS" {
			return c.SendStatus(204)
		}

		return c.Next()
	}
}
