package auth

import (
	"fmt"
	"time"

	"dployr/pkg/auth"
	"dployr/pkg/shared"

	"github.com/golang-jwt/jwt/v4"
	"github.com/oklog/ulid/v2"
)

type Auth struct {
	cfg *shared.Config
}

func Init(cfg *shared.Config) *Auth {
	return &Auth{cfg: cfg}
}

func (a Auth) NewAccessToken(email, username string) (string, error) {
	exp := time.Now().Add(10 * time.Minute)

	claims := jwt.MapClaims{
		"email":      email,
		"username":   username,
		"token_type": "access",
		"exp":        exp.Unix(),
		"iat":        time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(a.cfg.PrivateKey)
}

func (a Auth) NewRefreshToken(email, username, lifespan string) (string, error) {
	var exp time.Time

	switch lifespan {
	case "":
		exp = time.Now().Add(24 * time.Hour)
	case "never":
		exp = time.Now().Add(10 * 365 * 24 * time.Hour)
	default:
		d, err := time.ParseDuration(lifespan)
		if err != nil {
			return "", err
		}
		exp = time.Now().Add(d)
	}

	claims := jwt.MapClaims{
		"email":      email,
		"username":   username,
		"token_type": "refresh",
		"exp":        exp.Unix(),
		"iat":        time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(a.cfg.PrivateKey)
}

func (a Auth) NewBootstrapToken(instanceId string) (string, error) {
	exp := time.Now().Add(15 * time.Minute)
	nonce := ulid.Make().String() // Single-use identifier

	claims := jwt.MapClaims{
		"instance_id": instanceId,
		"token_type":  "bootstrap",
		"nonce":       nonce,
		"exp":         exp.Unix(),
		"iat":         time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(a.cfg.PrivateKey)
}

func (a Auth) ValidateToken(inputToken string) (*auth.Claims, error) {
	token, err := jwt.ParseWithClaims(inputToken, &auth.Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return a.cfg.PublicKey, nil
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
