package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newContextCmd(makeDeps makeDepsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "manage active cluster context",
		Long: `Switch and inspect the active cluster used by all commands.

The active cluster name is stored in config.toml and used as the
default for all cluster-scoped commands.`,
	}

	cmd.AddCommand(newContextUseCmd(makeDeps))
	cmd.AddCommand(newContextListCmd(makeDeps))
	cmd.AddCommand(newContextShowCmd(makeDeps))
	return cmd
}

func newContextUseCmd(makeDeps makeDepsFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "use <cluster-name>",
		Short: "set the active cluster by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}

			name := args[0]

			// Verify the cluster exists and the user has access; grab its ID.
			c, err := d.client.GetCluster(context.Background(), name)
			if err != nil {
				return fmt.Errorf("cluster %q not found or inaccessible: %w", name, err)
			}

			d.cfg.ActiveCluster = c.Name
			d.cfg.ActiveClusterID = c.ID
			if err := d.cfg.Save(); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Printf("active cluster set to %q\n", name)
			return nil
		},
	}
}

func newContextListCmd(makeDeps makeDepsFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list all accessible clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}

			clusters, err := d.client.ListClusters(context.Background())
			if err != nil {
				return err
			}

			if len(clusters) == 0 {
				fmt.Println("no clusters found")
				return nil
			}

			rows := make([][]string, len(clusters))
			for i, c := range clusters {
				active := ""
				if c.Name == d.cfg.ActiveCluster {
					active = "*"
				}
				rows[i] = []string{active, c.Name, timeAgo(c.CreatedAt)}
			}
			d.out.Table([]string{"", "NAME", "CREATED"}, rows)
			return nil
		},
	}
}

func newContextShowCmd(makeDeps makeDepsFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "show the current active cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if d.cfg.ActiveCluster == "" {
				fmt.Println("no active cluster — use 'dployr context use <name>'")
				return nil
			}
			fmt.Printf("active cluster: %s\n", d.cfg.ActiveCluster)
			return nil
		},
	}
}
