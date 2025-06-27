package logger

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"dployr.io/pkg/api/platform"
	"dployr.io/pkg/models"
	"dployr.io/pkg/repository"
)

type Logger struct {
	id         string
	sseManager *platform.SSEManager
}

func NewLogger(id string, m *platform.SSEManager) *Logger {
	return &Logger{
		id:         id,
		sseManager: m,
	}
}

func (l *Logger) Info(ctx context.Context, r *repository.Event, phase, message string) {
	l.log(ctx, r, "info", phase, message)
}

func (l *Logger) Warn(ctx context.Context, r *repository.Event, phase, message string) {
	l.log(ctx, r, "warn", phase, message)
}

func (l *Logger) Error(ctx context.Context, r *repository.Event, phase, message string) {
	l.log(ctx, r, "error", phase, message)
}

func (l *Logger) log(ctx context.Context, r *repository.Event, level, phase, message string) {
	logData := models.LogData{
		Level:   level,
		Message: message,
	}

	dataJSON, _ := json.Marshal(logData)
	metadataJSON, _ := json.Marshal(models.Metadata{
		CLIVersion: "0.1.0",
		UserAgent:  "dployr/0.1.0",
	})

	entry := models.Event{
		Timestamp:   time.Now(),
		Type:        models.LogEvent,
		AggregateID: l.id,
		Data:        dataJSON,
		Metadata:    metadataJSON,
		Version:     1,
	}

	// Send to SSE clients listening to this build/deployment
	formattedMessage := l.formatLogMessage(level, phase, message)
	l.sseManager.SendToBuild(l.id, formattedMessage)

	// Store in database
	entryJSON, _ := json.Marshal(entry)

	r.Create(ctx, &models.Event{
		ID:          uuid.New().String(),
		Type:        models.LogEvent,
		AggregateID: l.id,
		Data:        entryJSON,
		Metadata:    metadataJSON,
		Version:     1,
	})
}

// formatLogMessage formats the log message for SSE streaming
func (l *Logger) formatLogMessage(level, phase, message string) string {
	timestamp := time.Now().Format("15:04:05")

	// Create a structured log message
	logEntry := map[string]interface{}{
		"id":        l.id,
		"timestamp": timestamp,
		"level":     level,
		"phase":     phase,
		"message":   message,
	}

	jsonBytes, _ := json.Marshal(logEntry)
	return string(jsonBytes)
}

// SendToClient sends a log message to a specific client
func (l *Logger) SendToClient(clientID, userID, message string) {
	l.sseManager.SendToClient(clientID, userID, message)
}
