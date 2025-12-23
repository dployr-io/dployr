// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"

	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/store"
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
	GetService(ctx context.Context, name string) (*store.Service, error)
	ListServices(ctx context.Context, userID string, limit, offset int) ([]*store.Service, error)
	DeleteService(ctx context.Context, name string) error
}

func NewServicer(c *shared.Config, l *shared.Logger, s store.ServiceStore, a HandleService) *Servicer {
	return &Servicer{
		config: c,
		logger: l,
		store:  s,
		api:    a,
	}
}
