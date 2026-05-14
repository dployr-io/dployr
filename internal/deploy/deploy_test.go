// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package deploy

import (
	"context"
	"sync"
	"testing"

	coredeploy "github.com/dployr-io/dployr/pkg/core/deploy"
	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/store"
	"github.com/oklog/ulid/v2"
)

type mockDeployStore struct {
	mu          sync.Mutex
	deployments map[string]*store.Deployment
}

func (m *mockDeployStore) UpsertDeployment(_ context.Context, d *store.Deployment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deployments[d.ID] = d
	return nil
}

func (m *mockDeployStore) GetDeployment(_ context.Context, id string) (*store.Deployment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.deployments[id], nil
}

func (m *mockDeployStore) ListDeployments(_ context.Context, _, _ int) ([]*store.Deployment, error) {
	return nil, nil
}

func (m *mockDeployStore) UpdateDeploymentStatus(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockDeployStore) snapshot() map[string]*store.Deployment {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(map[string]*store.Deployment, len(m.deployments))
	for k, v := range m.deployments {
		out[k] = v
	}
	return out
}

type mockDisp struct {
	mu        sync.Mutex
	submitted []string
}

func (m *mockDisp) Submit(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.submitted = append(m.submitted, id)
}

func (m *mockDisp) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.submitted)
}

// --- helpers ---

func newDeployCtx() context.Context {
	ctx := context.Background()
	ctx = shared.WithTrace(ctx, ulid.Make().String())
	ctx = shared.WithUser(ctx, "user-test-01")
	return ctx
}

func newDeployer(role store.NodeRole) (*Deployer, *mockDeployStore, *mockDisp) {
	ds := &mockDeployStore{deployments: make(map[string]*store.Deployment)}
	disp := &mockDisp{}
	d := Init(
		&shared.Config{Role: role},
		shared.NewLogger(),
		ds,
		disp,
	)
	return d, ds, disp
}

func imageReq() *coredeploy.DeployRequest {
	return &coredeploy.DeployRequest{
		Name:    "my-app",
		Type:    "web",
		Source:  string(store.SourceImage),
		Runtime: "golang",
		Image:   "registry.example.com/my-app:abc123",
	}
}

func remoteReq() *coredeploy.DeployRequest {
	req := imageReq()
	req.Source = string(store.SourceRemote)
	req.Image = ""
	req.Remote = store.RemoteObj{Url: "https://github.com/example/app", Branch: "main"}
	return req
}

// Deploy() must reject source=remote on instance nodes before touching the DB.
func TestDeploy_SourceRemoteRejectedOnInstanceNode(t *testing.T) {
	d, ds, disp := newDeployer(store.NodeRoleInstance)
	ctx := newDeployCtx()

	_, err := d.Deploy(ctx, remoteReq())

	if err == nil {
		t.Fatal("expected error for source=remote on instance node, got nil")
	}
	if disp.count() != 0 {
		t.Errorf("expected 0 dispatcher submissions, got %d", disp.count())
	}
	if len(ds.snapshot()) != 0 {
		t.Error("expected no deployment record written to store")
	}
}

// Deploy() must accept source=remote on build nodes.
func TestDeploy_SourceRemoteAcceptedOnBuildNode(t *testing.T) {
	d, ds, disp := newDeployer(store.NodeRoleBuild)
	ctx := newDeployCtx()

	resp, err := d.Deploy(ctx, remoteReq())

	if err != nil {
		t.Fatalf("expected no error on build node, got: %v", err)
	}
	if resp == nil || !resp.Success {
		t.Fatal("expected successful response")
	}
	if disp.count() != 1 {
		t.Errorf("expected 1 dispatcher submission, got %d", disp.count())
	}
	if len(ds.snapshot()) != 1 {
		t.Error("expected deployment record in store")
	}
}

// Deploy() must accept source=image on instance nodes.
func TestDeploy_SourceImageAcceptedOnInstanceNode(t *testing.T) {
	d, _, disp := newDeployer(store.NodeRoleInstance)
	ctx := newDeployCtx()

	resp, err := d.Deploy(ctx, imageReq())

	if err != nil {
		t.Fatalf("expected no error for source=image on instance node, got: %v", err)
	}
	if resp == nil || !resp.Success {
		t.Fatal("expected successful response")
	}
	if disp.count() != 1 {
		t.Errorf("expected 1 dispatcher submission, got %d", disp.count())
	}
}

// Publish() must always store source=image and the provided image ref,
// regardless of what source was in the original payload.
func TestPublish_ForcesSrcImageAndQueues(t *testing.T) {
	d, ds, disp := newDeployer(store.NodeRoleInstance)
	ctx := newDeployCtx()

	const builtImage = "registry.example.com/my-app:built-sha"

	// Simulate base dispatching builds/publish:post — payload may still carry
	// the original source=remote fields from the build task.
	req := &coredeploy.PublishRequest{
		Image:   builtImage,
		Payload: *remoteReq(), // source=remote in original payload
	}

	resp, err := d.Publish(ctx, req)
	if err != nil {
		t.Fatalf("Publish() failed: %v", err)
	}
	if resp == nil || !resp.Success {
		t.Fatal("expected successful publish response")
	}
	if disp.count() != 1 {
		t.Errorf("expected 1 dispatcher submission, got %d", disp.count())
	}

	// The stored deployment must have source=image and the built image ref.
	stored := ds.snapshot()
	if len(stored) != 1 {
		t.Fatalf("expected 1 stored deployment, got %d", len(stored))
	}
	for _, dep := range stored {
		if dep.Blueprint.Source != store.SourceImage {
			t.Errorf("stored deployment source = %q, want %q", dep.Blueprint.Source, store.SourceImage)
		}
		if dep.Blueprint.Image != builtImage {
			t.Errorf("stored deployment image = %q, want %q", dep.Blueprint.Image, builtImage)
		}
	}
}
