package main

import (
	"fmt"

	"github.com/OpenFluke/discover"
)

// GlobalNetworks holds all constructed networks — accessible from anywhere
var (
	GlobalNetworks []NamedNetwork
	disco          *discover.Discover // Use D.I.S.C.O.V.E.R. for world/scene info
)

func main() {
	// This matches the Actor network used in Biofoundry agents
	/*layers := []struct{ Width, Height int }{
		{6, 1},   // StateDim
		{128, 1}, // Hidden layer 1
		{128, 1}, // Hidden layer 2
		{3, 1},   // ActionDim
	}
	acts := []string{"linear", "relu", "relu", "tanh"}
	full := []bool{true, true, true, true}

	// Build all variants for all types (unchanged)
	buildAllNetworks(layers, acts, full)*/
	/*TryToConnect()

	// ---- D.I.S.C.O.V.E.R. scene scan ----
	cfg := discover.Config{
		Hosts:      []string{"localhost"}, // Use your environment hosts
		StartPort:  14000,
		PortStep:   3,
		NumPods:    1,
		AuthPass:   "my_secure_password",
		Delimiter:  "<???DONE???---",
		TimeoutSec: 10,
	}
	disco = discover.NewDiscover(cfg)
	disco.ScanAll()
	disco.PrintSummary()

	// Use planet and cube data for spawning or agent setup
	planetCenters := disco.ExtractPlanetCenters()
	fmt.Println("Planet centers for agent spawning:", planetCenters)

	// Example: Generate spawn points around the first planet found
	var firstPlanet string
	for name := range disco.Planets {
		firstPlanet = name
		break
	}
	if firstPlanet != "" {
		spawnPoints, _ := disco.GenerateSpawnPositions(firstPlanet, 8, 100.0)
		fmt.Printf("Sample spawn points around %s: %v\n", firstPlanet, spawnPoints)
	}*/

	cfg, err := LoadExperimentConfig("experiment_config.json")
	if err != nil {
		fmt.Println("❌ Failed to load experiment config:", err)
	} else {
		fmt.Printf("✅ Loaded Experiment: %s\n", cfg.Name)
		fmt.Printf("   Description: %s\n", cfg.Description)
		fmt.Printf("   Numerical Types: %v\n", cfg.NumericalTypes)
		fmt.Printf("   Planets: %v\n", cfg.Planets)
		fmt.Printf("   Spectrum: %d steps, max stddev %.4f\n", cfg.SpectrumSteps, cfg.SpectrumMaxStdDev)
		fmt.Println("   Auto-launch enabled?", cfg.AutoLaunch)

		experimentConfig = cfg
		ensureInitialModelSetup(cfg)
	}

	if cfg.AutoState {
		// Try to load existing best model state
		//fmt.Println("Auto starting")
		go RunEpisodeLoop(experimentConfig)
	}

	go startWebSocketServer() // Starts WebSocket server on port 9001
	go startStatusPoller()
	go startStatusBroadcastLoop()

	host()
}
