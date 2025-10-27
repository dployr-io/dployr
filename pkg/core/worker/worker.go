package worker

import (
	"log/slog"

	"dployr/pkg/shared"
	"dployr/pkg/store"
)

type Worker struct {
	maxConcurrent int
	logger        *slog.Logger
	store         store.DeploymentStore
	config        *shared.Config
	semaphore     chan struct{}
	activeJobs    map[string]bool
	queue         chan string
}

func New(m int, c *shared.Config, l *slog.Logger, s store.DeploymentStore) *Worker {
	return &Worker{
		maxConcurrent: m,
		logger:        l,
		store:         s,
		config:        c,
		semaphore:     make(chan struct{}, m),
		activeJobs:    make(map[string]bool),
		queue:         make(chan string, 100),
	}
}

// Submit implements the JobSubmitter interface
func (w *Worker) Submit(id string) {
	w.queue <- id
}
