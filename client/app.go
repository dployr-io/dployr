package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
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
	Image         string    `json:"image,omitempty"`
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
	User       *User     `json:"user"`
	Status     string    `json:"status,omitempty"`
}

type Log struct {
	Host      string    `json:"host"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func authUrl() string {
	endpoint := os.Getenv("AUTH_ENDPOINT")
	if endpoint == "" {
		panic("AUTH_ENDPOINT environment variable is not set")
	}

	return endpoint
}

func (a *App) SignIn(provider string) AuthResponse {
	authURL := authUrl() + "/auth/" + provider

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
	resp, err := http.Post(authUrl()+"/v1/logout", "", nil)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

func (a *App) GetCurrentUser() *User {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", authUrl()+"/api/auth/session", nil)

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
			User:      a.GetCurrentUser(),
			Status:    "success",
		},
		{
			ID:        "KhddGNCdf",
			Branch:    "develop",
			Duration:  90,
			Message:   "Second deployment",
			CreatedAt: time.Now(),
			User:      a.GetCurrentUser(),
			Status:    "failed",
		},
	}
}

func (a *App) GetLogs() []Log {
	return []Log{
		{
			Host:      "https://api.example.com",
			Message:   "Updated user model to include image field",
			Status:    "GET 200",
			CreatedAt: time.Now(),
		},
		{
			Host:      "https://api.example.com",
			Message:   "Temporary redirect to new endpoint",
			Status:    "POST 307",
			CreatedAt: time.Now().Add(2 * time.Minute),
		},
		{
			Host:      "https://api.example.com",
			Message:   "Deployment failed due to timeout",
			Status:    "GET 500",
			CreatedAt: time.Now().Add(3 * time.Minute),
		},
	}
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
