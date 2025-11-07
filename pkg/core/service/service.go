package service

import (
	"context"
	"dployr/pkg/shared"
	"dployr/pkg/store"
	"log/slog"
)

type SvcState string

const (
	SvcRunning SvcState = "running"
	SvcStopped SvcState = "stopped"
	SvcUnknown SvcState = "unknown"
)

type Servicer struct {
	config *shared.Config
	logger *slog.Logger
	store  store.ServiceStore
	api    HandleService
}

type HandleService interface {
	CreateService(ctx context.Context, svc *store.Service) (store.Service, error)
	GetService(ctx context.Context, id string) (*store.Service, error)
	ListServices(ctx context.Context, id string, limit, offset int) ([]*store.Service, error)
	UpdateService(ctx context.Context, id string, status store.Service) error
}

func NewServicer(c *shared.Config, l *slog.Logger, s store.ServiceStore, a HandleService) *Servicer {
	return &Servicer{
		config: c,
		logger: l,
		store:  s,
		api:    a,
	}
}
