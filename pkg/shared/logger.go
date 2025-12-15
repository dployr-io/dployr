// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package shared

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	maxLogSize    = 300 * 1024 * 1024 // 300MB
	maxLogAge     = 24 * time.Hour    // 24 hours
	maxLogBackups = 5                 // 5 backups
)

var (
	defaultLogger     *Logger
	defaultLoggerOnce sync.Once
)

// Logger wraps slog with multi-sink (stdout + file) and context-aware logging.
type Logger struct {
	logger *slog.Logger
}

// rotatingWriter implements io.Writer with size and time-based rotation.
type rotatingWriter struct {
	mu        sync.Mutex
	filePath  string
	file      *os.File
	size      int64
	createdAt time.Time
}

// newRotatingWriter creates a new rotating file writer.
func newRotatingWriter(filePath string) (*rotatingWriter, error) {
	rw := &rotatingWriter{filePath: filePath}
	if err := rw.openOrCreate(); err != nil {
		return nil, err
	}
	return rw, nil
}

// openOrCreate opens an existing log file or creates a new one.
func (rw *rotatingWriter) openOrCreate() error {
	if err := os.MkdirAll(filepath.Dir(rw.filePath), 0o755); err != nil {
		return err
	}

	info, err := os.Stat(rw.filePath)
	if err == nil {
		rw.size = info.Size()
		rw.createdAt = info.ModTime()
		// Check if existing file needs rotation on startup
		if rw.size >= maxLogSize || time.Since(rw.createdAt) >= maxLogAge {
			if err := rw.rotate(); err != nil {
				return err
			}
		}
	} else if os.IsNotExist(err) {
		rw.size = 0
		rw.createdAt = time.Now()
	} else {
		return err
	}

	f, err := os.OpenFile(rw.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	rw.file = f
	return nil
}

// Write implements io.Writer with rotation checks.
func (rw *rotatingWriter) Write(p []byte) (n int, err error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	// Check if rotation is needed
	if rw.size+int64(len(p)) >= maxLogSize || time.Since(rw.createdAt) >= maxLogAge {
		if err := rw.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = rw.file.Write(p)
	rw.size += int64(n)
	return n, err
}

// rotate closes the current file, renames it with a timestamp, and opens a new one.
func (rw *rotatingWriter) rotate() error {
	if rw.file != nil {
		rw.file.Close()
	}

	// Generate rotated filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	ext := filepath.Ext(rw.filePath)
	base := strings.TrimSuffix(rw.filePath, ext)
	rotatedPath := fmt.Sprintf("%s.%s%s", base, timestamp, ext)

	if err := os.Rename(rw.filePath, rotatedPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	rw.cleanOldBackups()

	f, err := os.OpenFile(rw.filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	rw.file = f
	rw.size = 0
	rw.createdAt = time.Now()
	return nil
}

// cleanOldBackups removes old rotated log files beyond maxLogBackups.
func (rw *rotatingWriter) cleanOldBackups() {
	dir := filepath.Dir(rw.filePath)
	ext := filepath.Ext(rw.filePath)
	base := filepath.Base(strings.TrimSuffix(rw.filePath, ext))
	pattern := base + ".*" + ext

	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil || len(matches) <= maxLogBackups {
		return
	}

	sort.Sort(sort.Reverse(sort.StringSlice(matches)))
	for _, match := range matches[maxLogBackups:] {
		os.Remove(match)
	}
}

// Close closes the underlying file.
func (rw *rotatingWriter) Close() error {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	if rw.file != nil {
		return rw.file.Close()
	}
	return nil
}

// NewLogger creates a unified logger that writes structured JSON to both stdout and a file.
// Log files are rotated every 24 hours or when they exceed 300MB, whichever comes first.
func NewLogger() *Logger {
	defaultLoggerOnce.Do(func() {
		logFile := getLogFilePath()

		rw, err := newRotatingWriter(logFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create rotating log writer: %v\n", err)
			rw = nil
		}

		var writers []io.Writer
		writers = append(writers, os.Stdout)
		if rw != nil {
			writers = append(writers, rw)
		}

		multi := io.MultiWriter(writers...)
		handler := slog.NewJSONHandler(multi, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})

		defaultLogger = &Logger{
			logger: slog.New(handler),
		}
	})
	return defaultLogger
}

func getLogFilePath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("PROGRAMDATA"), "dployr", "logs", "app.log")
	case "darwin":
		return "/usr/local/var/log/dployrd/app.log"
	default:
		return "/var/log/dployrd/app.log"
	}
}

// WithContext returns a logger enriched with context fields (request_id, trace_id, user_id, instance_id).
func (l *Logger) WithContext(ctx context.Context) *Logger {
	requestID := RequestID(ctx)
	traceID := TraceID(ctx)
	user, _ := UserFromContext(ctx)
	userID := ""
	if user != nil {
		userID = user.ID
	}

	instanceID := ""
	if v := ctx.Value("instance_id"); v != nil {
		if id, ok := v.(string); ok {
			instanceID = id
		}
	}

	attrs := []any{
		slog.String("request_id", requestID),
		slog.String("trace_id", traceID),
		slog.String("user_id", userID),
	}
	if instanceID != "" {
		attrs = append(attrs, slog.String("instance_id", instanceID))
	}

	return &Logger{logger: l.logger.With(attrs...)}
}

// Debug logs a debug-level message.
func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

// Info logs an info-level message.
func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

// Warn logs a warning-level message.
func (l *Logger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

// Error logs an error-level message.
func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

// With returns a logger with additional fields.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{logger: l.logger.With(args...)}
}

// LogWithContext is a legacy helper; use logger.WithContext(ctx) instead.
// Kept for backward compatibility during migration.
func LogWithContext(ctx context.Context) *Logger {
	return NewLogger().WithContext(ctx)
}

// File-based logging helpers for deployment-specific logs
// These write to separate files per deployment/task

func LogInfoF(name, dir, message string) error {
	logger := NewLogger().With("deployment_id", name, "log_type", "deployment")
	logger.Info(message)
	return nil
}

func LogWarnF(name, dir, message string) error {
	logger := NewLogger().With("deployment_id", name, "log_type", "deployment")
	logger.Warn(message)
	return nil
}

func LogErrF(name, dir string, err error) error {
	logger := NewLogger().With("deployment_id", name, "log_type", "deployment")
	logger.Error(err.Error(), "error", err)
	return nil
}
