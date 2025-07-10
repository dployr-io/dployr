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
