package main

import (
	"bufio"
	"bytes"
	"dployr/pkg/core/deploy"
	"dployr/pkg/core/proxy"
	"dployr/pkg/core/utils"
	"dployr/pkg/shared"
	"dployr/pkg/store"
	"dployr/pkg/version"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

func getServiceManagerText() string {
	switch runtime.GOOS {
	case "linux":
		return "systemd on Linux"
	case "darwin":
		return "launchd on macOS"
	case "windows":
		return "NSSM on Windows"
	default:
		return "systemd on Linux, launchd on macOS, or NSSM on Windows"
	}
}

func main() {
	cfg, err := shared.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	addr := fmt.Sprintf("http://%s:%d", cfg.Address, cfg.Port)

	rootCmd := &cobra.Command{
		Use:   "dployr",
		Short: "dployr - your app, your server, your rules!",
		Long: `
Your app, your server, your rules!

Turn any machine into a deployment platform. Deploy applications from Git repositories or Docker images with automatic reverse proxy, SSL certificates, and service management.

` + "dployr" + ` gives developers a self-hosted alternative to managed platforms.  
It combines a lightweight daemon, a CLI client, and powerful integrations to automate deployment pipelines across operating systems.

---

` + "dployr" + ` consists of two main components:

- dployr — Command-line client  
- dployrd — Background daemon that handles deployment execution, service management, and API endpoints

- SQLite for persistence  
- Caddy for automatic HTTPS and reverse proxy  
- ` + getServiceManagerText() + ` for service management  
All components are written in Go and packaged as standalone binaries.`,
	}

	// version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "show version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			info := version.Get()

			jsonFlag, _ := cmd.Flags().GetBool("json")
			if jsonFlag {
				data, err := json.Marshal(info)
				if err != nil {
					return err
				}
				fmt.Println(string(data))
			} else {
				fmt.Println(info.String())
			}
			return nil
		},
	}
	versionCmd.Flags().BoolP("json", "j", false, "output in JSON format")
	rootCmd.AddCommand(versionCmd)

	// login
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "authenticate your account",
		Long: `Authenticate your dployr account and save the token locally.

For regular users:
  dployr login --email user@company.com

For first-time owner registration (requires secret key):
  dployr login --email admin@company.com --secret your-secret-key`,
		RunE: func(cmd *cobra.Command, args []string) error {
			email, _ := cmd.Flags().GetString("email")
			expiry, _ := cmd.Flags().GetString("expiry")
			secret, _ := cmd.Flags().GetString("secret")

			if email == "" {
				return fmt.Errorf("email is required")
			}

			reqBody := map[string]string{
				"email":    email,
				"lifespan": expiry,
			}

			if secret != "" {
				reqBody["secret"] = secret
			} else {
				// Retrieve current user
				if username := os.Getenv("USER"); username != "" {
					reqBody["username"] = username
				} else if username := os.Getenv("USERNAME"); username != "" {
					reqBody["username"] = username
				}
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
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
				ExpiresAt    string `json:"expires_at"`
				User         string `json:"user"`
				Role         string `json:"role"`
			}
			if err := json.NewDecoder(res.Body).Decode(&authResp); err != nil {
				return fmt.Errorf("failed to parse response: %v", err)
			}

			// save token to ~/.dployr/token.json
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("could not resolve user home directory: %v", err)
			}
			configPath := homeDir + "/.dployr/token.json"

			if err := os.Mkdir(homeDir+"/.dployr", 0700); err != nil && !os.IsExist(err) {
				return fmt.Errorf("could not create config directory: %v", err)
			}

			cfg := map[string]string{
				"access_token":  authResp.AccessToken,
				"refresh_token": authResp.RefreshToken,
				"expires_at":    authResp.ExpiresAt,
				"user":          authResp.User,
				"role":          authResp.Role,
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
	loginCmd.Flags().StringP("secret", "", "", "Secret key")
	loginCmd.Flags().StringP("server", "", addr, "Server URL")
	rootCmd.AddCommand(loginCmd)

	// create deployment
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
			if resp.StatusCode == http.StatusUnauthorized {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("invalid or expired token you need to sign in with 'dployr login' first: %s", string(body))
			}
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
	deployCmd.Flags().StringP("runtime", "r", "", "Runtime type: static, golang, php, python, nodejs, ruby, dotnet, java, docker, k3s, custom (required)")
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

	// list deployments
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

			if resp.StatusCode == http.StatusUnauthorized {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("invalid or expired token you need to sign in with 'dployr login' first: %s", string(body))
			}

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

	// logs command
	logsCmd := &cobra.Command{
		Use:   "logs [deployment-id]",
		Short: "view logs from deployment",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token, _ := shared.GetToken()

			var deploymentID string
			var deploymentName string
			if len(args) > 0 {
				deploymentID = args[0]
			} else {
				// Get the latest deployment if no ID provided
				r, err := http.NewRequest("GET", addr+"/deployments?limit=1", nil)
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

				if resp.StatusCode == http.StatusUnauthorized {
					body, _ := io.ReadAll(resp.Body)
					return fmt.Errorf("invalid or expired token you need to sign in with 'dployr login' first: %s", string(body))
				}

				if resp.StatusCode != http.StatusOK {
					body, _ := io.ReadAll(resp.Body)
					return fmt.Errorf("failed to get latest deployment with status %d: %s", resp.StatusCode, string(body))
				}

				var deployments []store.Deployment
				if err := json.NewDecoder(resp.Body).Decode(&deployments); err != nil {
					return fmt.Errorf("failed to parse deployments: %v", err)
				}

				if len(deployments) == 0 {
					fmt.Println("No deployments found. Create a deployment first with 'dployr deploy'")
					return nil
				}

				deploymentID = deployments[0].ID
				deploymentName = deployments[0].Blueprint.Name
				fmt.Printf("Showing logs for latest deployment: %s (%s)\n\n", deploymentID, deploymentName)
			}

			// Get log file path - logs are stored on disk, not via API
			var dataDir string
			if runtime.GOOS == "windows" {
				dataDir = filepath.Join(os.Getenv("PROGRAMDATA"), "dployr")
			} else {
				dataDir = "/var/lib/dployrd"
			}

			logPath := filepath.Join(dataDir, ".dployr", "logs", fmt.Sprintf("%s.log", strings.ToLower(deploymentID)))

			// Check if log file exists
			if _, err := os.Stat(logPath); os.IsNotExist(err) {
				fmt.Printf("No log file found for deployment: %s\n", deploymentID)
				fmt.Printf("Expected log file at: %s\n", logPath)
				fmt.Println("This could mean:")
				fmt.Println("  - The deployment hasn't started yet")
				fmt.Println("  - The deployment ID is invalid")
				fmt.Println("  - The deployment failed before logging began")
				return nil
			}

			// Open and read the log file
			file, err := os.Open(logPath)
			if err != nil {
				return fmt.Errorf("failed to open log file: %v", err)
			}
			defer file.Close()

			// Color codes for different log levels
			const (
				colorReset  = "\033[0m"
				colorRed    = "\033[31m"
				colorYellow = "\033[33m"
				colorBlue   = "\033[34m"
				colorGray   = "\033[90m"
				colorGreen  = "\033[32m"
			)

			lines, _ := cmd.Flags().GetInt("lines")

			scanner := bufio.NewScanner(file)
			var logLines []string

			// Read all lines first if we need to limit them
			if lines > 0 {
				for scanner.Scan() {
					logLines = append(logLines, scanner.Text())
				}

				// Show only the last N lines
				start := 0
				if len(logLines) > lines {
					start = len(logLines) - lines
				}

				for i := start; i < len(logLines); i++ {
					utils.PrintColoredLogLine(logLines[i], colorReset, colorRed, colorYellow, colorBlue, colorGray, colorGreen)
				}
			} else {
				// Print all lines with coloring
				for scanner.Scan() {
					line := scanner.Text()
					utils.PrintColoredLogLine(line, colorReset, colorRed, colorYellow, colorBlue, colorGray, colorGreen)
				}
			}

			if err := scanner.Err(); err != nil {
				return fmt.Errorf("error reading log file: %v", err)
			}

			return nil
		},
	}

	logsCmd.Flags().IntP("lines", "n", 0, "Number of lines to show from the end of the logs")
	rootCmd.AddCommand(logsCmd)

	servicesCmd := &cobra.Command{
		Use:   "services",
		Short: "manage services",
		Long:  "list and manage dployr services",
	}

	// list services
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

			if resp.StatusCode == http.StatusUnauthorized {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("invalid or expired token you need to sign in with 'dployr login' first: %s", string(body))
			}

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

	// get service
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

			if resp.StatusCode == http.StatusUnauthorized {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("invalid or expired token you need to sign in with 'dployr login' first: %s", string(body))
			}

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

	proxyCmd := &cobra.Command{
		Use:   "proxy",
		Short: "manage proxy service",
		Long:  "manage proxy configurations and service status",
	}

	// proxy status
	proxyStatusCmd := &cobra.Command{
		Use:   "status",
		Short: "get proxy service status",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, _ := shared.GetToken()

			r, err := http.NewRequest("GET", addr+"/proxy/status", nil)
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

			if resp.StatusCode == http.StatusUnauthorized {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("invalid or expired token you need to sign in with 'dployr login' first: %s", string(body))
			}

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("failed to get proxy status with status %d: %s", resp.StatusCode, string(body))
			}

			var status proxy.ProxyStatusResponse
			if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
				return fmt.Errorf("failed to parse response: %v", err)
			}

			fmt.Printf("%s", status.Status)
			return nil
		},
	}

	// proxy restart
	proxyRestartCmd := &cobra.Command{
		Use:   "restart",
		Short: "restart proxy service",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, _ := shared.GetToken()

			r, err := http.NewRequest("GET", addr+"/proxy/restart", nil)
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

			if resp.StatusCode == http.StatusUnauthorized {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("invalid or expired token you need to sign in with 'dployr login' first: %s", string(body))
			}

			if resp.StatusCode != http.StatusNoContent {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("failed to restart proxy with status %d: %s", resp.StatusCode, string(body))
			}

			fmt.Println("proxy service restarted successfully")
			return nil
		},
	}

	// proxy setup
	proxySetupCmd := &cobra.Command{
		Use:   "setup",
		Short: "setup proxy with app configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, _ := shared.GetToken()

			domain, _ := cmd.Flags().GetString("domain")
			upstream, _ := cmd.Flags().GetString("upstream")
			root, _ := cmd.Flags().GetString("root")
			template, _ := cmd.Flags().GetString("template")

			if domain == "" {
				return fmt.Errorf("domain is required")
			}
			if upstream == "" && template != "static" {
				return fmt.Errorf("upstream is required for non-static templates")
			}
			if template == "" {
				template = "reverse_proxy"
			}

			apps := map[string]proxy.App{
				domain: {
					Domain:   domain,
					Upstream: upstream,
					Root:     root,
					Template: proxy.TemplateType(template),
				},
			}

			jsonData, err := json.Marshal(apps)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %v", err)
			}

			r, err := http.NewRequest("POST", addr+"/proxy/setup", bytes.NewBuffer(jsonData))
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

			if resp.StatusCode == http.StatusUnauthorized {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("invalid or expired token you need to sign in with 'dployr login' first: %s", string(body))
			}

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("proxy setup failed with status %d: %s", resp.StatusCode, string(body))
			}

			fmt.Printf("proxy setup completed for domain: %s\n", domain)
			return nil
		},
	}

	proxySetupCmd.Flags().StringP("domain", "d", "", "Domain name (required)")
	proxySetupCmd.Flags().StringP("upstream", "u", "", "Upstream server address")
	proxySetupCmd.Flags().StringP("root", "r", "", "Root directory for static files")
	proxySetupCmd.Flags().StringP("template", "t", "reverse_proxy", "Template type: static, reverse_proxy, php_fastcgi")

	// proxy add
	proxyAddCmd := &cobra.Command{
		Use:   "add",
		Short: "add new app to proxy configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, _ := shared.GetToken()

			domain, _ := cmd.Flags().GetString("domain")
			upstream, _ := cmd.Flags().GetString("upstream")
			root, _ := cmd.Flags().GetString("root")
			template, _ := cmd.Flags().GetString("template")

			if domain == "" {
				return fmt.Errorf("domain is required")
			}
			if upstream == "" && template != "static" {
				return fmt.Errorf("upstream is required for non-static templates")
			}
			if template == "" {
				template = "reverse_proxy"
			}

			apps := []proxy.App{
				{
					Domain:   domain,
					Upstream: upstream,
					Root:     root,
					Template: proxy.TemplateType(template),
				},
			}

			data, err := json.Marshal(apps)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %v", err)
			}

			r, err := http.NewRequest("POST", addr+"/proxy/add", bytes.NewBuffer(data))
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

			if resp.StatusCode == http.StatusUnauthorized {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("invalid or expired token you need to sign in with 'dployr login' first: %s", string(body))
			}

			if resp.StatusCode != http.StatusNoContent {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("failed to add proxy app with status %d: %s", resp.StatusCode, string(body))
			}

			fmt.Printf("proxy app added successfully for domain: %s\n", domain)
			return nil
		},
	}

	proxyAddCmd.Flags().StringP("domain", "d", "", "Domain name (required)")
	proxyAddCmd.Flags().StringP("upstream", "u", "", "Upstream server address")
	proxyAddCmd.Flags().StringP("root", "r", "", "Root directory for static files")
	proxyAddCmd.Flags().StringP("template", "t", "reverse_proxy", "Template type: static, reverse_proxy, php_fastcgi")

	// proxy remove
	proxyRemoveCmd := &cobra.Command{
		Use:   "remove [domain...]",
		Short: "remove apps from proxy configuration",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token, _ := shared.GetToken()

			domains := args

			data, err := json.Marshal(domains)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %v", err)
			}

			r, err := http.NewRequest("POST", addr+"/proxy/remove", bytes.NewBuffer(data))
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

			if resp.StatusCode == http.StatusUnauthorized {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("invalid or expired token you need to sign in with 'dployr login' first: %s", string(body))
			}

			if resp.StatusCode != http.StatusNoContent {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("failed to remove proxy apps with status %d: %s", resp.StatusCode, string(body))
			}

			fmt.Printf("proxy apps removed successfully for domains: %v\n", domains)
			return nil
		},
	}

	// proxy list
	proxyListCmd := &cobra.Command{
		Use:   "list",
		Short: "list proxy app configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, _ := shared.GetToken()

			r, err := http.NewRequest("GET", addr+"/proxy/apps", nil)
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

			if resp.StatusCode == http.StatusUnauthorized {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("invalid or expired token you need to sign in with 'dployr login' first: %s", string(body))
			}

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("failed to list proxy apps with status %d: %s", resp.StatusCode, string(body))
			}

			var apps map[string]proxy.App
			if err := json.NewDecoder(resp.Body).Decode(&apps); err != nil {
				return fmt.Errorf("failed to parse response: %v", err)
			}

			if len(apps) == 0 {
				fmt.Println("no proxy apps configured")
				return nil
			}

			fmt.Printf("\nfound %d proxy app(s):\n\n", len(apps))
			for domain, app := range apps {
				fmt.Printf("  domain:   %s\n", domain)
				fmt.Printf("  template: %s\n", app.Template)
				if app.Upstream != "" {
					fmt.Printf("  upstream: %s\n", app.Upstream)
				}
				if app.Root != "" {
					fmt.Printf("  root:     %s\n", app.Root)
				}
				fmt.Println()
			}

			return nil
		},
	}

	proxyCmd.AddCommand(proxyStatusCmd)
	proxyCmd.AddCommand(proxyRestartCmd)
	proxyCmd.AddCommand(proxySetupCmd)
	proxyCmd.AddCommand(proxyAddCmd)
	proxyCmd.AddCommand(proxyRemoveCmd)
	proxyCmd.AddCommand(proxyListCmd)
	rootCmd.AddCommand(proxyCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
