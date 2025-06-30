package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheckHandler returns a simple health check response
func HealthCheckHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	}
}
