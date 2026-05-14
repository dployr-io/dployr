// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"errors"
	"strings"
	"testing"

	"github.com/dployr-io/dployr/pkg/shared"
)

// fakeSvcMgr implements svc_runtime.ServiceManager for testing.
type fakeSvcMgr struct {
	stopErr  error
	startErr error
	iceErr   error
	stopped  []string
	started  []string
	iced     []string
}

func (f *fakeSvcMgr) Stop(name string) error                   { f.stopped = append(f.stopped, name); return f.stopErr }
func (f *fakeSvcMgr) Start(name string) error                  { f.started = append(f.started, name); return f.startErr }
func (f *fakeSvcMgr) Ice(name string) error                    { f.iced = append(f.iced, name); return f.iceErr }
func (f *fakeSvcMgr) Remove(name string) error                 { return nil }
func (f *fakeSvcMgr) Status(name string) (string, error)       { return "", nil }
func (f *fakeSvcMgr) HealthStatus(name string) (string, error) { return "", nil }
func (f *fakeSvcMgr) Install(name, desc, runCmd, workDir string, envVars map[string]string) error {
	return nil
}

func makeServicer(mgr *fakeSvcMgr) *Servicer {
	return &Servicer{
		cfg:    &shared.Config{},
		logger: shared.NewLogger(),
		svcMgr: mgr,
	}
}

func makeNilMgrServicer() *Servicer {
	return &Servicer{
		cfg:    &shared.Config{},
		logger: shared.NewLogger(),
		svcMgr: nil,
	}
}

func TestSleepService_NilManager(t *testing.T) {
	s := makeNilMgrServicer()
	if err := s.SleepService("my-svc"); err == nil {
		t.Error("expected error when svcMgr is nil")
	}
}

func TestSleepService_Success(t *testing.T) {
	mgr := &fakeSvcMgr{}
	s := makeServicer(mgr)
	if err := s.SleepService("my-svc"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mgr.stopped) != 1 {
		t.Errorf("expected Stop called once, got %d", len(mgr.stopped))
	}
}

func TestSleepService_StopError(t *testing.T) {
	mgr := &fakeSvcMgr{stopErr: errors.New("systemd failure")}
	s := makeServicer(mgr)
	err := s.SleepService("my-svc")
	if err == nil {
		t.Fatal("expected error from Stop failure")
	}
	if !strings.Contains(err.Error(), "my") {
		t.Errorf("error should mention service name, got: %v", err)
	}
}

// --- WakeService ---

func TestWakeService_NilManager(t *testing.T) {
	s := makeNilMgrServicer()
	if err := s.WakeService("my-svc"); err == nil {
		t.Error("expected error when svcMgr is nil")
	}
}

func TestWakeService_Success(t *testing.T) {
	mgr := &fakeSvcMgr{}
	s := makeServicer(mgr)
	if err := s.WakeService("my-svc"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mgr.started) != 1 {
		t.Errorf("expected Start called once, got %d", len(mgr.started))
	}
}

func TestWakeService_StartError(t *testing.T) {
	mgr := &fakeSvcMgr{startErr: errors.New("unit not found")}
	s := makeServicer(mgr)
	if err := s.WakeService("my-svc"); err == nil {
		t.Fatal("expected error from Start failure")
	}
}

func TestIceService_NilManager(t *testing.T) {
	s := makeNilMgrServicer()
	if err := s.IceService("my-svc"); err == nil {
		t.Error("expected error when svcMgr is nil")
	}
}

func TestIceService_Success(t *testing.T) {
	mgr := &fakeSvcMgr{}
	s := makeServicer(mgr)
	if err := s.IceService("my-svc"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mgr.iced) != 1 {
		t.Errorf("expected Ice called once, got %d", len(mgr.iced))
	}
}

func TestIceService_IceError(t *testing.T) {
	mgr := &fakeSvcMgr{iceErr: errors.New("container not found")}
	s := makeServicer(mgr)
	if err := s.IceService("my-svc"); err == nil {
		t.Fatal("expected error from Ice failure")
	}
}
