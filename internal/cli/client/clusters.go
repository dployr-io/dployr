package client

import (
	"context"
	"fmt"
)

type clusterListData struct {
	Clusters []Cluster `json:"clusters"`
}

type clusterData struct {
	Cluster Cluster `json:"cluster"`
}

type invitesData struct {
	Invites []PendingInvite `json:"invites"`
}

// ListClusters returns all clusters the authenticated user belongs to.
func (c *Client) ListClusters(ctx context.Context) ([]Cluster, error) {
	r, err := get[clusterListData](ctx, c, "/clusters", nil)
	if err != nil {
		return nil, err
	}
	return r.Clusters, nil
}

// GetCluster returns a cluster by name by scanning the user's cluster list.
// There is no dedicated GET /:id endpoint — name lookup is done client-side.
func (c *Client) GetCluster(ctx context.Context, name string) (Cluster, error) {
	clusters, err := c.ListClusters(ctx)
	if err != nil {
		return Cluster{}, err
	}
	for _, cl := range clusters {
		if cl.Name == name {
			return cl, nil
		}
	}
	return Cluster{}, fmt.Errorf("cluster %q not found", name)
}

// UpdateCluster renames a cluster. Resolves the cluster name to its ID first
// since the rename endpoint requires the cluster UUID.
func (c *Client) UpdateCluster(ctx context.Context, name, newName string) (Cluster, error) {
	cl, err := c.GetCluster(ctx, name)
	if err != nil {
		return Cluster{}, err
	}
	r, err := patch[clusterData](ctx, c, "/clusters/"+cl.ID, map[string]string{"name": newName})
	if err != nil {
		return Cluster{}, err
	}
	return r.Cluster, nil
}

// ListClusterUsers returns all users in a cluster identified by name.
func (c *Client) ListClusterUsers(ctx context.Context, clusterName string) ([]ClusterUser, error) {
	cl, err := c.GetCluster(ctx, clusterName)
	if err != nil {
		return nil, err
	}
	r, err := get[paginatedItems[ClusterUser]](ctx, c, fmt.Sprintf("/clusters/%s/users", cl.ID), nil)
	if err != nil {
		return nil, err
	}
	return r.Items, nil
}

// AddClusterUsers invites users to a cluster with a given role.
func (c *Client) AddClusterUsers(ctx context.Context, clusterName string, emails []string, role string) error {
	cl, err := c.GetCluster(ctx, clusterName)
	if err != nil {
		return err
	}
	return postNoContent(ctx, c, fmt.Sprintf("/clusters/%s/users", cl.ID), nil, map[string]any{
		"emails": emails,
		"role":   role,
	})
}

// RemoveClusterUsers removes users from a cluster.
func (c *Client) RemoveClusterUsers(ctx context.Context, clusterName string, userIDs []string) error {
	cl, err := c.GetCluster(ctx, clusterName)
	if err != nil {
		return err
	}
	return postNoContent(ctx, c, fmt.Sprintf("/clusters/%s/users/remove", cl.ID), nil, map[string]any{
		"userIds": userIDs,
	})
}

// ListInvites returns pending cluster invites for the authenticated user.
func (c *Client) ListInvites(ctx context.Context) ([]PendingInvite, error) {
	r, err := get[invitesData](ctx, c, "/clusters/users/invites", nil)
	if err != nil {
		return nil, err
	}
	return r.Invites, nil
}

// AcceptInvite accepts a cluster invite identified by cluster name.
func (c *Client) AcceptInvite(ctx context.Context, clusterName string) error {
	cl, err := c.GetCluster(ctx, clusterName)
	if err != nil {
		return err
	}
	return postNoContent(ctx, c, fmt.Sprintf("/clusters/%s/users/invites/accept", cl.ID), nil, nil)
}

// DeclineInvite declines a cluster invite identified by cluster name.
func (c *Client) DeclineInvite(ctx context.Context, clusterName string) error {
	cl, err := c.GetCluster(ctx, clusterName)
	if err != nil {
		return err
	}
	return postNoContent(ctx, c, fmt.Sprintf("/clusters/%s/users/invites/decline", cl.ID), nil, nil)
}
