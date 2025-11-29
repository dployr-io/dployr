// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"net/http"

	"github.com/dployr-io/dployr/pkg/shared"
)

type ctxKey string

const claimsCtxKey ctxKey = "claims"

type Role string

const (
	RoleViewer    Role = "viewer"
	RoleDeveloper Role = "developer"
	RoleAdmin     Role = "admin"
	RoleOwner     Role = "owner"
	RoleAgent     Role = "agent" // M2M role for daemon-to-base communication
)

type Middleware struct {
	auth Authenticator
}

func NewMiddleware(auth Authenticator) *Middleware {
	return &Middleware{
		auth: auth,
	}
}

func (m *Middleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		ctx := r.Context()
		logger := shared.LogWithContext(ctx)
		if authHeader == "" {
			e := shared.Errors.Auth.Unauthorized
			logger.Error("authorization header missing")
			shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
			return
		}

		if len(authHeader) <= len("Bearer ") || authHeader[:len("Bearer ")] != "Bearer " {
			e := shared.Errors.Auth.Unauthorized
			logger.Error("authorization header has invalid format")
			shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
			return
		}
		token := authHeader[len("Bearer "):]
		claims, err := m.auth.ValidateToken(ctx, token)
		if err != nil {
			e := shared.Errors.Auth.Unauthorized
			logger.Error("token validation failed", "error", err)
			shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
			return
		}

		ctx = context.WithValue(ctx, shared.CtxUserIDKey, claims.Subject)
		ctx = context.WithValue(ctx, claimsCtxKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func isPermitted(actual, required string) bool {
	a, okA := parseRole(actual)
	r, okR := parseRole(required)
	if !okA || !okR {
		return false
	}
	return roleAllows(a, r)
}

func parseRole(s string) (Role, bool) {
	switch s {
	case string(RoleViewer):
		return RoleViewer, true
	case string(RoleDeveloper):
		return RoleDeveloper, true
	case string(RoleAdmin):
		return RoleAdmin, true
	case string(RoleOwner):
		return RoleOwner, true
	case string(RoleAgent):
		return RoleAgent, true
	default:
		return "", false
	}
}

// roleAllows encodes which roles can act as which others in a single place.
// Higher-privilege roles are explicitly listed instead of using numeric weights.
func roleAllows(actual, required Role) bool {
	if actual == required {
		return true
	}
	switch actual {
	case RoleOwner:
		return required == RoleAdmin || required == RoleDeveloper || required == RoleViewer
	case RoleAdmin:
		return required == RoleDeveloper || required == RoleViewer
	case RoleDeveloper:
		return required == RoleViewer
	case RoleAgent:
		// Agent can act as viewer for read-only tasks
		return required == RoleViewer
	default:
		return false
	}
}

// IsPermitted is a public wrapper for checking role permissions.
func IsPermitted(actual, required string) bool {
	return isPermitted(actual, required)
}

func (m *Middleware) RequireRole(required string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := shared.LogWithContext(ctx)
			claims, ok := ctx.Value(claimsCtxKey).(*Claims)
			if !ok {
				e := shared.Errors.Runtime.InternalServer
				logger.Error("missing claims in context during permission check")
				shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
				return
			}

			if !isPermitted(claims.Perm, required) {
				e := shared.Errors.Auth.Forbidden
				logger.Warn("permission denied", "required", required, "actual", claims.Perm)
				shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *Middleware) Trace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := shared.EnrichContext(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
