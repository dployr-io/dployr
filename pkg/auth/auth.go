// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"

	"github.com/golang-jwt/jwt/v4"
)

// Claims represents the token structure used across the system
type Claims struct {
	Subject    string   `json:"sub,omitempty"`
	InstanceID string   `json:"instance_id,omitempty"`
	Perm       string   `json:"perm,omitempty"` // one of: viewer, developer, admin, owner
	Scopes     []string `json:"scopes,omitempty"`
	ExpiresAt  int64    `json:"exp"`
	IssuedAt   int64    `json:"iat"`
	jwt.RegisteredClaims
}

type Authenticator interface {
	ValidateToken(ctx context.Context, inputToken string) (*Claims, error)
}
