package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func host() {
	// Try to load .env file if present (ignored if not found)
	_ = godotenv.Load()

	// Default port
	port := "8123"
	if v := os.Getenv("PORT"); v != "" {
		port = v
	}

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ello world")
	})

	log.Printf("Starting Fiber server on port %s\n", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
