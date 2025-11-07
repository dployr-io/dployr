package shared

import (
	"encoding/base64"
	"encoding/json"
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