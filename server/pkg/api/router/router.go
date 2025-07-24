package router

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"dployr.io/pkg/api/auth"
	"dployr.io/pkg/api/deployments"
	"dployr.io/pkg/api/logs"
	"dployr.io/pkg/api/middleware"
	"dployr.io/pkg/api/observability"
	"dployr.io/pkg/api/platform"
	"dployr.io/pkg/api/projects"
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

	r.LoadHTMLGlob("public/*.html")
	r.StaticFile("/favicon.ico", "./public/favicon.ico")

	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", nil)
	})


	health := observability.NewHealthManager(ssh, queue)
	r.GET("/health", health.HealthHandler())

	// Swagger endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := r.Group("/v1")
	{
		// Public routes
		_auth := v1.Group("/auth")
		_auth.Use(auth.JWTAuth(j))
		{
			_auth.POST("/request-code", auth.RequestMagicCodeHandler(
				ar.UserRepo,
				ar.MagicTokenRepo,
				rl,
			))
			_auth.POST("/verify-code", auth.VerifyMagicCodeHandler(
				j,
				ar.UserRepo,
				ar.MagicTokenRepo,
			))
			_auth.POST("/refresh", auth.RefreshTokenHandler(
				j,
				ar.RefreshTokenRepo,
			))
		}

		v1.POST("/ssh/connect", ssh.SshConnectHandler())
		v1.GET("/ws/ssh/:session-id", ssh.SshWebSocketHandler())

		// Protected routes
		api := v1.Group("/api")
		api.Use(auth.JWTAuth(j))
		{
			// Projects
			api.GET("/projects", projects.RetrieveProjectsHandler(ar.ProjectRepo))
			api.POST("/projects", projects.CreateProjectHandler(ar.ProjectRepo, rl))
			api.PUT("/projects/:id", projects.UpdateProjectHandler(ar.ProjectRepo, rl))
			api.DELETE("/projects/:id", projects.DeleteProjectHandler(ar.ProjectRepo))

			// Deployments
			api.GET("/deployments", deployments.RetrieveDeploymentsHandler(ar.DeploymentRepo))
			api.POST("/deployments", deployments.CreateDeploymentHandler(ar.DeploymentRepo, rl))
			api.GET("/deployments/:id", deployments.RetrieveDeploymentHandler(ar.DeploymentRepo))

			// Logs
			api.GET("/logs/stream", logs.StreamLogsHandler(ar.LogRepo))
		}
	}

	return r
}
