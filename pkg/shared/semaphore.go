// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package shared

import "context"

// Semaphore provides context-aware concurrency limiting.
type Semaphore struct {
	ch chan struct{}
}

// NewSemaphore creates a new semaphore with the given max concurrent acquisitions.
func NewSemaphore(max int) *Semaphore {
	return &Semaphore{ch: make(chan struct{}, max)}
}

// Acquire blocks until a semaphore slot is available or context is cancelled.
func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.ch <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release returns a semaphore slot.
func (s *Semaphore) Release() {
	select {
	case <-s.ch:
	default:
	}
}
