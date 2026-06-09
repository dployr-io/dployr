package client

import (
	"context"
	"net/url"
	"strconv"
)

type deploymentData struct {
	Deployment Deployment `json:"deployment"`
}

// ListDeployments returns deployments in the active cluster.
func (c *Client) ListDeployments(ctx context.Context, limit int) ([]Deployment, error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if c.cluster != "" {
		q.Set("clusterId", c.cluster)
	}
	r, err := get[paginatedItems[Deployment]](ctx, c, "/deployments", q)
	if err != nil {
		return nil, err
	}
	return r.Items, nil
}

// GetDeployment returns a single deployment by ID.
func (c *Client) GetDeployment(ctx context.Context, id string) (Deployment, error) {
	r, err := get[deploymentData](ctx, c, "/deployments/"+id, nil)
	if err != nil {
		return Deployment{}, err
	}
	return r.Deployment, nil
}

// CreateDeployment submits a new deployment and returns the task ID.
func (c *Client) CreateDeployment(ctx context.Context, req CreateDeploymentRequest) (CreateDeploymentResult, error) {
	q := url.Values{}
	if c.cluster != "" {
		q.Set("clusterId", c.cluster)
	}
	resp, err := c.do(ctx, "POST", "/deployments", q, req)
	if err != nil {
		return CreateDeploymentResult{}, err
	}
	return decodeResponse[CreateDeploymentResult](resp)
}

// DeleteDeployment removes a deployment by ID.
func (c *Client) DeleteDeployment(ctx context.Context, id string) error {
	return del(ctx, c, "/deployments/"+id)
}
