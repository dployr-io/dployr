// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"fmt"

	"github.com/dployr-io/dployr/internal/svc_runtime"
	"github.com/dployr-io/dployr/pkg/core/proxy"
	"github.com/dployr-io/dployr/pkg/core/utils"
	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/store"
)

type Servicer struct {
	cfg      *shared.Config
	logger   *shared.Logger
	store    store.ServiceStore
	proxyAPI proxy.HandleProxy
	svcMgr   svc_runtime.ServiceManager
}

func Init(cfg *shared.Config, logger *shared.Logger, store store.ServiceStore, proxyAPI proxy.HandleProxy) *Servicer {
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
	}
}

func (s *Servicer) GetService(ctx context.Context, id string) (*store.Service, error) {
	return s.store.GetService(ctx, id)
}

func (s *Servicer) ListServices(ctx context.Context, userID string, limit, offset int) ([]*store.Service, error) {
	return s.store.ListServices(ctx, limit, offset)
}

func (s *Servicer) DeleteService(ctx context.Context, id string) error {
	svc, err := s.store.GetService(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}
	if svc == nil {
		return fmt.Errorf("service not found")
	}

	serviceName := utils.FormatName(svc.Name)

	if s.svcMgr != nil {
		s.logger.Info("stopping systemd service", "service", serviceName)
		if err := s.svcMgr.Stop(serviceName); err != nil {
			s.logger.Warn("failed to stop service (may not exist)", "service", serviceName, "error", err)
		}

		s.logger.Info("removing systemd service", "service", serviceName)
		if err := s.svcMgr.Remove(serviceName); err != nil {
			s.logger.Warn("failed to remove service (may not exist)", "service", serviceName, "error", err)
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

	s.logger.Info("deleting service from database", "service_id", id)
	if err := s.store.DeleteService(ctx, id); err != nil {
		return fmt.Errorf("failed to delete service from database: %w", err)
	}

	s.logger.Info("service deleted successfully", "service_id", id, "service_name", svc.Name)
	return nil
}
