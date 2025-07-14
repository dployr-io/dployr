// data/service.go
package data

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"dployr/core/types"
)

type DataService struct{}

func NewDataService() *DataService {
	return &DataService{}
}

func (d *DataService) GetDeployments() []types.Deployment {
	return []types.Deployment{
		{
			ID:        "GYjp7NCai",
			Branch:    "main",
			Duration:  120,
			Message:   "Initial deployment",
			CreatedAt: time.Now(),
			Status:    "success",
		},
		{
			ID:        "KhddGNCdf",
			Branch:    "develop",
			Duration:  90,
			Message:   "Second deployment",
			CreatedAt: time.Now(),
			Status:    "failed",
		},
	}
}

func (d *DataService) GetLogs() []types.LogEntry {
	var logs []types.LogEntry

	hosts := []string{
		"https://api.example.com",
		"https://auth.example.com",
		"https://cdn.example.com",
		"https://gateway.example.com",
		"https://admin.example.com",
		"https://webhooks.example.com",
	}

	messages := []string{
		"Updated user model to include image field",
		"Temporary redirect to new endpoint",
		"Deployment failed due to timeout",
		"Successfully authenticated user",
		"Cache miss for user profile",
		"Database connection established",
		"Rate limit exceeded for API key",
		"Image upload completed successfully",
		"Invalid token provided",
		"Resource not found in database",
		"Webhook delivery successful",
		"Payment processing completed",
		"Email notification sent",
		"File compression finished",
		"Session expired for user",
		"Health check passed",
		"Backup process initiated",
		"Configuration updated",
		"SSL certificate renewed",
		"Memory usage threshold exceeded",
	}

	statuses := []string{
		"GET 200", "POST 201", "PUT 200", "DELETE 204",
		"GET 404", "POST 400", "PUT 422", "DELETE 403",
		"GET 500", "POST 502", "PUT 503", "DELETE 500",
		"GET 307", "POST 301", "PUT 302",
	}

	baseTime := time.Now()

	for i := 0; i < 300; i++ {
		minutesBack := rand.Intn(43200)
		logTime := baseTime.Add(-time.Duration(minutesBack) * time.Minute)

		status := statuses[rand.Intn(len(statuses))]
		var level string

		statusCode := status[len(status)-3:]
		switch {
		case strings.HasPrefix(statusCode, "2"):
			level = "success"
		case strings.HasPrefix(statusCode, "3"):
			level = "warning"
		case strings.HasPrefix(statusCode, "4"), strings.HasPrefix(statusCode, "5"):
			level = "error"
		default:
			level = "info"
		}

		logs = append(logs, types.LogEntry{
			Id:        fmt.Sprintf("717172%04d", 8220+i),
			Level:     level,
			Host:      hosts[rand.Intn(len(hosts))],
			Message:   messages[rand.Intn(len(messages))],
			Status:    status,
			CreatedAt: logTime,
		})
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].CreatedAt.After(logs[j].CreatedAt)
	})

	return logs
}

func (d *DataService) GetProjects() []types.Project {
	return []types.Project{
		{
			Name:        "taxi-navigator",
			Description: "A web application for navigating taxi routes",
			URL:         "github.com/tommy/taxi-navigator",
			Icon:        "https://picsum.photos/200/200",
			Date:        time.Now(),
			Provider:    "github",
		},
		{
			Name:        "docker-study",
			Description: "A study project for Docker",
			URL:         "github.com/tommy/docker-study",
			Icon:        "https://picsum.photos/200/200",
			Date:        time.Now().AddDate(0, 0, -30),
			Provider:    "github",
		},
		{
			Name:        "ml-project",
			Description: "A machine learning project",
			URL:         "gitlab.com/tommy/ml-project",
			Icon:        "https://picsum.photos/200/200",
			Date:        time.Now().AddDate(0, 0, -70),
			Provider:    "gitlab",
		},
		{
			Name:        "Xmas-Frenzy",
			Description: "A festive project for the holiday season",
			URL:         "unity.com/tommy/xmas-frenzy",
			Icon:        "https://picsum.photos/200/200",
			Date:        time.Now().AddDate(0, 0, -210),
			Provider:    "unity",
		},
	}
}
