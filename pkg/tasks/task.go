// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package tasks

import (
	"encoding/json"
)

// Task represents a remotely-triggered action from base.
type Task struct {
	ID      string
	Type    string // TaskAddress format: "path:method" e.g. "system/status:get"
	Payload json.RawMessage
	Status  string
}

// Result represents the outcome of executing a task.
type Result struct {
	ID     string `json:"id"`
	Status string `json:"status"` // "done" or "failed"
	Result any    `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}
