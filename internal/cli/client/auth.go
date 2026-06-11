package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// RequestEmailOTP sends an OTP to the given email address and returns whether
// the account requires TOTP instead of an email code.
// Corresponds to POST /v1/auth/login/email.
func (c *Client) RequestEmailOTP(ctx context.Context, email string) (requireTotp bool, err error) {
	q := url.Values{"client": {"cli"}}
	resp, err := c.do(ctx, http.MethodPost, "/auth/login/email", q, LoginEmailRequest{Email: email})
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return false, readAPIError(resp)
	}

	var body struct {
		Data struct {
			RequireTotp bool `json:"requireTotp"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	return body.Data.RequireTotp, nil
}

// VerifyEmailOTP submits the OTP code and returns the session cookie on success.
// Corresponds to POST /v1/auth/login/email/verify.
func (c *Client) VerifyEmailOTP(ctx context.Context, email, code string) (string, error) {
	resp, err := c.do(ctx, http.MethodPost, "/auth/login/email/verify", nil, VerifyEmailRequest{
		Email: email,
		Code:  code,
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", readAPIError(resp)
	}

	// The backend sets a session cookie on successful login.
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "session" || cookie.Name == "dployr_session" {
			return cookie.String(), nil
		}
	}

	// Fall back to raw Set-Cookie header if no named match.
	if raw := resp.Header.Get("Set-Cookie"); raw != "" {
		return raw, nil
	}

	return "", fmt.Errorf("login succeeded but no session cookie was returned")
}

// Logout invalidates the current session.
// Corresponds to GET /v1/auth/logout.
func (c *Client) Logout(ctx context.Context) error {
	resp, err := c.do(ctx, http.MethodGet, "/auth/logout", nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	return readAPIError(resp)
}

type meData struct {
	User Me `json:"user"`
}

// Me returns the current authenticated user's profile.
// Corresponds to GET /v1/users/me.
func (c *Client) Me(ctx context.Context) (Me, error) {
	r, err := get[meData](ctx, c, "/users/me", nil)
	if err != nil {
		return Me{}, err
	}
	return r.User, nil
}
