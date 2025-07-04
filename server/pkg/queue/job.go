package queue

import (
	"context"
	"time"

	"github.com/riverqueue/river"
)

type BaseJobArgs struct {
	JobID     string      `json:"job_id"`
	UserID    string      `json:"user_id,omitempty"`
	Data      interface{} `json:"data"`
	CreatedAt time.Time   `json:"created_at"`
}

// Generic worker interface for type safety
type JobWorker[T JobArgs] interface {
	Work(ctx context.Context, job *river.Job[T]) error
}

// JobArgs interface - all job types must implement
type JobArgs interface {
	Kind() string
	GetBaseArgs() BaseJobArgs
}
