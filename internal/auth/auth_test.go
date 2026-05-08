// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pkgAuth "github.com/dployr-io/dployr/pkg/auth"
	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/store"
	"github.com/golang-jwt/jwt/v4"
)

const testKID = "test-key-1"

func generateKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

// signToken mints a JWT signed with key using the given claims.
func signToken(t *testing.T, key *rsa.PrivateKey, claims pkgAuth.Claims) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tok.Header["kid"] = testKID
	s, err := tok.SignedString(key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return s
}

// jwksServer spins up an httptest server that serves the public key as JWKS.
func jwksServer(t *testing.T, pub *rsa.PublicKey) *httptest.Server {
	t.Helper()
	nBytes := pub.N.Bytes()
	eBytes := big.NewInt(int64(pub.E)).Bytes()

	body := map[string]any{
		"keys": []map[string]any{
			{
				"kty": "RSA",
				"kid": testKID,
				"use": "sig",
				"alg": "RS256",
				"n":   base64.RawURLEncoding.EncodeToString(nBytes),
				"e":   base64.RawURLEncoding.EncodeToString(eBytes),
			},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// mockInstanceStore satisfies store.InstanceStore for test control.
type mockInstanceStore struct {
	instance *store.Instance
	token    string
	bToken   string
}

func (m *mockInstanceStore) GetInstance(_ context.Context) (*store.Instance, error) {
	return m.instance, nil
}
func (m *mockInstanceStore) RegisterInstance(_ context.Context, _ *store.Instance) error { return nil }
func (m *mockInstanceStore) UpdateLastInstalledAt(_ context.Context) error               { return nil }
func (m *mockInstanceStore) SetBootstrapToken(_ context.Context, t string) error {
	m.bToken = t
	return nil
}
func (m *mockInstanceStore) GetBootstrapToken(_ context.Context) (string, error) {
	return m.bToken, nil
}
func (m *mockInstanceStore) SetAccessToken(_ context.Context, t string) error {
	m.token = t
	return nil
}
func (m *mockInstanceStore) GetAccessToken(_ context.Context) (string, error) { return m.token, nil }

// newAuth builds an Auth wired to a real JWKS server.
func newAuth(t *testing.T, srv *httptest.Server, inst *store.Instance) *Auth {
	t.Helper()
	cfg := &shared.Config{BaseURL: srv.URL}
	ms := &mockInstanceStore{instance: inst}
	return Init(cfg, ms)
}

func TestContainsAudience(t *testing.T) {
	cases := []struct {
		aud      []string
		expected string
		want     bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a", "b", "c"}, "d", false},
		{[]string{}, "a", false},
		{nil, "a", false},
	}
	for _, c := range cases {
		got := containsAudience(c.aud, c.expected)
		if got != c.want {
			t.Errorf("containsAudience(%v, %q) = %v, want %v", c.aud, c.expected, got, c.want)
		}
	}
}

func TestValidateToken_EmptyToken(t *testing.T) {
	a := &Auth{keys: make(map[string]*rsa.PublicKey)}
	_, err := a.ValidateToken(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestValidateToken_MissingKID(t *testing.T) {
	key := generateKey(t)
	// Sign without setting kid header.
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, pkgAuth.Claims{})
	delete(tok.Header, "kid")
	s, _ := tok.SignedString(key)

	a := &Auth{keys: make(map[string]*rsa.PublicKey)}
	_, err := a.ValidateToken(context.Background(), s)
	if err == nil {
		t.Fatal("expected error for missing kid")
	}
}

func TestValidateToken_NonRSAlgorithm(t *testing.T) {
	// HS256 token — must be rejected before any key lookup.
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, pkgAuth.Claims{})
	tok.Header["kid"] = testKID
	s, _ := tok.SignedString([]byte("secret"))

	a := &Auth{keys: make(map[string]*rsa.PublicKey), cfg: &shared.Config{}}
	_, err := a.ValidateToken(context.Background(), s)
	if err == nil {
		t.Fatal("expected error for non-RS algorithm")
	}
}

func TestValidateToken_Valid(t *testing.T) {
	key := generateKey(t)
	srv := jwksServer(t, &key.PublicKey)

	inst := &store.Instance{InstanceID: "inst-1"}
	a := newAuth(t, srv, inst)

	claims := pkgAuth.Claims{
		InstanceID: "inst-1",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	tok := signToken(t, key, claims)

	got, err := a.ValidateToken(context.Background(), tok)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.InstanceID != "inst-1" {
		t.Errorf("InstanceID = %q, want inst-1", got.InstanceID)
	}
}

func TestValidateToken_Expired(t *testing.T) {
	key := generateKey(t)
	srv := jwksServer(t, &key.PublicKey)
	a := newAuth(t, srv, &store.Instance{})

	// Claims.ExpiresAt is int64 (shadows RegisteredClaims.ExpiresAt).
	// The manual check in ValidateToken reads this int64 field.
	claims := pkgAuth.Claims{
		ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(),
	}
	tok := signToken(t, key, claims)

	_, err := a.ValidateToken(context.Background(), tok)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestValidateToken_IssuerMismatch(t *testing.T) {
	key := generateKey(t)
	srv := jwksServer(t, &key.PublicKey)
	inst := &store.Instance{Issuer: "https://base.dployr.dev"}
	a := newAuth(t, srv, inst)

	claims := pkgAuth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "https://evil.example.com",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		},
	}
	tok := signToken(t, key, claims)

	_, err := a.ValidateToken(context.Background(), tok)
	if err == nil {
		t.Fatal("expected error for issuer mismatch")
	}
}

func TestValidateToken_AudienceMismatch(t *testing.T) {
	key := generateKey(t)
	srv := jwksServer(t, &key.PublicKey)
	inst := &store.Instance{Audience: "dployr-daemon"}
	a := newAuth(t, srv, inst)

	claims := pkgAuth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  []string{"wrong-audience"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		},
	}
	tok := signToken(t, key, claims)

	_, err := a.ValidateToken(context.Background(), tok)
	if err == nil {
		t.Fatal("expected error for audience mismatch")
	}
}

func TestValidateToken_InstanceIDMismatch(t *testing.T) {
	key := generateKey(t)
	srv := jwksServer(t, &key.PublicKey)
	inst := &store.Instance{InstanceID: "inst-correct"}
	a := newAuth(t, srv, inst)

	claims := pkgAuth.Claims{
		InstanceID: "inst-wrong",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		},
	}
	tok := signToken(t, key, claims)

	_, err := a.ValidateToken(context.Background(), tok)
	if err == nil {
		t.Fatal("expected error for instance ID mismatch")
	}
}

func TestValidateToken_WrongSigningKey(t *testing.T) {
	key := generateKey(t)
	otherKey := generateKey(t)
	srv := jwksServer(t, &key.PublicKey) // server has key1's public key
	a := newAuth(t, srv, &store.Instance{})

	claims := pkgAuth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		},
	}
	// Signed with otherKey — verification against key must fail.
	tok := signToken(t, otherKey, claims)

	_, err := a.ValidateToken(context.Background(), tok)
	if err == nil {
		t.Fatal("expected error for wrong signing key")
	}
}

func TestJWKSCache_StaleTriggersRefresh(t *testing.T) {
	key := generateKey(t)
	fetchCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fetchCount++
		nBytes := key.PublicKey.N.Bytes()
		eBytes := big.NewInt(int64(key.PublicKey.E)).Bytes()
		body := fmt.Sprintf(`{"keys":[{"kty":"RSA","kid":%q,"use":"sig","alg":"RS256","n":%q,"e":%q}]}`,
			testKID,
			base64.RawURLEncoding.EncodeToString(nBytes),
			base64.RawURLEncoding.EncodeToString(eBytes),
		)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	a := Init(&shared.Config{BaseURL: srv.URL}, &mockInstanceStore{instance: &store.Instance{}})

	// Mark lastFetch as 6 minutes ago — beyond the 5-minute cache window.
	a.mu.Lock()
	a.lastFetch = time.Now().Add(-6 * time.Minute)
	a.mu.Unlock()

	claims := pkgAuth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		},
	}
	tok := signToken(t, key, claims)
	_, _ = a.ValidateToken(context.Background(), tok)

	if fetchCount == 0 {
		t.Error("expected JWKS refresh when cache is stale")
	}
}
