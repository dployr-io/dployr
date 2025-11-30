// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/dployr-io/dployr/pkg/core/logs"
	"github.com/dployr-io/dployr/pkg/shared"
)

// WSHandler handles WebSocket connections for log streaming.
type WSHandler struct {
	handler *Handler
	logger  *shared.Logger
	mu      sync.RWMutex
	streams map[string]context.CancelFunc // streamID -> cancel function
}

// NewWSHandler creates a new WebSocket handler for log streaming.
func NewWSHandler(logger *shared.Logger) *WSHandler {
	return &WSHandler{
		handler: NewHandler(logger),
		logger:  logger,
		streams: make(map[string]context.CancelFunc),
	}
}

// HandleWebSocket handles incoming WebSocket connections for log streaming.
func (h *WSHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := h.logger.WithContext(ctx)

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		CompressionMode: websocket.CompressionContextTakeover,
	})
	if err != nil {
		logger.Error("failed to accept websocket", "error", err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	logger.Info("log websocket connected")

	connCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Send initial ready message
	if err := wsjson.Write(connCtx, conn, map[string]any{
		"kind":    "ready",
		"message": "log stream ready",
	}); err != nil {
		logger.Error("failed to send ready message", "error", err)
		return
	}

	// Handle incoming messages
	for {
		var msg map[string]any
		if err := wsjson.Read(connCtx, conn, &msg); err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				logger.Info("log websocket closed normally")
				return
			}
			logger.Error("failed to read websocket message", "error", err)
			return
		}

		kind, ok := msg["kind"].(string)
		if !ok {
			h.sendError(connCtx, conn, "invalid message: missing 'kind' field")
			continue
		}

		switch kind {
		case "log_subscribe":
			h.handleStart(connCtx, conn, msg)
		case "log_unsubscribe":
			h.handleStop(connCtx, conn, msg)
		case "ping":
			h.handlePing(connCtx, conn)
		default:
			h.sendError(connCtx, conn, fmt.Sprintf("unknown message kind: %s", kind))
		}
	}
}

// handleStart starts a new log stream.
func (h *WSHandler) handleStart(ctx context.Context, conn *websocket.Conn, msg map[string]any) {
	streamID, _ := msg["streamId"].(string)
	logType, _ := msg["logType"].(string)
	mode, _ := msg["mode"].(string)
	startFrom, _ := msg["startFrom"].(float64)
	limit, _ := msg["limit"].(float64)

	if streamID == "" {
		h.sendError(ctx, conn, "missing streamId")
		return
	}
	if logType == "" {
		h.sendError(ctx, conn, "missing logType")
		return
	}

	// Default to tail mode
	streamMode := logs.StreamModeTail
	if mode == "historical" {
		streamMode = logs.StreamModeHistorical
	}

	// Default startFrom to -1 for tailing from end
	startFromInt := int64(startFrom)
	if startFromInt == 0 && streamMode == logs.StreamModeTail {
		startFromInt = -1
	}

	h.logger.Info("starting log stream", "stream_id", streamID, "log_type", logType, "mode", streamMode)

	// Check if stream already exists
	h.mu.Lock()
	if cancel, exists := h.streams[streamID]; exists {
		cancel() // Stop existing stream
		delete(h.streams, streamID)
	}

	streamCtx, streamCancel := context.WithCancel(ctx)
	h.streams[streamID] = streamCancel
	h.mu.Unlock()

	// Send acknowledgment
	if err := wsjson.Write(ctx, conn, map[string]any{
		"kind":     "started",
		"streamId": streamID,
		"mode":     string(streamMode),
	}); err != nil {
		h.logger.Error("failed to send started message", "error", err)
		streamCancel()
		return
	}

	// Start streaming in background
	go func() {
		defer func() {
			h.mu.Lock()
			delete(h.streams, streamID)
			h.mu.Unlock()
		}()

		opts := logs.StreamOptions{
			StreamID:  streamID,
			LogType:   logType,
			Mode:      streamMode,
			StartFrom: startFromInt,
			Limit:     int(limit),
		}

		err := h.handler.StreamLogs(streamCtx, opts, func(chunk logs.LogChunk) error {
			return h.sendChunk(streamCtx, conn, chunk)
		})

		if err != nil && err != context.Canceled {
			h.logger.Error("log streaming failed", "error", err, "stream_id", streamID)
			h.sendError(streamCtx, conn, fmt.Sprintf("stream error: %v", err))
		}
	}()
}

// handleStop stops an active log stream.
func (h *WSHandler) handleStop(ctx context.Context, conn *websocket.Conn, msg map[string]any) {
	streamID, _ := msg["streamId"].(string)
	if streamID == "" {
		h.sendError(ctx, conn, "missing streamId")
		return
	}

	h.mu.Lock()
	cancel, exists := h.streams[streamID]
	if exists {
		cancel()
		delete(h.streams, streamID)
	}
	h.mu.Unlock()

	if exists {
		h.logger.Info("stopped log stream", "stream_id", streamID)
		wsjson.Write(ctx, conn, map[string]any{
			"kind":     "stopped",
			"streamId": streamID,
		})
	} else {
		h.sendError(ctx, conn, fmt.Sprintf("stream not found: %s", streamID))
	}
}

// handlePing responds to ping messages.
func (h *WSHandler) handlePing(ctx context.Context, conn *websocket.Conn) {
	wsjson.Write(ctx, conn, map[string]any{
		"kind": "pong",
		"time": time.Now().Unix(),
	})
}

// sendChunk sends a log chunk to the client.
func (h *WSHandler) sendChunk(ctx context.Context, conn *websocket.Conn, chunk logs.LogChunk) error {
	msg := map[string]any{
		"kind":     "log_chunk",
		"streamId": chunk.StreamID,
		"logType":  chunk.LogType,
		"entries":  chunk.Entries,
		"eof":      chunk.EOF,
		"hasMore":  chunk.HasMore,
		"offset":   chunk.Offset,
	}
	return wsjson.Write(ctx, conn, msg)
}

// sendError sends an error message to the client.
func (h *WSHandler) sendError(ctx context.Context, conn *websocket.Conn, message string) {
	wsjson.Write(ctx, conn, map[string]any{
		"kind":    "error",
		"message": message,
	})
}
