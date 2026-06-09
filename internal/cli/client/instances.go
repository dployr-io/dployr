package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// ListInstances returns instances visible to the authenticated user.
func (c *Client) ListInstances(ctx context.Context, limit int) ([]Instance, error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if c.cluster != "" {
		q.Set("clusterId", c.cluster)
	}
	r, err := get[paginatedItems[Instance]](ctx, c, "/instances", q)
	if err != nil {
		return nil, err
	}
	return r.Items, nil
}

// GetInstance returns a single instance by ID.
// The server returns { "success": true, "data": <instance> } with the instance
// directly as the data value (no intermediate field name).
func (c *Client) GetInstance(ctx context.Context, id string) (Instance, error) {
	return get[Instance](ctx, c, "/instances/"+id, nil)
}

// PingInstance sends a health ping to the instance daemon.
func (c *Client) PingInstance(ctx context.Context, name string) error {
	return postNoContent(ctx, c, fmt.Sprintf("/instances/%s/ping", name), nil)
}

// DeleteInstance removes an instance.
func (c *Client) DeleteInstance(ctx context.Context, id string) error {
	return del(ctx, c, "/instances/"+id)
}

// SystemInstall triggers dployr installation on the instance daemon.
func (c *Client) SystemInstall(ctx context.Context, instanceID string) error {
	return postNoContent(ctx, c, fmt.Sprintf("/instances/%s/system/install", instanceID), nil)
}

// SystemReboot reboots the underlying machine.
func (c *Client) SystemReboot(ctx context.Context, instanceID string) error {
	return postNoContent(ctx, c, fmt.Sprintf("/instances/%s/system/reboot", instanceID), nil)
}

// SystemRestart restarts the dployrd daemon on the instance.
func (c *Client) SystemRestart(ctx context.Context, instanceID string) error {
	return postNoContent(ctx, c, fmt.Sprintf("/instances/%s/system/restart", instanceID), nil)
}

// RotateInstanceToken rotates the access token for an instance.
func (c *Client) RotateInstanceToken(ctx context.Context, instanceID string) error {
	return postNoContent(ctx, c, fmt.Sprintf("/instances/%s/tokens/rotate", instanceID), nil)
}

// AddInstanceDomain configures a custom domain on an instance.
func (c *Client) AddInstanceDomain(ctx context.Context, name, domain string) error {
	return postNoContent(ctx, c, fmt.Sprintf("/instances/%s/domain", name), map[string]string{
		"domain": domain,
	})
}
