package projects

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"dployr.io/pkg/api/middleware"
	"dployr.io/pkg/models"
	"dployr.io/pkg/repository"
	"github.com/gin-gonic/gin"
)

type CreateProjectRequest struct {
	Name          string            `json:"name" binding:"required"`
	GitRepo       string            `json:"git_repo" binding:"required"`
	Domain        string            `json:"domain"`
	Environment   json.RawMessage `json:"environment,omitempty"`
	DeploymentURL string            `json:"deployment_url,omitempty"`
	Status        string            `json:"status"`
	HostConfigs   json.RawMessage    `json:"host_configs,omitempty"`
}

type UpdateProjectRequest struct {
	Name          string            `json:"name"`
	GitRepo       string            `json:"git_repo"`
	Domain        string            `json:"domain"`
	Environment   json.RawMessage `json:"environment,omitempty"`
	DeploymentURL string            `json:"deployment_url,omitempty"`
	Status        string            `json:"status"`
	HostConfigs   json.RawMessage    `json:"host_configs,omitempty"`
}

// CreateProjectHandler creates a new project
// @Summary Create a new project
// @Description Create a new deployment project with the provided configuration
// @Tags projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateProjectRequest true "Project creation request"
// @Success 201 {object} gin.H "Project created successfully"
// @Failure 400 {object} gin.H "Invalid request"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 429 {object} gin.H "Too many requests"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /api/projects [post]
func CreateProjectHandler(projectRepo *repository.ProjectRepo, rl *middleware.RateLimiter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
        clientIP := ctx.ClientIP()

		var req CreateProjectRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Rate limiting
		if !rl.IsAllowed(fmt.Sprintf("%s-create-project", clientIP)) {
			ctx.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please wait before attempting to create project.",
			})
			return
		}

		project := &models.Project{
			Name:          req.Name,
			GitRepo:       req.GitRepo,
		}

		if req.Environment != nil {
			var envData map[string]string
			json.Unmarshal(req.Environment, &envData)
			project.Environment = &models.JSON[any]{Data: envData}
		}
		
		if req.HostConfigs != nil {
			var envData map[string]string
			json.Unmarshal(req.Environment, &envData)
			project.Environment = &models.JSON[any]{Data: envData} 
		}

		if err := projectRepo.Create(ctx, project); err != nil {
			log.Printf("Error creating project: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
			return
		}

		ctx.JSON(http.StatusCreated, project)
	}
}

// RetrieveProjectsHandler retrieves all projects
// @Summary Get all projects
// @Description Retrieve a list of all projects for the authenticated user
// @Tags projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} gin.H "Projects retrieved successfully"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /api/projects [get]
func RetrieveProjectsHandler(projectRepo *repository.ProjectRepo) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		projects, err := projectRepo.GetAll(ctx)
		if err != nil {
			log.Printf("Error retrieving projects: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve projects"})
			return
		}

		ctx.JSON(http.StatusOK, projects)
	}
}

// RetrieveProjectHandler retrieves a specific project by ID
// @Summary Get a project by ID
// @Description Retrieve a specific project by its ID
// @Tags projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Success 200 {object} gin.H "Project retrieved successfully"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 404 {object} gin.H "Project not found"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /api/projects/{id} [get]
func RetrieveProjectHandler(projectRepo *repository.ProjectRepo) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")

		project, err := projectRepo.GetByID(ctx, id)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}

		ctx.JSON(http.StatusOK, project)
	}
}

// UpdateProjectHandler updates an existing project
// @Summary Update a project
// @Description Update an existing project with new configuration
// @Tags projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Param request body UpdateProjectRequest true "Project update request"
// @Success 200 {object} gin.H "Project updated successfully"
// @Failure 400 {object} gin.H "Invalid request"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 404 {object} gin.H "Project not found"
// @Failure 429 {object} gin.H "Too many requests"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /api/projects/{id} [put]
func UpdateProjectHandler(projectRepo *repository.ProjectRepo, rl *middleware.RateLimiter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		clientIP := ctx.ClientIP()
		id := ctx.Param("id")

		var req UpdateProjectRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Rate limiting
		if !rl.IsAllowed(fmt.Sprintf("%s-create-project", clientIP)) {
			ctx.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please wait before attempting to create project.",
			})
			return
		}

		project, err := projectRepo.GetByID(ctx, id)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}

		// Update fields
		if req.Name != "" {
			project.Name = req.Name
		}
		if req.GitRepo != "" {
			project.GitRepo = req.GitRepo
		}

		if req.Environment != nil {
			var envData map[string]string
			json.Unmarshal(req.Environment, &envData)
			project.Environment = &models.JSON[interface{}]{Data: envData}
		}
		
		if req.HostConfigs != nil {
			var envData map[string]string
			json.Unmarshal(req.HostConfigs, &envData)
			project.HostConfigs = &models.JSON[interface{}]{Data: envData}
		}


		if err := projectRepo.Update(ctx, project); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
			return
		}

		ctx.JSON(http.StatusOK, project)
	}
}

// DeleteProjectHandler deletes a project
// @Summary Deletes a project
// @Description Deletes a project from list of projects 
// @Tags projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} gin.H "Project deleted successfully"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /api/projects [get]
func DeleteProjectHandler(projectRepo *repository.ProjectRepo) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")

		_, err := projectRepo.GetByID(ctx, id)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}

		if err := projectRepo.Delete(ctx, id); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
	}
}