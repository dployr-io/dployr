package commands

import (
	"context"
	"fmt"

	"github.com/dployr-io/dployr/internal/cli/client"
	"github.com/dployr-io/dployr/internal/cli/config"
	"github.com/dployr-io/dployr/internal/cli/output"
	"github.com/spf13/cobra"
)

// deps bundles the shared dependencies threaded through every command.
type deps struct {
	cfg    *config.Config
	client *client.Client
	out    *output.Writer
}

// New builds and returns the root cobra command with all subcommands attached.
func New() *cobra.Command {
	var (
		outputFormat string
		clusterID    string
	)

	root := &cobra.Command{
		Use:   "dployr",
		Short: "dployr — deploy anywhere, your rules",
		Long: `CLI for the dployr platform.

Manage clusters, services, deployments, and instances from your terminal.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", output.FormatFlag)
	root.PersistentFlags().StringVarP(&clusterID, "cluster", "c", "", "Cluster name (defaults to active cluster in config)")

	makeDeps := func(cmd *cobra.Command) (*deps, error) {
		cfg, err := config.Load()
		if err != nil {
			return nil, fmt.Errorf("load config: %w", err)
		}

		outFmt, err := output.ParseFormat(outputFormat)
		if err != nil {
			return nil, err
		}
		w := output.New(outFmt)

		cl := client.New(cfg)

		// Resolve the cluster UUID for this invocation.
		// Priority: --cluster flag > config active_cluster_id > auto-select (fetched once, then cached).
		clusterUUID := cfg.ActiveClusterID
		if clusterID != "" {
			// Flag overrides config; resolve name → UUID without touching config.
			clusters, err := cl.ListClusters(context.Background())
			if err != nil {
				return nil, fmt.Errorf("fetch clusters: %w", err)
			}
			for _, c := range clusters {
				if c.Name == clusterID {
					clusterUUID = c.ID
					break
				}
			}
			if clusterUUID == cfg.ActiveClusterID {
				return nil, fmt.Errorf("cluster %q not found — run 'dployr context list' to see available clusters", clusterID)
			}
		} else if clusterUUID == "" && cfg.IsAuthenticated() {
			// No cluster in config yet: fetch, auto-select, and persist so future commands skip this.
			clusters, err := cl.ListClusters(context.Background())
			if err != nil {
				return nil, fmt.Errorf("fetch clusters: %w", err)
			}
			var selected *client.Cluster
			for i := range clusters {
				if clusters[i].Role == "owner" {
					selected = &clusters[i]
					break
				}
			}
			if selected == nil && len(clusters) > 0 {
				selected = &clusters[0]
			}
			if selected != nil {
				clusterUUID = selected.ID
				cfg.ActiveCluster = selected.Name
				cfg.ActiveClusterID = selected.ID
				_ = cfg.Save()
				fmt.Printf("active cluster set to %q\n", selected.Name)
			}
		}

		if clusterUUID != "" {
			cl = cl.WithCluster(clusterUUID)
		}

		return &deps{cfg: cfg, client: cl, out: w}, nil
	}

	// Store makeDeps in root context so subcommands can call it.
	ctx := context.WithValue(context.Background(), depsKey{}, makeDeps)
	root.SetContext(ctx)

	root.AddCommand(newVersionCmd())
	root.AddCommand(newAuthCmd(makeDeps))
	root.AddCommand(newContextCmd(makeDeps))
	root.AddCommand(newClustersCmd(makeDeps))
	root.AddCommand(newServicesCmd(makeDeps))
	root.AddCommand(newDeploymentsCmd(makeDeps))
	root.AddCommand(newInstancesCmd(makeDeps))
	root.AddCommand(newLogsCmd(makeDeps))

	return root
}

type depsKey struct{}

// requireAuth checks that the user is logged in and returns an error if not.
func requireAuth(cfg *config.Config) error {
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated — run 'dployr auth login' first")
	}
	return nil
}

// requireCluster checks that an active cluster name is set.
func requireCluster(d *deps) error {
	if d.client.Cluster() == "" {
		return fmt.Errorf("no active cluster — set one with 'dployr context use <name>' or --cluster flag")
	}
	return nil
}

// makeDepsFunc is the type of the makeDeps closure passed to command constructors.
type makeDepsFunc func(cmd *cobra.Command) (*deps, error)
