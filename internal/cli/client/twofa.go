package client

import (
	"context"
	"net/http"
)

// TwoFAStatus mirrors GET /v1/auth/2fa/status.
type TwoFAStatus struct {
	Method               string `json:"method"`
	TotpEnabled          bool   `json:"totpEnabled"`
	BackupCodesRemaining int    `json:"backupCodesRemaining"`
}

// Get2FAStatus returns the current user's 2FA configuration.
func (c *Client) Get2FAStatus(ctx context.Context) (TwoFAStatus, error) {
	return get[TwoFAStatus](ctx, c, "/auth/2fa/status", nil)
}

// Verify2FA submits a TOTP code (or backup code) and marks the session as 2FA-verified.
// Corresponds to POST /v1/auth/2fa/verify.
func (c *Client) Verify2FA(ctx context.Context, code string) error {
	resp, err := c.do(ctx, http.MethodPost, "/auth/2fa/verify", nil, map[string]string{"code": code})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return readAPIError(resp)
}
