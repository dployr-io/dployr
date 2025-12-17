// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"

	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/store"
)

type Servicer struct {
	cfg    *shared.Config
	logger *shared.Logger
	store  store.ServiceStore
}

func Init(cfg *shared.Config, logger *shared.Logger, store store.ServiceStore) *Servicer {
	return &Servicer{
		cfg:    cfg,
		logger: logger,
		store:  store,
	}
}

func (s *Servicer) GetService(ctx context.Context, id string) (*store.Service, error) {
	return s.store.GetService(ctx, id)
}

func (s *Servicer) ListServices(ctx context.Context, userID string, limit, offset int) ([]*store.Service, error) {
	return s.store.ListServices(ctx, limit, offset)
}

func (s *Servicer) DeleteService(ctx context.Context, id string) error {
	return s.store.DeleteService(ctx, id)
}
