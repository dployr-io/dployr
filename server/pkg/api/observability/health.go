package observability

import (
	"net/http"
	"time"

	"dployr.io/pkg/api/platform"
	"dployr.io/pkg/config"
	"dployr.io/pkg/queue"
	"github.com/gin-gonic/gin"
)

type HealthManager struct {
	sshManager  *platform.SshManager
	queue       *queue.Queue
}

func NewHealthManager(sshManager *platform.SshManager, queue *queue.Queue) *HealthManager {
	return &HealthManager{sshManager: sshManager, queue: queue}
}

// HealthHandler returns the health status of the application
// @Summary Health check
// @Description Get the current health status and statistics of the application
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} gin.H "Health status retrieved successfully"
// @Router /health [get]
func (api *HealthManager) HealthHandler() gin.HandlerFunc {
    return func(ctx *gin.Context) {
        ctx.JSON(http.StatusOK, gin.H{
			"version": config.Version,  
            "status": "healthy",
            "ssh_sessions": api.sshManager.SessionCount(),
            "timestamp": time.Now().Unix(),
        })
    }
}
