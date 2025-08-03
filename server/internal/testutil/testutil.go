// server/internal/testutil/testutil.go
package testutil

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// JSONRequest creates JSON request body
func JSONRequest(t *testing.T, v any) *strings.Reader {
	t.Helper()

	data, err := json.Marshal(v)
	require.NoError(t, err)

	return strings.NewReader(string(data))
}

// ParseJSONResponse parses JSON response
func ParseJSONResponse(t *testing.T, resp *http.Response, v any) {
	t.Helper()

	defer resp.Body.Close()
	err := json.NewDecoder(resp.Body).Decode(v)
	require.NoError(t, err)
}

// TokenResponse represents a JWT token response for testing
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// Claims represents JWT claims for testing
type Claims struct {
	UserID string `json:"user_id"`
}

// MockJWTManager for testing
type MockJWTManager struct {
	GenerateTokenPairFunc func(userID string) (*TokenResponse, error)
	ValidateTokenFunc     func(token string) (*Claims, error)
}

func (m *MockJWTManager) GenerateTokenPair(userID string) (*TokenResponse, error) {
	if m.GenerateTokenPairFunc != nil {
		return m.GenerateTokenPairFunc(userID)
	}
	return &TokenResponse{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresIn:    900,
	}, nil
}

func (m *MockJWTManager) ValidateToken(token string) (*Claims, error) {
	if m.ValidateTokenFunc != nil {
		return m.ValidateTokenFunc(token)
	}
	return &Claims{UserID: "test-user-id"}, nil
}

// AssertErrorContains checks if error contains expected text
func AssertErrorContains(t *testing.T, err error, expected string) {
	t.Helper()
	require.Error(t, err)
	require.Contains(t, err.Error(), expected)
}

// WaitForCondition waits for a condition to be true
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("condition not met within timeout")
}
