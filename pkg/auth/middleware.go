package auth

import (
	"context"
	"dployr/pkg/shared"
	"dployr/pkg/store"
	"net/http"
)

type Middleware struct {
	auth      Authenticator
	userStore store.UserStore
}

func NewMiddleware(auth Authenticator, userStore store.UserStore) *Middleware {
	return &Middleware{
		auth:      auth,
		userStore: userStore,
	}
}

func (m *Middleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if len(authHeader) <= len("Bearer ") || authHeader[:len("Bearer ")] != "Bearer " {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		token := authHeader[len("Bearer "):]
		claims, err := m.auth.ValidateToken(token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		user, err := m.userStore.GetUserByEmail(r.Context(), claims.Email)
		if err != nil {
			http.Error(w, "user not found", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), shared.CtxUserIDKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) Trace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := shared.EnrichContext(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
