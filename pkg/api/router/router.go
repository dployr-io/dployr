package router

import (
	"encoding/gob"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"dployr.io/pkg/api"
	"dployr.io/pkg/api/auth"
	"dployr.io/pkg/api/platform"
	"dployr.io/pkg/queue"
)

type Router struct {
	Auth       *auth.Auth
	QM         *queue.QueueManager
	SSEManager *platform.SSEManager
}

// New registers the routes and returns the router.
func New(r *Router) *gin.Engine {
	router := gin.Default()

	// To store custom types in our cookies,
	// we must first register them using gob.Register
	gob.Register(map[string]interface{}{})

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("auth-session", store))

	router.LoadHTMLGlob("dashboard/public/index.html")

	router.Static("/public", "dashboard/public")

	router.GET("/health", api.HealthCheckHandler())

	router.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", nil)
	})

	// API v1 routes
	v1 := router.Group("/v1")
	{
		v1.GET("/login", r.Auth.LoginHandler())
		v1.GET("/callback", r.Auth.CallbackHandler())
		v1.GET("/user", r.Auth.UserHandler())
		v1.GET("/logout", r.Auth.LogoutHandler())

		// SSE endpoint for build logs streaming
		v1.GET("/builds/:buildId/logs/stream", platform.BuildLogsStreamHandler(r.SSEManager))

		// Test endpoint to simulate build logs
		v1.POST("/builds/:buildId/test", platform.TestBuildLogsHandler(r.SSEManager))
	}

	return router
}
