package auth

import (
	"github.com/golang-jwt/jwt/v4"
)

type TokenType string

type Claims struct {
	Email      string    `json:"email,omitempty"`
	Username   string    `json:"username,omitempty"`
	InstanceId string    `json:"instance_id,omitempty"`
	TokenType  TokenType `json:"token_type"`
	Nonce      string    `json:"nonce,omitempty"` // For bootstrap tokens
	ExpiresAt  int64     `json:"exp"`
	IssuedAt   int64     `json:"iat"`
	jwt.RegisteredClaims
}

type Authenticator interface {
	NewAccessToken(email, username string) (string, error)
	NewRefreshToken(email, username string, lifespan string) (string, error)
	NewBootstrapToken(instanceId string) (string, error) // New method
	ValidateToken(inputToken string) (*Claims, error)
}
