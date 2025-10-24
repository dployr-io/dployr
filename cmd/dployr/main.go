// cmd/dployr/main.go
package main

import (
	"bytes"
	"dployr/pkg/shared"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	cfg, err := shared.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	addr := fmt.Sprintf("http://%s:%d", cfg.Address, cfg.Port)

	// Root command
	rootCmd := &cobra.Command{
		Use:   "dployr",
		Short: "dployr - Your app, your server, your rules!",
		Long:  `Manage deployments, blueprints, and runtimes for dployr environments.`,
	}

	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with the dployr server",
		RunE: func(cmd *cobra.Command, args []string) error {
			email, _ := cmd.Flags().GetString("email")
			expiry, _ := cmd.Flags().GetString("expiry")

			if email == "" {
				return fmt.Errorf("email is required")
			}

            if expiry == "" {
                expiry = "15m"
            }

			// Prepare request payload
			reqBody := map[string]string{
				"email":  email,
				"expiry": expiry,
			}
			jsonData, err := json.Marshal(reqBody)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %v", err)
			}

			// Make HTTP request to auth endpoint
			resp, err := http.Post(addr+"/auth/request", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				return fmt.Errorf("failed to connect to server: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("login failed with status: %d", resp.StatusCode)
			}

			// Parse response
			var authResp struct {
				Token     string `json:"token"`
				ExpiresAt string `json:"expires_at"`
				User      string `json:"user"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
				return fmt.Errorf("failed to parse response: %v", err)
			}

			// save token to ~/.dployr/config.json
			configDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("could not resolve user home directory: %v", err)
			}
			configPath := configDir + "/.dployr/config.json"

			if err := os.Mkdir(configDir+"/.dployr", 0700); err != nil && !os.IsExist(err) {
				return fmt.Errorf("could not create config directory: %v", err)
			}

			cfg := map[string]string{
				"token": authResp.Token,
				"expires_at": authResp.ExpiresAt,
				"user": authResp.User,
			}
			configData, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return fmt.Errorf("could not marshal config: %v", err)
			}
			if err := os.WriteFile(configPath, configData, 0600); err != nil {
				return fmt.Errorf("could not write config file: %v", err)
			}
			fmt.Printf("Token saved to %s\n", configPath)
			return nil
		},
	}
	loginCmd.Flags().StringP("email", "", "", "Your email")
	loginCmd.Flags().StringP("expiry", "", "", "Expiry time")
	loginCmd.Flags().StringP("server", "", addr, "Server URL")
	rootCmd.AddCommand(loginCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
