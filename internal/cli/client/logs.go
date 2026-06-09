package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// LogQuery parameters for historical log retrieval.
// The API endpoint and exact field names will be confirmed once the
// centralized logging implementation in dployr-base is finalized.
type LogQuery struct {
	ServiceID string
	Source    string    // build | runtime | all (default: all)
	Since     time.Time // zero means no filter
	Limit     int       // 0 means server default
}

// GetLogs fetches historical log entries for a service.
//
// TODO: Wire to GET /v1/services/:id/logs once the centralized log API is finalized.
// Expected query params: source, since (RFC3339), limit.
func (c *Client) GetLogs(ctx context.Context, q LogQuery) ([]LogChunk, error) {
	query := url.Values{}
	if q.Source != "" && q.Source != "all" {
		query.Set("source", q.Source)
	}
	if !q.Since.IsZero() {
		query.Set("since", q.Since.UTC().Format(time.RFC3339))
	}
	if q.Limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", q.Limit))
	}

	path := fmt.Sprintf("/services/%s/logs", q.ServiceID)
	resp, err := c.do(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("service %q not found", q.ServiceID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, readAPIError(resp)
	}

	var chunks []LogChunk
	if err := json.NewDecoder(resp.Body).Decode(&chunks); err != nil {
		return nil, fmt.Errorf("decode log response: %w", err)
	}
	return chunks, nil
}

// StreamLogs opens a server-sent events (or newline-delimited JSON) stream for
// real-time log output from both build and instance nodes.
//
// TODO: The streaming endpoint and wire format will be confirmed once the
// centralized logging implementation in dployr-base is finalized.
// Candidate endpoints:
//   - GET /v1/services/:id/logs?follow=true  (SSE / ndjson)
//   - WS  /v1/ws/logs/:serviceId             (WebSocket, LogSubscribeRequest protocol)
//
// onChunk is called for each received LogChunk. The stream ends when ctx is cancelled
// or the server closes the connection.
func (c *Client) StreamLogs(ctx context.Context, serviceID, source string, onChunk func(LogChunk)) error {
	query := url.Values{
		"follow": {"true"},
	}
	if source != "" && source != "all" {
		query.Set("source", source)
	}

	path := fmt.Sprintf("/services/%s/logs", serviceID)
	resp, err := c.do(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("service %q not found", serviceID)
	}
	if resp.StatusCode == http.StatusNotImplemented {
		return fmt.Errorf("centralized log streaming is not yet available on this server")
	}
	if resp.StatusCode != http.StatusOK {
		return readAPIError(resp)
	}

	// Read newline-delimited JSON chunks.
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var chunk LogChunk
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			// Non-JSON line (SSE comment or keep-alive) — skip.
			continue
		}
		onChunk(chunk)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("log stream error: %w", err)
	}
	return nil
}
