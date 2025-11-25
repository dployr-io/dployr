package service

import (
	"context"

	"dployr/pkg/shared"
	"dployr/pkg/store"
)

type SvcState string

const (
	SvcRunning SvcState = "running"
	SvcStopped SvcState = "stopped"
	SvcUnknown SvcState = "unknown"
)

type Servicer struct {
	config *shared.Config
	logger *shared.Logger
	store  store.ServiceStore
	api    HandleService
}

type ListServicesResponse struct {
	Services []*store.Service `json:"services"`
	Total    int              `json:"total"`
}

type HandleService interface {
	CreateService(ctx context.Context, svc *store.Service) (store.Service, error)
	GetService(ctx context.Context, id string) (*store.Service, error)
	ListServices(ctx context.Context, id string, limit, offset int) ([]*store.Service, error)
	UpdateService(ctx context.Context, id string, status store.Service) (*store.Service, error)
}

func NewServicer(c *shared.Config, l *shared.Logger, s store.ServiceStore, a HandleService) *Servicer {
	return &Servicer{
		config: c,
		logger: l,
		store:  s,
		api:    a,
	}
}
