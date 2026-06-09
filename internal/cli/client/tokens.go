package client

import "context"

type tokenListData struct {
	Tokens []ApiToken `json:"tokens"`
}

// CreateToken creates a new scoped personal access token.
// The returned CreatedApiToken includes the plaintext — it is shown only once.
func (c *Client) CreateToken(ctx context.Context, name string, scopes []string, expiresIn *int) (CreatedApiToken, error) {
	body := map[string]any{
		"name":   name,
		"scopes": scopes,
	}
	if expiresIn != nil {
		body["expiresIn"] = *expiresIn
	}
	return post[CreatedApiToken](ctx, c, "/auth/tokens", body)
}

// ListTokens returns all personal access tokens for the authenticated user.
func (c *Client) ListTokens(ctx context.Context) ([]ApiToken, error) {
	r, err := get[tokenListData](ctx, c, "/auth/tokens", nil)
	if err != nil {
		return nil, err
	}
	return r.Tokens, nil
}

// RevokeToken deletes a token by ID.
func (c *Client) RevokeToken(ctx context.Context, id string) error {
	return del(ctx, c, "/auth/tokens/"+id)
}
