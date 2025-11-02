package auth

import "github.com/golang-jwt/jwt/v4"

type Claims struct {
	Email     string `json:"email"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
	jwt.RegisteredClaims
}

type Authenticator interface {
	NewToken(email, lifespan string) (string, error) 
	ValidateToken(inputToken string) (*Claims, error)
}
