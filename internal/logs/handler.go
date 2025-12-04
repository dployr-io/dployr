// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/dployr-io/dployr/pkg/core/logs"
	"github.com/dployr-io/dployr/pkg/shared"
)

// Handler implements the LogStreamer interface.
type Handler struct {
	logger *shared.Logger
}

// NewHandler creates a new log stream handler.
func NewHandler(logger *shared.Logger) *Handler {
	return &Handler{logger: logger}
}

// StreamLogs streams logs based on the provided options.
// Supports both tail mode (follow new logs) and historical mode (read from offset).
func (h *Handler) StreamLogs(ctx context.Context, opts logs.StreamOptions, sendChunk func(chunk logs.LogChunk) error) error {
	logPath := h.getLogPath(opts.Path)
	h.logger.Debug("starting log stream", "stream_id", opts.StreamID, "path", opts.Path, "mode", opts.Mode, "resolved_path", logPath)

	file, err := os.Open(logPath)
	if err != nil {
		h.logger.Error("failed to open log file", "error", err, "path", logPath)
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	if opts.Mode == logs.StreamModeHistorical {
		return h.streamHistorical(ctx, file, opts, sendChunk)
	}
	return h.streamTail(ctx, file, opts, sendChunk)
}

// streamHistorical reads a fixed number of log entries from a specific offset.
func (h *Handler) streamHistorical(ctx context.Context, file *os.File, opts logs.StreamOptions, sendChunk func(chunk logs.LogChunk) error) error {
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	fileSize := fileInfo.Size()

	// Determine start position
	startPos := opts.StartFrom
	if startPos < 0 {
		startPos = fileSize // Start from end
	}
	if startPos > fileSize {
		startPos = fileSize
	}

	// Seek to start position
	if _, err := file.Seek(startPos, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek: %w", err)
	}

	reader := bufio.NewReader(file)
	var entries []logs.LogEntry
	limit := opts.Limit
	if limit <= 0 {
		limit = 100 // Default limit
	}

	// Max chunk size in bytes (reserve space for JSON overhead)
	const maxChunkBytes = 8 * 1024 * 1024
	estimatedSize := 0

	for len(entries) < limit {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			h.logger.Error("failed to read log line", "error", err)
			break
		}

		entry := h.parseLogLine(line)
		if entry != nil {
			entrySize := len(entry.RawLine) + 200
			if estimatedSize+entrySize > maxChunkBytes && len(entries) > 0 {
				currentPos, _ := file.Seek(0, io.SeekCurrent)
				if err := sendChunk(logs.LogChunk{
					StreamID: opts.StreamID,
					Path:     opts.Path,
					Entries:  entries,
					EOF:      false,
					HasMore:  true,
					Offset:   currentPos,
				}); err != nil {
					return err
				}
				entries = nil
				estimatedSize = 0
			}
			entries = append(entries, *entry)
			estimatedSize += entrySize
		}
	}

	// Get current position
	currentPos, _ := file.Seek(0, io.SeekCurrent)
	hasMore := currentPos < fileSize

	if len(entries) > 0 || !hasMore {
		if err := sendChunk(logs.LogChunk{
			StreamID: opts.StreamID,
			Path:     opts.Path,
			Entries:  entries,
			EOF:      !hasMore,
			HasMore:  hasMore,
			Offset:   currentPos,
		}); err != nil {
			return err
		}
	}

	return nil
}

// streamTail reads from current position and follows new log entries.
// Batches chunks with 250ms OR 50 lines (whichever comes first).
func (h *Handler) streamTail(ctx context.Context, file *os.File, opts logs.StreamOptions, sendChunk func(chunk logs.LogChunk) error) error {
	startPos := opts.StartFrom
	if startPos < 0 {
		if _, err := file.Seek(0, io.SeekEnd); err != nil {
			return fmt.Errorf("failed to seek to end: %w", err)
		}
	} else {
		if _, err := file.Seek(startPos, io.SeekStart); err != nil {
			return fmt.Errorf("failed to seek: %w", err)
		}
	}

	reader := bufio.NewReader(file)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var entries []logs.LogEntry
	const batchSize = 50
	const maxChunkBytes = 8 * 1024 * 1024
	const baseBuffer = 250 * time.Millisecond
	const maxBuffer = 2 * time.Second
	estimatedSize := 0

	// Batching state
	var batchTimer *time.Timer
	var batchTimeout <-chan time.Time
	currentBuffer := baseBuffer
	backoffMultiplier := 1

	flushBatch := func(isEOF bool) error {
		if len(entries) == 0 {
			return nil
		}
		currentPos, _ := file.Seek(0, io.SeekCurrent)
		if err := sendChunk(logs.LogChunk{
			StreamID: opts.StreamID,
			Path:     opts.Path,
			Entries:  entries,
			EOF:      isEOF,
			Offset:   currentPos,
		}); err != nil {
			return err
		}
		entries = nil
		estimatedSize = 0
		backoffMultiplier++
		currentBuffer = baseBuffer * time.Duration(backoffMultiplier)
		if currentBuffer > maxBuffer {
			currentBuffer = maxBuffer
		}
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			if batchTimer != nil {
				batchTimer.Stop()
			}
			if err := flushBatch(true); err != nil {
				h.logger.Error("failed to send final chunk", "error", err)
			}
			h.logger.Debug("log stream stopped", "stream_id", opts.StreamID, "reason", "context_done")
			return ctx.Err()

		case <-ticker.C:
			// Read available lines
			for {
				line, err := reader.ReadString('\n')
				if err == io.EOF {
					break
				}
				if err != nil {
					h.logger.Error("failed to read log line", "error", err)
					break
				}

				entry := h.parseLogLine(line)
				if entry != nil {
					entrySize := len(entry.RawLine) + 200
					entries = append(entries, *entry)
					estimatedSize += entrySize

					// Start batch timer on first entry
					if len(entries) == 1 {
						batchTimer = time.NewTimer(currentBuffer)
						batchTimeout = batchTimer.C
					}

					// Flush immediately if batch is full (50 lines or size limit)
					if len(entries) >= batchSize || estimatedSize >= maxChunkBytes {
						if batchTimer != nil {
							batchTimer.Stop()
							batchTimer = nil
							batchTimeout = nil
						}
						if err := flushBatch(false); err != nil {
							h.logger.Error("failed to send chunk", "error", err)
							return err
						}
						// Reset backoff on activity
						backoffMultiplier = 1
						currentBuffer = baseBuffer
					}
				}
			}

		case <-batchTimeout:
			// Time window expired, flush batch
			batchTimer = nil
			batchTimeout = nil
			if err := flushBatch(false); err != nil {
				h.logger.Error("failed to send batch on timeout", "error", err)
				return err
			}
		}
	}
}

// parseLogLine attempts to parse a JSON log line into a LogEntry.
// Returns nil if the line is not valid JSON.
func (h *Handler) parseLogLine(line string) *logs.LogEntry {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	var entry logs.LogEntry
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		// Not valid JSON, store as raw line with fallback structure
		return &logs.LogEntry{
			Time:    time.Now().Format(time.RFC3339),
			Level:   "INFO",
			Msg:     line,
			RawLine: line,
			Attrs:   map[string]interface{}{"raw": true},
		}
	}

	entry.RawLine = line
	return &entry
}

// getLogPath returns the file path for the specified relative path under the log root.
func (h *Handler) getLogPath(path string) string {
	var baseDir string
	switch runtime.GOOS {
	case "windows":
		baseDir = filepath.Join(os.Getenv("PROGRAMDATA"), "dployr", "logs")
	case "darwin":
		baseDir = "/usr/local/var/log/dployrd"
	default:
		baseDir = "/var/log/dployrd"
	}

	// Normalize the requested path and ensure it is relative
	clean := filepath.Clean(path)
	if clean == "." || clean == "" {
		clean = "app"
	}

	// Default extension: treat "app" -> "app.log", but allow explicit .log
	if !strings.HasSuffix(clean, ".log") {
		clean = clean + ".log"
	}

	full := filepath.Join(baseDir, clean)
	// Prevent escaping the base directory
	if !strings.HasPrefix(full, baseDir+string(os.PathSeparator)) && full != baseDir {
		return filepath.Join(baseDir, "app.log")
	}

	return full
}
