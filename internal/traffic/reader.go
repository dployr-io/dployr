// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package traffic

import (
	"bufio"
	"encoding/json"
	"os"
	"time"
)

// LogEntry represents a single line from Caddy's JSON access log.
// Fields match Caddy's default structured log output.
type LogEntry struct {
	TS      float64    `json:"ts"` // Unix timestamp with fractional seconds
	Request LogRequest `json:"request"`
	Status  int        `json:"status"`
}

// LogRequest holds the per-request fields we care about for signal computation.
type LogRequest struct {
	RemoteIP string `json:"remote_ip"`
	URI      string `json:"uri"`
	Method   string `json:"method"`
}

// ReadLastHour reads a Caddy JSON access log file and returns all entries
// whose timestamp falls within the past hour. Lines that cannot be parsed
// are silently skipped so a single malformed line never blocks analysis.
func ReadLastHour(logPath string) ([]LogEntry, error) {
	f, err := os.Open(logPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cutoff := float64(time.Now().Add(-time.Hour).Unix())
	var entries []LogEntry

	scanner := bufio.NewScanner(f)
	// Caddy log lines can be long when headers are included; 1 MB buffer is safe.
	scanner.Buffer(make([]byte, 1<<20), 1<<20)

	for scanner.Scan() {
		var e LogEntry
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			continue
		}
		if e.TS >= cutoff {
			entries = append(entries, e)
		}
	}

	return entries, scanner.Err()
}
