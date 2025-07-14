package auth

import (
	"crypto/rand"
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

// GenerateToken creates a new JWT token for a user
func (j *JWTManager) GenerateToken(userID string) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "dployr.io",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
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
func RequestMagicCodeHandler(userRepo *repository.UserRepo, tokenRepo *repository.MagicTokenRepo, codeRateLimiter *middleware.RateLimiter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req MagicCodeRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Rate limiting
		if !codeRateLimiter.IsAllowed(req.Email) {
			ctx.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please wait before requesting another code.",
			})
			return
		}

		// Generate 6-digit alphanumeric code
		code, err := generateMagicCode()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate code"})
			return
		}

		magicToken, err := tokenRepo.GetByEmail(ctx, req.Email)

		if err != nil || magicToken == nil {
			log.Printf("No existing magic token found for email: %s", err)

			// Store code with 15-minute expiration
			magicToken = &models.MagicToken{
				Email:     req.Email,
				Code:      code,
				ExpiresAt: time.Now().Add(15 * time.Minute),
				Name:      req.Name,
				Used:      false,
			}

			err = tokenRepo.Create(ctx, magicToken)
			if err != nil {
				log.Printf("Error creating magic token: %v", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate login code"})
				return
			}
		}

		// Send email with code
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
			user = &models.User{
				Email: req.Email,
				Name: name,
			}

			err = userRepo.Create(ctx, user)


			if err != nil {
				log.Printf("Error creating user: %v", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
				return
			}
		}

		// Generate JWT session token
		sessionToken, err := generateJWT(j, user.Id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"token": sessionToken,
			"user":  user,
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


// Placeholder for JWT generation
func generateJWT(j *JWTManager, userID string) (string, error) {
	return j.GenerateToken(userID)
}
