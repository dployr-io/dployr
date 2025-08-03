// server/pkg/api/auth/auth_test.go
package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dployr.io/internal/testdb"
	"dployr.io/internal/testutil"
	"dployr.io/pkg/api/middleware"
	"dployr.io/pkg/models"
	"dployr.io/pkg/repository"
)

func TestJWTManager_GenerateTokenPair(t *testing.T) {
	jwtManager := NewJWTManager()
	userID := "test-user-123"

	tokens, err := jwtManager.GenerateTokenPair(userID)

	require.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.Equal(t, 900, tokens.ExpiresIn)
}

func TestJWTManager_ValidateToken(t *testing.T) {
	jwtManager := NewJWTManager()
	userID := "test-user-123"

	// Generate a token first
	tokens, err := jwtManager.GenerateTokenPair(userID)
	require.NoError(t, err)

	// Validate the token
	claims, err := jwtManager.ValidateToken(tokens.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
}

func TestJWTManager_ValidateToken_Invalid(t *testing.T) {
	jwtManager := NewJWTManager()

	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"invalid token", "invalid-token"},
		{"malformed jwt", "not.a.jwt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := jwtManager.ValidateToken(tt.token)
			assert.Error(t, err)
		})
	}
}

func TestJWTAuth_Middleware(t *testing.T) {
	jwtManager := NewJWTManager()
	middleware := JWTAuth(jwtManager)

	tests := []struct {
		name           string
		authorization  string
		expectedStatus int
	}{
		{
			name:           "missing authorization header",
			authorization:  "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid bearer format",
			authorization:  "InvalidFormat token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token",
			authorization:  "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("GET", "/", nil)
			if tt.authorization != "" {
				req.Header.Set("Authorization", tt.authorization)
			}
			c.Request = req

			middleware(c)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestJWTAuth_Middleware_ValidToken(t *testing.T) {
	jwtManager := NewJWTManager()
	userID := "test-user-123"

	// Generate valid token
	tokens, err := jwtManager.GenerateTokenPair(userID)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Track if the next handler was called
	nextCalled := false
	router.Use(JWTAuth(jwtManager))
	router.GET("/test", func(c *gin.Context) {
		nextCalled = true
		// Verify the user_id was set by the middleware
		assert.Equal(t, userID, c.GetString("user_id"))
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequestMagicCodeHandler(t *testing.T) {
	// Setup test database and repositories
	db := testdb.SetupTestDB(t, testdb.MockMigrations{})
	userRepo := repository.NewUserRepo(db)
	tokenRepo := repository.NewMagicTokenRepo(db)

	// Setup rate limiter
	rl := middleware.NewRateLimiter(3, 15*time.Minute)

	// Setup gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/auth/request-code", RequestMagicCodeHandler(userRepo, tokenRepo, rl))

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "valid request",
			requestBody: MagicCodeRequest{
				Email: "test@example.com",
				Name:  "Test User",
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Check your email",
		},
		{
			name: "invalid email",
			requestBody: MagicCodeRequest{
				Email: "invalid-email",
				Name:  "Test User",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing email",
			requestBody:    MagicCodeRequest{Name: "Test User"},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := testutil.JSONRequest(t, tt.requestBody)
			req := httptest.NewRequest("POST", "/auth/request-code", body)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestVerifyMagicCodeHandler(t *testing.T) {
	// Setup test database and repositories
	db := testdb.SetupTestDB(t, testdb.MockMigrations{})
	userRepo := repository.NewUserRepo(db)
	tokenRepo := repository.NewMagicTokenRepo(db)
	jwtManager := NewJWTManager()

	// Create test magic token
	magicToken := &models.MagicToken{
		Email:     "test@example.com",
		Code:      "ABC123",
		Name:      "Test User",
		Used:      false,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	err := tokenRepo.Create(context.Background(), magicToken)
	require.NoError(t, err)

	// Setup gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/auth/verify-code", VerifyMagicCodeHandler(jwtManager, userRepo, tokenRepo))

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "valid code",
			requestBody: MagicCodeVerify{
				Email: "test@example.com",
				Code:  "ABC123",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid code",
			requestBody: MagicCodeVerify{
				Email: "test@example.com",
				Code:  "INVALID",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid email format",
			requestBody: MagicCodeVerify{
				Email: "invalid-email",
				Code:  "ABC123",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := testutil.JSONRequest(t, tt.requestBody)
			req := httptest.NewRequest("POST", "/auth/verify-code", body)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestGenerateMagicCode(t *testing.T) {
	for i := 0; i < 100; i++ {
		code, err := generateMagicCode()
		require.NoError(t, err)
		assert.Len(t, code, 6)
		assert.Regexp(t, "^[A-Z0-9]{6}$", code)
	}
}
