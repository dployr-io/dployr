// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package system

// ProcessInfo represents a single process's resource usage matching top output format.
type ProcessInfo struct {
	PID      int     `json:"pid"`
	User     string  `json:"user"`
	Priority int     `json:"priority"`
	Nice     int     `json:"nice"`
	VirtMem  int64   `json:"virt_mem"`
	ResMem   int64   `json:"res_mem"`
	ShrMem   int64   `json:"shr_mem"`
	State    string  `json:"state"`
	CPUPct   float64 `json:"cpu_pct"`
	MEMPct   float64 `json:"mem_pct"`
	Time     string  `json:"time"`
	Command  string  `json:"command"`
}

// CPUStats represents CPU usage breakdown
type CPUStats struct {
	User   float64 `json:"user"`
	System float64 `json:"system"`
	Nice   float64 `json:"nice"`
	Idle   float64 `json:"idle"`
	Wait   float64 `json:"wait"`
	HI     float64 `json:"hi"`
	SI     float64 `json:"si"`
	ST     float64 `json:"st"`
}

// MemoryStats represents memory usage (in MiB).
type MemoryStats struct {
	Total       float64 `json:"total"`
	Free        float64 `json:"free"`
	Used        float64 `json:"used"`
	BufferCache float64 `json:"buffer_cache"`
}

// SwapStats represents swap usage (in MiB).
type SwapStats struct {
	Total     float64 `json:"total"`
	Free      float64 `json:"free"`
	Used      float64 `json:"used"`
	Available float64 `json:"available"`
}

// LoadAverage represents system load averages.
type LoadAverage struct {
	One     float64 `json:"one"`
	Five    float64 `json:"five"`
	Fifteen float64 `json:"fifteen"`
}

// TaskStats represents task/process counts by state.
type TaskStats struct {
	Total    int `json:"total"`
	Running  int `json:"running"`
	Sleeping int `json:"sleeping"`
	Stopped  int `json:"stopped"`
	Zombie   int `json:"zombie"`
}

// TopHeader represents the header information from top.
type TopHeader struct {
	Time    string      `json:"time"`
	Uptime  string      `json:"uptime"`
	Users   int         `json:"users"`
	LoadAvg LoadAverage `json:"load_avg"`
}

// SystemTop represents a snapshot of system resource usage (like `top` output).
type SystemTop struct {
	Header    TopHeader     `json:"header"`
	Tasks     TaskStats     `json:"tasks"`
	CPU       CPUStats      `json:"cpu"`
	Memory    MemoryStats   `json:"memory"`
	Swap      SwapStats     `json:"swap"`
	Processes []ProcessInfo `json:"processes"`
}

// TopRequest represents query parameters for the /system/top endpoint.
type TopRequest struct {
	SortBy string `json:"sort_by"` // "cpu" or "mem"
	Limit  int    `json:"limit"`   // max processes to return
}
