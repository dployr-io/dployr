// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dployr-io/dployr/pkg/core/logs"
	"github.com/dployr-io/dployr/pkg/shared"
)

func newTestHandler(t *testing.T, dataDir string) *Handler {
	t.Helper()
	h := &Handler{
		logger:  shared.NewLogger(),
		dataDir: dataDir,
		config: &shared.Config{
			LogMaxChunkBytes:    8 * 1024 * 1024,
			LogBatchSize:        50,
			LogBatchTimeout:     250 * time.Millisecond,
			LogMaxBatchTimeout:  2 * time.Second,
			LogPollInterval:     20 * time.Millisecond,
			LogMaxFileReadBytes: 100 * 1024 * 1024,
			LogMaxStreams:       100,
		},
	}
	streamSemOnce.Do(func() {
		streamSemaphore = shared.NewSemaphore(h.config.LogMaxStreams)
	})
	return h
}

// writeLogFile creates the .dployr/logs directory under dataDir and writes lines to <name>.log
func writeLogFile(t *testing.T, dataDir, name string, lines []string) string {
	t.Helper()
	dir := filepath.Join(dataDir, ".dployr", "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(dir, name+".log")
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeLogFile: %v", err)
	}
	return path
}

func allEntries(chunks []logs.LogChunk) []logs.LogEntry {
	var out []logs.LogEntry
	for _, c := range chunks {
		out = append(out, c.Entries...)
	}
	return out
}

// ---------------------------------------------------------------------------

func TestStreamLogs_EmptyDuration_ReadsFromBeginning(t *testing.T) {
	dir := t.TempDir()
	writeLogFile(t, dir, "my-svc", []string{
		`{"time":"2025-01-01T00:00:01Z","level":"INFO","msg":"line-1"}`,
		`{"time":"2025-01-01T00:00:02Z","level":"INFO","msg":"line-2"}`,
		`{"time":"2025-01-01T00:00:03Z","level":"INFO","msg":"line-3"}`,
	})

	h := newTestHandler(t, dir)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var chunks []logs.LogChunk
	_ = h.StreamLogs(ctx, logs.StreamOptions{StreamID: "s1", Path: "my-svc", Duration: ""}, func(c logs.LogChunk) error {
		chunks = append(chunks, c)
		cancel() // stop after first flush
		return nil
	})

	entries := allEntries(chunks)
	if len(entries) < 3 {
		t.Fatalf("expected at least 3 entries from beginning, got %d", len(entries))
	}
}

func TestStreamLogs_LiveDuration_DoesNotReplayOldEntries(t *testing.T) {
	dir := t.TempDir()
	oldTime := time.Now().Add(-10 * time.Minute)
	var oldLines []string
	for i := range 5 {
		oldLines = append(oldLines, fmt.Sprintf(`{"time":%q,"level":"INFO","msg":"old-%d"}`, oldTime.Format(time.RFC3339), i))
	}
	writeLogFile(t, dir, "my-svc", oldLines)

	h := newTestHandler(t, dir)
	ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
	defer cancel()

	var chunks []logs.LogChunk
	_ = h.StreamLogs(ctx, logs.StreamOptions{StreamID: "s-live", Path: "my-svc", Duration: "live"}, func(c logs.LogChunk) error {
		chunks = append(chunks, c)
		return nil
	})

	for _, e := range allEntries(chunks) {
		if strings.Contains(e.RawLine, "old-") {
			t.Errorf("live mode must not replay entries older than 5 minutes, got: %s", e.RawLine)
		}
	}
}

func TestStreamLogs_ContextCancellation_StopsStream(t *testing.T) {
	dir := t.TempDir()
	writeLogFile(t, dir, "my-svc", []string{`{"time":"2025-01-01T00:00:01Z","level":"INFO","msg":"ping"}`})

	h := newTestHandler(t, dir)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = h.StreamLogs(ctx, logs.StreamOptions{StreamID: "s1", Path: "my-svc", Duration: "live"}, func(logs.LogChunk) error {
			return nil
		})
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("StreamLogs did not stop within 2s after context cancellation")
	}
}

func TestStreamLogs_FileNotExistInitially_WaitsAndOpens(t *testing.T) {
	dir := t.TempDir()
	logDir := filepath.Join(dir, ".dployr", "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	h := newTestHandler(t, dir)

	go func() {
		time.Sleep(150 * time.Millisecond)
		path := filepath.Join(logDir, "late-svc.log")
		_ = os.WriteFile(path, []byte("{\"time\":\"2025-01-01T00:00:01Z\",\"level\":\"INFO\",\"msg\":\"appeared\"}\n"), 0o644)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var got []logs.LogEntry
	_ = h.StreamLogs(ctx, logs.StreamOptions{StreamID: "s2", Path: "late-svc", Duration: ""}, func(c logs.LogChunk) error {
		got = append(got, c.Entries...)
		cancel()
		return nil
	})

	if len(got) == 0 {
		t.Fatal("expected at least one entry after file appeared, got none")
	}
}

func TestStreamLogs_BatchesMultipleLines(t *testing.T) {
	dir := t.TempDir()
	var lines []string
	for i := range 20 {
		lines = append(lines, fmt.Sprintf(`{"time":"2025-01-01T00:00:%02dZ","level":"INFO","msg":"entry-%d"}`, i%60, i))
	}
	writeLogFile(t, dir, "batchy-svc", lines)

	h := newTestHandler(t, dir)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var chunks []logs.LogChunk
	_ = h.StreamLogs(ctx, logs.StreamOptions{StreamID: "s3", Path: "batchy-svc", Duration: ""}, func(c logs.LogChunk) error {
		chunks = append(chunks, c)
		cancel()
		return nil
	})

	entries := allEntries(chunks)
	if len(entries) == 0 {
		t.Fatal("expected batched entries, got none")
	}
	if len(chunks) >= 20 {
		t.Errorf("expected batched chunks (< 20), got %d separate chunks", len(chunks))
	}
}
