package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
)

type LogEntry struct {
	Id 		string    `json:"id"`
	Level   string    `json:"level"`
	Message string    `json:"message"`
	Time    time.Time `json:"timestamp"`
	Error   error    `json:"error,omitempty"`
}

func NewLogger() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return slog.New(handler)
}

func LogWithContext(ctx context.Context) *slog.Logger {
	requestID, _ := RequestFromContext(ctx)
	traceID, _ := TraceFromContext(ctx)
	user, _ := UserFromContext(ctx)

	attrs := []slog.Attr{
		slog.String("request_id", safeString(&requestID)),
		slog.String("trace_id", safeString(&traceID)),
		slog.String("user_id", safeString(&user.ID)),
	}

	args := make([]any, len(attrs))
	for i, a := range attrs {
		args[i] = a
	}

	return slog.With(args...)
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// LogF creates a log file and writes a LogEntry
func LogF(name, dir string, entry LogEntry) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logFile := filepath.Join(dir, fmt.Sprintf("%s.log", strings.ToLower(name)))
	file, openErr := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if openErr != nil {
		return fmt.Errorf("failed to create log file: %w", openErr)
	}
	defer file.Close()

	if entry.Time.IsZero() {
		entry.Time = time.Now()
	}

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(file, "[%s] %s: %s", entry.Level, entry.Time.Format(time.RFC3339), entry.Message)
		if entry.Error != nil {
			fmt.Fprintf(file, "%s", entry.Error)
		}
		fmt.Fprintln(file)
		return nil
	}

	fmt.Fprintln(file, string(data))
	return nil
}

// LogInfoF logs a info to file
func LogInfoF(name, dir, message string) error {
	entry := LogEntry{
		Id: 	ulid.Make().String(),
		Level:   "info",
		Message: message,
		Time:    time.Now(),
	}
	return LogF(name, dir, entry)
}

// LogInfoF logs a warning to file
func LogWarnF(name, dir, message string) error {
	entry := LogEntry{
		Id: 	ulid.Make().String(),
		Level:   "warn",
		Message: message,
		Time:    time.Now(),
	}
	return LogF(name, dir, entry)
}

// LogErrF logs an error to file
func LogErrF(name, dir string, err error) error {
	entry := LogEntry{
		Id: 	ulid.Make().String(),
		Level:   "error",
        Error: err,
		Time:    time.Now(),
	}
	return LogF(name, dir, entry)
}
