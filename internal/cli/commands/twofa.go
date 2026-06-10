package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/dployr-io/dployr/internal/cli/client"
)

// withTwoFA calls fn and, if the server responds with ErrTwoFARequired, prompts
// the user interactively for their TOTP code, verifies it, then retries fn once.
func withTwoFA(ctx context.Context, d *deps, fn func() error) error {
	err := fn()
	if err == nil {
		return nil
	}
	if !errors.Is(err, client.ErrTwoFARequired{}) {
		return err
	}

	fmt.Print("2FA verification required. Enter your authenticator code (or backup code): ")
	code, readErr := readLine()
	if readErr != nil {
		return fmt.Errorf("read 2FA code: %w", readErr)
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return fmt.Errorf("2FA code is required")
	}

	if verifyErr := d.client.Verify2FA(ctx, code); verifyErr != nil {
		return fmt.Errorf("2FA verification failed: %w", verifyErr)
	}

	return fn()
}
