package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
)

type JWKSHandler struct {
	publicKey *rsa.PublicKey
}

func NewJWKSHandler(publicKey *rsa.PublicKey) *JWKSHandler {
	return &JWKSHandler{publicKey: publicKey}
}

func (h *JWKSHandler) ServeJWKS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Convert RSA public key to JWK format
	n := base64.RawURLEncoding.EncodeToString(h.publicKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(h.publicKey.E)).Bytes())

	jwks := map[string]any{
		"keys": []map[string]string{
			{
				"kty": "RSA",
				"use": "sig",
				"alg": "RS256",
				"n":   n,
				"e":   e,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jwks)
}
