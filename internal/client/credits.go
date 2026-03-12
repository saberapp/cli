package client

import (
	"context"
	"io"
)

// CreditsBalance is returned by GET /v1/credits.
type CreditsBalance struct {
	RemainingCredits int `json:"remainingCredits"`
}

// GetCredits returns the current credit balance.
func (c *Client) GetCredits(ctx context.Context, rawDst io.Writer) (*CreditsBalance, error) {
	if rawDst != nil {
		return nil, c.Get(ctx, "/v1/credits", nil, rawDst)
	}
	var resp CreditsBalance
	if err := c.Get(ctx, "/v1/credits", &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}
