// cmd/dployr/main.go
package main

import (
	"bytes"
	"dployr/pkg/core"
	"dployr/pkg/shared"
	"dployr/pkg/store"
	"encoding/json"
	"fmt"
	"io"
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
			res, err := http.Post(addr+"/auth/request", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				return fmt.Errorf("failed to connect to server: %v", err)
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				return fmt.Errorf("login failed with status: %d", res.StatusCode)
			}

			// Parse response
			var authResp struct {
				Token     string `json:"token"`
				ExpiresAt string `json:"expires_at"`
				User      string `json:"user"`
			}
			if err := json.NewDecoder(res.Body).Decode(&authResp); err != nil {
				return fmt.Errorf("failed to parse response: %v", err)
			}

			// save token to ~/.dployr/config.json
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("could not resolve user home directory: %v", err)
			}
			configPath := homeDir + "/.dployr/config.json"

			if err := os.Mkdir(homeDir+"/.dployr", 0700); err != nil && !os.IsExist(err) {
				return fmt.Errorf("could not create config directory: %v", err)
			}

			cfg := map[string]string{
				"token":      authResp.Token,
				"expires_at": authResp.ExpiresAt,
				"user":       authResp.User,
			}
			configData, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return fmt.Errorf("could not marshal config: %v", err)
			}
			if err := os.WriteFile(configPath, configData, 0600); err != nil {
				return fmt.Errorf("could not write config file: %v", err)
			}
			fmt.Printf("token saved to %s\n", configPath)
			return nil
		},
	}
	loginCmd.Flags().StringP("email", "", "", "Your email")
	loginCmd.Flags().StringP("expiry", "", "", "Expiry time")
	loginCmd.Flags().StringP("server", "", addr, "Server URL")
	rootCmd.AddCommand(loginCmd)

	// Create deployment command
	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Create a new deployment",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load token from config
			configDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("could not resolve user home directory: %v", err)
			}
			configPath := configDir + "/.dployr/config.json"

			configData, err := os.ReadFile(configPath)
			if err != nil {
				return fmt.Errorf("could not read config file: %v. Please run 'dployr login' first", err)
			}

			var config map[string]string
			if err := json.Unmarshal(configData, &config); err != nil {
				return fmt.Errorf("could not parse config file: %v", err)
			}

			token, exists := config["token"]
			if !exists {
				return fmt.Errorf("no token found in config. Please run 'dployr login' first")
			}

			// Get command flags
			name, _ := cmd.Flags().GetString("name")
			description, _ := cmd.Flags().GetString("description")
			source, _ := cmd.Flags().GetString("source")
			runtime, _ := cmd.Flags().GetString("runtime")
			version, _ := cmd.Flags().GetString("version")
			runCmd, _ := cmd.Flags().GetString("run-cmd")
			buildCmd, _ := cmd.Flags().GetString("build-cmd")
			port, _ := cmd.Flags().GetInt("port")
			workingDir, _ := cmd.Flags().GetString("working-dir")
			staticDir, _ := cmd.Flags().GetString("static-dir")
			image, _ := cmd.Flags().GetString("image")
			domain, _ := cmd.Flags().GetString("domain")
			dnsProvider, _ := cmd.Flags().GetString("dns-provider")
			saveSpec, _ := cmd.Flags().GetBool("save-spec")
			envVars, _ := cmd.Flags().GetStringToString("env")

			// Remote flags
			remote, _ := cmd.Flags().GetString("remote")
			branch, _ := cmd.Flags().GetString("branch")
			commitHash, _ := cmd.Flags().GetString("commit-hash")

			// Validate required fields
			if name == "" {
				return fmt.Errorf("name is required")
			}
			if source == "" {
				return fmt.Errorf("source is required (remote or image)")
			}
			if runtime == "" {
				return fmt.Errorf("runtime is required")
			}

			// Build request payload
			req := core.DeployRequest{
				Name:        name,
				Description: description,
				Source:      source,
				Runtime: store.RuntimeObj{
					Type:    store.Runtime(runtime),
					Version: version,
				},
				RunCmd:      runCmd,
				BuildCmd:    buildCmd,
				Port:        port,
				WorkingDir:  workingDir,
				StaticDir:   staticDir,
				Image:       image,
				Domain:      domain,
				DNSProvider: dnsProvider,
				SaveSpec:    saveSpec,
				EnvVars:     envVars,
			}

			// Add remote if provided
			if source == "remote" {
				req.Remote = store.RemoteObj{
					Url:        remote,
					Branch:     branch,
					CommitHash: commitHash,
				}
			}

			jsonData, err := json.Marshal(req)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %v", err)
			}

			// Create HTTP request
			r, err := http.NewRequest("POST", addr+"/deployments", bytes.NewBuffer(jsonData))
			if err != nil {
				return fmt.Errorf("failed to create request: %v", err)
			}

			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Authorization", "Bearer "+token)

			// Make HTTP request
			client := &http.Client{}
			resp, err := client.Do(r)
			if err != nil {
				return fmt.Errorf("failed to connect to server: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusCreated {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("deployment creation failed with status %d: %s", resp.StatusCode, string(body))
			}

			// Parse response
			res := &core.DeployResponse{}
			if err := json.NewDecoder(resp.Body).Decode(res); err != nil {
				return fmt.Errorf("failed to parse response: %v", err)
			}

			fmt.Printf("   Deployment created successfully!\n")
			fmt.Printf("   ID: %s\n", res.ID)
			fmt.Printf("   Name: %s\n", res.Name)
			fmt.Printf("   Created: %s\n", res.CreatedAt)

			return nil
		},
	}

	// Add flags for deploy command
	deployCmd.Flags().StringP("name", "n", "", "Deployment name (required)")
	deployCmd.Flags().StringP("description", "d", "", "Deployment description")
	deployCmd.Flags().StringP("source", "s", "", "Source type: remote or image (required)")
	deployCmd.Flags().StringP("runtime", "r", "", "Runtime type: static, go, php, python, node-js, ruby, dotnet, java, docker, k3s, custom (required)")
	deployCmd.Flags().StringP("version", "", "", "Runtime version")
	deployCmd.Flags().StringP("run-cmd", "", "", "Command to run the application")
	deployCmd.Flags().StringP("build-cmd", "", "", "Command to build the application")
	deployCmd.Flags().IntP("port", "p", 0, "Port number for the application")
	deployCmd.Flags().StringP("working-dir", "", "", "Working directory")
	deployCmd.Flags().StringP("static-dir", "", "", "Static files directory")
	deployCmd.Flags().StringP("image", "", "", "Docker image name")
	deployCmd.Flags().StringP("domain", "", "", "Domain name")
	deployCmd.Flags().StringP("dns-provider", "", "", "DNS provider")
	deployCmd.Flags().BoolP("save-spec", "", false, "Save deployment specification")
	deployCmd.Flags().StringToStringP("env", "e", nil, "Environment variables (key=value pairs)")

	// Remote-specific flags
	deployCmd.Flags().StringP("remote", "", "", "Url to remote repository")
	deployCmd.Flags().StringP("branch", "", "", "Git branch")
	deployCmd.Flags().StringP("commit-hash", "", "", "Specific commit hash")

	rootCmd.AddCommand(deployCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
