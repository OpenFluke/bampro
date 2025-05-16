package main

// GlobalNetworks holds all constructed networks — accessible from anywhere
var GlobalNetworks []NamedNetwork

func main() {
	// This matches the Actor network used in Biofoundry agents
	layers := []struct{ Width, Height int }{
		{6, 1},   // StateDim
		{128, 1}, // Hidden layer 1
		{128, 1}, // Hidden layer 2
		{3, 1},   // ActionDim
	}
	acts := []string{"linear", "relu", "relu", "tanh"}
	full := []bool{true, true, true, true}

	// Build all variants for all types
	buildAllNetworks(layers, acts, full)
	TryToConnect()
	host()

}
