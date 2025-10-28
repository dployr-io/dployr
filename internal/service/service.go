package service

import (
	"context"
	"dployr/pkg/shared"
	"dployr/pkg/store"
	"log/slog"
)

type Servicer struct {
	cfg    *shared.Config
	logger *slog.Logger
	store  store.ServiceStore
}

func (s Servicer) CreateService(ctx context.Context, svc *store.Service) (*store.Service, error) {
	return s.store.CreateService(ctx, svc)
}

func (s Servicer) GetService(ctx context.Context, id string) (*store.Service, error) {
	return s.store.GetService(ctx, id)
}

func (s Servicer) ListServices(ctx context.Context, limit, offset int) ([]*store.Service, error) {
	return s.store.ListServices(ctx, limit, offset)
}

func (s Servicer) UpdateService(ctx context.Context, svc *store.Service) error {
	return s.store.UpdateService(ctx, svc)
}
