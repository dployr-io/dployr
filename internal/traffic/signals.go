// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package traffic

import (
	"math"
	"net"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dployr-io/dployr/pkg/core/utils"
)

// Signals holds the computed bot-detection signals for a single service
// over a rolling 1-hour observation window.
type Signals struct {
	// Domain is the Caddy virtual-host domain this log file belongs to.
	Domain string `json:"domain"`

	// WindowHours is the observation window used (always 1).
	WindowHours int `json:"window_hours"`

	// RequestCount is the total number of requests seen in the window.
	RequestCount int `json:"request_count"`

	// UniqueSubnets is the count of distinct /24 networks that sent requests.
	// Fewer than 3 suggests a single automated source.
	UniqueSubnets int `json:"unique_subnets"`

	// CadenceCV is the coefficient of variation (stddev/mean) of inter-request
	// gap durations in seconds. Values below 0.2 indicate metronomic (bot-like)
	// cadence. Set to 1.0 when fewer than 2 requests are present.
	CadenceCV float64 `json:"cadence_cv"`

	// UniquePaths is the count of distinct URI paths (query strings stripped).
	// Fewer than 2 suggests a bot hitting only / or /health.
	UniquePaths int `json:"unique_paths"`

	// LastRequestAt is the Unix millisecond timestamp of the most recent request
	// in the window, or 0 if the window is empty.
	LastRequestAt int64 `json:"last_request_at"`
}

// subnet24 extracts the /24 network prefix string from an IP address.
// Returns the raw IP string unchanged if it cannot be parsed (e.g. IPv6).
func subnet24(rawIP string) string {
	ip := net.ParseIP(rawIP)
	if ip == nil {
		return rawIP
	}
	v4 := ip.To4()
	if v4 == nil {
		// IPv6 — use a /48 approximation by taking the first 6 bytes as the prefix.
		return ip.String()[:strings.LastIndex(ip.String(), ":")]
	}
	parts := strings.Split(v4.String(), ".")
	if len(parts) < 3 {
		return rawIP
	}
	return strings.Join(parts[:3], ".")
}

// stripQuery removes the query string from a URI, returning only the path.
func stripQuery(uri string) string {
	if idx := strings.IndexByte(uri, '?'); idx != -1 {
		return uri[:idx]
	}
	return uri
}

// cadenceCV computes the coefficient of variation of inter-request gap durations
// given a sorted slice of Unix timestamps (fractional seconds).
// Returns 1.0 when there are fewer than 2 samples — not enough data to judge regularity.
func cadenceCV(timestamps []float64) float64 {
	if len(timestamps) < 2 {
		return 1.0
	}

	gaps := make([]float64, len(timestamps)-1)
	for i := range gaps {
		gaps[i] = timestamps[i+1] - timestamps[i]
	}

	var sum float64
	for _, g := range gaps {
		sum += g
	}
	mean := sum / float64(len(gaps))
	if mean == 0 {
		return 0
	}

	var variance float64
	for _, g := range gaps {
		d := g - mean
		variance += d * d
	}
	variance /= float64(len(gaps))

	return math.Sqrt(variance) / mean
}

// Compute derives bot-detection signals for a given domain from its Caddy access log.
// logDir is the directory where Caddy writes per-domain log files
// (typically ~/.dployr/logs/caddy/).
func Compute(domain, logDir string) Signals {
	sig := Signals{
		Domain:      domain,
		WindowHours: 1,
		CadenceCV:   1.0, // default: assume human-like when no data
	}

	logPath := filepath.Join(logDir, domain+".log")
	entries, err := ReadLastHour(logPath)
	if err != nil || len(entries) == 0 {
		return sig
	}

	sig.RequestCount = len(entries)

	subnets := make(map[string]struct{})
	paths := make(map[string]struct{})
	timestamps := make([]float64, 0, len(entries))
	var lastTS float64

	for _, e := range entries {
		subnets[subnet24(e.Request.RemoteIP)] = struct{}{}
		paths[stripQuery(e.Request.URI)] = struct{}{}
		timestamps = append(timestamps, e.TS)
		if e.TS > lastTS {
			lastTS = e.TS
		}
	}

	sig.UniqueSubnets = len(subnets)
	sig.UniquePaths = len(paths)
	sig.LastRequestAt = int64(lastTS * 1000)

	sort.Float64s(timestamps)
	sig.CadenceCV = cadenceCV(timestamps)

	return sig
}

// CaddyLogDir returns the canonical path to the Caddy per-domain log directory.
func CaddyLogDir() string {
	return filepath.Join(utils.GetDataDir(), ".dployr", "logs", "caddy")
}
