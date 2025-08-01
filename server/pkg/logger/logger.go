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
	deploymentID string
	wsClient     platform.WebSocketClient
}

func New(deploymentID string, ws platform.WebSocketClient) *Logger {
	return &Logger{deploymentID, ws}
}

func (l *Logger) Info(ctx context.Context, r *repository.EventRepo, phase, message string) {
	l.log(ctx, r, "info", phase, message)
}

func (l *Logger) Warn(ctx context.Context, r *repository.EventRepo, phase, message string) {
	l.log(ctx, r, "warn", phase, message)
}

func (l *Logger) Error(ctx context.Context, r *repository.EventRepo, phase, message string) {
	l.log(ctx, r, "error", phase, message)
}

func (l *Logger) log(ctx context.Context, r *repository.EventRepo, level, phase, message string) {
	entry := platform.LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Phase:     phase,
		Message:   message,
	}

	l.wsClient.StreamLog(l.deploymentID, entry)

	entryJSON, _ := json.Marshal(entry)
	metadataJSON, _ := json.Marshal(models.Metadata{})

	event := models.Event{
		ID:          uuid.New().String(),
		Type:        models.LogEvent,
		AggregateID: l.deploymentID,
		Data:        entryJSON,
		Metadata:    metadataJSON,
		Version:     1,
	}
	r.Create(ctx, &event)
}
