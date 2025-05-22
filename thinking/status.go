package main

import (
	"sync"
	"time"
)

type ExperimentStatus struct {
	Timestamp  time.Time `json:"timestamp"`
	Generation int       `json:"generation"`
	NumType    string    `json:"num_type"`
	Mode       string    `json:"mode"`
	Variant    int       `json:"variant"`
	Stage      string    `json:"stage"`
	Message    string    `json:"message"`
}

type ScoreRecord struct {
	Generation   int     `json:"generation"`
	NumType      string  `json:"num_type"`
	Mode         string  `json:"mode"`
	VariantIndex int     `json:"variant"`
	MeanProgress float64 `json:"mean_progress"`
}

var (
	StatusUpdates []ExperimentStatus
	statusMu      sync.Mutex
)

func AppendStatus(gen int, numType, mode string, variant int, stage, msg string) {
	statusMu.Lock()
	StatusUpdates = append(StatusUpdates, ExperimentStatus{
		Timestamp:  time.Now(),
		Generation: gen,
		NumType:    numType,
		Mode:       mode,
		Variant:    variant,
		Stage:      stage,
		Message:    msg,
	})
	statusMu.Unlock()
}
