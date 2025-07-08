package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"slices"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/bitbucket"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/gitlab"

	"dployr.io/pkg/logger"
	"dployr.io/pkg/queue"
	"dployr.io/pkg/repository"
)

type Auth struct {
	logger       *logger.Logger
	projectRepo  *repository.ProjectRepo
	eventRepo    *repository.EventRepo
	queueManager *queue.QueueManager
}

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Provider string `json:"provider"`
}

func InitAuth(projectRepo *repository.ProjectRepo, eventRepo *repository.EventRepo, queueManager *queue.QueueManager) *Auth {
	redirectURL := os.Getenv("OAUTH_REDIRECT_URL")
	if redirectURL == "" {
		redirectURL = "http://localhost:7879/auth"
	}

	// Initialize Goth providers
	var providers []goth.Provider

	// GitHub
	if githubKey := os.Getenv("GITHUB_CLIENT_ID"); githubKey != "" {
		providers = append(providers, github.New(
			githubKey,
			os.Getenv("GITHUB_CLIENT_SECRET"),
			fmt.Sprintf("%s/github/callback", redirectURL),
		))
	}

	// GitLab
	if gitlabKey := os.Getenv("GITLAB_CLIENT_ID"); gitlabKey != "" {
		providers = append(providers, gitlab.New(
			gitlabKey,
			os.Getenv("GITLAB_CLIENT_SECRET"),
			fmt.Sprintf("%s/gitlab/callback", redirectURL),
		))
	}

	// BitBucket
	if bitbucketKey := os.Getenv("BITBUCKET_CLIENT_ID"); bitbucketKey != "" {
		providers = append(providers, bitbucket.New(
			bitbucketKey,
			os.Getenv("BITBUCKET_CLIENT_SECRET"),
			fmt.Sprintf("%s/bitbucket/callback", redirectURL),
		))
	}

	// Set up Goth providers
	goth.UseProviders(providers...)

	// Configure Gothic to work with Gin
	gothic.GetProviderName = getProviderName

	return &Auth{
		projectRepo:  projectRepo,
		eventRepo:    eventRepo,
		queueManager: queueManager,
	}
}

// Custom function to extract provider name from Gin context
func getProviderName(req *http.Request) (string, error) {
	// Extract provider from URL path
	// Expected paths: /auth/{provider} or /auth/{provider}/callback
	path := req.URL.Path
	parts := strings.Split(path, "/")
	
	if len(parts) >= 3 && parts[1] == "auth" {
		return parts[2], nil
	}
	
	return "", fmt.Errorf("provider not found in path: %s", path)
}

func (a *Auth) LoginHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		provider := ctx.Param("provider")
		
		// Validate provider
		if !a.isValidProvider(provider) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "unsupported provider"})
			return
		}

		// Store provider in session for callback
		session := sessions.Default(ctx)
		session.Set("provider", provider)
		session.Save()

		// Set up the request path for Gothic
		ctx.Request.URL.Path = fmt.Sprintf("/auth/%s", provider)
		
		// Try to get existing user first
		if gothUser, err := gothic.CompleteUserAuth(ctx.Writer, ctx.Request); err == nil {
			// User already authenticated, convert and redirect
			user := a.convertGothUser(gothUser)
			session.Set("user", user)
			session.Set("access_token", gothUser.AccessToken)
			session.Save()
			
			ctx.Redirect(http.StatusTemporaryRedirect, "/v1/user")
			return
		}

		// Start new authentication
		gothic.BeginAuthHandler(ctx.Writer, ctx.Request)
	}
}

func (a *Auth) CallbackHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		provider := ctx.Param("provider")
		
		// Validate provider
		if !a.isValidProvider(provider) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "unsupported provider"})
			return
		}

		// Set up the request path for Gothic
		ctx.Request.URL.Path = fmt.Sprintf("/auth/%s/callback", provider)

		// Complete authentication with Gothic
		gothUser, err := gothic.CompleteUserAuth(ctx.Writer, ctx.Request)
		if err != nil {
			log.Printf("Auth error: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "authentication failed"})
			return
		}

		// Convert Goth user to our User struct
		user := a.convertGothUser(gothUser)

		// Store user in session
		session := sessions.Default(ctx)
		session.Set("access_token", gothUser.AccessToken)
		session.Set("user", user)
		session.Set("provider", provider)
		session.Save()

		// Setup user account asynchronously
		go a.setupUserAccount(user)

		ctx.Redirect(http.StatusTemporaryRedirect, "/v1/user")
	}
}

func (a *Auth) UserHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		user := session.Get("user")

		if user == nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		ctx.JSON(http.StatusOK, user)
	}
}

func (a *Auth) LogoutHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		provider := session.Get("provider")
		
		if provider != nil {
			// Set up request path for Gothic logout
			ctx.Request.URL.Path = fmt.Sprintf("/logout/%s", provider.(string))
			err := gothic.Logout(ctx.Writer, ctx.Request)
			if err != nil {
				log.Printf("Logout error: %v", err)
			}
		}

		// Clear session
		session.Clear()
		session.Save()
		
		ctx.JSON(http.StatusOK, gin.H{"message": "logged out"})
	}
}

func (a *Auth) setupUserAccount(user *User) {
	log.Printf("Setting up account for user: %s", user.ID)

	baseJob := queue.NewJob(user.ID, map[string]interface{}{
		"user_id":  user.ID,
		"provider": user.Provider,
	})

	jobArgs := queue.SetupUserJobArgs{
		BaseJobArgs: *baseJob,
	}

	if a.queueManager != nil {
		_, err := a.queueManager.GetClient().Insert(context.Background(), jobArgs, nil)
		if err != nil {
			log.Printf("Failed to enqueue setup user job: %v", err)
		}
	}
}

func (a *Auth) convertGothUser(gothUser goth.User) *User {
	user := &User{
		ID:       gothUser.UserID,
		Email:    gothUser.Email,
		Provider: gothUser.Provider,
	}

	// Handle name - try different fields
	if gothUser.Name != "" {
		user.Name = gothUser.Name
	} else if gothUser.NickName != "" {
		user.Name = gothUser.NickName
	} else if gothUser.FirstName != "" || gothUser.LastName != "" {
		user.Name = strings.TrimSpace(gothUser.FirstName + " " + gothUser.LastName)
	} else {
		user.Name = gothUser.Email // fallback to email
	}

	return user
}

func (a *Auth) isValidProvider(provider string) bool {
	validProviders := []string{"github", "gitlab", "bitbucket"}
	return slices.Contains(validProviders, provider)
}
