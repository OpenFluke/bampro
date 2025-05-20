package main

import (
	"fmt"
	"os"

	paragon "github.com/OpenFluke/PARAGON"
)

// NamedNetwork describes each built network
type NamedNetwork struct {
	TypeName string
	Mode     string
	Net      any
}

// Global array is declared in main.go
// var GlobalNetworks []NamedNetwork

type TypeModeBuilder struct {
	TypeName string
	BuildSet func(
		layers []struct{ Width, Height int },
		acts []string,
		full []bool,
	)
}

// ✅ Corrected function signatures
var allTypeModeBuilders = []TypeModeBuilder{
	{"int", func(layers []struct{ Width, Height int }, acts []string, full []bool) {
		buildVariants[int]("int", layers, acts, full)
	}},
	{"int8", func(layers []struct{ Width, Height int }, acts []string, full []bool) {
		buildVariants[int8]("int8", layers, acts, full)
	}},
	{"int16", func(layers []struct{ Width, Height int }, acts []string, full []bool) {
		buildVariants[int16]("int16", layers, acts, full)
	}},
	{"int32", func(layers []struct{ Width, Height int }, acts []string, full []bool) {
		buildVariants[int32]("int32", layers, acts, full)
	}},
	{"int64", func(layers []struct{ Width, Height int }, acts []string, full []bool) {
		buildVariants[int64]("int64", layers, acts, full)
	}},
	{"uint", func(layers []struct{ Width, Height int }, acts []string, full []bool) {
		buildVariants[uint]("uint", layers, acts, full)
	}},
	{"uint8", func(layers []struct{ Width, Height int }, acts []string, full []bool) {
		buildVariants[uint8]("uint8", layers, acts, full)
	}},
	{"uint16", func(layers []struct{ Width, Height int }, acts []string, full []bool) {
		buildVariants[uint16]("uint16", layers, acts, full)
	}},
	{"uint32", func(layers []struct{ Width, Height int }, acts []string, full []bool) {
		buildVariants[uint32]("uint32", layers, acts, full)
	}},
	{"uint64", func(layers []struct{ Width, Height int }, acts []string, full []bool) {
		buildVariants[uint64]("uint64", layers, acts, full)
	}},
	{"float32", func(layers []struct{ Width, Height int }, acts []string, full []bool) {
		buildVariants[float32]("float32", layers, acts, full)
	}},
	{"float64", func(layers []struct{ Width, Height int }, acts []string, full []bool) {
		buildVariants[float64]("float64", layers, acts, full)
	}},
}

func buildAllNetworks(
	layers []struct{ Width, Height int },
	acts []string,
	full []bool,
) {
	for _, tb := range allTypeModeBuilders {
		tb.BuildSet(layers, acts, full)
	}
}

func buildVariants[T paragon.Numeric](
	typeName string,
	layers []struct{ Width, Height int },
	acts []string,
	full []bool,
) {
	buildWithMode[T](typeName, "Standard", layers, acts, full, func(nn *paragon.Network[T]) {})

	buildWithMode[T](typeName, "Replay", layers, acts, full, func(nn *paragon.Network[T]) {
		layer := &nn.Layers[1]
		layer.ReplayEnabled = true
		layer.ReplayPhase = "after"
		layer.ReplayOffset = -1
		layer.MaxReplay = 1
	})

	buildWithMode[T](typeName, "DynamicReplay", layers, acts, full, func(nn *paragon.Network[T]) {
		layer := &nn.Layers[1]
		layer.ReplayEnabled = true
		layer.ReplayBudget = 3
		layer.ReplayGateFunc = func(_ [][]T) float64 { return 0.6 }
		layer.ReplayGateToReps = func(score float64) int {
			switch {
			case score > 0.8:
				return 3
			case score > 0.6:
				return 2
			default:
				return 1
			}
		}
	})
}

func buildWithMode[T paragon.Numeric](
	typeName, mode string,
	layers []struct{ Width, Height int },
	acts []string,
	full []bool,
	config func(*paragon.Network[T]),
) {
	nn := paragon.NewNetwork[T](layers, acts, full)
	config(nn)

	GlobalNetworks = append(GlobalNetworks, NamedNetwork{
		TypeName: typeName,
		Mode:     mode,
		Net:      nn,
	})

	fmt.Printf("🧱 Built: %-8s | Mode: %s\n", typeName, mode)
}

func (tb TypeModeBuilder) BuildSetWithSave(
	layers []struct{ Width, Height int },
	acts []string,
	full []bool,
	gen int,
) {
	switch tb.TypeName {
	case "int":
		buildVariantsWithSave[int](tb.TypeName, layers, acts, full, gen)
	case "int8":
		buildVariantsWithSave[int8](tb.TypeName, layers, acts, full, gen)
	case "int16":
		buildVariantsWithSave[int16](tb.TypeName, layers, acts, full, gen)
	case "int32":
		buildVariantsWithSave[int32](tb.TypeName, layers, acts, full, gen)
	case "int64":
		buildVariantsWithSave[int64](tb.TypeName, layers, acts, full, gen)
	case "uint":
		buildVariantsWithSave[uint](tb.TypeName, layers, acts, full, gen)
	case "uint8":
		buildVariantsWithSave[uint8](tb.TypeName, layers, acts, full, gen)
	case "uint16":
		buildVariantsWithSave[uint16](tb.TypeName, layers, acts, full, gen)
	case "uint32":
		buildVariantsWithSave[uint32](tb.TypeName, layers, acts, full, gen)
	case "uint64":
		buildVariantsWithSave[uint64](tb.TypeName, layers, acts, full, gen)
	case "float32":
		buildVariantsWithSave[float32](tb.TypeName, layers, acts, full, gen)
	case "float64":
		buildVariantsWithSave[float64](tb.TypeName, layers, acts, full, gen)
	}
}

func buildVariantsWithSave[T paragon.Numeric](
	typeName string,
	layers []struct{ Width, Height int },
	acts []string,
	full []bool,
	generation int,
) {
	buildWithModeAndSave[T](typeName, "Standard", layers, acts, full, generation, func(nn *paragon.Network[T]) {})

	buildWithModeAndSave[T](typeName, "Replay", layers, acts, full, generation, func(nn *paragon.Network[T]) {
		layer := &nn.Layers[1]
		layer.ReplayEnabled = true
		layer.ReplayPhase = "after"
		layer.ReplayOffset = -1
		layer.MaxReplay = 1
	})

	buildWithModeAndSave[T](typeName, "DynamicReplay", layers, acts, full, generation, func(nn *paragon.Network[T]) {
		layer := &nn.Layers[1]
		layer.ReplayEnabled = true
		layer.ReplayBudget = 3
		layer.ReplayGateFunc = func(_ [][]T) float64 { return 0.6 }
		layer.ReplayGateToReps = func(score float64) int {
			switch {
			case score > 0.8:
				return 3
			case score > 0.6:
				return 2
			default:
				return 1
			}
		}
	})
}

func buildWithModeAndSave[T paragon.Numeric](
	typeName, mode string,
	layers []struct{ Width, Height int },
	acts []string,
	full []bool,
	gen int,
	config func(*paragon.Network[T]),
) {
	nn := paragon.NewNetwork[T](layers, acts, full)
	config(nn)

	GlobalNetworks = append(GlobalNetworks, NamedNetwork{
		TypeName: typeName,
		Mode:     mode,
		Net:      nn,
	})

	fmt.Printf("🧱 Built: %-8s | Mode: %s\n", typeName, mode)

	saveDir := fmt.Sprintf("models/%d", gen)
	_ = os.MkdirAll(saveDir, 0755)
	savePath := fmt.Sprintf("%s/%s_%s.json", saveDir, typeName, mode)

	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		if err := nn.SaveJSON(savePath); err != nil {
			fmt.Printf("❌ Failed to save model %s: %v\n", savePath, err)
		} else {
			fmt.Printf("💾 Saved model: %s\n", savePath)
		}
	}
}
