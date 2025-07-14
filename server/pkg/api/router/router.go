package router

import (
	"encoding/gob"
	// "net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"dployr.io/pkg/api/middleware"
	"dployr.io/pkg/api/observability"
	"dployr.io/pkg/api/platform"
	"dployr.io/pkg/api/auth"
	"dployr.io/pkg/queue"
	"dployr.io/pkg/repository"
)


// New registers the routes and returns the router.
func New(
	ar *repository.AppRepos, 
	queue *queue.Queue, 
	ssh *platform.SshManager, 
	rl *middleware.RateLimiter, 
	j *auth.JWTManager,
) *gin.Engine {
	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Connection", "Upgrade"}
	r.Use(cors.New(config))

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

	// API v1 routes
	v1 := r.Group("/v1")
	{
		v1.POST("/auth/request-code", auth.RequestMagicCodeHandler(
			ar.UserRepo, 
			ar.TokenRepo, 
			rl,
		))
    	v1.POST("/auth/verify-code", auth.VerifyMagicCodeHandler(
			j,
			ar.UserRepo, 
			ar.TokenRepo,
		))

		v1.POST("/ssh/connect", ssh.SshConnectHandler())

		v1.GET("/ws/ssh/:session-id", ssh.SshWebSocketHandler())
	}

	return r
}
