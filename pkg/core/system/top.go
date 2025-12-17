// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package system

import "time"

// ProcessInfo represents a single process's resource usage.
type ProcessInfo struct {
	PID        int32   `json:"pid"`
	User       string  `json:"user"`
	Command    string  `json:"command"`
	CPUPercent float64 `json:"cpu_percent"`
	MemPercent float32 `json:"mem_percent"`
	RSS        int64   `json:"rss_bytes"`
	VMS        int64   `json:"vms_bytes"`
	State      string  `json:"state"`
}

// CPUStats represents CPU usage breakdown.
type CPUStats struct {
	User   float64 `json:"user"`
	System float64 `json:"system"`
	Idle   float64 `json:"idle"`
	IOWait float64 `json:"iowait"`
	Steal  float64 `json:"steal"`
}

// MemoryStats represents memory usage in bytes.
type MemoryStats struct {
	Total       uint64  `json:"total_bytes"`
	Used        uint64  `json:"used_bytes"`
	Free        uint64  `json:"free_bytes"`
	Available   uint64  `json:"available_bytes"`
	UsedPercent float64 `json:"used_percent"`
	SwapTotal   uint64  `json:"swap_total_bytes"`
	SwapUsed    uint64  `json:"swap_used_bytes"`
	SwapFree    uint64  `json:"swap_free_bytes"`
}

// LoadAverage represents system load averages.
type LoadAverage struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

// TaskStats represents task/process counts by state.
type TaskStats struct {
	Total    int `json:"total"`
	Running  int `json:"running"`
	Sleeping int `json:"sleeping"`
	Stopped  int `json:"stopped"`
	Zombie   int `json:"zombie"`
}

// SystemTop represents a snapshot of system resource usage (like `top` output).
type SystemTop struct {
	Timestamp time.Time     `json:"timestamp"`
	Uptime    uint64        `json:"uptime_seconds"`
	LoadAvg   LoadAverage   `json:"load_avg"`
	CPU       CPUStats      `json:"cpu"`
	Memory    MemoryStats   `json:"memory"`
	Tasks     TaskStats     `json:"tasks"`
	Processes []ProcessInfo `json:"processes"`
}

// TopRequest represents query parameters for the /system/top endpoint.
type TopRequest struct {
	SortBy string `json:"sort_by"` // "cpu" or "mem"
	Limit  int    `json:"limit"`   // max processes to return
}
