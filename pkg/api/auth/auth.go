package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"

	"dployr.io/pkg/config"
	"dployr.io/pkg/queue"
	"dployr.io/pkg/repository"
)

type Auth struct {
	*oidc.Provider

	projectRepo *repository.Project
	eventRepo   *repository.Event
	Qm          *queue.QueueManager
}

func InitAuth(projectRepo *repository.Project, eventRepo *repository.Event, queueManager *queue.QueueManager) *Auth {
	return &Auth{
		Provider:    config.GetOauth2Provider(),
		projectRepo: projectRepo,
		eventRepo:   eventRepo,
		Qm:          queueManager,
	}
}

// VerifyIDToken verifies that an *oauth2.Token is a valid *oidc.IDToken.
func (a *Auth) VerifyIDToken(ctx context.Context, t *oauth2.Token) (*oidc.IDToken, error) {
	rawIDToken, ok := t.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token field in oauth2 token")
	}

	oidcConfig := &oidc.Config{
		ClientID: os.Getenv("AUTH0_CLIENT_ID"),
	}

	return a.Verifier(oidcConfig).Verify(ctx, rawIDToken)
}

func (a *Auth) LoginHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		state, err := generateRandomState()
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		// Save the state inside the session.
		session := sessions.Default(ctx)
		session.Set("state", state)
		if err := session.Save(); err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		// Create OAuth2 config with dynamic redirect URL based on host
		oauth2Config := config.GetOauth2Config(ctx.Request.Host)
		authURL := oauth2Config.AuthCodeURL(state)

		ctx.Redirect(http.StatusTemporaryRedirect, authURL)
	}
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	state := base64.StdEncoding.EncodeToString(b)

	return state, nil
}

func (a *Auth) CallbackHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		if ctx.Query("state") != session.Get("state") {
			ctx.String(http.StatusBadRequest, "Invalid state parameter.")
			return
		}

		// Create OAuth2 config with dynamic redirect URL based on host
		oauth2Config := config.GetOauth2Config(ctx.Request.Host)

		// Exchange an authorization code for a token.
		token, err := oauth2Config.Exchange(ctx.Request.Context(), ctx.Query("code"))
		if err != nil {
			ctx.String(http.StatusUnauthorized, "Failed to convert an authorization code into a token.")
			return
		}

		idToken, err := a.VerifyIDToken(ctx.Request.Context(), token)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Failed to verify ID Token.")
			return
		}

		var profile map[string]interface{}
		if err := idToken.Claims(&profile); err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		session.Set("access_token", token.AccessToken)
		session.Set("profile", profile)

		go a.setupUserAccount(ctx)

		if err := session.Save(); err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		// Redirect to logged in page.
		ctx.Redirect(http.StatusTemporaryRedirect, "/v1/user")
	}
}

// TODO: Implement user page
// Handler for our logged-in user page.
func (a *Auth) UserHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		profile := session.Get("profile")

		if profile == nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		ctx.JSON(http.StatusOK, profile)
	}
}

func (a *Auth) LogoutHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logoutUrl, err := url.Parse("https://" + os.Getenv("AUTH0_DOMAIN") + "/v2/logout")
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		scheme := "http"
		if ctx.Request.TLS != nil {
			scheme = "https"
		}

		returnTo, err := url.Parse(scheme + "://" + ctx.Request.Host)
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		parameters := url.Values{}
		parameters.Add("returnTo", returnTo.String())
		parameters.Add("client_id", os.Getenv("AUTH0_CLIENT_ID"))
		logoutUrl.RawQuery = parameters.Encode()

		ctx.Redirect(http.StatusTemporaryRedirect, logoutUrl.String())
	}
}

func (a *Auth) setupUserAccount(ctx *gin.Context) {
	if profile, ok := sessions.Default(ctx).Get("profile").(map[string]interface{}); ok {
		if sub, ok := profile["sub"].(string); ok {
			// Split sub by pipe and get the right part
			parts := strings.Split(sub, "|")
			if len(parts) > 1 {
				id := parts[1]
				name := "User"
				if userName, ok := profile["name"].(string); ok {
					name = userName
				}

				log.Println("Setting up account for user" + id)

				if a.Qm != nil {
					tx, err := a.Qm.GetPool().Begin(ctx)
					if err != nil {
						log.Printf("Failed to begin transaction: %v", err)
						return
					}
					defer tx.Rollback(ctx)

					_, err = a.Qm.InsertCreateProjectJob(ctx, tx, id, fmt.Sprintf("job_%d", time.Now().UnixNano()), map[string]interface{}{
						"user_id":      id,
						"project_name": name + "'s Project",
					})
					if err != nil {
						log.Printf("Failed to insert create project job for user %s: %v", id, err)
						return
					}

					if err := tx.Commit(ctx); err != nil {
						log.Printf("Failed to commit transaction: %v", err)
						return
					}

					log.Printf("Successfully queued project creation for user %s", id)
				}
			}
		}
	}
}
