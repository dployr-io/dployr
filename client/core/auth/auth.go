// auth/auth.go
package auth

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"dployr/core/types"
)

type AuthService struct {
	baseURL string
}

func NewAuthService(baseURL string) *AuthService {
	return &AuthService{baseURL: baseURL}
}

func (a *AuthService) SignIn(provider string) types.AuthResponse {
	authURL := a.baseURL + "/auth/" + provider
	err := a.openBrowser(authURL)
	if err != nil {
		return types.AuthResponse{Success: false, Error: err.Error()}
	}
	return types.AuthResponse{Success: true}
}

func (a *AuthService) SignOut() bool {
	resp, err := http.Post(a.baseURL+"/v1/logout", "", nil)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func (a *AuthService) GetCurrentUser() *types.User {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", a.baseURL+"/api/auth/session", nil)

	if sessionCookie := a.getStoredSessionCookie(); sessionCookie != "" {
		req.Header.Set("Cookie", sessionCookie)
	}

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()

	var sessionResp struct {
		User *types.User `json:"user"`
	}

	json.NewDecoder(resp.Body).Decode(&sessionResp)
	return sessionResp.User
}

func (a *AuthService) StoreSession(token string) {
	os.WriteFile("session.cookie", []byte(token), 0600)
}

func (a *AuthService) openBrowser(url string) error {
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

	return exec.Command(cmd, args...).Start()
}

func (a *AuthService) getStoredSessionCookie() string {
	token, err := os.ReadFile("session.cookie")
	if err != nil {
		return ""
	}
	return "better-auth.session_token=" + string(token)
}