package main

import (
	"fmt"
	"os"
	"path/filepath"

	paragon "github.com/OpenFluke/PARAGON"
)

func ensureInitialModelSetup(cfg *ExperimentConfig) {
	modelsDir := "models"
	gen0Dir := filepath.Join(modelsDir, "0")
	resultsFile := filepath.Join(gen0Dir, "results.json")

	// 1. Ensure models/ exists
	if _, err := os.Stat(modelsDir); os.IsNotExist(err) {
		if err := os.Mkdir(modelsDir, 0755); err != nil {
			fmt.Printf("❌ Failed to create models directory: %v\n", err)
			return
		}
		fmt.Println("📁 Created models/ directory")
	}

	// 2. Ensure models/0/ exists
	if _, err := os.Stat(gen0Dir); os.IsNotExist(err) {
		if err := os.Mkdir(gen0Dir, 0755); err != nil {
			fmt.Printf("❌ Failed to create models/0/ directory: %v\n", err)
			return
		}
		fmt.Println("📁 Created models/0/ directory")
	}

	// 3. Check for models/0/results.json
	if _, err := os.Stat(resultsFile); os.IsNotExist(err) {
		fmt.Println("🧪 No results.json found — will create models now")
		RunInitialModelSetup(cfg, 0) // ✅ fix: added missing generation argument
	} else {
		fmt.Println("✅ Found existing results.json — skipping initial model creation")
	}
}

func RunInitialModelSetup(cfg *ExperimentConfig, generation int) {
	fmt.Printf("🔧 Starting model generation for Gen %d...\n", generation)

	layerDefs := make([]struct{ Width, Height int }, len(cfg.NetworkConfig.Layers))
	activations := make([]string, len(cfg.NetworkConfig.Layers))
	full := make([]bool, len(cfg.NetworkConfig.Layers))

	for i, layer := range cfg.NetworkConfig.Layers {
		layerDefs[i] = struct{ Width, Height int }{
			Width:  layer.Width,
			Height: layer.Height,
		}
		activations[i] = layer.Activation
		full[i] = true
	}

	for _, requestedType := range cfg.NumericalTypes {
		for _, builder := range allTypeModeBuilders {
			if builder.TypeName == requestedType {
				fmt.Printf("🧠 Building models for type: %s\n", requestedType)
				builder.BuildSetWithSave(layerDefs, activations, full, generation)
				break
			}
		}
	}
}

func loadAndRegister[T paragon.Numeric](typeName, mode, path string) {
	nn := &paragon.Network[T]{}
	if err := nn.LoadJSON(path); err != nil {
		fmt.Printf("❌ Failed to load %s: %v\n", path, err)
		return
	}
	GlobalNetworks = append(GlobalNetworks, NamedNetwork{
		TypeName: typeName,
		Mode:     mode,
		Net:      nn,
	})
	fmt.Printf("✅ Loaded model: %s\n", path)
}
