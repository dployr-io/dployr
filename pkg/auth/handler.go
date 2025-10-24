package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"dployr/pkg/store"
)

type AuthHandler struct {
	store store.UserStore
	logger   *slog.Logger
	auth  Authenticator
}

func NewAuthHandler(store store.UserStore, logger *slog.Logger, auth Authenticator) *AuthHandler {
	return &AuthHandler{
		store: store,
		logger: logger,
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
		msg := "missing email in request body"
		h.logger.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	u, err := h.store.FindOrCreateUser(req.Email)
	if err != nil {
		msg := "unable to create new user"
		h.logger.Error(msg, "error", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	jwt, err := h.auth.NewToken(req.Email, req.Expiry)
	if err != nil {
		msg := "failed to generate new token"
		h.logger.Error(msg, "error", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	claims, err := h.auth.ValidateToken(jwt)
	if err != nil {
		msg := "failed to validate generated token"
		h.logger.Error(msg, "error", err)
		http.Error(w, msg, http.StatusInternalServerError)
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
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		msg := "missing Authorization header"
		h.logger.Error(msg)
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		msg := "invalid auth header format"
		h.logger.Error(msg)
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}

	tokenStr := parts[1]
	claims, err := h.auth.ValidateToken(tokenStr)
	if err != nil {
		msg := "invalid or expired token"
		h.logger.Error(msg)
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(Claims{
		Email: claims.Email,
		ExpiresAt:   claims.ExpiresAt,
		IssuedAt: claims.IssuedAt,
	})
}
