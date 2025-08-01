package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	_ "embed"

	"github.com/golang-jwt/jwt/v5"

	"dployr.io/pkg/api/middleware"
	"dployr.io/pkg/mail"
	"dployr.io/pkg/models"
	"dployr.io/pkg/repository"
	"github.com/gin-gonic/gin"
)

type VerificationData struct {
	Name string
	Code string
}

type MagicCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name,omitempty"` // Optional for signup
}

type MagicCodeVerify struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

type JWTManager struct {
	secretKey []byte
}

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func JWTAuth(jwtManager *JWTManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			ctx.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
			ctx.Abort()
			return
		}

		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			ctx.Abort()
			return
		}

		// Set user info in context
		ctx.Set("user_id", claims.UserID)
		ctx.Next()
	}
}

// NewJWTManager creates a new JWT manager with a secure random key
func NewJWTManager() *JWTManager {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		panic("failed to generate JWT secret key")
	}
	return &JWTManager{secretKey: key}
}

func (j *JWTManager) GenerateTokenPair(userID string) (*TokenResponse, error) {
	accessToken, err := j.generateAccessToken(userID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := j.generateRefreshToken()
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900, // 15 minutes in seconds
	}, nil
}

// generateAccessToken creates a short-lived JWT
func (j *JWTManager) generateAccessToken(userID string) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "dployr.io",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// generateRefreshToken creates a random string (not a JWT)
func (j *JWTManager) generateRefreshToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ValidateToken validates and parses a JWT token
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RequestMagicCodeHandler handles both signup and login with rate limiting
// @Summary Request magic code for authentication
// @Description Request a 6-digit magic code to be sent to the user's email for authentication
// @Tags auth
// @Accept json
// @Produce json
// @Param request body MagicCodeRequest true "Magic code request"
// @Success 200 {object} gin.H "Magic code sent successfully"
// @Failure 400 {object} gin.H "Invalid request"
// @Failure 429 {object} gin.H "Too many requests"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /auth/request-code [post]
func RequestMagicCodeHandler(userRepo *repository.UserRepo, tokenRepo *repository.MagicTokenRepo, rl *middleware.RateLimiter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		clientIP := ctx.ClientIP()

		var req MagicCodeRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if !rl.IsAllowed(fmt.Sprintf("%s-request-code", clientIP)) {
			ctx.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please wait before requesting another code.",
			})
			return
		}

		code, err := generateMagicCode()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate code"})
			return
		}

		magicToken := models.MagicToken{
			Email:     req.Email,
			Code:      code,
			ExpiresAt: time.Now().Add(15 * time.Minute),
			Name:      req.Name,
			Used:      false,
		}

		err = tokenRepo.Upsert(ctx, magicToken, []string{"email"}, []string{"code", "expires_at", "name", "used"})
		if err != nil {
			log.Printf("Error upserting magic token: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate login code"})
			return
		}

		err = sendMagicCodeEmail(req.Email, code, req.Name)
		if err != nil {
			log.Printf("Error sending email: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": "Check your email for the 6-digit login code",
		})
	}
}

// VerifyMagicCodeHandler verifies the magic code and returns JWT token
// @Summary Verify magic code and authenticate
// @Description Verify the 6-digit magic code received via email and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body MagicCodeVerify true "Magic code verification"
// @Success 200 {object} gin.H "Authentication successful"
// @Failure 400 {object} gin.H "Invalid request"
// @Failure 401 {object} gin.H "Invalid or expired code"
// @Failure 500 {object} gin.H "Internal server error"
func VerifyMagicCodeHandler(j *JWTManager, userRepo *repository.UserRepo, tokenRepo *repository.MagicTokenRepo) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req MagicCodeVerify
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verify and consume code
		magicToken, err := tokenRepo.ConsumeCode(ctx, req.Email, req.Code)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired code"})
			return
		}

		name := magicToken.Name
		if name == "" {
			name = strings.Split(req.Email, "@")[0]
		}

		user, err := userRepo.GetByEmail(ctx, req.Email)
		if err != nil {
			log.Printf("Error getting user by email: %s", err)
			user = models.User{
				Email: req.Email,
				Name:  name,
			}

			err = userRepo.Create(ctx, &user)

			if err != nil {
				log.Printf("Error creating user: %v", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
				return
			}
		}

		// Generate JWT session token
		res, err := j.GenerateTokenPair(user.Id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"access_token":  res.AccessToken,
			"refresh_token": res.RefreshToken,
			"expires_in":    res.ExpiresIn,
			"user":          user,
		})
	}
}

// generateMagicCode creates a 6-digit alphanumeric code (uppercase)
func generateMagicCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 6)

	for i := range code {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		code[i] = charset[randomIndex.Int64()]
	}

	return string(code), nil
}

// sendMagicCodeEmail sends the 6-char code via email
func sendMagicCodeEmail(toAddr, code, toName string) error {
	emailBody := strings.ReplaceAll(string(mail.VerificationTemplate), "{{.Name}}", toName)
	emailBody = strings.ReplaceAll(emailBody, "{{.Code}}", code)
	return mail.SendEmail(toAddr, "Your Login Code", emailBody, toName)
}

// RefreshTokenHandler - endpoint to exchange refresh token for new access token
// @Summary Refresh access token
// @Description Exchange a valid refresh token for a new access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token request"
// @Success 200 {object} TokenResponse "New access token"
// @Failure 400 {object} gin.H "Invalid request"
// @Failure 401 {object} gin.H "Invalid or expired refresh token"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /auth/refresh [post]
func RefreshTokenHandler(j *JWTManager, refreshRepo *repository.RefreshTokenRepo) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req RefreshRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate and consume refresh token (marks as used)
		refreshToken, err := refreshRepo.ConsumeToken(ctx, req.RefreshToken)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
			return
		}

		// Generate new token pair
		tokenPair, err := j.GenerateTokenPair(refreshToken.UserId)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
			return
		}

		// Store new refresh token in database
		newRefreshToken := models.RefreshToken{
			Token:     tokenPair.RefreshToken,
			UserId:    refreshToken.UserId,
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
			Used:      false,
		}

		err = refreshRepo.Create(ctx, &newRefreshToken)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
			return
		}

		// Return new token pair (same format as login)
		ctx.JSON(http.StatusOK, gin.H{
			"access_token":  tokenPair.AccessToken,
			"refresh_token": tokenPair.RefreshToken,
			"expires_in":    tokenPair.ExpiresIn,
			"token_type":    "Bearer",
		})
	}
}
