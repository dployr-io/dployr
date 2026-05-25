// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"fmt"

	"github.com/docker/docker/errdefs"

	"github.com/dployr-io/dployr/internal/svc_runtime"
	"github.com/dployr-io/dployr/pkg/core/proxy"
	"github.com/dployr-io/dployr/pkg/core/utils"
	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/store"
)

// RedeployFunc recreates a missing container from its last known blueprint.
// Injected at startup so the service layer stays decoupled from deploy internals.
type RedeployFunc func(name string) error

type Servicer struct {
	cfg      *shared.Config
	logger   *shared.Logger
	store    store.ServiceStore
	proxyAPI proxy.HandleProxy
	svcMgr   svc_runtime.ServiceManager
	redeploy RedeployFunc // nil = redeploy not available
}

func Init(cfg *shared.Config, logger *shared.Logger, store store.ServiceStore, proxyAPI proxy.HandleProxy, redeploy RedeployFunc) *Servicer {
	svcMgr, err := svc_runtime.SvcRuntime()
	if err != nil {
		logger.Error("failed to initialize service manager", "error", err)
	}

	return &Servicer{
		cfg:      cfg,
		logger:   logger,
		store:    store,
		proxyAPI: proxyAPI,
		svcMgr:   svcMgr,
		redeploy: redeploy,
	}
}

func (s *Servicer) GetService(ctx context.Context, name string) (*store.Service, error) {
	return s.store.GetService(ctx, name)
}

func (s *Servicer) ListServices(ctx context.Context, userID string, limit, offset int) ([]*store.Service, error) {
	return s.store.ListServices(ctx, limit, offset)
}

func (s *Servicer) SleepService(name string) error {
	if s.svcMgr == nil {
		return fmt.Errorf("service manager not available")
	}
	svcName := utils.FormatName(name)
	s.logger.Info("sleeping service", "service", svcName)
	if err := s.svcMgr.Stop(svcName); err != nil {
		return fmt.Errorf("failed to stop service %s: %w", svcName, err)
	}
	return nil
}

func (s *Servicer) WakeService(name string) error {
	if s.svcMgr == nil {
		return fmt.Errorf("service manager not available")
	}
	svcName := utils.FormatName(name)
	s.logger.Info("waking service", "service", svcName)
	if err := s.svcMgr.Start(svcName); err != nil {
		if errdefs.IsNotFound(err) {
			return s.redeployService(svcName)
		}
		return fmt.Errorf("failed to start service %s: %w", svcName, err)
	}
	return nil
}

func (s *Servicer) redeployService(name string) error {
	if s.redeploy == nil {
		return fmt.Errorf("container %s not found and redeploy is not available", name)
	}
	s.logger.Info("container missing, redeploying service", "service", name)
	if err := s.redeploy(name); err != nil {
		return fmt.Errorf("failed to redeploy service %s: %w", name, err)
	}
	return nil
}

func (s *Servicer) IceService(name string) error {
	if s.svcMgr == nil {
		return fmt.Errorf("service manager not available")
	}
	svcName := utils.FormatName(name)
	s.logger.Info("icing service", "service", svcName)
	if err := s.svcMgr.Ice(svcName); err != nil {
		return fmt.Errorf("failed to ice service %s: %w", svcName, err)
	}
	return nil
}

func (s *Servicer) DeleteService(ctx context.Context, name string) error {
	svc, err := s.store.GetService(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}
	if svc == nil {
		return fmt.Errorf("service not found")
	}

	service_name := utils.FormatName(svc.Name)

	if s.svcMgr != nil {
		s.logger.Info("stopping systemd service", "service", service_name)
		if err := s.svcMgr.Stop(service_name); err != nil {
			s.logger.Warn("failed to stop service (may not exist)", "service", service_name, "error", err)
		}

		s.logger.Info("removing systemd service", "service", service_name)
		if err := s.svcMgr.Remove(service_name); err != nil {
			s.logger.Warn("failed to remove service (may not exist)", "service", service_name, "error", err)
		}
	}

	if s.proxyAPI != nil {
		apps := s.proxyAPI.GetApps()
		var domainsToRemove []string

		for _, app := range apps {
			if app.Upstream != "" && (app.Upstream == fmt.Sprintf("localhost:%d", svc.Port) ||
				app.Upstream == fmt.Sprintf("127.0.0.1:%d", svc.Port)) {
				domainsToRemove = append(domainsToRemove, app.Domain)
			}
		}

		if len(domainsToRemove) > 0 {
			s.logger.Info("removing proxy configurations", "domains", domainsToRemove)
			if err := s.proxyAPI.Remove(domainsToRemove); err != nil {
				s.logger.Warn("failed to remove proxy configurations", "domains", domainsToRemove, "error", err)
			}
		}
	}

	s.logger.Info("deleting service from database", "service_name", name)
	if err := s.store.DeleteService(ctx, name); err != nil {
		return fmt.Errorf("failed to delete service from database: %w", err)
	}

	s.logger.Info("service deleted successfully", "service_name", name)
	return nil
}
