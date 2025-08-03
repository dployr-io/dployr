// server/pkg/queue/queue_test.go
package queue

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewQueue(t *testing.T) {
	handler := CreateHandler()
	q := NewQueue(2, time.Second, handler)
	defer q.Stop()

	assert.NotNil(t, q)
	assert.Equal(t, 2, q.workers)
	assert.Equal(t, time.Second, q.debounceDur)
}

func TestQueue_Enqueue(t *testing.T) {
	var processedTasks []string
	var mu sync.Mutex

	handler := func(ctx context.Context, t *Task) error {
		mu.Lock()
		defer mu.Unlock()
		processedTasks = append(processedTasks, t.Id)
		return nil
	}

	q := NewQueue(1, 100*time.Millisecond, handler)
	defer q.Stop()

	// Enqueue a task
	task := &Task{
		Id:         "task-1",
		Payload:    "test payload",
		MaxRetries: 3,
	}
	q.Enqueue(task)

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	assert.Contains(t, processedTasks, "task-1")
}

func TestQueue_Debounce(t *testing.T) {
	var processedTasks []string
	var mu sync.Mutex

	handler := func(ctx context.Context, t *Task) error {
		mu.Lock()
		defer mu.Unlock()
		processedTasks = append(processedTasks, t.Id)
		return nil
	}

	q := NewQueue(1, 100*time.Millisecond, handler)
	defer q.Stop()

	// Enqueue the same task multiple times quickly
	task := &Task{
		Id:         "debounced-task",
		Payload:    "test payload",
		MaxRetries: 3,
	}

	for i := 0; i < 5; i++ {
		q.Enqueue(task)
		time.Sleep(10 * time.Millisecond) // Less than debounce duration
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	// Should only process once due to debouncing
	count := 0
	for _, id := range processedTasks {
		if id == "debounced-task" {
			count++
		}
	}
	assert.Equal(t, 1, count)
}

func TestQueue_Retry(t *testing.T) {
	var attempts int
	var mu sync.Mutex

	handler := func(ctx context.Context, t *Task) error {
		mu.Lock()
		defer mu.Unlock()
		attempts++
		if attempts < 3 {
			return fmt.Errorf("simulated failure")
		}
		return nil
	}

	q := NewQueue(1, 10*time.Millisecond, handler)
	defer q.Stop()

	task := &Task{
		Id:         "retry-task",
		Payload:    "test payload",
		MaxRetries: 3,
	}
	q.Enqueue(task)

	// Wait for retries
	time.Sleep(4 * time.Second)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 3, attempts)
}

func TestQueue_MaxRetries(t *testing.T) {
	var attempts int
	var mu sync.Mutex

	handler := func(ctx context.Context, t *Task) error {
		mu.Lock()
		defer mu.Unlock()
		attempts++
		return fmt.Errorf("always fails")
	}

	q := NewQueue(1, 10*time.Millisecond, handler)
	defer q.Stop()

	task := &Task{
		Id:         "failing-task",
		Payload:    "test payload",
		MaxRetries: 2,
	}
	q.Enqueue(task)

	// Wait for all retries
	time.Sleep(4 * time.Second)

	mu.Lock()
	defer mu.Unlock()
	// Should attempt MaxRetries + 1 times (initial + retries)
	assert.Equal(t, 3, attempts)
}

func TestQueue_Stop(t *testing.T) {
	handler := CreateHandler()
	q := NewQueue(2, time.Second, handler)

	// Stop should not block
	done := make(chan bool)
	go func() {
		q.Stop()
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Stop() did not complete within timeout")
	}
}

func TestCreateHandler(t *testing.T) {
	handler := CreateHandler()
	assert.NotNil(t, handler)

	// Test handler behavior
	task := &Task{
		Id:       "test-task",
		Payload:  "test",
		Attempts: 0,
	}

	// First two calls should fail
	err := handler(context.Background(), task)
	assert.Error(t, err)

	task.Attempts = 1
	err = handler(context.Background(), task)
	assert.Error(t, err)

	// Third call should succeed
	task.Attempts = 2
	err = handler(context.Background(), task)
	assert.NoError(t, err)
}
