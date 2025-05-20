package main

import (
	"encoding/json"
	"log"
)

// Universal wrapper for all WebSocket messages
type TypedMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Outbound message types
const (
	TypeStatusUpdate   = "status_update"
	TypeExperimentConf = "experiment_config"
	TypeExperimentDone = "experiment_done"
)

// SerializeTyped returns a JSON-encoded message of {type, data}
func SerializeTyped(msgType string, payload interface{}) []byte {
	wrapped := TypedMessage{
		Type: msgType,
		Data: payload,
	}
	data, err := json.Marshal(wrapped)
	if err != nil {
		log.Printf("❌ Failed to marshal %s message: %v", msgType, err)
		return nil
	}
	return data
}
