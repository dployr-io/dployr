package auth

import (
	"context"
	"dployr/pkg/shared"
	"net/http"
)

type Middleware struct {
	auth Authenticator
}

func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		token := authHeader[len("Bearer "):]
		claims, err := m.auth.ValidateToken(token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), shared.CtxUserIDKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
