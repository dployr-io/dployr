// cmd/dployr/main.go
package main

import (
	"bytes"
	"dployr/pkg/core/deploy"
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
		Short: "dployr - your app, your server, your rules!",
		Long:  `manage deployments, blueprints, and runtimes for dployr environments.`,
	}

	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "authenticate with the dployr server",
		RunE: func(cmd *cobra.Command, args []string) error {
			email, _ := cmd.Flags().GetString("email")
			expiry, _ := cmd.Flags().GetString("expiry")

			if email == "" {
				return fmt.Errorf("email is required")
			}
			if expiry == "" {
				expiry = "15m"
			}

			reqBody := map[string]string{
				"email":  email,
				"expiry": expiry,
			}
			jsonData, err := json.Marshal(reqBody)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %v", err)
			}

			res, err := http.Post(addr+"/auth/request", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				return fmt.Errorf("failed to connect to server: %v", err)
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				return fmt.Errorf("login failed with status: %d", res.StatusCode)
			}
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
		Short: "create a new deployment",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, _ := shared.GetToken()

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
			envVars, _ := cmd.Flags().GetStringToString("env")
			remote, _ := cmd.Flags().GetString("remote")
			branch, _ := cmd.Flags().GetString("branch")
			commitHash, _ := cmd.Flags().GetString("commit-hash")

			if name == "" {
				return fmt.Errorf("name is required")
			}
			if source == "" {
				return fmt.Errorf("source is required (remote or image)")
			}
			if runtime == "" {
				return fmt.Errorf("runtime is required")
			}

			req := deploy.DeployRequest{
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
				EnvVars:     envVars,
			}

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

			r, err := http.NewRequest("POST", addr+"/deployments", bytes.NewBuffer(jsonData))
			if err != nil {
				return fmt.Errorf("failed to create request: %v", err)
			}
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Authorization", "Bearer "+token)
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

			res := &deploy.DeployResponse{}
			if err := json.NewDecoder(resp.Body).Decode(res); err != nil {
				return fmt.Errorf("failed to parse response: %v", err)
			}

			fmt.Printf("   deployment created successfully!\n")
			fmt.Printf("   id: %s\n", res.ID)
			fmt.Printf("   name: %s\n", res.Name)
			fmt.Printf("   created: %s\n", res.CreatedAt)

			return nil
		},
	}

	deployCmd.Flags().StringP("name", "n", "", "Deployment name (required)")
	deployCmd.Flags().StringP("description", "d", "", "Deployment description")
	deployCmd.Flags().StringP("source", "s", "", "Source type: remote or image (required)")
	deployCmd.Flags().StringP("runtime", "r", "", "Runtime type: static, golang, php, python, node-js, ruby, dotnet, java, docker, k3s, custom (required)")
	deployCmd.Flags().StringP("version", "", "", "Runtime version")
	deployCmd.Flags().StringP("run-cmd", "", "", "Command to run the application")
	deployCmd.Flags().StringP("build-cmd", "", "", "Command to build the application")
	deployCmd.Flags().IntP("port", "p", 0, "Port number for the application")
	deployCmd.Flags().StringP("working-dir", "", "", "Working directory")
	deployCmd.Flags().StringP("static-dir", "", "", "Static files directory")
	deployCmd.Flags().StringP("image", "", "", "Docker image name")
	deployCmd.Flags().StringP("domain", "", "", "Domain name")
	deployCmd.Flags().StringP("dns-provider", "", "", "DNS provider")
	deployCmd.Flags().StringToStringP("env", "e", nil, "Environment variables (key=value pairs)")
	deployCmd.Flags().StringP("remote", "", "", "Url to remote repository")
	deployCmd.Flags().StringP("branch", "", "", "Git branch")
	deployCmd.Flags().StringP("commit-hash", "", "", "Specific commit hash")

	rootCmd.AddCommand(deployCmd)

	// List deployments command
	listDeploymentsCmd := &cobra.Command{
		Use:   "list",
		Short: "list previous deployments",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, _ := shared.GetToken()

			limit, _ := cmd.Flags().GetInt("limit")
			if limit <= 0 {
				limit = 10
			}

			r, err := http.NewRequest("GET", addr+"/deployments", nil)
			if err != nil {
				return fmt.Errorf("failed to create request: %v", err)
			}

			r.Header.Set("Authorization", "Bearer "+token)
			q := r.URL.Query()
			q.Add("limit", fmt.Sprintf("%d", limit))
			r.URL.RawQuery = q.Encode()
			client := &http.Client{}
			resp, err := client.Do(r)
			if err != nil {
				return fmt.Errorf("failed to connect to server: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("failed to list deployments with status %d: %s", resp.StatusCode, string(body))
			}

			var deployments []store.Deployment
			if err := json.NewDecoder(resp.Body).Decode(&deployments); err != nil {
				return fmt.Errorf("failed to parse response: %v", err)
			}

			if len(deployments) == 0 {
				fmt.Println("no deployments found")
				return nil
			}

			fmt.Printf("\nfound %d deployment(s):\n\n", len(deployments))
			for _, dep := range deployments {
				fmt.Printf("  id:       %s\n", dep.ID)
				fmt.Printf("  name:     %s\n", dep.Blueprint.Name)
				fmt.Printf("  status:   %s\n", dep.Status)
				fmt.Printf("  runtime:  %s\n", dep.Blueprint.Runtime.Type)
				if dep.Blueprint.Runtime.Version != "" {
					fmt.Printf("  version:  %s\n", dep.Blueprint.Runtime.Version)
				}
				fmt.Printf("  created:  %s\n", dep.CreatedAt.Format("2006-01-02 15:04:05"))
				fmt.Println()
			}

			return nil
		},
	}

	listDeploymentsCmd.Flags().IntP("limit", "l", 10, "Maximum number of deployments to show")
	rootCmd.AddCommand(listDeploymentsCmd)

	// Services command group
	servicesCmd := &cobra.Command{
		Use:   "services",
		Short: "manage services",
		Long:  "list and manage dployr services",
	}

	// List services command
	listServicesCmd := &cobra.Command{
		Use:   "list",
		Short: "list services",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, _ := shared.GetToken()

			limit, _ := cmd.Flags().GetInt("limit")
			if limit <= 0 {
				limit = 10
			}

			r, err := http.NewRequest("GET", addr+"/services", nil)
			if err != nil {
				return fmt.Errorf("failed to create request: %v", err)
			}

			r.Header.Set("Authorization", "Bearer "+token)
			q := r.URL.Query()
			q.Add("limit", fmt.Sprintf("%d", limit))
			r.URL.RawQuery = q.Encode()
			client := &http.Client{}
			resp, err := client.Do(r)
			if err != nil {
				return fmt.Errorf("failed to connect to server: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("failed to list services with status %d: %s", resp.StatusCode, string(body))
			}

			var services []store.Service
			if err := json.NewDecoder(resp.Body).Decode(&services); err != nil {
				return fmt.Errorf("failed to parse response: %v", err)
			}

			if len(services) == 0 {
				fmt.Println("no services found")
				return nil
			}

			fmt.Printf("\nfound %d service(s):\n\n", len(services))
			for _, svc := range services {
				fmt.Printf("  id:       %s\n", svc.ID)
				fmt.Printf("  name:     %s\n", svc.Name)
				fmt.Printf("  status:   %s\n", svc.Status)
				fmt.Printf("  type:     %s\n", svc.Description)
				if svc.Port > 0 {
					fmt.Printf("  port:     %d\n", svc.Port)
				}
				fmt.Printf("  created:  %s\n", svc.CreatedAt.Format("2006-01-02 15:04:05"))
				fmt.Println()
			}

			return nil
		},
	}

	listServicesCmd.Flags().IntP("limit", "l", 10, "Maximum number of services to show")
	servicesCmd.AddCommand(listServicesCmd)

	// Get service command
	getServiceCmd := &cobra.Command{
		Use:   "get [service-id]",
		Short: "get service details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token, _ := shared.GetToken()
			serviceID := args[0]

			r, err := http.NewRequest("GET", addr+"/services/"+serviceID, nil)
			if err != nil {
				return fmt.Errorf("failed to create request: %v", err)
			}

			r.Header.Set("Authorization", "Bearer "+token)
			client := &http.Client{}
			resp, err := client.Do(r)
			if err != nil {
				return fmt.Errorf("failed to connect to server: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("failed to get service with status %d: %s", resp.StatusCode, string(body))
			}

			var service store.Service
			if err := json.NewDecoder(resp.Body).Decode(&service); err != nil {
				return fmt.Errorf("failed to parse response: %v", err)
			}

			fmt.Printf("\nservice details:\n\n")
			fmt.Printf("  id:          %s\n", service.ID)
			fmt.Printf("  name:        %s\n", service.Name)
			fmt.Printf("  status:      %s\n", service.Status)
			if service.Port > 0 {
				fmt.Printf("  port:        %d\n", service.Port)
			}
			if service.Description != "" {
				fmt.Printf("  description: %s\n", service.Description)
			}
			fmt.Printf("  created:     %s\n", service.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("  updated:     %s\n", service.UpdatedAt.Format("2006-01-02 15:04:05"))

			return nil
		},
	}

	servicesCmd.AddCommand(getServiceCmd)
	rootCmd.AddCommand(servicesCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
