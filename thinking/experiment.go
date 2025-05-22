package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	paragon "github.com/OpenFluke/PARAGON"
	"github.com/shirou/gopsutil/v3/cpu"
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

	if cfg.LoadBalance {
		//go runBenchmarks(cfg)
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

func runBenchmarks(cfg *ExperimentConfig) {
	fmt.Println("starting benchmark")
	if len(cfg.NumericalTypes) == 0 {
		fmt.Println("⚠️  no numerical types – skipping benchmarks")
		return
	}

	baseInput := [][]float64{{0.1, 0.2, -0.3, 0.4, -0.5, 0.6}}
	aps := cfg.Movement.Translation.ActionsPerSecond
	if aps <= 0 {
		aps = 10
	}
	clones := cfg.MaxNeeded
	if clones <= 0 {
		clones = 200
	}

	duration := 10 * time.Second
	benchRoot := filepath.Join("models", "0", "benchmarks")
	_ = os.MkdirAll(benchRoot, 0o755)

	for _, t := range cfg.NumericalTypes {

		outDir := filepath.Join(benchRoot, t)
		jsonFile := filepath.Join(outDir, "benchmark.json")

		/*───── skip if result already present ─────*/
		if _, err := os.Stat(jsonFile); err == nil {
			fmt.Printf("⏭️  %s benchmark exists – skipping\n", t)
			continue
		}

		modelPath := filepath.Join("models", "0", fmt.Sprintf("%s_Standard.json", t))
		netAny, err := paragon.LoadNamedNetworkFromJSONFile(modelPath)
		if err != nil {
			fmt.Printf("❌ %s: %v – skip benchmark\n", modelPath, err)
			continue
		}

		/*───── CPU sampler ─────*/
		var (
			cpuLog []float64
			mu     sync.Mutex
			stop   = make(chan struct{})
		)
		go func() {
			tick := time.NewTicker(time.Second)
			defer tick.Stop()
			for {
				select {
				case <-tick.C:
					if v, _ := cpu.Percent(0, false); len(v) > 0 {
						mu.Lock()
						cpuLog = append(cpuLog, v[0])
						mu.Unlock()
					}
				case <-stop:
					return
				}
			}
		}()

		/*───── pulse flood ─────*/
		switch n := netAny.(type) {
		case *paragon.Network[float64]:
			n.ClonePulse(baseInput, clones, aps, duration)
		case *paragon.Network[float32]:
			n.ClonePulse(baseInput, clones, aps, duration)
		case *paragon.Network[int]:
			n.ClonePulse(baseInput, clones, aps, duration)
		case *paragon.Network[int8]:
			n.ClonePulse(baseInput, clones, aps, duration)
		case *paragon.Network[int16]:
			n.ClonePulse(baseInput, clones, aps, duration)
		case *paragon.Network[int32]:
			n.ClonePulse(baseInput, clones, aps, duration)
		case *paragon.Network[int64]:
			n.ClonePulse(baseInput, clones, aps, duration)
		case *paragon.Network[uint]:
			n.ClonePulse(baseInput, clones, aps, duration)
		case *paragon.Network[uint8]:
			n.ClonePulse(baseInput, clones, aps, duration)
		case *paragon.Network[uint16]:
			n.ClonePulse(baseInput, clones, aps, duration)
		case *paragon.Network[uint32]:
			n.ClonePulse(baseInput, clones, aps, duration)
		case *paragon.Network[uint64]:
			n.ClonePulse(baseInput, clones, aps, duration)
		default:
			fmt.Printf("⚠️  unsupported type %T – skip benchmark\n", netAny)
		}

		close(stop) // stop CPU sampler

		/*───── save results ─────*/
		mu.Lock()
		res := struct {
			Type      string    `json:"type"`
			Clones    int       `json:"clones"`
			APS       int       `json:"aps"`
			Seconds   int       `json:"seconds"`
			CPU       []float64 `json:"cpu_log"`
			Timestamp time.Time `json:"timestamp"`
		}{t, clones, aps, int(duration.Seconds()), cpuLog, time.Now()}
		mu.Unlock()

		_ = os.MkdirAll(outDir, 0o755)
		b, _ := json.MarshalIndent(res, "", "  ")
		_ = os.WriteFile(jsonFile, b, 0o644)
		fmt.Printf("✅  %s benchmark saved (%d clones, %d APS)\n", t, clones, aps)
	}

	fmt.Println("finished benchmark")
}
