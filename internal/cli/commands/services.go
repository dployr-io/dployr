package commands

import (
	"context"
	"fmt"

	"github.com/dployr-io/dployr/internal/cli/output"
	"github.com/spf13/cobra"
)

func newServicesCmd(makeDeps makeDepsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "services",
		Short: "manage services",
	}

	cmd.AddCommand(newServicesListCmd(makeDeps))
	cmd.AddCommand(newServicesGetCmd(makeDeps))
	cmd.AddCommand(newServicesStopCmd(makeDeps))
	cmd.AddCommand(newServicesStartCmd(makeDeps))
	cmd.AddCommand(newServicesDeleteCmd(makeDeps))
	return cmd
}

func newServicesListCmd(makeDeps makeDepsFunc) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list services in the active cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}

			services, err := d.client.ListServices(context.Background(), limit)
			if err != nil {
				return err
			}

			if d.out.Format() == output.FormatJSON {
				return d.out.JSON(services)
			}

			if len(services) == 0 {
				fmt.Println("no services found")
				return nil
			}

			rows := make([][]string, len(services))
			for i, s := range services {
				status := "running"
				if s.IcedAt != nil {
					status = "sleeping"
				}
				deployment := "-"
				if s.DeploymentName != nil {
					deployment = *s.DeploymentName
				} else if s.DeploymentID != nil {
					deployment = *s.DeploymentID
				}
				rows[i] = []string{s.Name, s.Type, status, deployment, timeAgo(s.CreatedAt), timeAgo(s.UpdatedAt)}
			}
			d.out.Table([]string{"NAME", "TYPE", "STATUS", "DEPLOYMENT", "CREATED", "UPDATED"}, rows)
			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 20, "maximum number of services to return")
	return cmd
}

func newServicesGetCmd(makeDeps makeDepsFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "get service details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}

			s, err := d.client.GetService(context.Background(), args[0])
			if err != nil {
				return err
			}

			if d.out.Format() == output.FormatJSON {
				return d.out.JSON(s)
			}

			status := "running"
			if s.IcedAt != nil {
				status = fmt.Sprintf("sleeping (%s)", timeAgoPtr(s.IcedAt))
			}

			d.out.Printf("id:         %s\n", s.ID)
			d.out.Printf("name:       %s\n", s.Name)
			d.out.Printf("type:       %s\n", s.Type)
			d.out.Printf("status:     %s\n", status)
			if s.DeploymentName != nil {
				d.out.Printf("deployment: %s\n", *s.DeploymentName)
			} else if s.DeploymentID != nil {
				d.out.Printf("deployment: %s\n", *s.DeploymentID)
			}
			d.out.Printf("created:    %s\n", timeAgo(s.CreatedAt))
			d.out.Printf("updated:    %s\n", timeAgo(s.UpdatedAt))
			return nil
		},
	}
}

func newServicesStopCmd(makeDeps makeDepsFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "stop <name>",
		Short: "stop (sleep) a service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}
			if err := d.client.StopService(context.Background(), args[0]); err != nil {
				return err
			}
			fmt.Printf("service %s stopped\n", args[0])
			return nil
		},
	}
}

func newServicesStartCmd(makeDeps makeDepsFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "start <name>",
		Short: "start a sleeping service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}
			if err := d.client.StartService(context.Background(), args[0]); err != nil {
				return err
			}
			fmt.Printf("service %s started\n", args[0])
			return nil
		},
	}
}

func newServicesDeleteCmd(makeDeps makeDepsFunc) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"rm", "remove"},
		Short:   "delete a service",
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
				fmt.Printf("delete service %s? this cannot be undone [y/N]: ", args[0])
				var confirm string
				fmt.Scanln(&confirm) //nolint:errcheck
				if confirm != "y" && confirm != "Y" {
					fmt.Println("aborted")
					return nil
				}
			}

			if err := d.client.DeleteService(context.Background(), args[0]); err != nil {
				return err
			}
			fmt.Printf("service %s deleted\n", args[0])
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation prompt")
	return cmd
}

// shortHash returns the first 8 characters of a commit hash for display.
func shortHash(h string) string {
	if len(h) > 8 {
		return h[:8]
	}
	return h
}
