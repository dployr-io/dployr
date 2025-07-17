package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"dployr.io/pkg/api/auth"
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

	health := observability.NewHealthManager(ssh, queue)
	r.GET("/health", health.HealthHandler())

	// Swagger endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := r.Group("/v1")
	{
		// Public routes
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

		// Protected routes
		api := v1.Group("/api")
		api.Use(auth.JWTAuth(j))
		{
			api.GET("/projects", projects.RetrieveProjectsHandler(ar.ProjectRepo))
			api.POST("/projects", projects.CreateProjectHandler(ar.ProjectRepo, rl))
			api.PUT("/projects/:id", projects.UpdateProjectHandler(ar.ProjectRepo, rl))
			api.DELETE("/projects/:id", projects.DeleteProjectHandler(ar.ProjectRepo))
		}
	}

	return r
}
