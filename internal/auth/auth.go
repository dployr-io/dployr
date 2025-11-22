package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	pkgAuth "dployr/pkg/auth"
	"dployr/pkg/shared"
	"dployr/pkg/store"

	"github.com/golang-jwt/jwt/v4"
)

type jwk struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwksResponse struct {
	Keys []jwk `json:"keys"`
}

type Auth struct {
	cfg       *shared.Config
	store     store.InstanceStore
	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey // kid -> key
	lastFetch time.Time
}

func Init(cfg *shared.Config, s store.InstanceStore) *Auth {
	return &Auth{
		cfg:   cfg,
		store: s,
		keys:  make(map[string]*rsa.PublicKey),
	}
}

func (a *Auth) ValidateToken(ctx context.Context, tokenStr string) (*pkgAuth.Claims, error) {
	if tokenStr == "" {
		return nil, errors.New("empty token")
	}

	// First parse to inspect header and get kid/alg.
	parser := &jwt.Parser{}
	token, _, err := parser.ParseUnverified(tokenStr, &pkgAuth.Claims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token header: %w", err)
	}

	head, ok := token.Header["kid"].(string)
	if !ok || head == "" {
		return nil, errors.New("missing kid in token header")
	}

	alg, _ := token.Header["alg"].(string)
	if alg == "" || !strings.HasPrefix(alg, "RS") {
		return nil, fmt.Errorf("unexpected signing alg: %s", alg)
	}

	pub, err := a.getKey(head)
	if err != nil {
		return nil, err
	}

	// Now parse and verify with claims.
	claims := &pkgAuth.Claims{}
	parsed, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != alg {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Method.Alg())
		}
		return pub, nil
	})
	if err != nil {
		return nil, fmt.Errorf("token verification failed: %w", err)
	}
	if !parsed.Valid {
		return nil, errors.New("invalid token")
	}

	now := time.Now().Unix()
	if claims.ExpiresAt != 0 && now > claims.ExpiresAt {
		return nil, errors.New("token expired")
	}

	inst, err := a.store.GetInstance(ctx)
	if err != nil {
		shared.LogWithContext(ctx).Error("failed to get instance for token", "error", err)
		return nil, err
	}

	if inst == nil {
		err := errors.New("instance not found")
		shared.LogWithContext(ctx).Error("token validation failed", "error", err)
		return nil, err
	}

	if inst.Issuer != "" && claims.Issuer != "" && claims.Issuer != inst.Issuer {
		err := errors.New("invalid token issuer")
		shared.LogWithContext(ctx).Error("token validation failed", "error", err)
		return nil, err
	}
	if inst.Audience != "" {
		if !containsAudience(claims.Audience, inst.Audience) {
			err := errors.New("invalid token audience")
			shared.LogWithContext(ctx).Error("token validation failed", "error", err)
			return nil, err
		}
	}

	if inst.InstanceID != "" && claims.InstanceID != "" && claims.InstanceID != inst.InstanceID {
		err := errors.New("token not intended for this instance")
		shared.LogWithContext(ctx).Error("token validation failed", "error", err)
		return nil, err
	}

	return claims, nil
}

func (a *Auth) getKey(kid string) (*rsa.PublicKey, error) {
	a.mu.RLock()
	key, ok := a.keys[kid]
	stale := time.Since(a.lastFetch) > 5*time.Minute
	a.mu.RUnlock()

	if ok && !stale {
		return key, nil
	}

	if err := a.refreshJWKS(); err != nil {
		return nil, err
	}

	a.mu.RLock()
	defer a.mu.RUnlock()
	key, ok = a.keys[kid]
	if !ok {
		return nil, fmt.Errorf("no key for kid %s", kid)
	}
	return key, nil
}

func (a *Auth) refreshJWKS() error {
	if a.cfg.BaseURL == "" {
		return errors.New("base_url is not configured")
	}

	resp, err := http.Get(a.cfg.BaseURL + "/.well-known/jwks.json")
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	var body jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	newKeys := make(map[string]*rsa.PublicKey)
	for _, k := range body.Keys {
		if k.Kty != "RSA" || k.N == "" || k.E == "" {
			continue
		}

		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			continue
		}
		e := big.NewInt(0).SetBytes(eBytes).Int64()
		pub := &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: int(e)}
		newKeys[k.Kid] = pub
	}

	if len(newKeys) == 0 {
		return errors.New("no usable keys in JWKS")
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	a.keys = newKeys
	a.lastFetch = time.Now()
	return nil
}

func containsAudience(aud []string, expected string) bool {
	return slices.Contains(aud, expected)
}
