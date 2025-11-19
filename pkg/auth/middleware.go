package auth

import (
	"context"
	"dployr/pkg/shared"
	"net/http"
)

type ctxKey string

const claimsCtxKey ctxKey = "claims"

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

		ctx := r.Context()
		ctx = context.WithValue(ctx, shared.CtxUserIDKey, claims.Subject)
		ctx = context.WithValue(ctx, claimsCtxKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func isPermitted(actual, required string) bool {
	order := map[string]int{
		"viewer":    1,
		"developer": 2,
		"admin":     3,
		"owner":     4,
	}

	a, okA := order[actual]
	r, okR := order[required]
	if !okA || !okR {
		return false
	}
	return a >= r
}

func (m *Middleware) RequireRole(required string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(claimsCtxKey).(*Claims)
			if !ok {
				http.Error(w, "claims not found in context", http.StatusInternalServerError)
				return
			}

			if !isPermitted(claims.Perm, required) {
				http.Error(w, "forbidden: insufficient role permissions", http.StatusForbidden)
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
