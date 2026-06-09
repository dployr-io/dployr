package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/dployr-io/dployr/internal/cli/output"
	"github.com/spf13/cobra"
)

func newClustersCmd(makeDeps makeDepsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clusters",
		Short: "manage clusters",
	}

	cmd.AddCommand(newClustersListCmd(makeDeps))
	cmd.AddCommand(newClustersGetCmd(makeDeps))
	cmd.AddCommand(newClustersUsersCmd(makeDeps))
	cmd.AddCommand(newClustersInvitesCmd(makeDeps))
	return cmd
}

func newClustersListCmd(makeDeps makeDepsFunc) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list all clusters",
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

			if d.out.Format() == output.FormatJSON {
				return d.out.JSON(clusters)
			}

			if len(clusters) == 0 {
				fmt.Println("no clusters found")
				return nil
			}

			rows := make([][]string, len(clusters))
			for i, c := range clusters {
				rows[i] = []string{c.Name, timeAgo(c.CreatedAt)}
			}
			d.out.Table([]string{"NAME", "CREATED"}, rows)
			return nil
		},
	}
}

func newClustersGetCmd(makeDeps makeDepsFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "get cluster details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}

			c, err := d.client.GetCluster(context.Background(), args[0])
			if err != nil {
				return err
			}

			if d.out.Format() == output.FormatJSON {
				return d.out.JSON(c)
			}

			d.out.Printf("name:    %s\n", c.Name)
			d.out.Printf("created: %s\n", timeAgo(c.CreatedAt))
			d.out.Printf("updated: %s\n", timeAgo(c.UpdatedAt))
			return nil
		},
	}
}

func newClustersUsersCmd(makeDeps makeDepsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "users",
		Short: "manage cluster members",
	}

	// list
	cmd.AddCommand(&cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list cluster members",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}
			if err := requireCluster(d); err != nil {
				return err
			}

			users, err := d.client.ListClusterUsers(context.Background(), d.client.Cluster())
			if err != nil {
				return err
			}

			if d.out.Format() == output.FormatJSON {
				return d.out.JSON(users)
			}

			if len(users) == 0 {
				fmt.Println("no users found")
				return nil
			}

			rows := make([][]string, len(users))
			for i, u := range users {
				rows[i] = []string{u.ID, u.Email, u.Name, u.Role}
			}
			d.out.Table([]string{"ID", "EMAIL", "NAME", "ROLE"}, rows)
			return nil
		},
	})

	// add
	addCmd := &cobra.Command{
		Use:   "add <email> [email...]",
		Short: "invite users to the cluster",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}
			if err := requireCluster(d); err != nil {
				return err
			}

			role, _ := cmd.Flags().GetString("role")
			if role == "" {
				role = "developer"
			}

			if err := d.client.AddClusterUsers(context.Background(), d.client.Cluster(), args, role); err != nil {
				return err
			}

			fmt.Printf("invited %s as %s\n", strings.Join(args, ", "), role)
			return nil
		},
	}
	addCmd.Flags().String("role", "developer", "role to assign: owner, admin, developer, viewer")
	cmd.AddCommand(addCmd)

	// remove
	cmd.AddCommand(&cobra.Command{
		Use:   "remove <user-id> [user-id...]",
		Short: "remove users from the cluster",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}
			if err := requireCluster(d); err != nil {
				return err
			}

			if err := d.client.RemoveClusterUsers(context.Background(), d.client.Cluster(), args); err != nil {
				return err
			}

			fmt.Printf("removed %d user(s) from cluster\n", len(args))
			return nil
		},
	})

	return cmd
}

func newClustersInvitesCmd(makeDeps makeDepsFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invites",
		Short: "manage cluster invites",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "list pending cluster invites",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}

			invites, err := d.client.ListInvites(context.Background())
			if err != nil {
				return err
			}

			if d.out.Format() == output.FormatJSON {
				return d.out.JSON(invites)
			}

			if len(invites) == 0 {
				fmt.Println("no pending invites")
				return nil
			}

			rows := make([][]string, len(invites))
			for i, inv := range invites {
				rows[i] = []string{inv.ClusterName, inv.OwnerName, inv.ClusterID}
			}
			d.out.Table([]string{"CLUSTER", "OWNER", "CLUSTER ID"}, rows)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "accept <cluster-name>",
		Short: "accept a cluster invite",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}
			if err := d.client.AcceptInvite(context.Background(), args[0]); err != nil {
				return err
			}
			fmt.Printf("joined cluster %s\n", args[0])
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "decline <cluster-name>",
		Short: "decline a cluster invite",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}
			if err := d.client.DeclineInvite(context.Background(), args[0]); err != nil {
				return err
			}
			fmt.Printf("declined invite to cluster %s\n", args[0])
			return nil
		},
	})

	return cmd
}
