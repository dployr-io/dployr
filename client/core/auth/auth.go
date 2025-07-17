// auth/auth.go
package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"

    "dployr.io/pkg/models"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}


func (a *AuthService) SignIn(host, email, name, password, privateKey string) (*models.MessageResponse, error) {
    url := fmt.Sprintf("http://%s:7879/v1/auth/request-code", host)

    payload := map[string]string{"name": name, "email": email}
    b, err := json.Marshal(payload)
    if err != nil {
        return nil, err
    }

    resp, err := http.Post(url, "application/json", bytes.NewReader(b))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode > 299 {
        errBody, err := io.ReadAll(resp.Body)
        if err != nil {
            return nil, err
        }
        return nil, fmt.Errorf("sign-in failed (%d): %s", resp.StatusCode, errBody)
    }

    var result models.MessageResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return &result, nil
}

func (a *AuthService) VerifyMagicCode(host, email, code string) (*any, error) {
	url := fmt.Sprintf("http://%s:7879/v1/auth/verify-code", host)

	payload := map[string]string{"code": code, "email": email}
    b, err := json.Marshal(payload)
    if err != nil {
        return nil, err
    }

    resp, err := http.Post(url, "application/json", bytes.NewReader(b))
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

    var result any
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return &result, nil
}

func SignOut() {
	// Remove from local storage
}

func GetCurrentUser() *models.User {
	// Retrieve from localstorage

	var sessionResp struct {
		User *models.User `json:"user"`
	}

	//json.NewDecoder(resp.Body).Decode(&sessionResp)
	return sessionResp.User
}

func StoreSession(token string) {
	os.WriteFile("session.cookie", []byte(token), 0600)
}

func openBrowser(url string) error {
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

func getStoredSessionCookie() string {
	token, err := os.ReadFile("session.cookie")
	if err != nil {
		return ""
	}
	return "better-auth.session_token=" + string(token)
}