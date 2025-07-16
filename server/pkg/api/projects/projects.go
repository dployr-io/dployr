package projects

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"dployr.io/pkg/models"
	"dployr.io/pkg/repository"
)

type CreateProjectRequest struct {
	Name          string            `json:"name" binding:"required"`
	GitRepo       string            `json:"git_repo" binding:"required"`
	Domain        string            `json:"domain"`
	Provider      string            `json:"provider" binding:"required"`
	Environment   map[string]string `json:"environment,omitempty"`
	DeploymentURL string            `json:"deployment_url,omitempty"`
	Status        string            `json:"status"`
	HostConfigs   map[string]any    `json:"host_configs,omitempty"`
}

type UpdateProjectRequest struct {
	Name          string            `json:"name"`
	GitRepo       string            `json:"git_repo"`
	Domain        string            `json:"domain"`
	Provider      string            `json:"provider"`
	Environment   map[string]string `json:"environment,omitempty"`
	DeploymentURL string            `json:"deployment_url,omitempty"`
	Status        string            `json:"status"`
	HostConfigs   map[string]any    `json:"host_configs,omitempty"`
}

func CreateProjectHandler(projectRepo *repository.ProjectRepo) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req CreateProjectRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		project := &models.Project{
			Name:          req.Name,
			GitRepo:       req.GitRepo,
			Domain:        req.Domain,
			Provider:      req.Provider,
			Environment:   req.Environment,
			DeploymentURL: req.DeploymentURL,
			Status:        req.Status,
			HostConfigs:   req.HostConfigs,
		}

		if err := projectRepo.Create(ctx, project); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
			return
		}

		ctx.JSON(http.StatusCreated, project)
	}
}

func RetrieveProjectsHandler(projectRepo *repository.ProjectRepo) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		projects, err := projectRepo.GetAll(ctx)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve projects"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"projects": projects})
	}
}

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

func UpdateProjectHandler(projectRepo *repository.ProjectRepo) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")

		var req UpdateProjectRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		if req.Domain != "" {
			project.Domain = req.Domain
		}
		if req.Provider != "" {
			project.Provider = req.Provider
		}
		if req.Environment != nil {
			project.Environment = req.Environment
		}
		if req.DeploymentURL != "" {
			project.DeploymentURL = req.DeploymentURL
		}
		if req.Status != "" {
			project.Status = req.Status
		}
		if req.HostConfigs != nil {
			project.HostConfigs = req.HostConfigs
		}

		if err := projectRepo.Update(ctx, project); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
			return
		}

		ctx.JSON(http.StatusOK, project)
	}
}

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