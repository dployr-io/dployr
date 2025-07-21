package data

import (
	"os"
	"os/exec"
	"runtime"

    "dployr.io/pkg/models"
)

func (d *DataService) SignIn(host, email, name string) (*models.MessageResponse, error) {
    resp, err := d.makeRequest("POST", "auth/request-code", host, "", nil, map[string]string{
        "name": name, 
        "email": email,
    })
	if err != nil {
		return nil, err
	}

	var result *models.MessageResponse
	return result, d.decodeResponse(resp, &result)
}

func (d *DataService) VerifyMagicCode(host, email, code string) (*models.AuthTokenPair, error) {
    resp, err := d.makeRequest("POST", "auth/verify-code", host, "", nil, map[string]string{
        "code": code, 
        "email": email,
    })
	if err != nil {
		return nil, err
	}

	var result *models.AuthTokenPair
	return result, d.decodeResponse(resp, &result)
}

func (d *DataService) RefreshToken(host, refreshToken string) (*models.AuthTokenPair, error) {
    resp, err := d.makeRequest("POST", "auth/refresh-token", host, "", nil, map[string]string{
        "refresh_token": refreshToken, 
    })
	if err != nil {
		return nil, err
	}

	var result *models.AuthTokenPair
	return result, d.decodeResponse(resp, &result)
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