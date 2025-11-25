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
	"sync"
	"sync/atomic"
	"time"

	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/tasks"

	"github.com/oklog/ulid/v2"
)

// AccessTokenProvider provides the current agent access token.
type AccessTokenProvider interface {
	GetAccessToken(ctx context.Context) (string, error)
}

// Executor runs tasks by routing them through existing HTTP handlers.
type Executor struct {
	logger  *shared.Logger
	handler http.Handler
	tokens  AccessTokenProvider
}

var pendingTasks int64

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
func NewExecutor(logger *shared.Logger, handler http.Handler, tokens AccessTokenProvider) *Executor {
	return &Executor{
		logger:  logger,
		handler: handler,
		tokens:  tokens,
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
	// Authorization: extract token from payload;
	var bearer string
	if len(task.Payload) > 0 {
		var p map[string]any
		if json.Unmarshal(task.Payload, &p) == nil {
			if t, ok := p["token"].(string); ok && strings.TrimSpace(t) != "" {
				bearer = strings.TrimSpace(t)
			}
		}
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
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
