package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func EnsureClientCertificate(instanceID string) (tls.Certificate, error) {
	certPath, keyPath := DefaultClientCertPaths()

	if fileExists(certPath) && fileExists(keyPath) {
		return tls.LoadX509KeyPair(certPath, keyPath)
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("generate ecdsa key: %w", err)
	}

	serial, err := crand.Int(crand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		serial = big.NewInt(time.Now().UnixNano())
	}

	tmpl := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: fmt.Sprintf("dployr-instance:%s", strings.TrimSpace(instanceID))},
		NotBefore:             time.Now().Add(-5 * time.Minute),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(crand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("create certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("marshal private key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes})

	dir := filepath.Dir(certPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return tls.Certificate{}, fmt.Errorf("mkdir %s: %w", dir, err)
	}

	if err := os.WriteFile(certPath, certPEM, 0o644); err != nil {
		return tls.Certificate{}, fmt.Errorf("write cert: %w", err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		return tls.Certificate{}, fmt.Errorf("write key: %w", err)
	}

	return tls.X509KeyPair(certPEM, keyPEM)
}

func DefaultClientCertPaths() (certPath, keyPath string) {
	var dir string
	switch runtime.GOOS {
	case "windows":
		dir = filepath.Join(os.Getenv("PROGRAMDATA"), "dployr")
	case "darwin":
		dir = "/usr/local/etc/dployr"
	default:
		dir = "/var/lib/dployrd"
	}
	return filepath.Join(dir, "client.crt"), filepath.Join(dir, "client.key")
}

func BuildPinnedTLSConfig(clientCert tls.Certificate, wsCertPath string, embeddedCACert string) (*tls.Config, error) {
	var pool *x509.CertPool

	if wsCertPath != "" {
		b, err := os.ReadFile(wsCertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read pinned cert from %s: %w", wsCertPath, err)
		}
		p := x509.NewCertPool()
		if !p.AppendCertsFromPEM(b) {
			return nil, fmt.Errorf("failed to parse pinned cert from %s", wsCertPath)
		}
		pool = p
	} else if embeddedCACert != "" {
		p := x509.NewCertPool()
		if !p.AppendCertsFromPEM([]byte(embeddedCACert)) {
			return nil, fmt.Errorf("failed to parse embedded WebSocket CA cert")
		}
		pool = p
	}

	return &tls.Config{
		RootCAs:      pool,
		Certificates: []tls.Certificate{clientCert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

func ComputeCertFingerprint(cert tls.Certificate) (string, error) {
	if len(cert.Certificate) == 0 {
		return "", fmt.Errorf("client certificate is empty")
	}

	parsed, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return "", fmt.Errorf("failed to parse client certificate: %w", err)
	}

	hash := sha256.Sum256(parsed.RawSubjectPublicKeyInfo)
	return base64.StdEncoding.EncodeToString(hash[:]), nil
}

func CertToPEM(cert tls.Certificate) (string, error) {
	if len(cert.Certificate) == 0 {
		return "", fmt.Errorf("client certificate is empty")
	}

	parsed, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return "", fmt.Errorf("failed to parse client certificate: %w", err)
	}

	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: parsed.Raw})), nil
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
