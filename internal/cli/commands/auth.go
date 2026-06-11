package commands

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dployr-io/dployr/internal/cli/config"
	"github.com/spf13/cobra"
)

func newAuthCmd(makeDeps makeDepsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "manage authentication",
	}
	cmd.AddCommand(newAuthLoginCmd(makeDeps))
	cmd.AddCommand(newAuthLogoutCmd(makeDeps))
	cmd.AddCommand(newAuthStatusCmd(makeDeps))
	cmd.AddCommand(newAuthOIDCCmd(makeDeps))
	cmd.AddCommand(newAuthTokensCmd(makeDeps))
	return cmd
}

func newAuthLoginCmd(makeDeps makeDepsFunc) *cobra.Command {
	var email string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "authenticate via email OTP",
		Long: `Authenticate with your dployr account using a one-time code sent to your email.

  dployr auth login --email you@company.com

For CI/CD environments use OIDC federation instead:

  dployr auth oidc --cluster <name>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if email == "" {
				return fmt.Errorf("--email is required")
			}
			return loginEmail(d, email)
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "your email address (required)")
	_ = cmd.MarkFlagRequired("email")
	return cmd
}

func loginEmail(d *deps, email string) error {
	ctx := context.Background()
	requireTotp, err := d.client.RequestEmailOTP(ctx, email)
	if err != nil {
		return fmt.Errorf("request OTP: %w", err)
	}

	if requireTotp {
		fmt.Print("enter your authenticator app code: ")
	} else {
		fmt.Printf("sending OTP to %s...\n", email)
		fmt.Print("enter the code from your email: ")
	}
	code, err := readLine()
	if err != nil {
		return fmt.Errorf("read code: %w", err)
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return fmt.Errorf("code is required")
	}

	cookie, err := d.client.VerifyEmailOTP(ctx, email, code)
	if err != nil {
		return fmt.Errorf("verify OTP: %w", err)
	}

	d.cfg.Auth = config.Auth{
		SessionCookie: cookie,
		UserEmail:     email,
		ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
	}
	if err := d.cfg.Save(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Printf("logged in as %s\n", email)
	return nil
}

func newAuthLogoutCmd(makeDeps makeDepsFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "revoke the current session",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if !d.cfg.IsAuthenticated() {
				fmt.Println("not logged in")
				return nil
			}

			_ = d.client.Logout(context.Background())

			d.cfg.ClearAuth()
			if err := d.cfg.Save(); err != nil {
				return fmt.Errorf("save config: %w", err)
			}
			fmt.Println("logged out")
			return nil
		},
	}
}

func newAuthStatusCmd(makeDeps makeDepsFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "show current authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}

			if !d.cfg.IsAuthenticated() {
				fmt.Println("not logged in")
				return nil
			}

			fmt.Printf("logged in as %s\n", d.cfg.Auth.UserEmail)
			return nil
		},
	}
}

func newAuthOIDCCmd(makeDeps makeDepsFunc) *cobra.Command {
	var clusterName string

	cmd := &cobra.Command{
		Use:   "oidc",
		Short: "authenticate using CI OIDC federation (GitHub Actions, GitLab CI, Bitbucket Pipelines)",
		Long: `Exchange a CI-issued OIDC token for a short-lived dployr session.

The OIDC token is auto-detected from the CI environment:

  GitHub Actions:  $ACTIONS_ID_TOKEN_REQUEST_URL / $ACTIONS_ID_TOKEN_REQUEST_TOKEN
  GitLab CI:       $CI_JOB_JWT_V2
  Bitbucket:       $BITBUCKET_STEP_OIDC_TOKEN

An OIDC binding must be pre-configured in the dployr dashboard linking your
repository and environment to a cluster. The session expires after 1 hour.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}

			token, provider, err := detectCIOIDCToken()
			if err != nil {
				return err
			}

			ctx := context.Background()
			sessionId, err := d.client.ExchangeOIDCToken(ctx, token)
			if err != nil {
				return fmt.Errorf("OIDC exchange failed: %w", err)
			}

			d.cfg.Auth = config.Auth{
				SessionCookie: "session=" + sessionId,
				ExpiresAt:     time.Now().Add(time.Hour),
			}
			if clusterName != "" {
				d.cfg.ActiveCluster = clusterName
			}
			if err := d.cfg.Save(); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Printf("authenticated via OIDC (%s)\n", provider)
			return nil
		},
	}

	cmd.Flags().StringVar(&clusterName, "cluster", "", "set active cluster after authentication")
	return cmd
}

// detectCIOIDCToken reads the CI OIDC token from the environment.
// For GitHub Actions, it fetches the token from the OIDC token endpoint.
func detectCIOIDCToken() (token, provider string, err error) {
	// GitLab CI
	if t := os.Getenv("CI_JOB_JWT_V2"); t != "" {
		return t, "gitlab", nil
	}

	// Bitbucket Pipelines
	if t := os.Getenv("BITBUCKET_STEP_OIDC_TOKEN"); t != "" {
		return t, "bitbucket", nil
	}

	// GitHub Actions — must request the token from the runtime API
	if reqURL := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL"); reqURL != "" {
		reqToken := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN")
		if reqToken == "" {
			return "", "", fmt.Errorf("ACTIONS_ID_TOKEN_REQUEST_TOKEN is not set — ensure id-token: write permission is granted")
		}
		t, fetchErr := fetchGitHubOIDCToken(reqURL, reqToken)
		if fetchErr != nil {
			return "", "", fmt.Errorf("fetch GitHub OIDC token: %w", fetchErr)
		}
		return t, "github", nil
	}

	return "", "", fmt.Errorf("no CI OIDC token found — set CI_JOB_JWT_V2 (GitLab), BITBUCKET_STEP_OIDC_TOKEN, or configure GitHub Actions id-token: write permissions")
}

func fetchGitHubOIDCToken(requestURL, requestToken string) (string, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, requestURL+"&audience=dployr", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+requestToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub OIDC token endpoint returned HTTP %d", resp.StatusCode)
	}

	var body struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	if body.Value == "" {
		return "", fmt.Errorf("GitHub OIDC token response contained no value")
	}
	return body.Value, nil
}

func readLine() (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("EOF on stdin")
}
