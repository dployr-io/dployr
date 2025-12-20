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
	Path     string     `json:"path"`
	Entries  []LogEntry `json:"entries"`
	EOF      bool       `json:"eof"`     // End of file
	HasMore  bool       `json:"hasMore"` // More logs available in history
	Offset   int64      `json:"offset"`  // Current offset in file
}

// StreamOptions configures log streaming behavior.
type StreamOptions struct {
	StreamID  string
	Path      string
	StartFrom int64  // Byte offset for pagination (0 if not resuming)
	Limit     int    // Max entries to return per chunk
	Duration  string // Controls behavior: "live" = tail from now, time duration = read history then tail
}

// LogStreamer defines the interface for streaming logs.
type LogStreamer interface {
	StreamLogs(ctx context.Context, opts StreamOptions, sendChunk func(chunk LogChunk) error) error
}
