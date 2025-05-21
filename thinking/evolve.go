// evolve.go
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	paragon "github.com/OpenFluke/PARAGON"
)

type SpectrumSaver struct {
	TypeName  string // e.g. "float32"
	Mode      string // e.g. "Replay"
	Gen       int    // e.g. 1
	Steps     int    // Number of spectrum mutations to generate
	MaxStdDev float64
	BaseModel any // the loaded *paragon.Network[T]
}

// Main loop for evolutionary generation
func RunEpisodeLoop(cfg *ExperimentConfig) {
	fmt.Println("🔁 Starting Episode Loop...")
	modes := []string{"Standard", "Replay", "DynamicReplay"}

	for gen := 0; gen < cfg.Episodes; gen++ {
		fmt.Printf("\n📦 Generation %d\n", gen)

		for _, numType := range cfg.NumericalTypes {
			for _, mode := range modes {
				// Step 1: define folder for variants
				mutatedDir := filepath.Join("models", fmt.Sprint(gen), fmt.Sprintf("mutated_%s_%s", numType, mode))
				if err := os.MkdirAll(mutatedDir, 0755); err != nil {
					fmt.Printf("❌ Could not create folder: %s\n", mutatedDir)
					continue
				}

				//modelPath := filepath.Join(mutatedDir, fmt.Sprintf("%s_%s.json", numType, mode))
				modelPath := ""
				// Step 2: load base model on generation 0
				if gen == 0 {
					modelPath = filepath.Join("models", strconv.Itoa(gen), fmt.Sprintf("%s_%s.json", numType, mode))
					/*netAny, err := paragon.LoadNamedNetworkFromJSONFile(modelPath)
					if err != nil {
						fmt.Printf("❌ Failed to load base model from %s: %v\n", modelPath, err)
						continue
					}*/

				}

				fmt.Println("Grabbing model from ", modelPath)
				fmt.Printf("🧪 [%s_%s] Checking mutations in: %s\n", numType, mode, mutatedDir)

				// then:
				if !hasAllVariants(mutatedDir, cfg.SpectrumSteps) {
					// Load model
					netAny, err := paragon.LoadNamedNetworkFromJSONFile(modelPath)
					if err != nil {
						fmt.Printf("❌ Failed to load base model from %s: %v\n", modelPath, err)
						continue
					}

					// Save missing variants
					saver := &SpectrumSaver{
						TypeName:  numType,
						Mode:      mode,
						Gen:       gen,
						Steps:     cfg.SpectrumSteps,
						MaxStdDev: cfg.SpectrumMaxStdDev,
						BaseModel: netAny,
					}

					if err := saver.SaveSpectrumVariants(); err != nil {
						fmt.Printf("❌ Spectrum save failed: %v\n", err)
						continue
					}
				} else {
					fmt.Printf("✅ All %d variants already exist in %s\n", cfg.SpectrumSteps, mutatedDir)
				}

				// ✅ Evaluate variants that don't yet have results
				var agentsToRun []Agent

				err := filepath.WalkDir(mutatedDir, func(path string, d fs.DirEntry, err error) error {
					if err != nil || d.IsDir() {
						return nil
					}
					if !strings.HasPrefix(d.Name(), "variant_") || !strings.HasSuffix(d.Name(), ".json") {
						return nil
					}

					// Check if results exist
					resultPath := strings.TrimSuffix(path, ".json") + "_results.json"
					if _, err := os.Stat(resultPath); err == nil {
						// Already evaluated
						return nil
					}

					// Load unevaluated variant
					netAny, err := paragon.LoadNamedNetworkFromJSONFile(path)
					if err != nil {
						fmt.Printf("❌ Failed to load variant %s: %v\n", d.Name(), err)
						return nil
					}

					variantID := extractVariantIndex(d.Name()) // helper function
					agentsToRun = append(agentsToRun, Agent{
						ID:         fmt.Sprintf("%s_%s_variant_%d", numType, mode, variantID),
						Generation: gen,
						VariantID:  variantID,
						Network: NamedNetwork{
							TypeName: numType,
							Mode:     mode,
							Net:      netAny,
						},
						Config: cfg,
					})
					return nil
				})
				if err != nil {
					fmt.Printf("❌ Failed scanning variant dir: %v\n", err)
				}

				if len(agentsToRun) > 0 {
					fmt.Printf("🏃 Running %d agent(s) for %s_%s...\n", len(agentsToRun), numType, mode)

					RunAgentsInPool(agentsToRun)
				} else {
					fmt.Printf("✅ All variants for %s_%s already evaluated.\n", numType, mode)
				}

			}
		}

		break // TEMP: Only process generation 0
	}
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
		var err error

		switch base.TypeName {
		case "float32":
			newNet, err = cloneAndMutate(base.Net.(*paragon.Network[float32]), maxStdDev, i)
		case "float64":
			newNet, err = cloneAndMutate(base.Net.(*paragon.Network[float64]), maxStdDev, i)
		case "int":
			newNet, err = cloneAndMutate(base.Net.(*paragon.Network[int]), maxStdDev, i)
		case "int8":
			newNet, err = cloneAndMutate(base.Net.(*paragon.Network[int8]), maxStdDev, i)
		case "int16":
			newNet, err = cloneAndMutate(base.Net.(*paragon.Network[int16]), maxStdDev, i)
		case "int32":
			newNet, err = cloneAndMutate(base.Net.(*paragon.Network[int32]), maxStdDev, i)
		case "int64":
			newNet, err = cloneAndMutate(base.Net.(*paragon.Network[int64]), maxStdDev, i)
		case "uint":
			newNet, err = cloneAndMutate(base.Net.(*paragon.Network[uint]), maxStdDev, i)
		case "uint8":
			newNet, err = cloneAndMutate(base.Net.(*paragon.Network[uint8]), maxStdDev, i)
		case "uint16":
			newNet, err = cloneAndMutate(base.Net.(*paragon.Network[uint16]), maxStdDev, i)
		case "uint32":
			newNet, err = cloneAndMutate(base.Net.(*paragon.Network[uint32]), maxStdDev, i)
		case "uint64":
			newNet, err = cloneAndMutate(base.Net.(*paragon.Network[uint64]), maxStdDev, i)
		default:
			fmt.Printf("⚠️ Unsupported type %s in spectrum\n", base.TypeName)
			continue
		}

		if err != nil {
			fmt.Printf("❌ Clone failed for %s: %v\n", base.TypeName, err)
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

func cloneAndMutate[T paragon.Numeric](net *paragon.Network[T], stddev float64, seed int) (*paragon.Network[T], error) {
	snap := net.ToS()
	var clone paragon.Network[T]
	if err := clone.FromS(snap); err != nil {
		return nil, err
	}
	clone.TypeName = net.TypeName // ✅ Set it here
	clone.PerturbWeights(stddev, seed)
	return &clone, nil
}

type BaseModelInfo struct {
	TypeName string
	Mode     string
	Path     string
}

func DiscoverBaseModelPathsForGen(gen int) []BaseModelInfo {
	dir := fmt.Sprintf("models/%d", gen)
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("❌ Failed to read model dir for Gen %d: %v\n", gen, err)
		return nil
	}

	var models []BaseModelInfo
	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".json" {
			continue
		}

		name := strings.TrimSuffix(f.Name(), ".json")
		parts := strings.Split(name, "_")
		if len(parts) != 2 {
			continue
		}

		models = append(models, BaseModelInfo{
			TypeName: parts[0],
			Mode:     parts[1],
			Path:     filepath.Join(dir, f.Name()),
		})
	}
	return models
}

func hasAllVariants(dir string, steps int) bool {
	for i := 0; i < steps; i++ {
		variantPath := filepath.Join(dir, fmt.Sprintf("variant_%d.json", i))
		if _, err := os.Stat(variantPath); os.IsNotExist(err) {
			return false // missing at least one
		}
	}
	return true
}

func (s *SpectrumSaver) SaveSpectrumVariants() error {
	outputDir := filepath.Join("models", fmt.Sprint(s.Gen), fmt.Sprintf("mutated_%s_%s", s.TypeName, s.Mode))
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	switch base := s.BaseModel.(type) {

	case *paragon.Network[float32]:
		return saveVariantsForType(base, s.TypeName, s.Mode, outputDir, s.Steps, s.MaxStdDev)

	case *paragon.Network[float64]:
		return saveVariantsForType(base, s.TypeName, s.Mode, outputDir, s.Steps, s.MaxStdDev)

	case *paragon.Network[int]:
		return saveVariantsForType(base, s.TypeName, s.Mode, outputDir, s.Steps, s.MaxStdDev)

	case *paragon.Network[int8]:
		return saveVariantsForType(base, s.TypeName, s.Mode, outputDir, s.Steps, s.MaxStdDev)

	case *paragon.Network[int16]:
		return saveVariantsForType(base, s.TypeName, s.Mode, outputDir, s.Steps, s.MaxStdDev)

	case *paragon.Network[int32]:
		return saveVariantsForType(base, s.TypeName, s.Mode, outputDir, s.Steps, s.MaxStdDev)

	case *paragon.Network[int64]:
		return saveVariantsForType(base, s.TypeName, s.Mode, outputDir, s.Steps, s.MaxStdDev)

	case *paragon.Network[uint]:
		return saveVariantsForType(base, s.TypeName, s.Mode, outputDir, s.Steps, s.MaxStdDev)

	case *paragon.Network[uint8]:
		return saveVariantsForType(base, s.TypeName, s.Mode, outputDir, s.Steps, s.MaxStdDev)

	case *paragon.Network[uint16]:
		return saveVariantsForType(base, s.TypeName, s.Mode, outputDir, s.Steps, s.MaxStdDev)

	case *paragon.Network[uint32]:
		return saveVariantsForType(base, s.TypeName, s.Mode, outputDir, s.Steps, s.MaxStdDev)

	case *paragon.Network[uint64]:
		return saveVariantsForType(base, s.TypeName, s.Mode, outputDir, s.Steps, s.MaxStdDev)

	default:
		return fmt.Errorf("unsupported network type: %T", base)
	}
}

func saveVariantsForType[T paragon.Numeric](
	base *paragon.Network[T],
	typeName string,
	mode string,
	outputDir string,
	steps int,
	stddev float64,
) error {
	for i := 0; i < steps; i++ {
		clone, err := cloneAndMutate(base, stddev, i)
		if err != nil {
			fmt.Printf("❌ Clone/mutation failed [%s_%s_mut%d]: %v\n", typeName, mode, i, err)
			continue
		}

		filePath := filepath.Join(outputDir, fmt.Sprintf("variant_%d.json", i))
		if err := clone.SaveJSON(filePath); err != nil {
			fmt.Printf("❌ Save failed for variant %d: %v\n", i, err)
		} else {
			fmt.Printf("💾 Saved variant: %s\n", filePath)
		}
	}
	return nil
}

func extractVariantIndex(name string) int {
	// Expects: variant_0.json, variant_1.json, ...
	name = strings.TrimSuffix(name, ".json")
	parts := strings.Split(name, "_")
	if len(parts) < 2 {
		return -1
	}
	n, err := strconv.Atoi(parts[1])
	if err != nil {
		return -1
	}
	return n
}
