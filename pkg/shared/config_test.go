// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package shared

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestIsTokenExpired(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected bool
	}{
		{
			name:     "invalid format",
			token:    "invalid",
			expected: true,
		},
		{
			name:     "expired token",
			token:    createTestToken(time.Now().Add(-1 * time.Hour).Unix()),
			expected: true,
		},
		{
			name:     "valid token",
			token:    createTestToken(time.Now().Add(10 * time.Minute).Unix()),
			expected: false,
		},
		{
			name:     "expiring soon",
			token:    createTestToken(time.Now().Add(30 * time.Second).Unix()),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTokenExpired(tt.token)
			if result != tt.expected {
				t.Errorf("isTokenExpired() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func createTestToken(exp int64) string {
	payload := map[string]interface{}{
		"exp": exp,
	}
	jsonBytes, _ := json.Marshal(payload)
	encoded := base64.RawURLEncoding.EncodeToString(jsonBytes)
	return "header." + encoded + ".signature"
}

func TestSanitizeSyncInterval(t *testing.T) {
	tests := []struct {
		name string
		in   time.Duration
		want time.Duration
	}{
		{"zero uses default", 0, 30 * time.Second},
		{"negative uses default", -1 * time.Second, 30 * time.Second},
		{"below minimum", 1 * time.Second, minSyncInterval},
		{"above maximum", 10 * time.Minute, maxSyncInterval},
		{"within bounds", 30 * time.Second, 30 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeSyncInterval(tt.in)
			if got != tt.want {
				t.Fatalf("sanitizeSyncInterval(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestGetEnvAsDuration(t *testing.T) {
	key := "TEST_SYNC_INTERVAL"
	os.Unsetenv(key)

	t.Run("uses default when unset", func(t *testing.T) {
		got := getEnvAsDuration(key, 45*time.Second)
		if got != sanitizeSyncInterval(45*time.Second) {
			t.Fatalf("expected default duration, got %v", got)
		}
	})

	t.Run("parses duration string", func(t *testing.T) {
		os.Setenv(key, "20s")
		defer os.Unsetenv(key)
		got := getEnvAsDuration(key, 45*time.Second)
		if got != sanitizeSyncInterval(20*time.Second) {
			t.Fatalf("expected 20s, got %v", got)
		}
	})

	t.Run("parses integer seconds", func(t *testing.T) {
		os.Setenv(key, "15")
		defer os.Unsetenv(key)
		got := getEnvAsDuration(key, 45*time.Second)
		if got != sanitizeSyncInterval(15*time.Second) {
			t.Fatalf("expected 15s, got %v", got)
		}
	})
}
