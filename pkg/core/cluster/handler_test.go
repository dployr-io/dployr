// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package cluster

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dployr-io/dployr/pkg/shared"
)

// stubProvisioner records calls and optionally returns an error.
type stubProvisioner struct {
	calls []setupCall
	err   error
}

type setupCall struct {
	clusterID     string
	memoryMB      int
	cpuMillicores int
}

func (s *stubProvisioner) Setup(clusterID string, memoryMB int, cpuMillicores int) error {
	s.calls = append(s.calls, setupCall{clusterID, memoryMB, cpuMillicores})
	return s.err
}

func newTestHandler(s Provisioner) *Handler {
	return NewHandler(s, shared.NewLogger())
}

func post(t *testing.T, h *Handler, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/clusters/setup", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.SetupCluster(rr, req)
	return rr
}

func TestSetupCluster_ValidRequest(t *testing.T) {
	stub := &stubProvisioner{}
	h := newTestHandler(stub)

	rr := post(t, h, map[string]any{
		"cluster_id":     "01JZZTEST001",
		"cluster_memory": 64,
		"cluster_cpu":    100,
	})

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	if len(stub.calls) != 1 {
		t.Fatalf("expected 1 Setup call, got %d", len(stub.calls))
	}
	c := stub.calls[0]
	if c.clusterID != "01JZZTEST001" {
		t.Errorf("expected cluster_id 01JZZTEST001, got %s", c.clusterID)
	}
	if c.memoryMB != 64 {
		t.Errorf("expected memoryMB 64, got %d", c.memoryMB)
	}
	if c.cpuMillicores != 100 {
		t.Errorf("expected cpuMillicores 100, got %d", c.cpuMillicores)
	}

	var resp map[string]bool
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp["success"] {
		t.Errorf("expected success:true in response")
	}
}

func TestSetupCluster_MissingClusterID(t *testing.T) {
	stub := &stubProvisioner{}
	h := newTestHandler(stub)

	rr := post(t, h, map[string]any{
		"cluster_memory": 64,
		"cluster_cpu":    100,
	})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	if len(stub.calls) != 0 {
		t.Errorf("Setup must not be called when cluster_id is missing")
	}
}

func TestSetupCluster_SetupError(t *testing.T) {
	stub := &stubProvisioner{err: errors.New("cgroup write failed")}
	h := newTestHandler(stub)

	rr := post(t, h, map[string]any{
		"cluster_id":     "01JZZTEST002",
		"cluster_memory": 64,
		"cluster_cpu":    100,
	})

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

func TestSetupCluster_MethodNotAllowed(t *testing.T) {
	stub := &stubProvisioner{}
	h := newTestHandler(stub)

	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		req := httptest.NewRequest(method, "/clusters/setup", nil)
		rr := httptest.NewRecorder()
		h.SetupCluster(rr, req)

		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("method %s: expected 405, got %d", method, rr.Code)
		}
	}
	if len(stub.calls) != 0 {
		t.Errorf("Setup must not be called for non-POST requests")
	}
}

func TestSetupCluster_InvalidBody(t *testing.T) {
	stub := &stubProvisioner{}
	h := newTestHandler(stub)

	req := httptest.NewRequest(http.MethodPost, "/clusters/setup", bytes.NewBufferString("not-json"))
	rr := httptest.NewRecorder()
	h.SetupCluster(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d", rr.Code)
	}
}
