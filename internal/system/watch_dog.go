// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	pollInterval  = 30 * time.Second
	pollTimeout   = 5 * time.Second
	pollThreshold = 3 // consecutive failures before marking degraded
)

// WatchDog probes each service's HTTP endpoint from the host and maintains
// a "healthy" | "degraded" result per service name.
type WatchDog struct {
	mu       sync.RWMutex
	results  map[string]string // service name → "healthy" | "degraded"
	failures map[string]int    // consecutive failure count per service
	client   *http.Client
}

func NewHealthPoller() *WatchDog {
	return &WatchDog{
		results:  make(map[string]string),
		failures: make(map[string]int),
		client: &http.Client{
			Timeout: pollTimeout,
		},
	}
}

// Get returns the last known health result for a service.
// Returns "degraded" if no result has been recorded yet (assume degraded until proven otherwise).
func (p *WatchDog) Get(serviceName string) string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if result, ok := p.results[serviceName]; ok {
		return result
	}
	return "degraded"
}

// Remove clears the health state for a service (called on decommission/stop).
func (p *WatchDog) Remove(serviceName string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.results, serviceName)
	delete(p.failures, serviceName)
}

// Run starts the polling loop. It calls getServices on each tick to discover
// what to probe. Pass the context to stop the loop.
func (p *WatchDog) Run(ctx context.Context, getServices func() []WatchTarget) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, target := range getServices() {
				go p.probe(target)
			}
		}
	}
}

// WatchTarget describes a single service to probe.
type WatchTarget struct {
	Name     string
	HostPort int
	Path     string // optional; defaults to "/". Must be an absolute path (e.g. "/healthz"), not a full URL.
}

// normalisePath returns a canonical absolute path from raw:
//   - ""           → "/"
//   - "foo/bar"    → "/foo/bar"   (A: missing leading slash — added)
//   - "/foo/bar/"  → "/foo/bar"   (C: trailing slash — stripped, except bare "/")
//
// Returns an error if raw looks like a full URL (contains "://").
func normalisePath(raw string) (string, error) {
	if strings.Contains(raw, "://") {
		return "", fmt.Errorf("health check path must be a path, not a full URL: %q", raw)
	}
	if raw == "" {
		return "/", nil
	}
	if !strings.HasPrefix(raw, "/") {
		raw = "/" + raw
	}
	if len(raw) > 1 {
		raw = strings.TrimRight(raw, "/")
	}
	return raw, nil
}

func (p *WatchDog) probe(target WatchTarget) {
	probePath, err := normalisePath(target.Path)
	if err != nil {
		p.mu.Lock()
		p.failures[target.Name]++
		if p.failures[target.Name] >= pollThreshold {
			p.results[target.Name] = "degraded"
		}
		p.mu.Unlock()
		return
	}
	url := fmt.Sprintf("http://localhost:%d%s", target.HostPort, probePath)

	resp, err := p.client.Get(url)
	if err == nil {
		resp.Body.Close()
	}

	healthy := err == nil && resp.StatusCode < 500

	p.mu.Lock()
	defer p.mu.Unlock()

	if healthy {
		p.failures[target.Name] = 0
		p.results[target.Name] = "healthy"
		return
	}

	p.failures[target.Name]++
	if p.failures[target.Name] >= pollThreshold {
		p.results[target.Name] = "degraded"
	}
}
