package queue

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Queue struct {
	workers     int
	debounceDur time.Duration

	handler Handler

	taskCh    chan *Task
	debounce  map[string]*time.Timer
	mu        sync.Mutex
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewQueue(workers int, debounceDur time.Duration, handler Handler) *Queue {
	ctx, cancel := context.WithCancel(context.Background())
	q := &Queue{
		workers:     workers,
		debounceDur: debounceDur,
		handler:     handler,
		taskCh:      make(chan *Task),
		debounce:    make(map[string]*time.Timer),
		ctx:         ctx,
		cancel:      cancel,
	}
	q.startWorkers()
	return q
}

func (q *Queue) startWorkers() {
    for i := range q.workers {
        q.wg.Add(1)
        go func(id int) {
            defer q.wg.Done()
            for {
                select {
                case <-q.ctx.Done():
                    return
                case task := <-q.taskCh:
                    q.process(task)
                }
            }
        }(i)
    }
}

// Enqueue with debounce
func (q *Queue) Enqueue(t *Task) {
    q.mu.Lock()
    defer q.mu.Unlock()

    if timer, ok := q.debounce[t.Id]; ok {
        timer.Reset(q.debounceDur)
        return
    }

    // schedule after debounceDur
    q.debounce[t.Id] = time.AfterFunc(q.debounceDur, func() {
        q.mu.Lock()
        delete(q.debounce, t.Id)
        q.mu.Unlock()
        q.taskCh <- t
    })
}

// Process + retry logic
func (q *Queue) process(t *Task) {
    err := q.handler(q.ctx, t)
    if err == nil {
        return
    }

    t.Attempts++
    if t.Attempts > t.MaxRetries {
        fmt.Printf("task %s failed after %d attempts: %v\n", t.Id, t.Attempts-1, err)
        return
    }

    backoff := time.Duration(1<<t.Attempts) * 500 * time.Millisecond
    time.AfterFunc(backoff, func() {
        q.taskCh <- t
    })
}

func (q *Queue) Stop() {
    q.cancel()
    q.wg.Wait()
}
