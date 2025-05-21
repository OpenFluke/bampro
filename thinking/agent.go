package main

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	paragon "github.com/OpenFluke/PARAGON"
)

type Agent struct {
	ID         string
	Generation int
	VariantID  string
	Network    NamedNetwork
	Config     *ExperimentConfig

	PlanetName string
	PlanetPos  Vec3
}

type Vec3 struct {
	X, Y, Z float64
}

func clamp(f, max float64) float64 {
	if f > max {
		return max
	} else if f < -max {
		return -max
	}
	return f
}

func runAgent(a Agent) {
	lifespan := time.Duration(a.Config.Movement.MaxLifespan) * time.Second
	actionRate := time.Second / time.Duration(a.Config.Movement.Translation.ActionsPerSecond)

	clampX := a.Config.Movement.Translation.Clamp.X
	clampY := a.Config.Movement.Translation.Clamp.Y
	clampZ := a.Config.Movement.Translation.Clamp.Z

	fmt.Printf("🧠 [%s] Starting agent loop for %.0fs at %d APS\n", a.ID, lifespan.Seconds(), a.Config.Movement.Translation.ActionsPerSecond)

	// Static dummy float64 input (shape: 1 row of 6 features)
	dummyInput := [][]float64{{10.1, -10.2, 10.3, -10.4, 10.5, -10.6}}

	position := Vec3{0, 0, 0}
	start := time.Now()
	ticker := time.NewTicker(actionRate)
	defer ticker.Stop()

	tickCount := 0

	for {
		select {
		case <-ticker.C:
			tickCount++

			switch net := a.Network.Net.(type) {
			case *paragon.Network[float32], *paragon.Network[float64],
				*paragon.Network[int], *paragon.Network[int8],
				*paragon.Network[int16], *paragon.Network[int32], *paragon.Network[int64],
				*paragon.Network[uint], *paragon.Network[uint8],
				*paragon.Network[uint16], *paragon.Network[uint32], *paragon.Network[uint64]:

				// Use reflect to call the interface-typed network
				// Forward pass
				reflect.ValueOf(net).MethodByName("Forward").Call([]reflect.Value{
					reflect.ValueOf(dummyInput),
				})

				// Get output slice
				result := reflect.ValueOf(net).MethodByName("GetOutput").Call(nil)
				if len(result) == 1 {
					output := result[0].Interface().([]float64)
					//fmt.Println(output)
					if len(output) >= 3 {
						dx := clamp(output[0], clampX)
						dy := clamp(output[1], clampY)
						dz := clamp(output[2], clampZ)
						position.X += dx
						position.Y += dy
						position.Z += dz
					}
				}

			default:
				fmt.Printf("⚠️ [%s] Unsupported model type: %s\n", a.ID, reflect.TypeOf(a.Network.Net))
			}

			/*if tickCount%10 == 0 {
				fmt.Printf("📍 [%s] Pos = (%.3f, %.3f, %.3f)\n", a.ID, position.X, position.Y, position.Z)
			}*/

		default:
			if time.Since(start) > lifespan {
				fmt.Printf("💀 [%s] Agent timed out after %.2f seconds\n", a.ID, time.Since(start).Seconds())
				return
			}
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func RunAgentsInPool(agents []Agent) {
	var wg sync.WaitGroup
	for _, agent := range agents {
		wg.Add(1)
		go func(a Agent) {
			defer wg.Done()
			runAgent(a)
		}(agent)
	}
	wg.Wait()
}
