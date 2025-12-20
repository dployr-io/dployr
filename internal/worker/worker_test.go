// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package worker

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/store"
)

// Mock stores for testing
type mockDeploymentStore struct {
	mu          sync.Mutex
	deployments map[string]*store.Deployment
	statusCalls []string
}

func (m *mockDeploymentStore) CreateDeployment(ctx context.Context, d *store.Deployment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deployments[d.ID] = d
	return nil
}

func (m *mockDeploymentStore) GetDeployment(ctx context.Context, id string) (*store.Deployment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if d, ok := m.deployments[id]; ok {
		return d, nil
	}
	return nil, nil
}

func (m *mockDeploymentStore) ListDeployments(ctx context.Context, limit, offset int) ([]*store.Deployment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*store.Deployment
	for _, d := range m.deployments {
		result = append(result, d)
	}
	return result, nil
}

func (m *mockDeploymentStore) UpdateDeploymentStatus(ctx context.Context, id, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if d, ok := m.deployments[id]; ok {
		d.Status = store.Status(status)
		m.statusCalls = append(m.statusCalls, status)
	}
	return nil
}

func (m *mockDeploymentStore) statusCallsSnapshot() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, len(m.statusCalls))
	copy(out, m.statusCalls)
	return out
}

type mockServiceStore struct {
	services map[string]*store.Service
}

func (m *mockServiceStore) GetService(ctx context.Context, id string) (*store.Service, error) {
	if s, ok := m.services[id]; ok {
		return s, nil
	}
	return nil, nil
}

func (m *mockServiceStore) ListServices(ctx context.Context, limit, offset int) ([]*store.Service, error) {
	var result []*store.Service
	for _, s := range m.services {
		result = append(result, s)
	}
	return result, nil
}

func (m *mockServiceStore) SaveService(ctx context.Context, svc *store.Service) (*store.Service, error) {
	m.services[svc.ID] = svc
	return svc, nil
}

func (m *mockServiceStore) DeleteService(ctx context.Context, id string) error {
	delete(m.services, id)
	return nil
}

func TestWorker_New(t *testing.T) {
	cfg := &shared.Config{}
	logger := shared.NewLogger()
	deployStore := &mockDeploymentStore{deployments: make(map[string]*store.Deployment)}
	svcStore := &mockServiceStore{services: make(map[string]*store.Service)}

	worker := New(5, cfg, logger, deployStore, svcStore)

	if worker == nil {
		t.Fatal("expected non-nil worker")
	}

	if worker.maxConcurrent != 5 {
		t.Errorf("expected maxConcurrent 5, got %d", worker.maxConcurrent)
	}

	if worker.semaphore == nil {
		t.Error("expected non-nil semaphore")
	}
}

func TestWorker_Submit(t *testing.T) {
	cfg := &shared.Config{}
	logger := shared.NewLogger()
	deployStore := &mockDeploymentStore{deployments: make(map[string]*store.Deployment)}
	svcStore := &mockServiceStore{services: make(map[string]*store.Service)}

	worker := New(2, cfg, logger, deployStore, svcStore)

	t.Run("submit job to queue", func(t *testing.T) {
		worker.Submit("test-job-1")

		// Give it a moment to be queued
		time.Sleep(10 * time.Millisecond)

		if len(worker.queue) != 1 {
			t.Errorf("expected 1 job in queue, got %d", len(worker.queue))
		}
	})

	t.Run("submit multiple jobs", func(t *testing.T) {
		worker2 := New(2, cfg, logger, deployStore, svcStore)

		worker2.Submit("job-1")
		worker2.Submit("job-2")
		worker2.Submit("job-3")

		time.Sleep(10 * time.Millisecond)

		if len(worker2.queue) != 3 {
			t.Errorf("expected 3 jobs in queue, got %d", len(worker2.queue))
		}
	})
}

func TestWorker_ActiveJobs(t *testing.T) {
	cfg := &shared.Config{}
	logger := shared.NewLogger()
	deployStore := &mockDeploymentStore{deployments: make(map[string]*store.Deployment)}
	svcStore := &mockServiceStore{services: make(map[string]*store.Service)}

	worker := New(2, cfg, logger, deployStore, svcStore)

	t.Run("mark job as active", func(t *testing.T) {
		worker.markActive("job-1")

		if !worker.isRunning("job-1") {
			t.Error("expected job-1 to be marked as active")
		}
	})

	t.Run("mark job as inactive", func(t *testing.T) {
		worker.markActive("job-2")
		worker.markInactive("job-2")

		if worker.isRunning("job-2") {
			t.Error("expected job-2 to be marked as inactive")
		}
	})

	t.Run("check non-existent job", func(t *testing.T) {
		if worker.isRunning("non-existent") {
			t.Error("expected non-existent job to not be running")
		}
	})

	t.Run("concurrent job tracking", func(t *testing.T) {
		worker.markActive("job-a")
		worker.markActive("job-b")
		worker.markActive("job-c")

		if !worker.isRunning("job-a") || !worker.isRunning("job-b") || !worker.isRunning("job-c") {
			t.Error("expected all jobs to be marked as active")
		}

		worker.markInactive("job-b")

		if !worker.isRunning("job-a") || worker.isRunning("job-b") || !worker.isRunning("job-c") {
			t.Error("expected only job-b to be inactive")
		}
	})
}

func TestWorker_Semaphore(t *testing.T) {
	cfg := &shared.Config{}
	logger := shared.NewLogger()
	deployStore := &mockDeploymentStore{deployments: make(map[string]*store.Deployment)}
	svcStore := &mockServiceStore{services: make(map[string]*store.Service)}

	t.Run("semaphore limits concurrent jobs", func(t *testing.T) {
		worker := New(2, cfg, logger, deployStore, svcStore)
		ctx := context.Background()

		// Try to acquire 2 slots (should succeed)
		if err := worker.semaphore.Acquire(ctx); err != nil {
			t.Fatalf("expected first acquire to succeed, got error: %v", err)
		}
		if err := worker.semaphore.Acquire(ctx); err != nil {
			t.Fatalf("expected second acquire to succeed, got error: %v", err)
		}

		// Try to acquire third slot (should block, test with select)
		select {
		case <-time.After(100 * time.Millisecond):
			// This is expected - semaphore is full and blocks
		default:
			t.Error("expected semaphore to block on third acquire")
		}

		// Release one slot
		worker.semaphore.Release()

		// Now third acquire should succeed
		if err := worker.semaphore.Acquire(ctx); err != nil {
			t.Fatalf("expected acquire after release to succeed, got error: %v", err)
		}
	})
}

func TestWorker_StatusUpdates(t *testing.T) {
	cfg := &shared.Config{}
	logger := shared.NewLogger()
	deployStore := &mockDeploymentStore{
		deployments: make(map[string]*store.Deployment),
		statusCalls: []string{},
	}
	svcStore := &mockServiceStore{services: make(map[string]*store.Service)}

	// Create a test deployment
	deployment := &store.Deployment{
		ID:     "test-deploy-1",
		Status: store.StatusPending,
		Blueprint: store.Blueprint{
			Name:    "test-app",
			Runtime: store.RuntimeObj{Type: store.RuntimeNodeJS, Version: "18"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	deployStore.CreateDeployment(context.Background(), deployment)

	worker := New(1, cfg, logger, deployStore, svcStore)
	ctx := context.Background()

	t.Run("status updates during execution", func(t *testing.T) {
		// Execute will try to update status (though it will fail due to missing dependencies)
		go func() {
			worker.execute(ctx, "test-deploy-1")
		}()

		// Give it time to update status
		time.Sleep(100 * time.Millisecond)

		calls := deployStore.statusCallsSnapshot()

		// Check that status update was called
		if len(calls) == 0 {
			t.Error("expected at least one status update call")
		}

		// First call should be in_progress
		if len(calls) > 0 && calls[0] != string(store.StatusInProgress) {
			t.Errorf("expected first status update to be 'in_progress', got %s", calls[0])
		}
	})
}

func TestWorker_ContextCancellation(t *testing.T) {
	cfg := &shared.Config{}
	logger := shared.NewLogger()
	deployStore := &mockDeploymentStore{deployments: make(map[string]*store.Deployment)}
	svcStore := &mockServiceStore{services: make(map[string]*store.Service)}

	worker := New(2, cfg, logger, deployStore, svcStore)

	t.Run("worker stops on context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		done := make(chan bool)
		go func() {
			worker.Start(ctx)
			done <- true
		}()

		// Let worker start
		time.Sleep(50 * time.Millisecond)

		// Cancel context
		cancel()

		// Wait for worker to stop
		select {
		case <-done:
			// Worker stopped successfully
		case <-time.After(1 * time.Second):
			t.Error("worker did not stop within timeout")
		}
	})
}

func TestWorker_DuplicateJobPrevention(t *testing.T) {
	cfg := &shared.Config{}
	logger := shared.NewLogger()
	deployStore := &mockDeploymentStore{deployments: make(map[string]*store.Deployment)}
	svcStore := &mockServiceStore{services: make(map[string]*store.Service)}

	worker := New(2, cfg, logger, deployStore, svcStore)

	// Mark a job as active
	worker.markActive("duplicate-job")

	t.Run("duplicate job is skipped", func(t *testing.T) {
		if !worker.isRunning("duplicate-job") {
			t.Error("expected job to be marked as running")
		}

		// The worker.Start would skip this job
		// We're testing the isRunning check here
		shouldSkip := worker.isRunning("duplicate-job")
		if !shouldSkip {
			t.Error("expected duplicate job to be identified as running")
		}
	})
}
