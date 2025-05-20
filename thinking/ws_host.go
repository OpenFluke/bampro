package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

var wsClients = make(map[*websocket.Conn]bool)

func startWebSocketServer() {
	app := fiber.New()

	app.Use("/ws/status", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/status", websocket.New(func(c *websocket.Conn) {
		wsClients[c] = true
		defer func() {
			delete(wsClients, c)
			c.Close()
		}()

		for {
			msg := fmt.Sprintf(`{"message":"Server time: %s"}`, time.Now().Format("15:04:05"))
			if err := c.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				break
			}
			time.Sleep(3 * time.Second)
		}
	}))

	log.Println("📡 Fiber WebSocket server running on :9001")
	if err := app.Listen(":9001"); err != nil {
		log.Fatalf("WebSocket Fiber server failed: %v", err)
	}
}
