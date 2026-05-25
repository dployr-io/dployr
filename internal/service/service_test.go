// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"errors"
	"strings"
	"testing"

	"github.com/docker/docker/errdefs"
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
func (f *fakeSvcMgr) Status(name string) (string, error)   { return "", nil }
func (f *fakeSvcMgr) ExitCode(name string) (int, error)    { return 0, nil }
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

func makeServicerWithRedeploy(mgr *fakeSvcMgr, fn RedeployFunc) *Servicer {
	return &Servicer{
		cfg:      &shared.Config{},
		logger:   shared.NewLogger(),
		svcMgr:   mgr,
		redeploy: fn,
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

func TestWakeService_ContainerMissing_Redeployed(t *testing.T) {
	redeployed := ""
	mgr := &fakeSvcMgr{startErr: errdefs.NotFound(errors.New("No such container: my-svc"))}
	s := makeServicerWithRedeploy(mgr, func(name string) error {
		redeployed = name
		return nil
	})
	if err := s.WakeService("my-svc"); err != nil {
		t.Fatalf("expected redeploy to succeed, got: %v", err)
	}
	if redeployed == "" {
		t.Error("expected redeploy to be called")
	}
}

func TestWakeService_ContainerMissing_RedeployFails(t *testing.T) {
	mgr := &fakeSvcMgr{startErr: errdefs.NotFound(errors.New("No such container: my-svc"))}
	s := makeServicerWithRedeploy(mgr, func(name string) error {
		return errors.New("image pull failed")
	})
	err := s.WakeService("my-svc")
	if err == nil {
		t.Fatal("expected error when redeploy fails")
	}
	if !strings.Contains(err.Error(), "redeploy") {
		t.Errorf("error should mention redeploy, got: %v", err)
	}
}

func TestWakeService_ContainerMissing_NoRedeployFunc(t *testing.T) {
	mgr := &fakeSvcMgr{startErr: errdefs.NotFound(errors.New("No such container: my-svc"))}
	s := makeServicer(mgr) // no redeploy func
	err := s.WakeService("my-svc")
	if err == nil {
		t.Fatal("expected error when redeploy func is nil")
	}
	if !strings.Contains(err.Error(), "redeploy") {
		t.Errorf("error should mention redeploy, got: %v", err)
	}
}

func TestWakeService_OtherErrorDoesNotRedeploy(t *testing.T) {
	redeployCalled := false
	mgr := &fakeSvcMgr{startErr: errors.New("permission denied")}
	s := makeServicerWithRedeploy(mgr, func(name string) error {
		redeployCalled = true
		return nil
	})
	if err := s.WakeService("my-svc"); err == nil {
		t.Fatal("expected error for non-NotFound failure")
	}
	if redeployCalled {
		t.Error("redeploy must not be called for non-NotFound errors")
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
