package client

import (
	"context"
	"encoding/json"
	"net/http"
)

type exchangeOIDCRequest struct {
	Token string `json:"token"`
}

type exchangeOIDCResponse struct {
	Data struct {
		SessionId string `json:"sessionId"`
	} `json:"data"`
}

// ExchangeOIDCToken exchanges a CI-issued OIDC token for a dployr session ID.
// Corresponds to POST /v1/auth/oidc/exchange.
func (c *Client) ExchangeOIDCToken(ctx context.Context, token string) (string, error) {
	resp, err := c.do(ctx, http.MethodPost, "/auth/oidc/exchange", nil, exchangeOIDCRequest{Token: token})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", readAPIError(resp)
	}

	var body exchangeOIDCResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	if body.Data.SessionId == "" {
		return "", nil
	}
	return body.Data.SessionId, nil
}
