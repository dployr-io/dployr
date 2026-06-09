package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/dployr-io/dployr/internal/cli/output"
	"github.com/spf13/cobra"
)

func newAuthTokensCmd(makeDeps makeDepsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tokens",
		Short: "manage personal access tokens",
	}
	cmd.AddCommand(newTokensCreateCmd(makeDeps))
	cmd.AddCommand(newTokensListCmd(makeDeps))
	cmd.AddCommand(newTokensRevokeCmd(makeDeps))
	return cmd
}

func newTokensCreateCmd(makeDeps makeDepsFunc) *cobra.Command {
	var (
		name      string
		scopes    []string
		expiresIn int
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "create a personal access token",
		Long: `Create a scoped personal access token (dpat_).

The token is shown only once — save it immediately.

Available scopes:
  oidc:bind   Allow registering OIDC bindings (e.g. GitHub Actions bootstrap)

Example:
  dployr auth tokens create --name "github-actions" --scope oidc:bind`,
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}
			if len(scopes) == 0 {
				return fmt.Errorf("--scope is required (e.g. --scope oidc:bind)")
			}

			var exp *int
			if expiresIn > 0 {
				exp = &expiresIn
			}

			tok, err := d.client.CreateToken(context.Background(), name, scopes, exp)
			if err != nil {
				return err
			}

			if d.out.Format() == output.FormatJSON {
				return d.out.JSON(tok)
			}

			fmt.Printf("token created\n\n")
			fmt.Printf("  id:      %s\n", tok.ID)
			fmt.Printf("  name:    %s\n", tok.Name)
			fmt.Printf("  scopes:  %s\n", strings.Join(tok.Scopes, ", "))
			if tok.ExpiresAt != nil {
				fmt.Printf("  expires: %s\n", timeAgoPtr(tok.ExpiresAt))
			}
			fmt.Printf("\n  token: %s\n", tok.Token)
			fmt.Printf("\nStore this token securely — it will not be shown again.\n")
			fmt.Printf("Set it as DPLOYR_TOKEN in your CI secrets.\n")
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "token name (required)")
	cmd.Flags().StringArrayVar(&scopes, "scope", nil, "scope to grant (repeatable, e.g. --scope oidc:bind)")
	cmd.Flags().IntVar(&expiresIn, "expires-in", 0, "expiry in seconds from now (0 = no expiry)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newTokensListCmd(makeDeps makeDepsFunc) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list personal access tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}

			tokens, err := d.client.ListTokens(context.Background())
			if err != nil {
				return err
			}

			if d.out.Format() == output.FormatJSON {
				return d.out.JSON(tokens)
			}

			if len(tokens) == 0 {
				fmt.Println("no tokens found")
				return nil
			}

			rows := make([][]string, len(tokens))
			for i, t := range tokens {
				expires := "never"
				if t.ExpiresAt != nil {
					expires = timeAgoPtr(t.ExpiresAt)
				}
				lastUsed := "-"
				if t.LastUsedAt != nil {
					lastUsed = timeAgoPtr(t.LastUsedAt)
				}
				rows[i] = []string{t.ID, t.Name, strings.Join(t.Scopes, ","), expires, lastUsed, timeAgo(t.CreatedAt)}
			}
			d.out.Table([]string{"ID", "NAME", "SCOPES", "EXPIRES", "LAST USED", "CREATED"}, rows)
			return nil
		},
	}
}

func newTokensRevokeCmd(makeDeps makeDepsFunc) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "revoke <id>",
		Aliases: []string{"delete", "rm"},
		Short:   "revoke a personal access token",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}

			if !force {
				fmt.Printf("revoke token %s? this cannot be undone [y/N]: ", args[0])
				var confirm string
				fmt.Scanln(&confirm) //nolint:errcheck
				if confirm != "y" && confirm != "Y" {
					fmt.Println("aborted")
					return nil
				}
			}

			if err := d.client.RevokeToken(context.Background(), args[0]); err != nil {
				return err
			}
			fmt.Printf("token %s revoked\n", args[0])
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation prompt")
	return cmd
}
