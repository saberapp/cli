package client

import (
	"context"
	"fmt"
	"io"
	"time"
)

// DataPoint is a single insight from a signal summary.
type DataPoint struct {
	Description        string   `json:"description"`
	ReferenceQuestions []string `json:"referenceQuestions"`
	Qualification      string   `json:"qualification"`
	Sources            []Source `json:"sources"`
}

// SummaryResponse is the response from POST /v1/companies/signals/summaries.
type SummaryResponse struct {
	Summary []DataPoint `json:"summary"`
}

// SummaryRecord is a single summary in the list response.
type SummaryRecord struct {
	ID           string      `json:"id"`
	Summary      []DataPoint `json:"summary"`
	Status       string      `json:"status"`
	SignalsCount int         `json:"signalsCount"`
	CreatedAt    time.Time   `json:"createdAt"`
	UpdatedAt    time.Time   `json:"updatedAt"`
}

// SummariesResponse wraps a paginated list of summaries.
type SummariesResponse struct {
	Results []SummaryRecord `json:"results"`
	Total   int             `json:"total"`
	Limit   int             `json:"limit"`
	Offset  int             `json:"offset"`
	Count   int             `json:"count"`
}

// GenerateSummaryRequest is the payload for POST /v1/companies/signals/summaries.
type GenerateSummaryRequest struct {
	Domain string `json:"domain"`
}

// GenerateSummary generates an AI summary for a domain.
func (c *Client) GenerateSummary(ctx context.Context, req GenerateSummaryRequest, rawDst io.Writer) (*SummaryResponse, error) {
	if rawDst != nil {
		return nil, c.Post(ctx, "/v1/companies/signals/summaries", req, nil, rawDst)
	}
	var resp SummaryResponse
	if err := c.Post(ctx, "/v1/companies/signals/summaries", req, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListSummaries lists all summaries for a domain.
func (c *Client) ListSummaries(ctx context.Context, domain string, limit, offset int, rawDst io.Writer) (*SummariesResponse, error) {
	path := fmt.Sprintf("/v1/companies/signals/summaries?domain=%s&limit=%d&offset=%d", domain, limit, offset)
	if rawDst != nil {
		return nil, c.Get(ctx, path, nil, rawDst)
	}
	var resp SummariesResponse
	if err := c.Get(ctx, path, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}
