package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// App struct
type App struct {
	ctx context.Context
}

type User struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	Avatar         string   `json:"avatar,omitempty"`
	EmailVerified bool      `json:"emailVerified"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type Project struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	Icon        string    `json:"icon"`
	Date        time.Time `json:"date"`
	Provider    string    `json:"provider"`
}

type Domain struct {
	Id 					string `json:"id"`
	Subdomain			string `json:"subdomain"`
    Provider            string `json:"provider"`
    AutoSetupAvailable  bool   `json:"auto_setup_available"`
    ManualRecords       string `json:"manual_records,omitempty"`
	Verified            bool `json:"verified"`
	UpdatedAt 		    time.Time `json:"updatedAt"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	err := godotenv.Load()
	if err != nil {
		panic("Failed to load .env file")
	}
	return &App{}
}

type AuthResponse struct {
	Success bool
	User    *User
	Error   string
}

type Deployment struct {
	ID         string    `json:"id"`
	CommitHash string    `json:"commitHash"`
	Branch     string    `json:"branch"`
	Duration   int       `json:"duration"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"createdAt"`
	Status     string    `json:"status,omitempty"`
}

type LogEntry struct {
	Id        string    `json:"id"`
	Host      string    `json:"host"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	Level     string    `json:"level"`
	CreatedAt time.Time `json:"createdAt"`
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func getBaseUrl() string {
	endpoint := os.Getenv("BASE_URL")
	if endpoint == "" {
		panic("BASE_URL environment variable is not set")
	}

	return endpoint
}

func (a *App) SignIn(provider string) AuthResponse {
	authURL := getBaseUrl() + "/auth/" + provider

	err := a.openBrowser(authURL)
	if err != nil {
		return AuthResponse{Success: false, Error: err.Error()}
	}

	return AuthResponse{Success: true}
}

func (a *App) openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}

	fmt.Printf("Attempting to open: %s\n", url)
	fmt.Printf("Command: %s %v\n", cmd, args)

	err := exec.Command(cmd, args...).Start()
	if err != nil {
		fmt.Printf("Browser open failed: %v\n", err)
	}
	return err
}

func (a *App) SignOut() bool {
	resp, err := http.Post(getBaseUrl()+"/v1/logout", "", nil)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

func (a *App) GetCurrentUser() *User {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", getBaseUrl()+"/api/auth/session", nil)

	// Forward stored session cookie
	if sessionCookie := a.getStoredSessionCookie(); sessionCookie != "" {
		req.Header.Set("Cookie", sessionCookie)
	}

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()

	var sessionResp struct {
		User *User `json:"user"`
	}

	json.NewDecoder(resp.Body).Decode(&sessionResp)
	return sessionResp.User
}

func (a *App) StoreSession(token string) {
	os.WriteFile("session.cookie", []byte(token), 0600)
}

func (a *App) getStoredSessionCookie() string {
	token, err := os.ReadFile("session.cookie")
	if err != nil {
		return ""
	}
	return "better-auth.session_token=" + string(token)
}

func (a *App) GetDeployments() []Deployment {
	return []Deployment{
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

func (a *App) GetLogs() []LogEntry {
	var logs []LogEntry
	
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
		// Progressive time going backwards, max 30 days
		minutesBack := rand.Intn(43200) // 30 days * 24 hours * 60 minutes
		logTime := baseTime.Add(-time.Duration(minutesBack) * time.Minute)
		
		status := statuses[rand.Intn(len(statuses))]
		var level string
		
		// Determine level based on status code
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
		
		logs = append(logs, LogEntry{
			Id:        fmt.Sprintf("717172%04d", 8220+i),
			Level:     level,
			Host:      hosts[rand.Intn(len(hosts))],
			Message:   messages[rand.Intn(len(messages))],
			Status:    status,
			CreatedAt: logTime,
		})
	}
	
	// Sort by CreatedAt descending (newest first)
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].CreatedAt.After(logs[j].CreatedAt)
	})
	
	return logs
}

func (a *App) GetProjects() []Project {
	return []Project{
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

// HTTP GET request example
func (a *App) FetchData(url string) (string, error) {
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return "", err
    }
    
    res, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err
    }
    defer res.Body.Close()
    
    body, err := io.ReadAll(res.Body)
    if err != nil {
        return "", err
    }
    
    return string(body), nil
}

// HTTP POST request example
func (a *App) PostData(url string, data interface{}) (string, error) {
    jsonData, _ := json.Marshal(data)
    
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return "", err
    }
    req.Header.Set("Content-Type", "application/json")
    
    res, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err
    }
    defer res.Body.Close()
    
    body, err := io.ReadAll(res.Body)
    return string(body), err
}

// HTTP PUT request example
func (a *App) UpdateData(url string, data interface{}) (string, error) {
    jsonData, _ := json.Marshal(data)
    
    req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return "", err
    }
    req.Header.Set("Content-Type", "application/json")
    
    res, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err
    }
    defer res.Body.Close()
    
    body, err := io.ReadAll(res.Body)
    return string(body), err
}

// HTTP DELETE request example
func (a *App) DeleteData(url string) (string, error) {
    req, err := http.NewRequest("DELETE", url, nil)
    if err != nil {
        return "", err
    }
    
    res, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err
    }
    defer res.Body.Close()
    
    body, err := io.ReadAll(res.Body)
    return string(body), err
}

func (a *App) AddDomain(domain string, projectID string) (Domain, error) {
    // Get project details
	_, err := a.PostData(getBaseUrl() + "/foo/bar",  map[string]interface{}{
		domain: "foo.bar",
	})

	// Network simulation
	time.Sleep(3 * time.Second)

	res := Domain{
		Provider: "cloudflare",
		AutoSetupAvailable: true,
		ManualRecords: generateManualInstructions(domain, "202.121.80.311"),
	}

    return res, err
}

func generateManualInstructions(domain, serverIP string) string {
    return fmt.Sprintf(`
A Record:
Name: @
Value: %s
TTL: 300

CNAME Record:
Name: www
Value: %s
TTL: 300
`, serverIP, domain)
}

func (a *App) GetDomains() []Domain {
	return []Domain{
		{
			Id: 	"39189134002340941",
			Subdomain: "foo.bar",
			Provider: "namecheap",
			AutoSetupAvailable: true,
			Verified: false,
			UpdatedAt: time.Now(),
		},
		{
			Id: 	"39189134002340940",
			Subdomain: "29500390932930390332.dployr.io",
			Provider: "cloudflare",
			AutoSetupAvailable: true,
			Verified: true,
			UpdatedAt: time.Now(),
		},
	}
}