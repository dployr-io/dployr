package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"dployr/pkg/store"
)

type AuthHandler struct {
	store  store.UserStore
	logger *slog.Logger
	auth   Authenticator
	config interface{ GetSecret() string }
}

func NewAuthHandler(store store.UserStore, logger *slog.Logger, auth Authenticator, config interface{ GetSecret() string }) *AuthHandler {
	return &AuthHandler{
		store:  store,
		logger: logger,
		auth:   auth,
		config: config,
	}
}

func (h *AuthHandler) validateOwnerSecret(providedSecret string) bool {
	return providedSecret == h.config.GetSecret()
}

func (h *AuthHandler) GenerateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Lifespan string `json:"lifespan"`           // e.g. "15m", "1h"
		Secret   string `json:"secret,omitempty"`   // Optional: for first owner user
		Username string `json:"username,omitempty"` // System username for group checking
	}
	json.NewDecoder(r.Body).Decode(&req)

	const maxEmailLength = 254
	if len(req.Email) > maxEmailLength {
		msg := "email address too long"
		h.logger.Error(msg)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": msg,
		})
		return
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		msg := "invalid email format"
		h.logger.Error(msg)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": msg,
		})
		return
	}
	if req.Email == "" {
		msg := "missing email in request body"
		h.logger.Error(msg)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": msg,
		})
		return
	}

	if req.Lifespan != "" && req.Lifespan != "never" {
		if _, err := time.ParseDuration(req.Lifespan); err != nil {
			msg := "invalid lifespan format, expected duration like '15m', '1h', '24h' or 'never'"
			h.logger.Error(msg, "lifespan", req.Lifespan)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"error": msg,
			})
			return
		}
	}

	var role store.Role = store.RoleViewer // Default role

	if req.Secret != "" {
		if !h.validateOwnerSecret(req.Secret) {
			msg := "invalid secret key"
			h.logger.Error(msg)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{
				"error": msg,
			})
			return
		}

		hasOwner, err := h.store.HasOwner()
		if err != nil {
			h.logger.Warn("unable to check if owner exists, proceeding with owner creation", "error", err)
		}

		if !hasOwner || err != nil {
			role = store.RoleOwner
			h.logger.Info("attempting to create user as owner", "email", req.Email)
		} else {
			h.logger.Warn("owner already exists, creating user as viewer instead", "email", req.Email)
			role = store.RoleViewer
		}
	} else {
		// For non-owner users, check system groups if username provided
		if req.Username != "" {
			systemRole, roleErr := GetUserSystemRole(req.Username)
			if roleErr != nil {
				h.logger.Warn("failed to get system role, defaulting to viewer", "username", req.Username, "error", roleErr)
				role = store.RoleViewer
			} else {
				role = systemRole
				h.logger.Info("determined system role for user", "email", req.Email, "username", req.Username, "role", role)
			}
		} else {
			role = store.RoleViewer
			h.logger.Info("no username provided, defaulting to viewer", "email", req.Email)
		}
	}

	u, err := h.store.FindOrCreateUser(req.Email, role)
	if err != nil {
		msg := "unable to create new user"
		h.logger.Error(msg)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": msg,
		})
		return
	}

	username := req.Username
	if username == "" {
		username = "unknown"
	}

	// Generate access token (10 minutes)
	accessToken, err := h.auth.NewAccessToken(req.Email, username)
	if err != nil {
		msg := "failed to generate access token"
		h.logger.Error(msg)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": msg,
		})
		return
	}

	// Generate refresh token (24+ hours)
	refreshToken, err := h.auth.NewRefreshToken(req.Email, username, req.Lifespan)
	if err != nil {
		msg := "failed to generate refresh token"
		h.logger.Error(msg)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": msg,
		})
		return
	}

	// Get expiry info from refresh token
	refreshClaims, err := h.auth.ValidateToken(refreshToken)
	if err != nil {
		msg := "failed to validate refresh token"
		h.logger.Error(msg)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": msg,
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_at":    time.Unix(refreshClaims.ExpiresAt, 0),
		"user":          u.Email,
		"role":          u.Role,
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
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"error": msg,
		})
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		msg := "invalid auth header format"
		h.logger.Error(msg)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"error": msg,
		})
		return
	}

	tokenStr := parts[1]
	claims, err := h.auth.ValidateToken(tokenStr)
	if err != nil {
		msg := "invalid or expired token"
		h.logger.Error(msg)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"error": msg,
		})
		return
	}

	json.NewEncoder(w).Encode(Claims{
		Email:     claims.Email,
		ExpiresAt: claims.ExpiresAt,
		IssuedAt:  claims.IssuedAt,
	})
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.RefreshToken == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "refresh_token required",
		})
		return
	}

	// Validate refresh token
	claims, err := h.auth.ValidateToken(req.RefreshToken)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid refresh token",
		})
		return
	}

	// Ensure it's actually a refresh token
	if claims.TokenType != "refresh" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid token type",
		})
		return
	}

	// Generate new access token
	newAccessToken, err := h.auth.NewAccessToken(claims.Email, claims.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to generate access token",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"access_token": newAccessToken,
		"token_type":   "Bearer",
	})
}
