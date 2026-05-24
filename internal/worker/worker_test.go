// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package worker

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/dployr-io/dployr/internal/deploy"
	"github.com/dployr-io/dployr/pkg/core/proxy"
	"github.com/dployr-io/dployr/pkg/core/utils"
	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/store"
)

// mockProxyAPI implements proxy.HandleProxy for testing.
type mockProxyAPI struct {
	mu     sync.Mutex
	added  []map[string]proxy.App
	addErr error
}

func (m *mockProxyAPI) Setup(apps map[string]proxy.App) error { return nil }
func (m *mockProxyAPI) Status() proxy.ProxyStatus             { return proxy.ProxyStatus{} }
func (m *mockProxyAPI) GetApps() []proxy.App                  { return nil }
func (m *mockProxyAPI) Restart() error                        { return nil }
func (m *mockProxyAPI) Remove(domains []string) error         { return nil }
func (m *mockProxyAPI) Add(apps map[string]proxy.App) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.added = append(m.added, apps)
	return m.addErr
}
func (m *mockProxyAPI) snapshot() []map[string]proxy.App {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]map[string]proxy.App, len(m.added))
	copy(out, m.added)
	return out
}

// Mock stores for testing
type mockDeploymentStore struct {
	mu          sync.Mutex
	deployments map[string]*store.Deployment
	statusCalls []string
}

func (m *mockDeploymentStore) UpsertDeployment(ctx context.Context, d *store.Deployment) error {
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

func (m *mockServiceStore) UpsertService(ctx context.Context, svc *store.Service) (*store.Service, error) {
	m.services[svc.ID] = svc
	return svc, nil
}

func (m *mockServiceStore) DeleteService(ctx context.Context, id string) error {
	delete(m.services, id)
	return nil
}

type mockInstanceStore struct {
	accessToken string
}

func (m *mockInstanceStore) GetAccessToken(ctx context.Context) (string, error) {
	return m.accessToken, nil
}

func (m *mockInstanceStore) GetBootstrapToken(ctx context.Context) (string, error) {
	return "bootstrap-token", nil
}

func (m *mockInstanceStore) SetAccessToken(ctx context.Context, token string) error {
	m.accessToken = token
	return nil
}

func (m *mockInstanceStore) SetBootstrapToken(ctx context.Context, token string) error {
	return nil
}

func (m *mockInstanceStore) RegisterInstance(ctx context.Context, i *store.Instance) error {
	return nil
}

func (m *mockInstanceStore) UpdateLastInstalledAt(ctx context.Context) error {
	return nil
}

func (m *mockInstanceStore) GetInstance(ctx context.Context) (*store.Instance, error) {
	return nil, nil
}

func TestWorker_New(t *testing.T) {
	cfg := &shared.Config{}
	logger := shared.NewLogger()
	deployStore := &mockDeploymentStore{deployments: make(map[string]*store.Deployment)}
	svcStore := &mockServiceStore{services: make(map[string]*store.Service)}
	instStore := &mockInstanceStore{accessToken: "test-token"}

	worker := New(5, cfg, logger, deployStore, svcStore, instStore, nil)

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
	instStore := &mockInstanceStore{accessToken: "test-token"}

	worker := New(2, cfg, logger, deployStore, svcStore, instStore, nil)

	t.Run("submit job to queue", func(t *testing.T) {
		worker.Submit("test-job-1")

		// Give it a moment to be queued
		time.Sleep(10 * time.Millisecond)

		if len(worker.queue) != 1 {
			t.Errorf("expected 1 job in queue, got %d", len(worker.queue))
		}
	})

	t.Run("submit multiple jobs", func(t *testing.T) {
		worker2 := New(2, cfg, logger, deployStore, svcStore, instStore, nil)

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
	instStore := &mockInstanceStore{accessToken: "test-token"}

	worker := New(2, cfg, logger, deployStore, svcStore, instStore, nil)

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
	instStore := &mockInstanceStore{accessToken: "test-token"}

	t.Run("semaphore limits concurrent jobs", func(t *testing.T) {
		worker := New(2, cfg, logger, deployStore, svcStore, instStore, nil)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Try to acquire 2 slots (should succeed)
		if err := worker.semaphore.Acquire(ctx); err != nil {
			t.Fatalf("expected first acquire to succeed, got error: %v", err)
		}
		if err := worker.semaphore.Acquire(ctx); err != nil {
			t.Fatalf("expected second acquire to succeed, got error: %v", err)
		}

		// Try to acquire third slot (should block)
		blockCtx, blockCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer blockCancel()

		acquired := make(chan bool, 1)
		go func() {
			if err := worker.semaphore.Acquire(blockCtx); err == nil {
				acquired <- true
			}
		}()

		select {
		case <-acquired:
			t.Error("expected semaphore to block on third acquire")
		case <-time.After(100 * time.Millisecond):
			// This is expected - semaphore is full and blocks
		}

		// Release one slot
		worker.semaphore.Release()

		// Now third acquire should succeed (the blocked goroutine will get it)
		select {
		case <-acquired:
			// Expected - the blocked goroutine acquired the slot
		case <-time.After(200 * time.Millisecond):
			t.Error("expected blocked acquire to succeed after release")
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
	instStore := &mockInstanceStore{accessToken: "test-token"}

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

	deployStore.UpsertDeployment(context.Background(), deployment)

	worker := New(1, cfg, logger, deployStore, svcStore, instStore, nil)
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
	instStore := &mockInstanceStore{accessToken: "test-token"}

	worker := New(2, cfg, logger, deployStore, svcStore, instStore, nil)

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

// TestWorker_SourceRemoteGuard verifies that an instance node (non-build role)
// rejects a source=remote deployment instead of attempting to clone and build.
// This guards against a routing anomaly where a build task leaks to a regular instance.
func TestWorker_SourceRemoteGuard(t *testing.T) {
	logger := shared.NewLogger()
	svcStore := &mockServiceStore{services: make(map[string]*store.Service)}
	instStore := &mockInstanceStore{accessToken: "test-token"}

	t.Run("instance node fails source=remote deployment", func(t *testing.T) {
		cfg := &shared.Config{Role: store.NodeRoleInstance}
		deployStore := &mockDeploymentStore{deployments: make(map[string]*store.Deployment)}

		deployment := &store.Deployment{
			ID:     "remote-deploy-1",
			Status: store.StatusPending,
			Blueprint: store.Blueprint{
				Name:    "my-app",
				Source:  store.SourceRemote,
				Remote:  store.RemoteObj{Url: "https://github.com/example/app", Branch: "main"},
				Runtime: store.RuntimeObj{Type: store.RuntimeNodeJS, Version: "20"},
			},
		}
		deployStore.UpsertDeployment(context.Background(), deployment)

		w := New(1, cfg, logger, deployStore, svcStore, instStore, nil)
		_, err := w.runDeployment(context.Background(), "remote-deploy-1")

		if err == nil {
			t.Fatal("expected error when instance node receives source=remote — got nil")
		}
		if !containsAny(err.Error(), "source=remote", "routing error", "NodeRoleBuild") {
			t.Errorf("expected routing error message, got: %s", err.Error())
		}
	})

	t.Run("build node accepts source=remote deployment", func(t *testing.T) {
		cfg := &shared.Config{Role: store.NodeRoleBuild}
		deployStore := &mockDeploymentStore{deployments: make(map[string]*store.Deployment)}

		deployment := &store.Deployment{
			ID:     "remote-deploy-build",
			Status: store.StatusPending,
			Blueprint: store.Blueprint{
				Name:    "my-app",
				Source:  store.SourceRemote,
				Remote:  store.RemoteObj{Url: "https://github.com/example/app", Branch: "main"},
				Runtime: store.RuntimeObj{Type: store.RuntimeNodeJS, Version: "20"},
			},
		}
		deployStore.UpsertDeployment(context.Background(), deployment)

		w := New(1, cfg, logger, deployStore, svcStore, instStore, nil)
		_, err := w.runDeployment(context.Background(), "remote-deploy-build")

		// Build will fail (no real git/docker in test), but NOT with the routing guard error.
		// The guard only fires on instance nodes — a build node should proceed past it.
		if err != nil && containsAny(err.Error(), "expected source=image", "routing error") {
			t.Errorf("build node must not trigger the instance routing guard, got: %s", err.Error())
		}
	})

	t.Run("instance node accepts source=image deployment", func(t *testing.T) {
		cfg := &shared.Config{Role: store.NodeRoleInstance}
		deployStore := &mockDeploymentStore{deployments: make(map[string]*store.Deployment)}

		deployment := &store.Deployment{
			ID:     "image-deploy-1",
			Status: store.StatusPending,
			Blueprint: store.Blueprint{
				Name:    "my-app",
				Source:  store.SourceImage,
				Image:   "registry.digitalocean.com/dployr/my-app:1234",
				Runtime: store.RuntimeObj{Type: store.RuntimeNodeJS, Version: "20"},
			},
		}
		deployStore.UpsertDeployment(context.Background(), deployment)

		w := New(1, cfg, logger, deployStore, svcStore, instStore, nil)
		_, err := w.runDeployment(context.Background(), "image-deploy-1")

		// Will fail at docker pull (no daemon in CI), but must NOT fail with the routing guard message.
		if err != nil && containsAny(err.Error(), "expected source=image", "routing error", "source=remote") {
			t.Errorf("instance node must accept source=image, got unexpected error: %s", err.Error())
		}
	})
}

// TestWorker_PublishPath verifies the full publish callback path on an instance node:
// Publish() → Deploy() stores source=image → Submit() queues it → execute() transitions
// status to in_progress, then fails at docker pull (no daemon in test env) — never
// hitting the routing guard that fires for source=remote.
func TestWorker_PublishPath(t *testing.T) {
	cfg := &shared.Config{Role: store.NodeRoleInstance}
	logger := shared.NewLogger()
	svcStore := &mockServiceStore{services: make(map[string]*store.Service)}
	instStore := &mockInstanceStore{accessToken: "test-token"}
	deployStore := &mockDeploymentStore{
		deployments: make(map[string]*store.Deployment),
		statusCalls: []string{},
	}

	// This deployment was created by Publish() on the instance node:
	// source is already overridden to image before the worker ever sees it.
	dep := &store.Deployment{
		ID:     "publish-path-01",
		Status: store.StatusPending,
		Blueprint: store.Blueprint{
			Name:    "my-app",
			Source:  store.SourceImage,
			Image:   "registry.example.com/my-app:built-sha",
			Runtime: store.RuntimeObj{Type: store.RuntimeNodeJS, Version: "20"},
		},
	}
	deployStore.UpsertDeployment(context.Background(), dep)

	w := New(1, cfg, logger, deployStore, svcStore, instStore, nil)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		w.execute(ctx, "publish-path-01")
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("execute() did not complete within timeout")
	}

	calls := deployStore.statusCallsSnapshot()
	if len(calls) == 0 {
		t.Fatal("expected at least one status update, got none")
	}
	if calls[0] != string(store.StatusInProgress) {
		t.Errorf("first status update = %q, want %q", calls[0], store.StatusInProgress)
	}

	// Re-run runDeployment to confirm the routing guard does not fire for source=image.
	_, err := w.runDeployment(ctx, "publish-path-01")
	if err != nil && containsAny(err.Error(), "expected source=image", "routing error", "source=remote") {
		t.Errorf("routing guard must not fire for source=image on instance node: %s", err.Error())
	}
}

func newWorkerWithProxy(p proxy.HandleProxy) *Worker {
	cfg := &shared.Config{}
	return &Worker{
		cfg:        cfg,
		logger:     shared.NewLogger(),
		proxyAPI:   p,
		activeJobs: make(map[string]bool),
	}
}

func TestRegisterProxyRoute_Web(t *testing.T) {
	mock := &mockProxyAPI{}
	w := newWorkerWithProxy(mock)

	err := w.registerProxyRoute(&store.Service{
		Name: "my-app",
		Type: store.TypeWeb,
		Port: 8080,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	calls := mock.snapshot()
	if len(calls) != 1 {
		t.Fatalf("expected 1 Add call, got %d", len(calls))
	}
	app := calls[0]["my-app.dployr.run"]
	if app.Template != proxy.TemplateReverseProxy {
		t.Errorf("template = %q, want reverse_proxy", app.Template)
	}
	if app.Upstream != "localhost:8080" {
		t.Errorf("upstream = %q, want localhost:8080", app.Upstream)
	}
}

func TestRegisterProxyRoute_WebDefaultPort(t *testing.T) {
	mock := &mockProxyAPI{}
	w := newWorkerWithProxy(mock)

	w.registerProxyRoute(&store.Service{Name: "api", Type: store.TypeWeb, Port: 0})

	calls := mock.snapshot()
	app := calls[0]["api.dployr.run"]
	if app.Upstream != "localhost:3000" {
		t.Errorf("upstream = %q, want localhost:3000 for zero port", app.Upstream)
	}
}

func TestRegisterProxyRoute_Static(t *testing.T) {
	mock := &mockProxyAPI{}
	w := newWorkerWithProxy(mock)

	// WorkingDir is relative (e.g. "nodejs"); registerProxyRoute computes the
	// absolute host path from GetDataDir() + service name + WorkingDir.
	err := w.registerProxyRoute(&store.Service{
		Name:       "my-site",
		Type:       store.TypeStatic,
		WorkingDir: "src",
		StaticDir:  "dist",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	absWorkDir := filepath.Join(utils.GetDataDir(), ".dployr", "services", "my-site", "src")
	wantRoot := deploy.ResolveStaticDir(absWorkDir, "dist")

	calls := mock.snapshot()
	app := calls[0]["my-site.dployr.run"]
	if app.Template != proxy.TemplateStatic {
		t.Errorf("template = %q, want static", app.Template)
	}
	if app.Root != wantRoot {
		t.Errorf("root = %q, want %q", app.Root, wantRoot)
	}
	if app.Upstream != "" {
		t.Errorf("upstream = %q, want empty for static", app.Upstream)
	}
}

func TestRegisterProxyRoute_StaticNoSubDir(t *testing.T) {
	mock := &mockProxyAPI{}
	w := newWorkerWithProxy(mock)

	w.registerProxyRoute(&store.Service{
		Name:      "my-site",
		Type:      store.TypeStatic,
		StaticDir: "",
	})

	wantRoot := deploy.ResolveStaticDir(filepath.Join(utils.GetDataDir(), ".dployr", "services", "my-site"), "")
	app := mock.snapshot()[0]["my-site.dployr.run"]
	if app.Root != wantRoot {
		t.Errorf("root = %q, want %q when WorkingDir and StaticDir are empty", app.Root, wantRoot)
	}
}

func TestRegisterProxyRoute_Failure(t *testing.T) {
	mock := &mockProxyAPI{addErr: errors.New("caddy config invalid")}
	w := newWorkerWithProxy(mock)

	err := w.registerProxyRoute(&store.Service{
		Name: "broken-app",
		Type: store.TypeWeb,
		Port: 3000,
	})
	if err == nil {
		t.Fatal("expected error from proxy Add failure, got nil")
	}
	if !errors.Is(err, mock.addErr) && !containsStr(err.Error(), "caddy config invalid") {
		t.Errorf("error = %q, expected to contain proxy error", err)
	}
}

func TestRegisterProxyRoute_NilProxy(t *testing.T) {
	w := newWorkerWithProxy(nil)
	// Must not panic and must return nil when no proxy is configured.
	if err := w.registerProxyRoute(&store.Service{Name: "x", Type: store.TypeWeb}); err != nil {
		t.Errorf("expected nil when proxyAPI is nil, got %v", err)
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}

func containsAny(s string, substrings ...string) bool {
	for _, sub := range substrings {
		if len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}

// TestBuildServiceRecord_WorkingDirRelative is a regression test for the bug where
// the service record was stored with the absolute host path (dir) instead of the
// relative user-supplied working_dir. The UI and workload sync must never see
// internal host paths like /var/lib/dployrd/.dployr/services/ronaldo/nodejs.
func TestBuildServiceRecord_WorkingDirRelative(t *testing.T) {
	dep := &store.Deployment{
		ID: "dep-01",
		Blueprint: store.Blueprint{
			Name:       "My App",
			Source:     store.SourceRemote,
			Type:       store.TypeWeb,
			WorkingDir: "nodejs",
			StaticDir:  "public",
			Runtime:    store.RuntimeObj{Type: store.RuntimeNodeJS, Version: "20"},
			Remote:     store.RemoteObj{Url: "https://github.com/example/app", Branch: "main"},
			Port:       3000,
		},
	}

	svc := buildServiceRecord(dep, "my-app")

	if svc.WorkingDir != "nodejs" {
		t.Errorf("WorkingDir = %q, want %q (relative) — got an absolute path leak", svc.WorkingDir, "nodejs")
	}
	if svc.StaticDir != "public" {
		t.Errorf("StaticDir = %q, want %q", svc.StaticDir, "public")
	}
	if svc.ID != "my-app" {
		t.Errorf("ID = %q, want my-app", svc.ID)
	}
	if svc.DeploymentId != "dep-01" {
		t.Errorf("DeploymentId = %q, want dep-01", svc.DeploymentId)
	}
	if svc.Port != 3000 {
		t.Errorf("Port = %d, want 3000", svc.Port)
	}
}

// TestBuildServiceRecord_EmptyWorkingDir verifies that a blueprint with no working_dir
// produces an empty string in the service record — not a default or computed path.
func TestBuildServiceRecord_EmptyWorkingDir(t *testing.T) {
	dep := &store.Deployment{
		ID: "dep-02",
		Blueprint: store.Blueprint{
			Name:   "simple-app",
			Source: store.SourceImage,
			Type:   store.TypeWeb,
			Image:  "registry.example.com/simple:latest",
		},
	}

	svc := buildServiceRecord(dep, "simple-app")

	if svc.WorkingDir != "" {
		t.Errorf("WorkingDir = %q, want empty string when blueprint has none", svc.WorkingDir)
	}
}

// TestWorker_ExecuteFailed_SetsStatusFailed verifies that when runDeployment fails
// (e.g. docker is unavailable), execute() transitions the deployment to "failed".
func TestWorker_ExecuteFailed_SetsStatusFailed(t *testing.T) {
	cfg := &shared.Config{Role: store.NodeRoleInstance}
	logger := shared.NewLogger()
	svcStore := &mockServiceStore{services: make(map[string]*store.Service)}
	instStore := &mockInstanceStore{accessToken: "test-token"}
	deployStore := &mockDeploymentStore{
		deployments: make(map[string]*store.Deployment),
		statusCalls: []string{},
	}

	dep := &store.Deployment{
		ID:     "fail-dep-01",
		Status: store.StatusPending,
		Blueprint: store.Blueprint{
			Name:   "broken-app",
			Source: store.SourceImage,
			Image:  "registry.example.com/app:latest",
		},
	}
	deployStore.UpsertDeployment(context.Background(), dep)

	w := New(1, cfg, logger, deployStore, svcStore, instStore, nil)

	done := make(chan struct{})
	go func() {
		w.execute(context.Background(), "fail-dep-01")
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("execute() did not complete within timeout")
	}

	calls := deployStore.statusCallsSnapshot()
	if len(calls) < 2 {
		t.Fatalf("expected at least 2 status updates (in_progress + failed/completed), got %v", calls)
	}
	if calls[0] != string(store.StatusInProgress) {
		t.Errorf("first status = %q, want in_progress", calls[0])
	}
	last := calls[len(calls)-1]
	if last != string(store.StatusFailed) && last != string(store.StatusCompleted) {
		t.Errorf("final status = %q, want failed or completed", last)
	}
}

func TestWorker_DuplicateJobPrevention(t *testing.T) {
	cfg := &shared.Config{}
	logger := shared.NewLogger()
	deployStore := &mockDeploymentStore{deployments: make(map[string]*store.Deployment)}
	svcStore := &mockServiceStore{services: make(map[string]*store.Service)}
	instStore := &mockInstanceStore{accessToken: "test-token"}

	worker := New(2, cfg, logger, deployStore, svcStore, instStore, nil)

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
