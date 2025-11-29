// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"context"
	"encoding/json"
)

// LogEntry represents a single structured log entry.
// This matches the slog JSON output format.
type LogEntry struct {
	Time    string                 `json:"time"`
	Level   string                 `json:"level"`
	Msg     string                 `json:"msg"`
	Attrs   map[string]interface{} `json:"-"`
	RawLine string                 `json:"-"` // Original line for fallback
}

// UnmarshalJSON custom unmarshaler to capture all fields including dynamic ones.
func (e *LogEntry) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if t, ok := raw["time"].(string); ok {
		e.Time = t
	}
	if l, ok := raw["level"].(string); ok {
		e.Level = l
	}
	if m, ok := raw["msg"].(string); ok {
		e.Msg = m
	}

	if e.Time == "" {
		if ts, ok := raw["timestamp"].(string); ok {
			e.Time = ts
		}
	}
	if e.Msg == "" {
		if msg, ok := raw["message"].(string); ok {
			e.Msg = msg
		}
	}

	e.Attrs = make(map[string]interface{})
	for k, v := range raw {
		if k != "time" && k != "level" && k != "msg" && k != "timestamp" && k != "message" {
			e.Attrs[k] = v
		}
	}

	return nil
}

// MarshalJSON custom marshaler to flatten Attrs into the main object.
func (e LogEntry) MarshalJSON() ([]byte, error) {
	result := make(map[string]interface{})
	result["time"] = e.Time
	result["level"] = e.Level
	result["msg"] = e.Msg

	// Merge attrs
	for k, v := range e.Attrs {
		result[k] = v
	}

	return json.Marshal(result)
}

// LogChunk represents a batch of log entries to be sent via WebSocket.
type LogChunk struct {
	StreamID string     `json:"streamId"`
	LogType  string     `json:"logType"`
	Entries  []LogEntry `json:"entries"`
	EOF      bool       `json:"eof"`     // End of file
	HasMore  bool       `json:"hasMore"` // More logs available in history
	Offset   int64      `json:"offset"`  // Current offset in file
}

// StreamOptions configures log streaming behavior.
type StreamOptions struct {
	StreamID  string
	LogType   string
	Mode      StreamMode // "tail" or "historical"
	StartFrom int64      // Byte offset to start from (0 = beginning, -1 = end)
	Limit     int        // Max entries to return (for historical mode)
	Follow    bool       // Continue tailing after reaching end (tail mode)
}

// StreamMode defines how logs should be streamed.
type StreamMode string

const (
	StreamModeTail       StreamMode = "tail"       // Start from end, follow new logs
	StreamModeHistorical StreamMode = "historical" // Read from offset, don't follow
)

// LogStreamer defines the interface for streaming logs.
type LogStreamer interface {
	StreamLogs(ctx context.Context, opts StreamOptions, sendChunk func(chunk LogChunk) error) error
}
