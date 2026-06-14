// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package version_resolver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// EOLClient fetches release cycle data for a given product.
// The product name matches endoflife.date's URL slug (e.g. "python", "nodejs").
type EOLClient interface {
	Cycles(product string) ([]Cycle, error)
}

// Cycle is a single release line as returned by endoflife.date.
type Cycle struct {
	Cycle       string   `json:"cycle"`
	Latest      string   `json:"latest"`
	EOL         EOLDate  `json:"eol"`
	ReleaseDate string   `json:"releaseDate"`
	LTS         FlexBool `json:"lts"`
}

// FlexBool unmarshals a JSON value that is either a boolean or a date string
// (e.g. Node.js LTS end dates). Any truthy value is treated as true.
type FlexBool struct{ v bool }

func NewFlexBool(v bool) FlexBool { return FlexBool{v: v} }

func (f FlexBool) Bool() bool { return f.v }

func (f *FlexBool) UnmarshalJSON(b []byte) error {
	switch string(b) {
	case "false", "null":
		f.v = false
	case "true":
		f.v = true
	default:
		// Date string like "2026-04-30" — present means LTS is active.
		f.v = true
	}
	return nil
}

// EOLDate is either "no end-of-life" (the zero value) or a calendar date.
// The endoflife.date API encodes this as either the boolean false or a
// "YYYY-MM-DD" string, so a custom unmarshaler is required.
type EOLDate struct {
	t *time.Time
}

// NewEOLDate constructs an EOLDate from a "YYYY-MM-DD" string.
// Pass "" to represent a release with no scheduled end-of-life.
func NewEOLDate(date string) EOLDate {
	if date == "" {
		return EOLDate{}
	}
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return EOLDate{}
	}
	return EOLDate{t: &t}
}

func (e *EOLDate) UnmarshalJSON(b []byte) error {
	switch string(b) {
	case "false", "null":
		e.t = nil
		return nil
	case "true":
		// EOL but no specific date recorded — treat as already expired.
		past := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		e.t = &past
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("eol field: %w", err)
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return fmt.Errorf("eol date %q: %w", s, err)
	}
	e.t = &t
	return nil
}

// Expired reports whether the end-of-life date has already passed.
func (e EOLDate) Expired() bool {
	return e.t != nil && time.Now().After(*e.t)
}

func (e EOLDate) String() string {
	if e.t == nil {
		return ""
	}
	return e.t.Format("2006-01-02")
}

// httpEOLClient fetches from endoflife.date with an in-memory TTL cache.
type httpEOLClient struct {
	mu    sync.RWMutex
	cache map[string]cacheEntry
	ttl   time.Duration
	http  *http.Client
}

type cacheEntry struct {
	cycles    []Cycle
	fetchedAt time.Time
}

// NewHTTPClient returns an EOLClient backed by endoflife.date with a 24-hour cache.
func NewHTTPClient() EOLClient {
	return &httpEOLClient{
		cache: make(map[string]cacheEntry),
		ttl:   24 * time.Hour,
		http:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *httpEOLClient) Cycles(product string) ([]Cycle, error) {
	c.mu.RLock()
	e, ok := c.cache[product]
	c.mu.RUnlock()

	if ok && time.Since(e.fetchedAt) < c.ttl {
		return e.cycles, nil
	}

	cycles, err := c.fetch(product)
	if err != nil {
		if ok {
			return e.cycles, nil // serve stale rather than failing mid-build
		}
		return nil, err
	}

	c.mu.Lock()
	c.cache[product] = cacheEntry{cycles: cycles, fetchedAt: time.Now()}
	c.mu.Unlock()

	return cycles, nil
}

func (c *httpEOLClient) fetch(product string) ([]Cycle, error) {
	resp, err := c.http.Get("https://endoflife.date/api/" + product + ".json")
	if err != nil {
		return nil, fmt.Errorf("endoflife.date: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("endoflife.date: %d for product %q", resp.StatusCode, product)
	}

	var out []Cycle
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("endoflife.date decode: %w", err)
	}
	return out, nil
}
