package client

import (
	"context"
	"io"
)

// Connector represents a configured integration connector.
type Connector struct {
	Source string `json:"source"` // salesNavigator, hubspotApp
	Status string `json:"status"` // connected, disconnected, needs-authentication, ready-to-connect, error
}

// ConnectorsListResponse is returned by GET /v1/connectors.
type ConnectorsListResponse struct {
	Connectors []Connector `json:"connectors"`
}

// ListConnectors returns all configured connectors and their status.
func (c *Client) ListConnectors(ctx context.Context, rawDst io.Writer) (*ConnectorsListResponse, error) {
	if rawDst != nil {
		return nil, c.Get(ctx, "/v1/connectors", nil, rawDst)
	}
	var resp ConnectorsListResponse
	if err := c.Get(ctx, "/v1/connectors", &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}
