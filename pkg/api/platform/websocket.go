package platform

import (
	"time"
)

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Phase     string    `json:"phase"`
	Message   string    `json:"message"`
	Stream    string    `json:"stream"`
}

type WebSocketClient interface {
	StreamLog(deploymentID string, entry LogEntry)
}
