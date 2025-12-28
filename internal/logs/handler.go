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
	"sync"
	"time"

	"github.com/dployr-io/dployr/pkg/core/logs"
	"github.com/dployr-io/dployr/pkg/shared"
)

var (
	streamSemaphore *shared.Semaphore
	streamSemOnce   sync.Once
)

// Handler implements the LogStreamer interface.
type Handler struct {
	logger *shared.Logger
	config *shared.Config
}

// NewHandler creates a new log stream handler.
func NewHandler(logger *shared.Logger) *Handler {
	cfg, err := shared.LoadConfig()
	if err != nil {
		logger.Warn("failed to load config, using defaults", "error", err)
		cfg = &shared.Config{
			LogMaxChunkBytes:    8 * 1024 * 1024,
			LogBatchSize:        50,
			LogBatchTimeout:     250 * time.Millisecond,
			LogMaxBatchTimeout:  2 * time.Second,
			LogPollInterval:     100 * time.Millisecond,
			LogMaxFileReadBytes: 100 * 1024 * 1024,
			LogMaxStreams:       100,
		}
	}

	streamSemOnce.Do(func() {
		streamSemaphore = shared.NewSemaphore(cfg.LogMaxStreams)
	})

	return &Handler{
		logger: logger,
		config: cfg,
	}
}

// StreamLogs streams logs based on the provided options.
// Duration controls behavior: "live" = tail from now, time duration = read history then tail.
func (h *Handler) StreamLogs(ctx context.Context, opts logs.StreamOptions, sendChunk func(chunk logs.LogChunk) error) error {
	// Rate limiting: acquire semaphore slot
	if err := streamSemaphore.Acquire(ctx); err != nil {
		return fmt.Errorf("stream rate limit: %w", err)
	}
	defer streamSemaphore.Release()

	logPath := h.getLogPath(opts.Path)
	h.logger.Debug("starting log stream", "stream_id", opts.StreamID, "path", opts.Path, "duration", opts.Duration, "resolved_path", logPath)

	var cutoffTime time.Time
	if opts.Duration != "" && opts.Duration != "live" {
		if d, err := parseDuration(opts.Duration); err == nil {
			cutoffTime = time.Now().Add(-d)
			h.logger.Debug("time-based filter enabled", "duration", opts.Duration, "cutoff", cutoffTime)
		}
	}

	return h.streamTail(ctx, logPath, opts, cutoffTime, sendChunk)
}

// streamTail reads from current position and follows new log entries with file rotation detection.
// Handles both historical reads (with time-based cutoff) and live tailing.
func (h *Handler) streamTail(ctx context.Context, logPath string, opts logs.StreamOptions, cutoffTime time.Time, sendChunk func(chunk logs.LogChunk) error) error {
	file, err := os.Open(logPath)
	if err != nil {
		h.logger.Error("failed to open log file", "error", err, "path", logPath)
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Track file identity for rotation detection
	var lastInode uint64
	if stat, err := file.Stat(); err == nil {
		if sys := stat.Sys(); sys != nil {
			if runtime.GOOS != "windows" {
				lastInode = getInode(sys)
			}
		}
	}

	// Determine start position
	if opts.StartFrom > 0 {
		// Resume from specific offset (pagination)
		if _, err := file.Seek(opts.StartFrom, io.SeekStart); err != nil {
			return fmt.Errorf("failed to seek: %w", err)
		}
	} else if opts.Duration == "live" {
		// Live mode: show last 5 minutes, fallback to end if not found
		cutoffTime = time.Now().Add(-5 * time.Minute)
	}

	// Apply time-based cutoff or default positioning
	if !cutoffTime.IsZero() {
		fileInfo, err := file.Stat()
		if err != nil {
			return fmt.Errorf("failed to stat file: %w", err)
		}

		h.logger.Debug("finding time-based cutoff position", "cutoff", cutoffTime)
		foundPos, err := h.findTimeCutoffPosition(file, fileInfo.Size(), cutoffTime)
		if err != nil {
			h.logger.Warn("failed to find cutoff position, using fallback", "error", err)
			if opts.Duration == "live" {
				file.Seek(0, io.SeekEnd) // Live mode fallback: end of file
			} else {
				file.Seek(0, io.SeekStart) // Time filter fallback: beginning
			}
		} else {
			h.logger.Debug("found cutoff position", "offset", foundPos)
			if _, err := file.Seek(foundPos, io.SeekStart); err != nil {
				return fmt.Errorf("failed to seek: %w", err)
			}
		}
	} else if opts.StartFrom == 0 {
		// Empty duration (deployment logs): start from beginning
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("failed to seek to start: %w", err)
		}
	}

	reader := bufio.NewReader(file)
	ticker := time.NewTicker(h.config.LogPollInterval)
	defer ticker.Stop()

	var entries []logs.LogEntry
	estimatedSize := int64(0)

	// Batching state
	var batchTimer *time.Timer
	var batchTimeout <-chan time.Time

	flushBatch := func(isEOF bool) error {
		if len(entries) == 0 {
			return nil
		}
		currentPos, err := file.Seek(0, io.SeekCurrent)
		if err != nil {
			h.logger.Error("failed to get file position", "error", err)
			currentPos = 0
		}
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
			// Check for file rotation
			if runtime.GOOS != "windows" && lastInode > 0 {
				if rotated, newInode := h.checkFileRotation(logPath, lastInode); rotated {
					h.logger.Info("log file rotated, reopening", "path", logPath)
					file.Close()

					newFile, err := os.Open(logPath)
					if err != nil {
						h.logger.Error("failed to reopen rotated file", "error", err)
						return fmt.Errorf("failed to reopen rotated file: %w", err)
					}
					file = newFile
					reader = bufio.NewReader(file)
					lastInode = newInode
				}
			}

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
					if !cutoffTime.IsZero() && !h.isAfterCutoff(entry, cutoffTime) {
						continue
					}

					entrySize := int64(len(entry.RawLine)) + h.config.LogEntryJSONOverhead
					entries = append(entries, *entry)
					estimatedSize += entrySize

					if len(entries) == 1 {
						batchTimer = time.NewTimer(h.config.LogBatchTimeout)
						batchTimeout = batchTimer.C
					}

					if len(entries) >= h.config.LogBatchSize || estimatedSize >= h.config.LogMaxChunkBytes {
						if batchTimer != nil {
							batchTimer.Stop()
							batchTimer = nil
							batchTimeout = nil
						}
						if err := flushBatch(false); err != nil {
							h.logger.Error("failed to send chunk", "error", err)
							return err
						}
					}
				}
			}

		case <-batchTimeout:
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

// getLogPath returns the file path for the specified relative path.
// Supported path formats:
//   - "app" or empty: system log file
//   - "service:<name>": service runtime logs (stdout/stderr)
//   - "<deployment-id>": deployment-specific log file
func (h *Handler) getLogPath(path string) string {
	clean := filepath.Clean(path)
	if clean == "." || clean == "" {
		clean = "app"
	}

	// System log (app.log)
	if clean == "app" {
		var sysLogDir string
		switch runtime.GOOS {
		case "windows":
			sysLogDir = filepath.Join(os.Getenv("PROGRAMDATA"), "dployr", "logs")
		case "darwin":
			sysLogDir = "/usr/local/var/log/dployrd"
		default:
			sysLogDir = "/var/log/dployrd"
		}
		return filepath.Join(sysLogDir, "app.log")
	}

	// Service runtime logs (service:<name>)
	if after, ok := strings.CutPrefix(clean, "service:"); ok {
		svcName := after
		if !strings.HasSuffix(svcName, ".log") {
			svcName = svcName + ".log"
		}
		switch runtime.GOOS {
		case "windows":
			return filepath.Join(os.Getenv("PROGRAMDATA"), "dployr", ".dployr", "logs", svcName)
		default:
			return filepath.Join("/home/dployrd/.dployr/logs", svcName)
		}
	}

	// Deployment logs
	var dataDir string
	switch runtime.GOOS {
	case "windows":
		dataDir = filepath.Join(os.Getenv("PROGRAMDATA"), "dployr")
	default:
		dataDir = "/var/lib/dployrd"
	}

	clean = strings.ToLower(clean)

	if !strings.HasSuffix(clean, ".log") {
		clean = clean + ".log"
	}

	return filepath.Join(dataDir, ".dployr", "logs", clean)
}

// parseDuration converts duration strings like "5m", "1h", "24h" to time.Duration.
func parseDuration(s string) (time.Duration, error) {
	switch s {
	case "5m":
		return 5 * time.Minute, nil
	case "15m":
		return 15 * time.Minute, nil
	case "30m":
		return 30 * time.Minute, nil
	case "1h":
		return 1 * time.Hour, nil
	case "3h":
		return 3 * time.Hour, nil
	case "6h":
		return 6 * time.Hour, nil
	case "12h":
		return 12 * time.Hour, nil
	case "24h":
		return 24 * time.Hour, nil
	default:
		return time.ParseDuration(s)
	}
}

// isAfterCutoff checks if a log entry's timestamp is after the cutoff time.
func (h *Handler) isAfterCutoff(entry *logs.LogEntry, cutoff time.Time) bool {
	if entry.Time == "" {
		return false
	}

	entryTime, err := time.Parse(time.RFC3339, entry.Time)
	if err != nil {
		entryTime, err = time.Parse(time.RFC3339Nano, entry.Time)
		if err != nil {
			return false
		}
	}

	return entryTime.After(cutoff) || entryTime.Equal(cutoff)
}

// findTimeCutoffPosition uses binary search to find the approximate file position
// where logs start matching the cutoff time.
func (h *Handler) findTimeCutoffPosition(file *os.File, fileSize int64, cutoff time.Time) (int64, error) {
	if fileSize == 0 {
		return 0, nil
	}

	// Sample from the end to estimate log density
	sampleSize := min(int64(64*1024), fileSize)

	if _, err := file.Seek(fileSize-sampleSize, io.SeekStart); err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(file)
	var sampleEntries []struct {
		offset int64
		time   time.Time
	}

	currentOffset := fileSize - sampleSize
	for scanner.Scan() && len(sampleEntries) < 100 {
		line := scanner.Text()
		entry := h.parseLogLine(line)
		if entry != nil && entry.Time != "" {
			if t, err := time.Parse(time.RFC3339, entry.Time); err == nil {
				sampleEntries = append(sampleEntries, struct {
					offset int64
					time   time.Time
				}{currentOffset, t})
			} else if t, err := time.Parse(time.RFC3339Nano, entry.Time); err == nil {
				sampleEntries = append(sampleEntries, struct {
					offset int64
					time   time.Time
				}{currentOffset, t})
			}
		}
		currentOffset += int64(len(line) + 1)
	}

	if len(sampleEntries) < 2 {
		return 0, fmt.Errorf("insufficient sample data")
	}

	// Find first entry after cutoff in sample
	for _, entry := range sampleEntries {
		if entry.time.After(cutoff) || entry.time.Equal(cutoff) {
			safeOffset := max(entry.offset-(10*1024), 0) // 10KB safety margin
			return safeOffset, nil
		}
	}

	return 0, nil
}

// checkFileRotation detects if a log file has been rotated by comparing inodes.
func (h *Handler) checkFileRotation(path string, lastInode uint64) (bool, uint64) {
	stat, err := os.Stat(path)
	if err != nil {
		return false, lastInode
	}

	if sys := stat.Sys(); sys != nil {
		currentInode := getInode(sys)
		if currentInode != lastInode {
			return true, currentInode
		}
	}

	return false, lastInode
}
