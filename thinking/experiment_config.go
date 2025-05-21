package main

import (
	"encoding/json"
	"os"
)

// Top-level config
type ExperimentConfig struct {
	Name                      string           `json:"name"`
	Description               string           `json:"description"`
	NumericalTypes            []string         `json:"numerical_types"`
	Planets                   []string         `json:"planets"`
	Episodes                  int              `json:"episodes"`
	CheckpointReward          int              `json:"checkpoint_reward"`
	EnableCheckpointing       bool             `json:"enable_checkpointing"`
	SpectrumSteps             int              `json:"spectrum_steps"`
	SpectrumMaxStdDev         float64          `json:"spectrum_max_stddev"`
	MutationStrategy          MutationStrategy `json:"mutation_strategy"`
	NetworkConfig             NetworkConfig    `json:"network_config"`
	Movement                  MovementConfig   `json:"movement"`
	Scoring                   ScoringConfig    `json:"scoring"`
	Evaluation                EvaluationConfig `json:"evaluation"`
	AutoLaunch                bool             `json:"auto_launch"`
	Notes                     string           `json:"notes"`
	AutoState                 bool             `json:"auto_state"`
	EvaluationSpawnsPerPlanet int              `json:"evaluation_spawns_per_planet"`
}

// Nested structs
type MutationStrategy struct {
	ApplyNoiseTo   string `json:"apply_noise_to"`
	NoiseType      string `json:"noise_type"`
	ReuseBestModel bool   `json:"reuse_best_model"`
}

type NetworkConfig struct {
	Layers []Layer `json:"layers"`
}
type Layer struct {
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Activation string `json:"activation"`
}

type MovementConfig struct {
	Translation MovementSubConfig `json:"translation"`
	Rotation    MovementSubConfig `json:"rotation"`
	MaxLifespan int               `json:"max_lifespan_seconds"`
}
type MovementSubConfig struct {
	Clamp            Vector3 `json:"clamp"`
	ActionsPerSecond int     `json:"actions_per_second"`
}
type Vector3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type ScoringConfig struct {
	Type               string `json:"type"`
	EvaluateEveryTick  bool   `json:"evaluate_every_tick"`
	Method             string `json:"method"`
	RewardFormula      string `json:"reward_formula"`
	AccumulateOverLife bool   `json:"accumulate_over_life"`
	ScoreIfTimeout     bool   `json:"score_if_timeout"`
	Normalize          bool   `json:"normalize"`
	Notes              string `json:"notes"`
}

type EvaluationConfig struct {
	PerPlanetTracking  bool `json:"per_planet_tracking"`
	SaveFinalDistances bool `json:"save_final_distances"`
	SaveCheckpointHits bool `json:"save_checkpoint_hits"`
}

// Load and parse the config
func LoadExperimentConfig(path string) (*ExperimentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg ExperimentConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
