// data/service.go
package data

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"time"

	"dployr.io/pkg/models"
)

type DataService struct{}

func NewDataService() *DataService {
	return &DataService{}
}

func (d *DataService) GetDeployments() []models.Deployment {
	return []models.Deployment{
		{
			CommitHash:        "GYjp7NCai",
			Branch:    "main",
			Duration:  120,
			Message:   "Initial deployment",
			CreatedAt: time.Now(),
			Status:    "success",
		},
		{
			CommitHash:        "KhddGNCdf",
			Branch:    "develop",
			Duration:  90,
			Message:   "Second deployment",
			CreatedAt: time.Now(),
			Status:    "failed",
		},
	}
}

func (d *DataService) GetLogs() []models.LogEntry {
	var logs []models.LogEntry

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

		logs = append(logs, models.LogEntry{
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

func (d *DataService) GetProjects(host, token string) ([]models.Project, error) {
	url := fmt.Sprintf("http://%s:7879/v1/api/projects", host)

	req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }

    req.Header.Set("Authorization", "Bearer "+token)
    resp, err := http.DefaultClient.Do(req)
	
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode > 299 {
        errBody, err := io.ReadAll(resp.Body)
        if err != nil {
            return nil, err
        }
        return nil, fmt.Errorf("verification failed (%d): %s", resp.StatusCode, errBody)
    }

    var result []models.Project
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return result, nil
}
