package router

import (
	"encoding/gob"
	// "net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"dployr.io/pkg/api/auth"
	"dployr.io/pkg/api/observability"
	"dployr.io/pkg/api/platform"
	"dployr.io/pkg/queue"
)

// New registers the routes and returns the router.
func New(auth *auth.Auth, queue *queue.Queue, ssh *platform.SshManager) *gin.Engine {
	r := gin.Default()

	// To store custom types in our cookies,
	// we must first register them using gob.Register
	gob.Register(map[string]interface{}{})

	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("auth-session", store))

	// r.LoadHTMLGlob("public/*.html")

	// r.StaticFile("/favicon.ico", "./public/favicon.ico")
	// r.StaticFile("/styles.css", "./public/styles.css")
	// r.StaticFile("/scripts.js", "./public/scripts.js")

	// r.GET("/", func(ctx *gin.Context) {
	// 	ctx.HTML(http.StatusOK, "index.html", nil)
	// })

	// r.GET("/about", func(ctx *gin.Context) {
	// 	ctx.HTML(http.StatusOK, "about.html", nil)
	// })

	health := observability.NewHealthManager(ssh, queue)
	r.GET("/health", health.HealthHandler())


	r.GET("/auth/:provider", auth.LoginHandler())
	r.GET("/auth/:provider/callback", auth.CallbackHandler())

	// API v1 routes
	v1 := r.Group("/v1")
	{
		v1.GET("/callback", auth.CallbackHandler())
		v1.GET("/user", auth.UserHandler())
		v1.GET("/logout", auth.LogoutHandler())

		v1.POST("/ssh/connect", ssh.SshConnectHandler())

		v1.GET("/ws/ssh/:sessionID", ssh.SshWebSocketHandler())
	}

	return r
}
