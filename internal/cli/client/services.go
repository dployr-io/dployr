package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

type serviceData struct {
	Service Service `json:"service"`
}

// ListServices returns services in the active cluster.
func (c *Client) ListServices(ctx context.Context, limit int) ([]Service, error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if c.cluster != "" {
		q.Set("clusterId", c.cluster)
	}
	r, err := get[paginatedItems[Service]](ctx, c, "/services", q)
	if err != nil {
		return nil, err
	}
	return r.Items, nil
}

// GetService returns a service by ID.
func (c *Client) GetService(ctx context.Context, id string) (Service, error) {
	r, err := get[serviceData](ctx, c, "/services/"+id, nil)
	if err != nil {
		return Service{}, err
	}
	return r.Service, nil
}

// StopService stops (sleeps) a service.
func (c *Client) StopService(ctx context.Context, id string) error {
	return postNoContent(ctx, c, fmt.Sprintf("/services/%s/stop", id), nil, nil)
}

// StartService wakes a sleeping service.
func (c *Client) StartService(ctx context.Context, id string) error {
	return postNoContent(ctx, c, fmt.Sprintf("/services/%s/start", id), nil, nil)
}

// DeleteService deletes a service and its associated resources.
func (c *Client) DeleteService(ctx context.Context, id string) error {
	return del(ctx, c, "/services/"+id)
}
