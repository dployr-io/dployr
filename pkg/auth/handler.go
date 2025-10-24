package auth

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"dployr/pkg/store"
)

type AuthHandler struct {
	store store.UserStore
	auth  Authenticator
}

func NewAuthHandler(store store.UserStore, auth Authenticator) *AuthHandler {
	return &AuthHandler{
		store: store,
		auth:  auth,
	}
}

func (h *AuthHandler) GenerateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email  string `json:"email"`
		Expiry string `json:"expiry"` // e.g. "15m", "1h", "never"
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Email == "" {
		http.Error(w, "missing email", http.StatusBadRequest)
		return
	}

	u, err := h.store.FindOrCreateUser(req.Email)
	if err != nil {
		http.Error(w, "user not found", http.StatusInternalServerError)
		return
	}

	jwt, err := h.auth.NewToken(req.Email, req.Expiry)
	if err != nil {
		http.Error(w, "failed to generate new token", http.StatusInternalServerError)
		return
	}

	claims, err := h.auth.ValidateToken(jwt)
	if err != nil {
		http.Error(w, "failed to validate generated token", http.StatusInternalServerError)
		return
	}

	exp := time.Unix(claims.ExpiresAt, 0)

	json.NewEncoder(w).Encode(map[string]any{
		"token":      jwt,
		"expires_at": exp,
		"user":       u.Email,
	})
}

func (h *AuthHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "missing Authorization header", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "invalid auth header format", http.StatusUnauthorized)
		return
	}

	tokenStr := parts[1]
	claims, err := h.auth.ValidateToken(tokenStr)
	if err != nil {
		http.Error(w, "invalid or expired token", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"valid": true,
		"email": claims.Email,
		"exp":   claims.ExpiresAt,
	})
}
