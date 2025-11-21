package store

import (
	"context"
	"time"
)

// Instance table for this dployr server
type Instance struct {
	ID              string    `json:"id" db:"id"`
	Token           string    `json:"token" db:"token"`
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
	SetToken(ctx context.Context, token string) error
	GetToken(ctx context.Context) (string, error)
}
