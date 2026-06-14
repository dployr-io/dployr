package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/dployr-io/dployr/internal/cli/client"
	"github.com/dployr-io/dployr/internal/cli/output"
	"github.com/spf13/cobra"
)

func newDeploymentsCmd(makeDeps makeDepsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deployments",
		Short: "manage deployments",
	}

	cmd.AddCommand(newDeploymentsListCmd(makeDeps))
	cmd.AddCommand(newDeploymentsGetCmd(makeDeps))
	cmd.AddCommand(newDeploymentsCreateCmd(makeDeps))
	cmd.AddCommand(newDeploymentsDeleteCmd(makeDeps))
	return cmd
}

func newDeploymentsListCmd(makeDeps makeDepsFunc) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list deployments in the active cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}

			deployments, err := d.client.ListDeployments(context.Background(), limit)
			if err != nil {
				return err
			}

			if d.out.Format() == output.FormatJSON {
				return d.out.JSON(deployments)
			}

			if len(deployments) == 0 {
				fmt.Println("no deployments found")
				return nil
			}

			rows := make([][]string, len(deployments))
			for i, dep := range deployments {
				rt := dep.RuntimeType
				if dep.RuntimeVersion != "" {
					rt += "@" + dep.RuntimeVersion
				}
				rows[i] = []string{
					dep.Name,
					dep.Status,
					rt,
					dep.Source,
					timeAgo(dep.CreatedAt),
					timeAgo(dep.UpdatedAt),
				}
			}
			d.out.Table([]string{"NAME", "STATUS", "RUNTIME", "SOURCE", "CREATED", "UPDATED"}, rows)
			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 20, "maximum number of deployments to return")
	return cmd
}

func newDeploymentsGetCmd(makeDeps makeDepsFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "get deployment details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}

			dep, err := d.client.GetDeployment(context.Background(), args[0])
			if err != nil {
				return err
			}

			if d.out.Format() == output.FormatJSON {
				return d.out.JSON(dep)
			}

			d.out.Printf("id:          %s\n", dep.ID)
			d.out.Printf("name:        %s\n", dep.Name)
			d.out.Printf("status:      %s\n", dep.Status)
			d.out.Printf("source:      %s\n", dep.Source)
			d.out.Printf("runtime:     %s", dep.RuntimeType)
			if dep.RuntimeVersion != "" {
				d.out.Printf(" (%s)", dep.RuntimeVersion)
			}
			d.out.Print("")
			if dep.Domain != "" {
				d.out.Printf("domain:      %s\n", dep.Domain)
			}
			if dep.RemoteURL != "" {
				d.out.Printf("remote:      %s\n", dep.RemoteURL)
				if dep.RemoteBranch != "" {
					d.out.Printf("branch:      %s\n", dep.RemoteBranch)
				}
				if dep.RemoteCommitHash != "" {
					d.out.Printf("commit:      %s\n", shortHash(dep.RemoteCommitHash))
				}
			}
			d.out.Printf("created:     %s\n", timeAgo(dep.CreatedAt))
			if dep.FinishedAt != nil {
				d.out.Printf("finished:    %s\n", timeAgoPtr(dep.FinishedAt))
			}
			return nil
		},
	}
}

func newDeploymentsCreateCmd(makeDeps makeDepsFunc) *cobra.Command {
	var (
		name             string
		description      string
		serviceType      string
		source           string
		runtimeType      string
		runtimeVersion   string
		runCmd           string
		buildCmd         string
		port             int
		workingDir       string
		staticDir        string
		healthCheck      string
		image            string
		domain           string
		remoteURL        string
		remoteBranch     string
		remoteCommitHash string
		envVars          []string
		secrets          []string
		forceRebuild     bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "create a new deployment",
		Long: `Create a new deployment.

Examples:
  # Deploy from a git remote (build on buildnode):
  dployr deployments create --name my-api --source remote --runtime nodejs \
    --remote https://github.com/user/repo --branch main

  # Deploy a pre-built Docker image:
  dployr deployments create --name my-api --source image --runtime nodejs \
    --image registry.example.com/my-api:latest`,
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}

			if name == "" {
				return fmt.Errorf("--name is required")
			}
			if source == "" {
				return fmt.Errorf("--source is required (remote or image)")
			}
			if runtimeType == "" {
				return fmt.Errorf("--runtime is required")
			}
			if source == "remote" && remoteURL == "" {
				return fmt.Errorf("--remote is required when --source=remote")
			}
			if source == "image" && image == "" {
				return fmt.Errorf("--image is required when --source=image")
			}

			req := client.CreateDeploymentRequest{
				Name:             name,
				Description:      description,
				Type:             serviceType,
				Source:           source,
				RuntimeType:      runtimeType,
				RuntimeVersion:   runtimeVersion,
				RunCmd:           runCmd,
				BuildCmd:         buildCmd,
				Port:             port,
				WorkingDir:       workingDir,
				StaticDir:        staticDir,
				HealthCheck:      healthCheck,
				Image:            image,
				Domain:           domain,
				RemoteURL:        remoteURL,
				RemoteBranch:     remoteBranch,
				RemoteCommitHash: remoteCommitHash,
				EnvVars:          parseEnvVars(envVars),
				Secrets:          parseEnvVars(secrets),
				ForceRebuild:     forceRebuild,
			}

			result, err := d.client.CreateDeployment(context.Background(), req)
			if err != nil {
				return err
			}

			if d.out.Format() == output.FormatJSON {
				return d.out.JSON(result)
			}

			if result.Cached {
				fmt.Printf("deployment queued (cached image — skipped build)\n")
			} else {
				fmt.Printf("deployment queued\n")
			}
			fmt.Printf("  task:    %s\n", result.TaskID)
			fmt.Printf("  name:    %s\n", name)
			fmt.Printf("  status:  pending\n")
			fmt.Printf("\nfollow build logs with: dployr logs %s --build -f\n", name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "deployment name (required)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "deployment description")
	cmd.Flags().StringVarP(&serviceType, "type", "t", "", "service type: web, worker, static, or job")
	cmd.Flags().StringVarP(&source, "source", "s", "", "source type: remote or image (required)")
	cmd.Flags().StringVarP(&runtimeType, "runtime", "r", "", "runtime: golang, nodejs, python, php, ruby, dotnet, java (required)")
	cmd.Flags().StringVar(&runtimeVersion, "runtime-version", "", "runtime version (e.g. 20, 3.11, 1.22)")
	cmd.Flags().StringVar(&runCmd, "run-cmd", "", "command to start the application")
	cmd.Flags().StringVar(&buildCmd, "build-cmd", "", "command to build the application")
	cmd.Flags().IntVarP(&port, "port", "p", 0, "application port")
	cmd.Flags().StringVar(&workingDir, "working-dir", "", "working directory inside the container")
	cmd.Flags().StringVar(&staticDir, "static-dir", "", "directory to serve as static files")
	cmd.Flags().StringVar(&healthCheck, "health-check", "", "HTTP path for health checks (e.g. /health)")
	cmd.Flags().StringVar(&image, "image", "", "Docker image (required when --source=image)")
	cmd.Flags().StringVar(&domain, "domain", "", "custom domain name")
	cmd.Flags().StringVar(&remoteURL, "remote", "", "git remote URL (required when --source=remote)")
	cmd.Flags().StringVar(&remoteBranch, "branch", "", "git branch (default: repository default)")
	cmd.Flags().StringVar(&remoteCommitHash, "commit", "", "specific commit hash to deploy")
	cmd.Flags().StringArrayVarP(&envVars, "env", "e", nil, "environment variables in KEY=VALUE format (repeatable)")
	cmd.Flags().StringArrayVar(&secrets, "secret", nil, "secrets in KEY=VALUE format (repeatable, stored encrypted)")
	cmd.Flags().BoolVar(&forceRebuild, "force-rebuild", false, "force a fresh build even if a cached image exists")
	return cmd
}

func newDeploymentsDeleteCmd(makeDeps makeDepsFunc) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"rm", "remove"},
		Short:   "delete a deployment",
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
				fmt.Printf("delete deployment %s? [y/N]: ", args[0])
				var confirm string
				fmt.Scanln(&confirm) //nolint:errcheck
				if confirm != "y" && confirm != "Y" {
					fmt.Println("aborted")
					return nil
				}
			}

			if err := d.client.DeleteDeployment(context.Background(), args[0]); err != nil {
				return err
			}
			fmt.Printf("deployment %s deleted\n", args[0])
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation prompt")
	return cmd
}

// parseEnvVars converts KEY=VALUE strings into a map.
func parseEnvVars(pairs []string) map[string]string {
	if len(pairs) == 0 {
		return nil
	}
	m := make(map[string]string, len(pairs))
	for _, p := range pairs {
		k, v, _ := strings.Cut(p, "=")
		if k != "" {
			m[k] = v
		}
	}
	return m
}
