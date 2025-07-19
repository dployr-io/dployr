package deployments

import (
	"fmt"
	"log"
	"net/http"

	"dployr.io/pkg/api/middleware"
	"dployr.io/pkg/models"
	"dployr.io/pkg/repository"
	"github.com/gin-gonic/gin"
)

type CreateDeploymentRequest struct {
	ProjectId  string `json:"project_id" binding:"required"`
	CommitHash string `json:"commit_hash" binding:"required"`
	Branch     string `json:"branch" binding:"required"`
	Duration   int    `json:"duration"`
	Message    string `json:"message"`
	Status     string `json:"status"`
}

// CreateDeploymentHandler creates a new deployment
// @Summary Create a new deployment
// @Description Create a new deployment for a project with the provided configuration
// @Tags deployments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateDeploymentRequest true "Deployment creation request"
// @Success 201 {object} gin.H "Deployment created successfully"
// @Failure 400 {object} gin.H "Invalid request"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 429 {object} gin.H "Too many requests"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /api/deployments [post]
func CreateDeploymentHandler(deploymentRepo *repository.DeploymentRepo, rl *middleware.RateLimiter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		clientIP := ctx.ClientIP()

		var req CreateDeploymentRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Rate limiting
		if !rl.IsAllowed(fmt.Sprintf("%s-create-deployment", clientIP)) {
			ctx.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please wait before attempting to create deployment.",
			})
			return
		}

		deployment := models.Deployment{
			CommitHash: req.CommitHash,
			Branch:     req.Branch,
			Message:    req.Message,
		}

		if err := deploymentRepo.Create(ctx, deployment); err != nil {
			log.Printf("Error creating deployment: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create deployment"})
			return
		}

		ctx.JSON(http.StatusCreated, deployment)
	}
}

// RetrieveDeploymentsHandler retrieves all deployments
// @Summary Get all deployments
// @Description Retrieve a list of all deployments for the authenticated user
// @Tags deployments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param project_id query string false "Filter by project ID"
// @Success 200 {object} gin.H "Deployments retrieved successfully"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /api/deployments [get]
func RetrieveDeploymentsHandler(deploymentRepo *repository.DeploymentRepo) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		projectId := ctx.Query("project_id")

		var deployments []models.Deployment
		var err error

		if projectId != "" {
			deployments, err = deploymentRepo.GetByProjectID(ctx, projectId)
		} else {
			deployments, err = deploymentRepo.GetAll(ctx)
		}

		if err != nil {
			log.Printf("Error retrieving deployments: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve deployments"})
			return
		}

		ctx.JSON(http.StatusOK, deployments)
	}
}

// RetrieveDeploymentHandler retrieves a specific deployment by commit_hash
// @Summary Get a deployment by commit_hash
// @Description Retrieve a specific deployment by its commit_hash
// @Tags deployments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param commit_hash path string true "Deployment commit_hash"
// @Success 200 {object} gin.H "Deployment retrieved successfully"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 404 {object} gin.H "Deployment not found"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /api/deployments/{commit_hash} [get]
func RetrieveDeploymentHandler(deploymentRepo *repository.DeploymentRepo) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		commit_hash := ctx.Param("commit_hash")

		deployment, err := deploymentRepo.GetByCommitHash(ctx, commit_hash)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Deployment not found"})
			return
		}

		ctx.JSON(http.StatusOK, deployment)
	}
}

// RollbackDeploymentHandler rolls back a deployment
// @Summary Rollback a deployment
// @Description Rollback a deployment by specified number of commits or 1 commit by default
// @Tags deployments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param commit_hash path string true "Deployment commit_hash"
// @Param request body RollbackDeploymentRequest false "Rollback request with optional steps"
// @Success 200 {object} gin.H "Deployment rolled back successfully"
// @Failure 400 {object} gin.H "Invalid request"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 404 {object} gin.H "Deployment not found"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /api/deployments/{commit_hash}/rollback [post]
// func RollbackDeploymentHandler(deploymentRepo *repository.DeploymentRepo, rl *middleware.RateLimiter) gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		clientIP := ctx.ClientIP()
// 		commitHash := ctx.Param("commit_hash")

// 		var req RollbackDeploymentRequest
// 		if err := ctx.ShouldBindJSON(&req); err != nil {
// 			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}

// 		// Rate limiting
// 		if !rl.IsAllowed(fmt.Sprintf("%s-rollback-deployment", clientIP)) {
// 			ctx.JSON(http.StatusTooManyRequests, gin.H{
// 				"error": "Too many requests. Please wait before attempting to rollback deployment.",
// 			})
// 			return
// 		}

// 		// Default to 1 step if not specified
// 		steps := req.Steps
// 		if steps <= 0 {
// 			steps = 1
// 		}

// 		// Verify the deployment exists
// 		currentDeployment, err := deploymentRepo.GetByCommitHash(ctx, commitHash)
// 		if err != nil {
// 			ctx.JSON(http.StatusNotFound, gin.H{"error": "Deployment not found"})
// 			return
// 		}

// 		// Get the deployment to rollback to
// 		targetDeployment, err := deploymentRepo.GetPreviousDeployment(ctx, currentDeployment.ProjectID, commitHash, steps)
// 		if err != nil {
// 			ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Cannot find deployment %d steps back", steps)})
// 			return
// 		}

// 		// Perform the rollback
// 		rollbackResult, err := deploymentRepo.Rollback(ctx, currentDeployment.ProjectID, targetDeployment.CommitHash, steps)
// 		if err != nil {
// 			log.Printf("Error rolling back deployment: %v", err)
// 			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to rollback deployment"})
// 			return
// 		}

// 		ctx.JSON(http.StatusOK, gin.H{
// 			"message":           "Deployment rolled back successfully",
// 			"from_commit":       commitHash,
// 			"to_commit":         targetDeployment.CommitHash,
// 			"steps_back":        steps,
// 			"rollback_result":   rollbackResult,
// 		})
// 	}
// }
