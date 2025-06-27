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
	"dployr.io/pkg/queue"
)

// New registers the routes and returns the router.
func New(auth *auth.Auth, qm *queue.QueueManager) *gin.Engine {
	r := gin.Default()

	// To store custom types in our cookies,
	// we must first register them using gob.Register
	gob.Register(map[string]interface{}{})

	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("auth-session", store))

	r.LoadHTMLGlob("dashboard/public/index.html")

	r.Static("/public", "dashboard/public")

	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", nil)
	})

	// API v1 routes
	v1 := r.Group("/v1")
	{
		v1.GET("/login", auth.LoginHandler())
		v1.GET("/callback", auth.CallbackHandler())
		v1.GET("/user", auth.UserHandler())
		v1.GET("/logout", auth.LogoutHandler())
	}

	if qm != nil {
		opts := &riverui.ServerOpts{
			Client: qm.GetClient(),
			DB:     qm.GetPool(),
			Logger: slog.Default(),
		}
		ui, err := riverui.NewServer(opts)
		if err != nil {
			gin.DefaultWriter.Write([]byte("Failed to create River UI server: " + err.Error() + "\n"))
			return r
		}

		// Serve River UI assets at root level to fix asset loading
		r.Any("/assets/*filepath", gin.WrapH(ui))

		// Serve River UI API endpoints at root level
		r.Any("/api/*path", gin.WrapH(ui))

		// Create an admin group for River UI
		adminGroup := r.Group("/admin")
		adminGroup.Any("/*path", gin.WrapH(http.StripPrefix("/admin", ui)))
	}

	return r
}
