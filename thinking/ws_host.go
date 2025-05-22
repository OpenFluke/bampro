package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

		// ✅ Send full status update array once on connect
		statusMu.Lock()
		fullStatus := make([]ExperimentStatus, len(StatusUpdates))
		copy(fullStatus, StatusUpdates)
		statusMu.Unlock()

		statusJSON := SerializeTyped(TypeExperimentRunning, fullStatus)
		if statusJSON != nil {
			if err := c.WriteMessage(websocket.TextMessage, statusJSON); err != nil {
				log.Println("❌ Failed to send full status on connect:", err)
				return
			}
		}

		// ✅ Send full score table once on connect
		scoreRecords := collectAllScores()
		scoreJSON := SerializeTyped(TypeScoresOverview, scoreRecords)
		if scoreJSON != nil {
			if err := c.WriteMessage(websocket.TextMessage, scoreJSON); err != nil {
				log.Println("❌ Failed to send scoring data:", err)
				return
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

func startStatusBroadcastLoop() {
	go func() {
		var lastSeen int

		for {
			time.Sleep(2 * time.Second)

			statusMu.Lock()
			if lastSeen >= len(StatusUpdates) {
				statusMu.Unlock()
				continue
			}
			newStatuses := StatusUpdates[lastSeen:]
			lastSeen = len(StatusUpdates)
			statusMu.Unlock()

			if len(newStatuses) > 0 {
				data := SerializeTyped(TypeExperimentRunning, newStatuses)
				if data != nil {
					for client := range wsClients {
						if err := client.WriteMessage(websocket.TextMessage, data); err != nil {
							client.Close()
							delete(wsClients, client)
						}
					}
				}
			}
		}
	}()
}

func collectAllScores() []ScoreRecord {
	var records []ScoreRecord

	for gen := 0; gen <= latestGeneration(); gen++ {
		dir := fmt.Sprintf("models/%d/total_results", gen)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".json") && entry.Name() != "full_results.json" {
				parts := strings.Split(strings.TrimSuffix(entry.Name(), ".json"), "_")
				if len(parts) != 2 {
					continue
				}
				numType, mode := parts[0], parts[1]
				raw, _ := os.ReadFile(filepath.Join(dir, entry.Name()))

				var variants []struct {
					Variant      string  `json:"variant"`
					MeanProgress float64 `json:"mean_progress"`
				}
				if err := json.Unmarshal(raw, &variants); err != nil {
					continue
				}

				for _, v := range variants {
					vid, _ := strconv.Atoi(v.Variant)
					records = append(records, ScoreRecord{
						Generation:   gen,
						NumType:      numType,
						Mode:         mode,
						VariantIndex: vid,
						MeanProgress: v.MeanProgress,
					})
				}
			}
		}
	}
	return records
}

func latestGeneration() int {
	entries, _ := os.ReadDir("models")
	highest := 0
	for _, e := range entries {
		if e.IsDir() {
			if n, err := strconv.Atoi(e.Name()); err == nil && n > highest {
				highest = n
			}
		}
	}
	return highest
}
