// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package worker

import (
	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/store"
)

type Worker struct {
	maxConcurrent int
	logger        *shared.Logger
	store         store.DeploymentStore
	config        *shared.Config
	semaphore     chan struct{}
	activeJobs    map[string]bool
	queue         chan string
}

func NewWorker(m int, c *shared.Config, l *shared.Logger, s store.DeploymentStore) *Worker {
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

func (w *Worker) Submit(id string) {
	w.queue <- id
}
