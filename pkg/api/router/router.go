package router

import (
	"encoding/gob"
	"log/slog"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"riverqueue.com/riverui"

	"dployr.io/pkg/api/auth"
	"dployr.io/pkg/api/platform"
	"dployr.io/pkg/queue"
)

type Router struct {
	Auth *auth.Auth
	QM   *queue.QueueManager
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

	if r.QM != nil {
		opts := &riverui.ServerOpts{
			Client: r.QM.GetClient(),
			DB:     r.QM.GetPool(),
			Logger: slog.Default(),
		}
		ui, err := riverui.NewServer(opts)
		if err != nil {
			gin.DefaultWriter.Write([]byte("Failed to create River UI server: " + err.Error() + "\n"))
			return router
		}

		// Serve River UI assets at root level to fix asset loading
		router.Any("/assets/*filepath", gin.WrapH(ui))

		// Serve River UI API endpoints at root level
		router.Any("/api/*path", gin.WrapH(ui))

		// Create an admin group for River UI
		adminGroup := router.Group("/admin")
		adminGroup.Any("/*path", gin.WrapH(http.StripPrefix("/admin", ui)))
	}

	return router
}
