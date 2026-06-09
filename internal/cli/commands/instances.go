package commands

import (
	"context"
	"fmt"

	"github.com/dployr-io/dployr/internal/cli/output"
	"github.com/spf13/cobra"
)

func newInstancesCmd(makeDeps makeDepsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instances",
		Short: "manage cluster instances",
	}

	cmd.AddCommand(newInstancesListCmd(makeDeps))
	cmd.AddCommand(newInstancesGetCmd(makeDeps))
	cmd.AddCommand(newInstancesPingCmd(makeDeps))
	cmd.AddCommand(newInstancesDeleteCmd(makeDeps))
	cmd.AddCommand(newInstancesSystemCmd(makeDeps))
	return cmd
}

func newInstancesListCmd(makeDeps makeDepsFunc) *cobra.Command {
	var limit int
	var role string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list instances",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}

			instances, err := d.client.ListInstances(context.Background(), limit)
			if err != nil {
				return err
			}

			// Client-side filter by role if specified.
			if role != "" {
				filtered := instances[:0]
				for _, inst := range instances {
					if inst.Role == role {
						filtered = append(filtered, inst)
					}
				}
				instances = filtered
			}

			if d.out.Format() == output.FormatJSON {
				return d.out.JSON(instances)
			}

			if len(instances) == 0 {
				fmt.Println("no instances found")
				return nil
			}

			rows := make([][]string, len(instances))
			for i, inst := range instances {
				region := inst.Region
				if region == "" {
					region = "-"
				}
				rows[i] = []string{inst.Tag, inst.Role, inst.Kind, inst.Status, region}
			}
			d.out.Table([]string{"TAG", "ROLE", "KIND", "STATUS", "REGION"}, rows)
			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 20, "maximum number of instances to return")
	cmd.Flags().StringVar(&role, "role", "", "filter by role: instance or build")
	return cmd
}

func newInstancesGetCmd(makeDeps makeDepsFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "get <tag>",
		Short: "get instance details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}

			inst, err := d.client.GetInstance(context.Background(), args[0])
			if err != nil {
				return err
			}

			if d.out.Format() == output.FormatJSON {
				return d.out.JSON(inst)
			}

			d.out.Printf("id:      %s\n", inst.ID)
			d.out.Printf("tag:     %s\n", inst.Tag)
			d.out.Printf("role:    %s\n", inst.Role)
			d.out.Printf("kind:    %s\n", inst.Kind)
			d.out.Printf("status:  %s\n", inst.Status)
			d.out.Printf("address: %s\n", inst.Address)
			if inst.Region != "" {
				d.out.Printf("region:  %s\n", inst.Region)
			}
			d.out.Printf("managed: %v\n", inst.Managed)
			d.out.Printf("created: %s\n", timeAgo(inst.CreatedAt))
			d.out.Printf("updated: %s\n", timeAgo(inst.UpdatedAt))
			return nil
		},
	}
}

func newInstancesPingCmd(makeDeps makeDepsFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "ping <instance-name>",
		Short: "ping an instance to check connectivity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}
			if err := d.client.PingInstance(context.Background(), args[0]); err != nil {
				return fmt.Errorf("ping failed: %w", err)
			}
			fmt.Printf("instance %s is reachable\n", args[0])
			return nil
		},
	}
}

func newInstancesDeleteCmd(makeDeps makeDepsFunc) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "delete <tag>",
		Aliases: []string{"rm"},
		Short:   "delete an instance",
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
				fmt.Printf("delete instance %s? [y/N]: ", args[0])
				var confirm string
				fmt.Scanln(&confirm) //nolint:errcheck
				if confirm != "y" && confirm != "Y" {
					fmt.Println("aborted")
					return nil
				}
			}

			if err := d.client.DeleteInstance(context.Background(), args[0]); err != nil {
				return err
			}
			fmt.Printf("instance %s deleted\n", args[0])
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation prompt")
	return cmd
}

// --- system subcommand ---

func newInstancesSystemCmd(makeDeps makeDepsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system",
		Short: "low-level instance system operations (requires admin)",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "install <tag>",
		Short: "trigger dployr installation on the instance daemon",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}
			if err := d.client.SystemInstall(context.Background(), args[0]); err != nil {
				return err
			}
			fmt.Printf("install triggered on instance %s\n", args[0])
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "reboot <tag>",
		Short: "reboot the instance machine",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}
			fmt.Printf("rebooting instance %s...\n", args[0])
			if err := d.client.SystemReboot(context.Background(), args[0]); err != nil {
				return err
			}
			fmt.Println("reboot command sent")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "restart <tag>",
		Short: "restart the dployrd daemon on an instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}
			if err := d.client.SystemRestart(context.Background(), args[0]); err != nil {
				return err
			}
			fmt.Printf("daemon restarted on instance %s\n", args[0])
			return nil
		},
	})

	return cmd
}
