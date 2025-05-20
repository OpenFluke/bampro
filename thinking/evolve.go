// evolve.go
package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	paragon "github.com/OpenFluke/PARAGON"
)

// Main loop for evolutionary generation
func RunEpisodeLoop(cfg *ExperimentConfig) {
	fmt.Println("🔁 Starting Episode Loop...")

	for gen := 0; gen < cfg.Episodes; gen++ {
		fmt.Printf("\n📦 Checking generation %d...\n", gen)

		genDir := fmt.Sprintf("models/%d", gen)
		resultsPath := filepath.Join(genDir, "results.json")

		// STEP 1: Check for existing results
		fmt.Printf("🔍 Looking for results.json at: %s\n", resultsPath)
		if _, err := os.Stat(resultsPath); os.IsNotExist(err) {
			fmt.Printf("🧪 No results.json for Gen %d – beginning processing...\n", gen)

			// STEP 2: Load base models
			fmt.Println("📥 Loading base models...")
			baseModels := LoadBaseModelsForGen(gen)
			fmt.Printf("📊 Found %d base models for mutation.\n", len(baseModels))

			// STEP 3: Generate mutated spectrum
			var readyModels []NamedNetwork
			for i, base := range baseModels {
				fmt.Printf("🌱 Mutating model %d/%d (%s_%s)...\n", i+1, len(baseModels), base.TypeName, base.Mode)
				spectrum := GenerateModelSpectrum(base, cfg.SpectrumSteps, cfg.SpectrumMaxStdDev)
				fmt.Printf("🔬 → Generated %d variants.\n", len(spectrum))
				readyModels = append(readyModels, spectrum...)
			}

			// STEP 4: Summary
			fmt.Printf("✅ Total mutated models generated for Gen %d: %d\n", gen, len(readyModels))

			// STEP 5: Placeholder for next phase (e.g., training, evaluation)
			fmt.Println("📡 Ready for training/evaluation... (WebSocket updates go here)")

			break // 🚧 TEMP: Only process one generation for now
		} else {
			fmt.Printf("✅ Gen %d already processed — skipping.\n", gen)
		}
	}

	fmt.Println("🛑 Episode loop exited (after first generation).")
}

func LoadBaseModelsForGen(gen int) []NamedNetwork {
	dir := fmt.Sprintf("models/%d", gen)
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("❌ Failed to read model dir for Gen %d: %v\n", gen, err)
		return nil
	}

	var models []NamedNetwork
	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".json" {
			continue
		}

		parts := strings.Split(strings.TrimSuffix(f.Name(), ".json"), "_")
		if len(parts) != 2 {
			continue
		}
		typeName, mode := parts[0], parts[1]
		fullPath := filepath.Join(dir, f.Name())

		switch typeName {
		case "int":
			loadAndRegister[int](typeName, mode, fullPath)
		case "int8":
			loadAndRegister[int8](typeName, mode, fullPath)
		case "int16":
			loadAndRegister[int16](typeName, mode, fullPath)
		case "int32":
			loadAndRegister[int32](typeName, mode, fullPath)
		case "int64":
			loadAndRegister[int64](typeName, mode, fullPath)
		case "uint":
			loadAndRegister[uint](typeName, mode, fullPath)
		case "uint8":
			loadAndRegister[uint8](typeName, mode, fullPath)
		case "uint16":
			loadAndRegister[uint16](typeName, mode, fullPath)
		case "uint32":
			loadAndRegister[uint32](typeName, mode, fullPath)
		case "uint64":
			loadAndRegister[uint64](typeName, mode, fullPath)
		case "float32":
			loadAndRegister[float32](typeName, mode, fullPath)
		case "float64":
			loadAndRegister[float64](typeName, mode, fullPath)
		default:
			fmt.Printf("⚠️ Unsupported type %s\n", typeName)
			continue
		}

		models = append(models, NamedNetwork{
			TypeName: typeName,
			Mode:     mode,
			Net:      GlobalNetworks[len(GlobalNetworks)-1].Net,
		})
	}

	return models
}

func GenerateModelSpectrum(base NamedNetwork, steps int, maxStdDev float64) []NamedNetwork {
	var mutated []NamedNetwork

	for i := 0; i < steps; i++ {
		var newNet any

		switch base.TypeName {
		case "float32":
			original := base.Net.(*paragon.Network[float32])
			snap := original.ToS()
			var clone paragon.Network[float32]
			if err := clone.FromS(snap); err != nil {
				fmt.Printf("❌ Failed to clone float32 network: %v\n", err)
				continue
			}
			ApplyNoiseToNetwork(&clone, maxStdDev)
			newNet = &clone

		case "float64":
			original := base.Net.(*paragon.Network[float64])
			snap := original.ToS()
			var clone paragon.Network[float64]
			if err := clone.FromS(snap); err != nil {
				fmt.Printf("❌ Failed to clone float64 network: %v\n", err)
				continue
			}
			ApplyNoiseToNetwork(&clone, maxStdDev)
			newNet = &clone

		case "int":
			original := base.Net.(*paragon.Network[int])
			snap := original.ToS()
			var clone paragon.Network[int]
			if err := clone.FromS(snap); err != nil {
				fmt.Printf("❌ Failed to clone int network: %v\n", err)
				continue
			}
			ApplyNoiseToNetwork(&clone, maxStdDev)
			newNet = &clone

		case "int8":
			original := base.Net.(*paragon.Network[int8])
			snap := original.ToS()
			var clone paragon.Network[int8]
			if err := clone.FromS(snap); err != nil {
				fmt.Printf("❌ Failed to clone int8 network: %v\n", err)
				continue
			}
			ApplyNoiseToNetwork(&clone, maxStdDev)
			newNet = &clone

		case "int16":
			original := base.Net.(*paragon.Network[int16])
			snap := original.ToS()
			var clone paragon.Network[int16]
			if err := clone.FromS(snap); err != nil {
				fmt.Printf("❌ Failed to clone int16 network: %v\n", err)
				continue
			}
			ApplyNoiseToNetwork(&clone, maxStdDev)
			newNet = &clone

		case "int32":
			original := base.Net.(*paragon.Network[int32])
			snap := original.ToS()
			var clone paragon.Network[int32]
			if err := clone.FromS(snap); err != nil {
				fmt.Printf("❌ Failed to clone int32 network: %v\n", err)
				continue
			}
			ApplyNoiseToNetwork(&clone, maxStdDev)
			newNet = &clone

		case "int64":
			original := base.Net.(*paragon.Network[int64])
			snap := original.ToS()
			var clone paragon.Network[int64]
			if err := clone.FromS(snap); err != nil {
				fmt.Printf("❌ Failed to clone int64 network: %v\n", err)
				continue
			}
			ApplyNoiseToNetwork(&clone, maxStdDev)
			newNet = &clone

		case "uint":
			original := base.Net.(*paragon.Network[uint])
			snap := original.ToS()
			var clone paragon.Network[uint]
			if err := clone.FromS(snap); err != nil {
				fmt.Printf("❌ Failed to clone uint network: %v\n", err)
				continue
			}
			ApplyNoiseToNetwork(&clone, maxStdDev)
			newNet = &clone

		case "uint8":
			original := base.Net.(*paragon.Network[uint8])
			snap := original.ToS()
			var clone paragon.Network[uint8]
			if err := clone.FromS(snap); err != nil {
				fmt.Printf("❌ Failed to clone uint8 network: %v\n", err)
				continue
			}
			ApplyNoiseToNetwork(&clone, maxStdDev)
			newNet = &clone

		case "uint16":
			original := base.Net.(*paragon.Network[uint16])
			snap := original.ToS()
			var clone paragon.Network[uint16]
			if err := clone.FromS(snap); err != nil {
				fmt.Printf("❌ Failed to clone uint16 network: %v\n", err)
				continue
			}
			ApplyNoiseToNetwork(&clone, maxStdDev)
			newNet = &clone

		case "uint32":
			original := base.Net.(*paragon.Network[uint32])
			snap := original.ToS()
			var clone paragon.Network[uint32]
			if err := clone.FromS(snap); err != nil {
				fmt.Printf("❌ Failed to clone uint32 network: %v\n", err)
				continue
			}
			ApplyNoiseToNetwork(&clone, maxStdDev)
			newNet = &clone

		case "uint64":
			original := base.Net.(*paragon.Network[uint64])
			snap := original.ToS()
			var clone paragon.Network[uint64]
			if err := clone.FromS(snap); err != nil {
				fmt.Printf("❌ Failed to clone uint64 network: %v\n", err)
				continue
			}
			ApplyNoiseToNetwork(&clone, maxStdDev)
			newNet = &clone

		default:
			fmt.Printf("⚠️ Unsupported type %s in spectrum\n", base.TypeName)
			continue
		}

		mutated = append(mutated, NamedNetwork{
			TypeName: base.TypeName,
			Mode:     base.Mode,
			Net:      newNet,
		})
	}

	return mutated
}

func ApplyNoiseToNetwork[T paragon.Numeric](net *paragon.Network[T], maxStdDev float64) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for l := 1; l < len(net.Layers); l++ { // Skip input layer
		for y := 0; y < net.Layers[l].Height; y++ {
			for x := 0; x < net.Layers[l].Width; x++ {
				neuron := net.Layers[l].Neurons[y][x]
				for i := range neuron.Inputs {
					noise := rng.NormFloat64() * maxStdDev
					neuron.Inputs[i].Weight += T(noise)
				}
			}
		}
	}
}
