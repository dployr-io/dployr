// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"context"
	"time"
)

// Instance table for this dployr server
type Instance struct {
	ID              string    `json:"id" db:"id"`
	BootstrapToken  string    `json:"bootstrap_token" db:"bootstrap_token"`
	AccessToken     string    `json:"access_token" db:"access_token"`
	InstanceID      string    `json:"instance_id" db:"instance_id"`
	Issuer          string    `json:"issuer" db:"issuer"`
	Audience        string    `json:"audience" db:"audience"`
	RegisteredAt    time.Time `json:"registered_at" db:"registered_at"`
	LastInstalledAt time.Time `json:"last_installed_at" db:"last_installed_at"`
}

// InstanceStore provides access to the instance record stored in SQLite.
type InstanceStore interface {
	// GetInstance returns the current instance record, if any.
	GetInstance(ctx context.Context) (*Instance, error)
	// RegisterInstance persists the instance row on first registration.
	RegisterInstance(ctx context.Context, i *Instance) error
	UpdateLastInstalledAt(ctx context.Context) error
	SetBootstrapToken(ctx context.Context, token string) error
	GetBootstrapToken(ctx context.Context) (string, error)
	SetAccessToken(ctx context.Context, token string) error
	GetAccessToken(ctx context.Context) (string, error)
}
