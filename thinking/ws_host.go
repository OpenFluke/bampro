package main

import (
	"encoding/json"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

var wsClients = make(map[*websocket.Conn]bool)
var experimentConfig *ExperimentConfig // populated in engine.go

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

		// ✅ Send config once on connection
		if experimentConfig != nil {
			configJSON := SerializeTyped(TypeExperimentConf, experimentConfig)
			if configJSON != nil {
				if err := c.WriteMessage(websocket.TextMessage, configJSON); err != nil {
					log.Println("❌ Failed to send config:", err)
					return
				}
			}
		}

		// ✅ Listen for incoming control messages
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Println("⚠️ WebSocket read error:", err)
				break
			}

			var incoming TypedMessage
			if err := json.Unmarshal(msg, &incoming); err != nil {
				log.Println("❌ Invalid WebSocket message:", string(msg))
				continue
			}

			switch incoming.Type {
			case "experiment_control":
				var ctrl struct {
					Action  string      `json:"action"`
					Payload interface{} `json:"payload"`
				}
				if err := mapToStruct(incoming.Data, &ctrl); err == nil {
					log.Printf("🧪 Received control: %s — %+v\n", ctrl.Action, ctrl.Payload)
					// TODO: implement run/stop/save/load logic using ctrl.Action and ctrl.Payload
				} else {
					log.Println("❌ Failed to map control data")
				}
			default:
				log.Println("🪐 Unknown WS message type:", incoming.Type)
			}
		}
	}))

	log.Println("📡 Fiber WebSocket server running on :9001")
	if err := app.Listen(":9001"); err != nil {
		log.Fatalf("WebSocket Fiber server failed: %v", err)
	}
}

// ✅ Helper to map generic interface{} into a typed struct
func mapToStruct(data interface{}, out interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, out)
}
