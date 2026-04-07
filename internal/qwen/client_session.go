package qwen

import (
	"context"

	"ds2api/internal/auth"
)

func (c *Client) CreateSession(_ context.Context, _ *auth.RequestAuth, _ int) (string, error) {
	return "", nil
}

func (c *Client) GetPow(_ context.Context, _ *auth.RequestAuth, _ int) (string, error) {
	return "", nil
}

func (c *Client) DeleteAllSessionsForToken(_ context.Context, _ string) error {
	return nil
}
