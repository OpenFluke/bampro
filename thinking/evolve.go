// evolve.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	paragon "github.com/OpenFluke/PARAGON"
	"github.com/OpenFluke/construct"
	"github.com/OpenFluke/discover"
)

type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// We define symbolic modes using Go constants and types.
type ExperimentMode interface {
	isMode()
	String() string
}

type StandardMode struct{}
type ReplayMode struct{}
type DynamicReplayMode struct{}

func (StandardMode) isMode()             {}
func (ReplayMode) isMode()               {}
func (DynamicReplayMode) isMode()        {}
func (StandardMode) String() string      { return "Standard" }
func (ReplayMode) String() string        { return "Replay" }
func (DynamicReplayMode) String() string { return "DynamicReplay" }

type Experiment[T Numeric, M ExperimentMode] struct {
	NumType    string
	Mode       M
	Config     *ExperimentConfig
	Gen        int
	Cubes      []*construct.Cube[T]
	ServerAddr string
	AuthPass   string
	Delimiter  string
}

type ExperimentRunner interface {
	SetGeneration(gen int)
	GenerateVariants()
	SpawnAgentNames()
	SpawnAgentsOnPlanets(variantNum int)
	UnfreezeAgents()
	RunAndMonitorAgents()
	DespawnAgents()
}

func (e *Experiment[T, M]) SetGeneration(gen int) {
	e.Gen = gen
}

func (e *Experiment[T, M]) GenerateVariants() {
	// your logic
	fmt.Println(e.Gen, e.NumType+e.Mode.String())
	modelPath := ""
	if e.Gen == 0 {
		modelPath = filepath.Join("models", strconv.Itoa(e.Gen), fmt.Sprintf("%s_%s.json", e.NumType, e.Mode.String()))
	}

	fmt.Println(modelPath)

	mutatedDir := filepath.Join("models", fmt.Sprint(e.Gen), fmt.Sprintf("mutated_%s_%s", e.NumType, e.Mode.String()))
	if err := os.MkdirAll(mutatedDir, 0755); err != nil {
		fmt.Printf("❌ Could not create folder: %s\n", mutatedDir)
		return
	}

	// Skip if all already exist
	if hasAllVariants(mutatedDir, e.Config.SpectrumSteps) {
		fmt.Printf("✅ All variants already exist in %s\n", mutatedDir)
	}

	selectedModel, err := paragon.LoadNamedNetworkFromJSONFile(modelPath)
	if err != nil {
		fmt.Printf("❌ Failed to load base model from %s: %v\n", modelPath, err)
		return
	}

	// 🚀 ASSERT that it's the correct *Network[T]
	net, ok := selectedModel.(*paragon.Network[T])
	if !ok {
		fmt.Printf("⚠️ Type mismatch: expected *Network[%T], got %T\n", *new(T), selectedModel)
		return
	}

	// Generate variants
	for i := 0; i < e.Config.SpectrumSteps; i++ {
		savePath := filepath.Join(mutatedDir, fmt.Sprintf("variant_%d.json", i))

		// 💡 Skip if already exists
		if _, err := os.Stat(savePath); err == nil {
			fmt.Printf("⚠️ Skipping variant %d — already exists\n", i)
			continue
		}

		// 🧬 Clone and mutate
		var clone paragon.Network[T]
		if err := clone.FromS(net.ToS()); err != nil {
			fmt.Printf("❌ Failed to clone base model for variant %d: %v\n", i, err)
			continue
		}
		clone.PerturbWeights(e.Config.SpectrumMaxStdDev, i)

		// 💾 Save
		if err := clone.SaveJSON(savePath); err != nil {
			fmt.Printf("❌ Variant %d failed to save: %v\n", i, err)
		} else {
			fmt.Printf("💾 Saved variant: %s\n", savePath)
		}
	}

}

func (e *Experiment[T, M]) SpawnAgentNames() {
	mutatedDir := filepath.Join("models", strconv.Itoa(e.Gen), fmt.Sprintf("mutated_%s_%s", e.NumType, e.Mode.String()))
	namesDir := filepath.Join(mutatedDir, "agent_names")

	// Ensure names directory exists
	if err := os.MkdirAll(namesDir, 0755); err != nil {
		fmt.Printf("❌ Could not create agent_names dir: %v\n", err)
		return
	}

	entries, err := os.ReadDir(mutatedDir)
	if err != nil {
		fmt.Printf("❌ Failed to read mutated dir: %v\n", err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "variant_") || !strings.HasSuffix(name, ".json") {
			continue
		}

		variantName := strings.TrimSuffix(name, ".json")
		namesFile := filepath.Join(namesDir, variantName+".json")

		if _, err := os.Stat(namesFile); err == nil {
			fmt.Printf("✅ Skipping %s — names file already exists\n", variantName)
			continue
		}

		fullPath := filepath.Join(mutatedDir, name)
		var unitNames []string

		for _, planetStr := range e.Config.Planets {
			for i := 0; i < e.Config.EvaluationSpawnsPerPlanet; i++ {
				unitName := discover.GenerateUnitID(fullPath, "openfluke.com", e.Gen, len(unitNames))
				unitNames = append(unitNames, unitName)
				fmt.Printf("🚀 Spawn: %s | Planet=%s | Rep=%d\n", unitName, planetStr, i)
			}
		}

		data, err := json.MarshalIndent(unitNames, "", "  ")
		if err != nil {
			fmt.Printf("❌ Failed to marshal names: %v\n", err)
			continue
		}
		if err := os.WriteFile(namesFile, data, 0644); err != nil {
			fmt.Printf("❌ Failed to write names file: %v\n", err)
		} else {
			fmt.Printf("💾 Saved unit names: %s\n", namesFile)
		}
	}
}

func (e *Experiment[T, M]) SpawnAgentsOnPlanets(variantNum int) {
	namesPath := filepath.Join(
		"models",
		strconv.Itoa(e.Gen),
		fmt.Sprintf("mutated_%s_%s", e.NumType, e.Mode.String()),
		"agent_names",
		fmt.Sprintf("variant_%d.json", variantNum),
	)

	data, err := os.ReadFile(namesPath)
	if err != nil {
		fmt.Printf("❌ Failed to load agent names from %s: %v\n", namesPath, err)
		return
	}

	var unitNames []string
	if err := json.Unmarshal(data, &unitNames); err != nil {
		fmt.Printf("❌ Failed to parse agent names JSON: %v\n", err)
		return
	}

	totalPlanets := len(e.Config.Planets)
	spawnsPerPlanet := e.Config.EvaluationSpawnsPerPlanet
	expected := totalPlanets * spawnsPerPlanet

	if len(unitNames) < expected {
		fmt.Printf("⚠️ Warning: Not enough unit names (%d provided, %d expected)\n", len(unitNames), expected)
	}

	const planetSpacing = 800.0
	const spawnRadius = 120.0
	idx := 0

	var cubes []*construct.Cube[T] // 👈 Hold all spawned cubes

	// Load model for this variant
	modelPath := filepath.Join(
		"models",
		strconv.Itoa(e.Gen),
		fmt.Sprintf("mutated_%s_%s", e.NumType, e.Mode.String()),
		fmt.Sprintf("variant_%d.json", variantNum),
	)

	modelAny, err := paragon.LoadNamedNetworkFromJSONFile(modelPath)
	if err != nil {
		fmt.Printf("❌ Failed to load variant model: %v\n", err)
		return
	}

	net, ok := modelAny.(*paragon.Network[T])
	if !ok {
		fmt.Printf("⚠️ Type assertion failed for model: %T\n", modelAny)
		return
	}

	for _, planetStr := range e.Config.Planets {
		pos, err := parseVec3(planetStr)
		if err != nil {
			fmt.Printf("⚠️ Invalid planet string %q: %v\n", planetStr, err)
			continue
		}

		center := []float64{
			pos.X * planetSpacing,
			pos.Y * planetSpacing,
			pos.Z * planetSpacing,
		}
		positions := discover.FibonacciSphere(spawnsPerPlanet, spawnRadius, center)

		fmt.Printf("🌍 Planet: %s (center: %.2f, %.2f, %.2f)\n", planetStr, center[0], center[1], center[2])

		for i := 0; i < spawnsPerPlanet && idx < len(unitNames); i++ {
			name := unitNames[idx]
			idx++
			spawn := positions[i]

			cube := &construct.Cube[T]{
				Name:       name,
				UnitName:   "AutoUnit",
				Position:   spawn,
				Model:      net,
				ServerAddr: e.ServerAddr,
				AuthPass:   e.AuthPass,
				Delimiter:  e.Delimiter,
				ClampMin:   -20.0,
				ClampMax:   20.0,
			}

			if err := cube.Spawn(); err != nil {
				fmt.Printf("❌ Spawn failed for %s: %v\n", name, err)
			} else {
				fmt.Printf("🚀 Spawned %s on %s at (%.2f, %.2f, %.2f)\n", name, planetStr, spawn[0], spawn[1], spawn[2])
				cubes = append(cubes, cube)
			}
		}
	}

	if idx < len(unitNames) {
		fmt.Printf("⚠️ %d unit names were unused\n", len(unitNames)-idx)
	}

	// 🔐 Save references to all cubes
	e.Cubes = cubes
}

func (e *Experiment[T, M]) UnfreezeAgents() {
	if len(e.Cubes) == 0 {
		fmt.Println("⚠️ No cubes to unfreeze.")
		return
	}

	construct := &construct.Construct[T]{
		ServerAddr: e.ServerAddr,
		AuthPass:   e.AuthPass,
		Delimiter:  e.Delimiter,
	}
	construct.UnfreezeAll()
}

func (e *Experiment[T, M]) DespawnAgents() {
	if len(e.Cubes) == 0 {
		fmt.Println("⚠️ No cubes to despawn.")
		return
	}

	fmt.Printf("💣 Despawning %d agent(s)...\n", len(e.Cubes))

	for _, cube := range e.Cubes {
		if err := cube.Despawn(); err != nil {
			fmt.Printf("❌ Failed to despawn %s: %v\n", cube.Name, err)
		} else {
			fmt.Printf("✅ Despawned %s\n", cube.Name)
		}
	}

	// Optional: clear out the cube references
	e.Cubes = nil
}

func ParseExperimentMode(modeStr string) (ExperimentMode, error) {
	switch modeStr {
	case "Standard":
		return StandardMode{}, nil
	case "Replay":
		return ReplayMode{}, nil
	case "DynamicReplay":
		return DynamicReplayMode{}, nil
	default:
		return nil, fmt.Errorf("unsupported mode: %s", modeStr)
	}
}

func CreateExperiments(cfg *ExperimentConfig) []ExperimentRunner {
	var all []ExperimentRunner

	// Default server connection setup
	host := os.Getenv("GAME_HOST")
	if host == "" {
		host = "localhost"
	}
	serverAddr := host + ":14000"
	authPass := "my_secure_password"
	delimiter := "<???DONE???---"

	for _, numType := range cfg.NumericalTypes {
		for _, modeStr := range cfg.Modes {
			mode, err := ParseExperimentMode(modeStr)
			if err != nil {
				fmt.Printf("⚠️ Skipping invalid mode %q: %v\n", modeStr, err)
				continue
			}

			switch numType {
			case "int":
				all = append(all, &Experiment[int, ExperimentMode]{
					NumType:    numType,
					Mode:       mode,
					Config:     cfg,
					ServerAddr: serverAddr,
					AuthPass:   authPass,
					Delimiter:  delimiter,
				})
			case "int8":
				all = append(all, &Experiment[int8, ExperimentMode]{NumType: numType, Mode: mode, Config: cfg, ServerAddr: serverAddr, AuthPass: authPass, Delimiter: delimiter})
			case "int16":
				all = append(all, &Experiment[int16, ExperimentMode]{NumType: numType, Mode: mode, Config: cfg, ServerAddr: serverAddr, AuthPass: authPass, Delimiter: delimiter})
			case "int32":
				all = append(all, &Experiment[int32, ExperimentMode]{NumType: numType, Mode: mode, Config: cfg, ServerAddr: serverAddr, AuthPass: authPass, Delimiter: delimiter})
			case "int64":
				all = append(all, &Experiment[int64, ExperimentMode]{NumType: numType, Mode: mode, Config: cfg, ServerAddr: serverAddr, AuthPass: authPass, Delimiter: delimiter})

			case "uint":
				all = append(all, &Experiment[uint, ExperimentMode]{NumType: numType, Mode: mode, Config: cfg, ServerAddr: serverAddr, AuthPass: authPass, Delimiter: delimiter})
			case "uint8":
				all = append(all, &Experiment[uint8, ExperimentMode]{NumType: numType, Mode: mode, Config: cfg, ServerAddr: serverAddr, AuthPass: authPass, Delimiter: delimiter})
			case "uint16":
				all = append(all, &Experiment[uint16, ExperimentMode]{NumType: numType, Mode: mode, Config: cfg, ServerAddr: serverAddr, AuthPass: authPass, Delimiter: delimiter})
			case "uint32":
				all = append(all, &Experiment[uint32, ExperimentMode]{NumType: numType, Mode: mode, Config: cfg, ServerAddr: serverAddr, AuthPass: authPass, Delimiter: delimiter})
			case "uint64":
				all = append(all, &Experiment[uint64, ExperimentMode]{NumType: numType, Mode: mode, Config: cfg, ServerAddr: serverAddr, AuthPass: authPass, Delimiter: delimiter})

			case "float32":
				all = append(all, &Experiment[float32, ExperimentMode]{NumType: numType, Mode: mode, Config: cfg, ServerAddr: serverAddr, AuthPass: authPass, Delimiter: delimiter})
			case "float64":
				all = append(all, &Experiment[float64, ExperimentMode]{NumType: numType, Mode: mode, Config: cfg, ServerAddr: serverAddr, AuthPass: authPass, Delimiter: delimiter})

			default:
				fmt.Printf("⚠️ Unknown numeric type: %s\n", numType)
			}
		}
	}

	return all
}

func (e *Experiment[T, M]) RunAndMonitorAgents() {
	if len(e.Cubes) == 0 {
		fmt.Println("⚠️ No agents to run.")
		return
	}

	type result struct {
		Name         string
		Planet       string
		PlanetCenter []float64
		Goal         []float64
		InitialPos   []float64
		FinalPos     []float64
		InitialDist  float64
		FinalDist    float64
		Progress     float64
		DeltaY       float64
	}

	planetSpacing := 800.0
	topOffset := []float64{0, 100, 0} // goal above center

	initialPos := make(map[string][]float64)
	planetLookup := make(map[string]string)
	goalLookup := make(map[string][]float64)

	// Prepare mappings
	idx := 0
	for _, planetStr := range e.Config.Planets {
		planetPos, err := parseVec3(planetStr)
		if err != nil {
			continue
		}
		center := []float64{
			planetPos.X * planetSpacing,
			planetPos.Y * planetSpacing,
			planetPos.Z * planetSpacing,
		}
		goal := []float64{
			center[0] + topOffset[0],
			center[1] + topOffset[1],
			center[2] + topOffset[2],
		}

		for i := 0; i < e.Config.EvaluationSpawnsPerPlanet; i++ {
			if idx >= len(e.Cubes) {
				break
			}
			cube := e.Cubes[idx]
			initialPos[cube.Name] = append([]float64{}, cube.Position...)
			planetLookup[cube.Name] = planetStr
			goalLookup[cube.Name] = goal
			idx++
		}
	}

	// Create construct and run pulsing
	c := &construct.Construct[T]{
		ServerAddr: e.ServerAddr,
		AuthPass:   e.AuthPass,
		Delimiter:  e.Delimiter,
		Cubes:      e.Cubes,
	}
	duration := 10 * time.Second
	fmt.Printf("⚡ Pulsing agents for %v...\n", duration)
	c.StartPulsing(10, duration)

	// Evaluate progress
	var results []result
	var progresses []float64

	for _, cube := range e.Cubes {
		_ = cube.RefreshPosition()
		start := initialPos[cube.Name]
		end := cube.Position
		goal := goalLookup[cube.Name]
		planet := planetLookup[cube.Name]

		initialDist := distance(start, goal)
		finalDist := distance(end, goal)
		progress := initialDist - finalDist

		results = append(results, result{
			Name:         cube.Name,
			Planet:       planet,
			PlanetCenter: goal, // technically goal is above center, but fine here
			Goal:         goal,
			InitialPos:   start,
			FinalPos:     end,
			InitialDist:  initialDist,
			FinalDist:    finalDist,
			Progress:     progress,
			DeltaY:       end[1] - start[1],
		})
		progresses = append(progresses, progress)
	}

	// Build summary
	summary := map[string]any{
		"mean_progress":   Mean(progresses),
		"median_progress": Median(progresses),
		"max_progress":    Max(progresses),
		"min_progress":    Min(progresses),
		"results":         results,
	}

	// Save
	resultsDir := filepath.Join("models", strconv.Itoa(e.Gen),
		fmt.Sprintf("mutated_%s_%s", e.NumType, e.Mode.String()), "results")
	_ = os.MkdirAll(resultsDir, 0755)

	summaryPath := filepath.Join(resultsDir, "summary.json")
	if err := os.WriteFile(summaryPath, mustMarshalIndent(summary), 0644); err != nil {
		fmt.Printf("❌ Failed to write summary: %v\n", err)
	} else {
		fmt.Printf("✅ Saved progress summary: %s\n", summaryPath)
	}
}

func mustMarshalIndent(v any) []byte {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return data
}

func RunEpisodeLoop(cfg *ExperimentConfig) {
	all := CreateExperiments(cfg)

	for gen := 0; gen < cfg.Episodes; gen++ {
		for _, exp := range all {
			exp.SetGeneration(gen)
			exp.GenerateVariants()
			exp.SpawnAgentNames()
			for i := 0; i < cfg.SpectrumSteps; i++ {
				exp.SpawnAgentsOnPlanets(i)
				exp.UnfreezeAgents()
				// 🕒 Allow agents to run for some time before despawning
				/*runDuration := 5 * time.Second
				fmt.Printf("⏳ Letting agents run for %s...\n", runDuration)
				time.Sleep(runDuration)*/
				exp.RunAndMonitorAgents()
				exp.DespawnAgents()
				//exP.RunExperiment()
				//exp.Despawn
			}
		}
		break
	}
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

func parseVec3(s string) (Vec3, error) {
	var v Vec3
	if _, err := fmt.Sscanf(s, "(%f,%f,%f)", &v.X, &v.Y, &v.Z); err != nil {
		return v, err
	}
	return v, nil
}

/*

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
					*netAny, err := paragon.LoadNamedNetworkFromJSONFile(modelPath)
					if err != nil {
						fmt.Printf("❌ Failed to load base model from %s: %v\n", modelPath, err)
						continue
					}*

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

					for _, pStr := range cfg.Planets {

						pos, err := parseVec3(pStr)
						if err != nil {
							fmt.Printf("⚠️ bad planet vec %s: %v\n", pStr, err)
							continue
						}

						variantID := d.Name() + pStr // helper function

						agentsToRun = append(agentsToRun, Agent{
							ID:         fmt.Sprintf("%s_%s_variant_%d", numType, mode, variantID),
							Generation: gen,
							VariantID:  variantID,
							Network: NamedNetwork{
								TypeName: numType,
								Mode:     mode,
								Net:      netAny,
							},
							Config:     cfg,
							PlanetPos:  pos,
							PlanetName: pStr,
						})
					}

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


*/
