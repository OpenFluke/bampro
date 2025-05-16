package main

import (
	"github.com/gofiber/fiber/v2"
)

func host() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ello world")
	})

	// Listen on port 8123
	app.Listen(":8123")
}
