package auth

import (
	"github.com/golang-jwt/jwt/v4"
)

type TokenType string

type Claims struct {
	Email     string    `json:"email"`
	Username  string    `json:"username"`   // System username for group checking
	TokenType TokenType `json:"token_type"` // "access" or "refresh"
	ExpiresAt int64     `json:"exp"`
	IssuedAt  int64     `json:"iat"`
	jwt.RegisteredClaims
}

type Authenticator interface {
	NewAccessToken(email, username string) (string, error)
	NewRefreshToken(email, username string, lifespan string) (string, error)
	ValidateToken(inputToken string) (*Claims, error)
}
