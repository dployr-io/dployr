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
	"sync"
)

var (
	defaultLogger     *Logger
	defaultLoggerOnce sync.Once
)

// Logger wraps slog with multi-sink (stdout + file) and context-aware logging.
type Logger struct {
	logger *slog.Logger
}

// NewLogger creates a unified logger that writes structured JSON to both stdout and a file.
func NewLogger() *Logger {
	defaultLoggerOnce.Do(func() {
		logFile := getLogFilePath()
		if err := os.MkdirAll(filepath.Dir(logFile), 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "failed to create log dir: %v\n", err)
		}

		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open log file: %v\n", err)
			f = nil
		}

		var writers []io.Writer
		writers = append(writers, os.Stdout)
		if f != nil {
			writers = append(writers, f)
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
