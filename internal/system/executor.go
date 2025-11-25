package system

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/tasks"

	"github.com/oklog/ulid/v2"
)

// Executor runs tasks by routing them through existing HTTP handlers.
type Executor struct {
	logger  *shared.Logger
	handler http.Handler
}

var pendingTasks int64

// NewExecutor creates a task executor that uses the web server's routes.
func NewExecutor(logger *shared.Logger, handler http.Handler) *Executor {
	return &Executor{
		logger:  logger,
		handler: handler,
	}
}

// Execute runs a task by converting it to an HTTP request and routing it internally.
func (e *Executor) Execute(ctx context.Context, task *tasks.Task) *tasks.Result {
	start := time.Now()
	atomic.AddInt64(&pendingTasks, 1)
	defer atomic.AddInt64(&pendingTasks, -1)

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
	ctx = shared.WithRequest(ctx, task.ID)
	ctx = shared.WithTrace(ctx, ulid.Make().String())
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

	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	logger.Debug("routing task to handler", "method", method, "path", path)
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

		return &tasks.Result{
			ID:     task.ID,
			Status: "done",
			Result: result,
		}
	}

	logger.Error("task failed", "status_code", resp.StatusCode, "duration_ms", duration.Milliseconds(), "response_bytes", len(respBody), "response", string(respBody))
	return &tasks.Result{
		ID:     task.ID,
		Status: "failed",
		Error:  fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)),
	}
}

func currentPendingTasks() int {
	return int(atomic.LoadInt64(&pendingTasks))
}
