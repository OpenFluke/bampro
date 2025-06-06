package main

import (
	"os"
	"time"

	"github.com/OpenFluke/discover"
	"github.com/gofiber/websocket/v2"
)

type GameStatus struct {
	Timestamp    string                 `json:"timestamp"`
	TotalCubes   int                    `json:"total_cubes"`
	TotalPlanets int                    `json:"total_planets"`
	Planets      []PlanetSummary        `json:"planets"`
	CubeHosts    map[string]int         `json:"cubes_by_host"`
	Extras       map[string]interface{} `json:"extras,omitempty"`
}

type PlanetSummary struct {
	Name string     `json:"name"`
	Pos  [3]float64 `json:"pos"`
	Host string     `json:"host"`
	Port int        `json:"port"`
}

func startStatusPoller() {
	go func() {
		for {
			time.Sleep(1 * time.Second)

			hostName := os.Getenv("GAME_HOST")
			if hostName == "" {
				hostName = "localhost"
			}

			cfg := discover.Config{
				Hosts:      []string{hostName},
				StartPort:  14000,
				PortStep:   3,
				NumPods:    1,
				AuthPass:   "my_secure_password",
				Delimiter:  "<???DONE???---",
				TimeoutSec: 5,
			}

			d := discover.NewDiscover(cfg)
			d.ScanAll()

			planetSummaries := make([]PlanetSummary, 0, len(d.Planets))
			for _, p := range d.Planets {
				planetSummaries = append(planetSummaries, PlanetSummary{
					Name: p.Name,
					Pos:  p.Coordinates,
					Host: p.Host,
					Port: p.Port,
				})
			}

			hostCount := make(map[string]int)
			for _, host := range d.Cubes {
				hostCount[host]++
			}

			status := GameStatus{
				Timestamp:    time.Now().Format(time.RFC3339),
				TotalCubes:   len(d.Cubes),
				TotalPlanets: len(d.Planets),
				Planets:      planetSummaries,
				CubeHosts:    hostCount,
			}

			data := SerializeTyped(TypeStatusUpdate, status)
			if data != nil {
				broadcastStatus(data)
			}
		}
	}()
}

func broadcastStatus(msg []byte) {
	for conn := range wsClients {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			conn.Close()
			delete(wsClients, conn)
		}
	}
}
