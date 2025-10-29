package auth

import (
	"fmt"
	"time"

	"dployr/pkg/auth"
	"dployr/pkg/shared"

	"github.com/golang-jwt/jwt/v4"
)

type Auth struct {
	cfg *shared.Config
}

func Init(cfg *shared.Config) *Auth {
	return &Auth{cfg: cfg}
}

func (a Auth) NewToken(email, lifespan string) (string, error) {
	exp := time.Now()
	if lifespan != "" && lifespan != "never" {
		d, _ := time.ParseDuration(lifespan)
		exp = exp.Add(d)
	} else if lifespan == "never" {
		exp = exp.Add(100 * 365 * 24 * time.Hour)
	}

	claims := jwt.MapClaims{
		"email": email,
		"exp":   exp.Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(a.cfg.SecretKey))
}

func (a Auth) ValidateToken(inputToken string) (*auth.Claims, error) {
	token, err := jwt.ParseWithClaims(inputToken, &auth.Claims{}, func(t *jwt.Token) (any, error) {
		return []byte(a.cfg.SecretKey), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*auth.Claims); ok && token.Valid {
		if time.Now().Unix() > claims.ExpiresAt {
			return nil, fmt.Errorf("token expired")
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
