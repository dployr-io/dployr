package logs

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"dployr.io/pkg/repository"
	"github.com/gin-gonic/gin"
)

// RetrieveLogsHandler retrieves all logs
// @Summary Get all logs
// @Description Retrieve a list of all logs for the authenticated user with optional filtering
// @Tags logs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param project_id query string false "Filter by project ID"
// @Param host query string false "Filter by host"
// @Param level query string false "Filter by log level (info, warning, error, debug)"
// @Param status query string false "Filter by status"
// @Param limit query int false "Limit number of results (default: 100, max: 1000)"
// @Param offset query int false "Offset for pagination (default: 0)"
// @Param start_date query string false "Filter logs from this date (RFC3339 format)"
// @Param end_date query string false "Filter logs until this date (RFC3339 format)"
// @Success 200 {object} gin.H "Logs retrieved successfully"
// @Failure 400 {object} gin.H "Invalid query parameters"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /api/logs [get]
func RetrieveLogsHandler(logRepo *repository.LogRepo) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Parse query parameters
		filters := repository.LogFilters{
			ProjectId: ctx.Query("project_id"),
			Host:      ctx.Query("host"),
			Level:     ctx.Query("level"),
			Status:    ctx.Query("status"),
		}

		// Parse pagination parameters
		limitStr := ctx.DefaultQuery("limit", "100")
		offsetStr := ctx.DefaultQuery("offset", "0")

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			limit = 100
		}
		if limit > 1000 {
			limit = 1000 // Cap at 1000 for performance
		}

		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			offset = 0
		}

		filters.Limit = limit
		filters.Offset = offset

		// Parse date filters
		if startDateStr := ctx.Query("start_date"); startDateStr != "" {
			if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
				filters.StartDate = &startDate
			} else {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use RFC3339 format (e.g., 2023-01-01T00:00:00Z)"})
				return
			}
		}

		if endDateStr := ctx.Query("end_date"); endDateStr != "" {
			if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
				filters.EndDate = &endDate
			} else {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use RFC3339 format (e.g., 2023-01-01T23:59:59Z)"})
				return
			}
		}

		logs, totalCount, err := logRepo.GetWithFilters(ctx, filters)
		if err != nil {
			log.Printf("Error retrieving logs: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve logs"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"logs":        logs,
			"total_count": totalCount,
			"limit":       limit,
			"offset":      offset,
		})
	}
}

// RetrieveLogHandler retrieves a specific log by ID
// @Summary Get a log by ID
// @Description Retrieve a specific log entry by its ID
// @Tags logs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Log ID"
// @Success 200 {object} gin.H "Log retrieved successfully"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 404 {object} gin.H "Log not found"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /api/logs/{id} [get]
func RetrieveLogHandler(logRepo *repository.LogRepo) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")

		logEntry, err := logRepo.GetByID(ctx, id)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Log not found"})
			return
		}

		ctx.JSON(http.StatusOK, logEntry)
	}
}


// StreamLogsHandler streams logs in real-time using Server-Sent Events
// @Summary Stream logs in real-time
// @Description Stream logs in real-time using Server-Sent Events (SSE)
// @Tags logs
// @Accept json
// @Produce text/plain
// @Security BearerAuth
// @Param project_id query string false "Filter by project ID"
// @Param host query string false "Filter by host"
// @Param level query string false "Filter by log level (info, warning, error, debug)"
// @Param status query string false "Filter by status"
// @Success 200 {string} string "Log stream"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /api/logs/stream [get]
func StreamLogsHandler(logRepo *repository.LogRepo) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Set headers for Server-Sent Events
		ctx.Header("Content-Type", "text/event-stream")
		ctx.Header("Cache-Control", "no-cache")
		ctx.Header("Connection", "keep-alive")
		ctx.Header("Access-Control-Allow-Origin", "*")

		// Parse filters
		filters := repository.LogFilters{
			ProjectId: ctx.Query("project_id"),
			Host:      ctx.Query("host"),
			Level:     ctx.Query("level"),
			Status:    ctx.Query("status"),
		}

		// Start streaming logs
		logStream, err := logRepo.StreamLogs(ctx, filters)
		if err != nil {
			log.Printf("Error starting log stream: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start log stream"})
			return
		}
		defer logStream.Close()

		// Stream logs to client
		for {
			select {
			case <-ctx.Request.Context().Done():
				return
			case logEntry := <-logStream.Channel():
				if logEntry != nil {
					ctx.SSEvent("log", logEntry)
					ctx.Writer.Flush()
				}
			}
		}
	}
}
