// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/dployr-io/dployr/internal/logs"
	pkgAuth "github.com/dployr-io/dployr/pkg/auth"
	corelogs "github.com/dployr-io/dployr/pkg/core/logs"
	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/tasks"
)

// AccessTokenProvider provides the current agent access token.
type AccessTokenProvider interface {
	GetAccessToken(ctx context.Context) (string, error)
}

// Executor runs tasks by routing them through existing HTTP handlers.
type Executor struct {
	logger   *shared.Logger
	handler  http.Handler
	tokens   AccessTokenProvider
	auth     pkgAuth.Authenticator
	wsConn   *websocket.Conn
	wsConnMu sync.RWMutex
}

var pendingTasks int64
var pendingSystemTasks int64 // tracks system/* tasks separately

type lastExecInfo struct {
	ID     string
	Status string
	DurMs  int64
	At     time.Time
}

var lastExec atomic.Value // stores lastExecInfo

var taskExecutedSuccessTotal uint64
var taskExecutedFailedTotal uint64

var taskExecBuckets = []float64{0.01, 0.05, 0.1, 0.25}
var taskExecCounts [5]uint64
var taskExecSum float64
var taskExecCount uint64
var taskExecMu sync.Mutex

// NewExecutor creates a task executor that uses the web server's routes.
func NewExecutor(logger *shared.Logger, handler http.Handler, tokens AccessTokenProvider, auth pkgAuth.Authenticator) *Executor {
	return &Executor{
		logger:  logger,
		handler: handler,
		tokens:  tokens,
		auth:    auth,
	}
}

// SetWSConn sets the WebSocket connection for log streaming.
func (e *Executor) SetWSConn(conn *websocket.Conn) {
	e.wsConnMu.Lock()
	defer e.wsConnMu.Unlock()
	e.wsConn = conn
}

// sendLogChunkToBase sends a log chunk to the base via WebSocket.
func (e *Executor) sendLogChunkToBase(ctx context.Context, chunk corelogs.LogChunk) error {
	e.wsConnMu.RLock()
	conn := e.wsConn
	e.wsConnMu.RUnlock()

	if conn == nil {
		return fmt.Errorf("no websocket connection")
	}

	msg := map[string]interface{}{
		"kind":     "log_chunk",
		"streamId": chunk.StreamID,
		"path":     chunk.Path,
		"entries":  chunk.Entries,
		"eof":      chunk.EOF,
		"hasMore":  chunk.HasMore,
		"offset":   chunk.Offset,
	}

	return wsjson.Write(ctx, conn, msg)
}

// handleLogStream handles the logs/stream:post task type.
func (e *Executor) handleLogStream(ctx context.Context, task *tasks.Task) *tasks.Result {
	var payload struct {
		Token     string `json:"token"`
		Path      string `json:"path"`
		StreamID  string `json:"streamId"`
		StartFrom int64  `json:"startFrom,omitempty"` // Byte offset for pagination
		Limit     int    `json:"limit,omitempty"`     // Max entries per chunk
		Duration  string `json:"duration,omitempty"`  // "live" or time duration ("5m", "1h", "24h")
	}

	if err := json.Unmarshal(task.Payload, &payload); err != nil {
		return &tasks.Result{
			ID:     task.ID,
			Status: "failed",
			Error:  fmt.Sprintf("invalid payload: %v", err),
		}
	}

	// Default duration to "live" if not specified
	duration := payload.Duration
	if duration == "" {
		duration = "live"
	}

	// Validate token before starting stream
	if strings.TrimSpace(payload.Token) == "" {
		return &tasks.Result{
			ID:     task.ID,
			Status: "failed",
			Error:  "missing token",
		}
	}
	if e.auth != nil {
		if _, err := e.auth.ValidateToken(ctx, strings.TrimSpace(payload.Token)); err != nil {
			e.logger.Error("log streaming token validation failed", "error", err)
			return &tasks.Result{
				ID:     task.ID,
				Status: "failed",
				Error:  "invalid token",
			}
		}
	}

	e.logger.Info("starting log stream", "stream_id", payload.StreamID, "path", payload.Path, "duration", duration, "start_from", payload.StartFrom)

	// Start streaming in background
	go func() {
		streamCtx := ctx
		logHandler := logs.NewHandler(e.logger)

		opts := corelogs.StreamOptions{
			StreamID:  payload.StreamID,
			Path:      payload.Path,
			StartFrom: payload.StartFrom,
			Limit:     payload.Limit,
			Duration:  duration,
		}

		err := logHandler.StreamLogs(streamCtx, opts, func(chunk corelogs.LogChunk) error {
			return e.sendLogChunkToBase(streamCtx, chunk)
		})

		if err != nil && !errors.Is(err, context.Canceled) {
			e.logger.Error("log streaming failed", "error", err, "stream_id", payload.StreamID)
		}
	}()

	return &tasks.Result{
		ID:     task.ID,
		Status: "done",
		Result: map[string]interface{}{
			"message":  "log streaming started",
			"duration": duration,
		},
	}
}

// Execute runs a task by converting it to an HTTP request and routing it internally.
func (e *Executor) Execute(ctx context.Context, task *tasks.Task) *tasks.Result {
	start := time.Now()
	atomic.AddInt64(&pendingTasks, 1)
	defer atomic.AddInt64(&pendingTasks, -1)

	// Track system/* tasks separately so install script can exclude them
	isSystemTask := strings.HasPrefix(task.Type, "system/")
	if isSystemTask {
		atomic.AddInt64(&pendingSystemTasks, 1)
		defer atomic.AddInt64(&pendingSystemTasks, -1)
	}

	// Handle logs/stream:post specially
	if task.Type == "logs/stream:post" {
		return e.handleLogStream(ctx, task)
	}

	parts := strings.SplitN(task.Type, ":", 2)
	if len(parts) != 2 {
		return &tasks.Result{
			ID:     task.ID,
			Status: "failed",
			Error:  fmt.Sprintf("invalid task type format: %s", task.Type),
		}
	}

	path := "/" + parts[0]
	method := strings.ToUpper(parts[1])
	logger := e.logger.WithContext(ctx)
	logger.Info("executing task", "type", task.Type, "method", method, "path", path, "payload_bytes", len(task.Payload))

	var body io.Reader
	if len(task.Payload) > 0 {
		body = bytes.NewReader(task.Payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, path, body)
	if err != nil {
		logger.Error("failed to create request", "error", err)
		return &tasks.Result{
			ID:     task.ID,
			Status: "failed",
			Error:  fmt.Sprintf("failed to create request: %v", err),
		}
	}

	logger.Info("raw task payload", "payload", string(task.Payload))

	req.Header.Set("Content-Type", "application/json")
	// Authorization: extract token from payload;
	var bearer string
	if len(task.Payload) > 0 {
		var p map[string]any
		if err := json.Unmarshal(task.Payload, &p); err != nil {
			logger.Error("failed to unmarshal payload for auth", "error", err)
		} else {
			logger.Info("parsed payload map", "payload_map", p)
			if t, ok := p["token"].(string); ok && strings.TrimSpace(t) != "" {
				bearer = strings.TrimSpace(t)
			}
		}
	}
	if bearer != "" {
		logger.Info("setting Authorization header", "bearer", bearer)
		req.Header.Set("Authorization", "Bearer "+bearer)
	} else {
		logger.Info("no bearer token found in payload")
	}
	rr := httptest.NewRecorder()
	logger.Debug("routing task", "task_id", task.ID, "method", method, "path", path)
	e.handler.ServeHTTP(rr, req)
	resp := rr.Result()
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	duration := time.Since(start)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logger.Info("task completed", "status_code", resp.StatusCode, "duration_ms", duration.Milliseconds(), "response_bytes", len(respBody))

		var result any
		if len(respBody) > 0 {
			json.Unmarshal(respBody, &result)
		}

		res := &tasks.Result{
			ID:     task.ID,
			Status: "done",
			Result: result,
		}
		lastExec.Store(lastExecInfo{ID: task.ID, Status: "done", DurMs: duration.Milliseconds(), At: time.Now()})
		atomic.AddUint64(&taskExecutedSuccessTotal, 1)
		recordTaskExecHistogram(duration)
		return res
	}

	logger.Error("task failed", "status_code", resp.StatusCode, "duration_ms", duration.Milliseconds(), "response_bytes", len(respBody), "response", string(respBody))
	res := &tasks.Result{
		ID:     task.ID,
		Status: "failed",
		Error:  fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)),
	}
	lastExec.Store(lastExecInfo{ID: task.ID, Status: "failed", DurMs: duration.Milliseconds(), At: time.Now()})
	atomic.AddUint64(&taskExecutedFailedTotal, 1)
	recordTaskExecHistogram(duration)
	return res
}

func currentPendingTasks() int {
	return int(atomic.LoadInt64(&pendingTasks))
}

// currentPendingTasksExcludingSystem returns pending task count excluding system/* tasks.
// This is used by the install script to avoid waiting for itself.
func currentPendingTasksExcludingSystem() int {
	total := atomic.LoadInt64(&pendingTasks)
	system := atomic.LoadInt64(&pendingSystemTasks)
	count := total - system
	if count < 0 {
		return 0
	}
	return int(count)
}

func getLastExec() (info *lastExecInfo) {
	if v := lastExec.Load(); v != nil {
		le := v.(lastExecInfo)
		return &le
	}
	return nil
}

func taskExecutionTotals() (success, failed uint64) {
	return atomic.LoadUint64(&taskExecutedSuccessTotal), atomic.LoadUint64(&taskExecutedFailedTotal)
}

type taskExecHistogram struct {
	Buckets []float64
	Counts  []uint64
	Sum     float64
	Count   uint64
}

func recordTaskExecHistogram(d time.Duration) {
	sec := d.Seconds()
	idx := len(taskExecBuckets)
	for i, b := range taskExecBuckets {
		if sec <= b {
			idx = i
			break
		}
	}
	taskExecMu.Lock()
	defer taskExecMu.Unlock()
	taskExecCounts[idx]++
	taskExecSum += sec
	taskExecCount++
}

func taskExecHistogramSnapshot() taskExecHistogram {
	taskExecMu.Lock()
	defer taskExecMu.Unlock()
	counts := make([]uint64, len(taskExecCounts))
	copy(counts, taskExecCounts[:])
	return taskExecHistogram{
		Buckets: append([]float64(nil), taskExecBuckets...),
		Counts:  counts,
		Sum:     taskExecSum,
		Count:   taskExecCount,
	}
}
