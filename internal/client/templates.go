package client

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"
)

// SignalTemplate represents a reusable signal template.
type SignalTemplate struct {
	ID                    string         `json:"id"`
	Version               int            `json:"version"`
	OrganizationID        string         `json:"organizationId"`
	Name                  string         `json:"name"`
	Description           string         `json:"description,omitempty"`
	Question              string         `json:"question"`
	AnswerType            string         `json:"answerType"`
	OutputSchema          map[string]any `json:"outputSchema,omitempty"`
	Weight                string         `json:"weight,omitempty"`
	QualificationCriteria map[string]any `json:"qualificationCriteria,omitempty"`
	CreatedByUserID       string         `json:"createdByUserId"`
	CreatedAt             time.Time      `json:"createdAt"`
	DeletedAt             *time.Time     `json:"deletedAt,omitempty"`
	Source                string         `json:"source"`
}

// SignalTemplatesResponse wraps a paginated list of signal templates.
type SignalTemplatesResponse struct {
	Items   []SignalTemplate `json:"items"`
	Total   int              `json:"total"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
	HasMore bool             `json:"hasMore"`
}

// CreateSignalTemplateRequest is the payload for POST /v1/companies/signals/templates.
type CreateSignalTemplateRequest struct {
	Name                  string         `json:"name"`
	Question              string         `json:"question"`
	Description           string         `json:"description,omitempty"`
	AnswerType            string         `json:"answerType,omitempty"`
	Weight                string         `json:"weight,omitempty"`
	QualificationCriteria map[string]any `json:"qualificationCriteria,omitempty"`
	OutputSchema          map[string]any `json:"outputSchema,omitempty"`
}

// UpdateSignalTemplateRequest is the payload for PATCH /v1/companies/signals/templates/{id}.
type UpdateSignalTemplateRequest struct {
	Name                  string         `json:"name,omitempty"`
	Question              string         `json:"question,omitempty"`
	Description           string         `json:"description,omitempty"`
	AnswerType            string         `json:"answerType,omitempty"`
	Weight                string         `json:"weight,omitempty"`
	QualificationCriteria map[string]any `json:"qualificationCriteria,omitempty"`
	OutputSchema          map[string]any `json:"outputSchema,omitempty"`
}

// CreateSignalTemplate creates a new signal template.
func (c *Client) CreateSignalTemplate(ctx context.Context, req CreateSignalTemplateRequest, rawDst io.Writer) (*SignalTemplate, error) {
	if rawDst != nil {
		return nil, c.Post(ctx, "/v1/companies/signals/templates", req, nil, rawDst)
	}
	var tmpl SignalTemplate
	if err := c.Post(ctx, "/v1/companies/signals/templates", req, &tmpl, nil); err != nil {
		return nil, err
	}
	return &tmpl, nil
}

// ListSignalTemplates lists signal templates for the organisation.
func (c *Client) ListSignalTemplates(ctx context.Context, limit, offset int, includeDeleted bool, rawDst io.Writer) (*SignalTemplatesResponse, error) {
	path := fmt.Sprintf("/v1/companies/signals/templates?limit=%d&offset=%d&includeDeleted=%t", limit, offset, includeDeleted)
	if rawDst != nil {
		return nil, c.Get(ctx, path, nil, rawDst)
	}
	var resp SignalTemplatesResponse
	if err := c.Get(ctx, path, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetSignalTemplate retrieves a single signal template by ID.
func (c *Client) GetSignalTemplate(ctx context.Context, id string, rawDst io.Writer) (*SignalTemplate, error) {
	path := "/v1/companies/signals/templates/" + url.PathEscape(id)
	if rawDst != nil {
		return nil, c.Get(ctx, path, nil, rawDst)
	}
	var tmpl SignalTemplate
	if err := c.Get(ctx, path, &tmpl, nil); err != nil {
		return nil, err
	}
	return &tmpl, nil
}

// UpdateSignalTemplate updates a signal template (creates a new version).
func (c *Client) UpdateSignalTemplate(ctx context.Context, id string, req UpdateSignalTemplateRequest, rawDst io.Writer) (*SignalTemplate, error) {
	path := "/v1/companies/signals/templates/" + url.PathEscape(id)
	if rawDst != nil {
		return nil, c.Patch(ctx, path, req, nil, rawDst)
	}
	var tmpl SignalTemplate
	if err := c.Patch(ctx, path, req, &tmpl, nil); err != nil {
		return nil, err
	}
	return &tmpl, nil
}

// DeleteSignalTemplate soft-deletes a signal template.
func (c *Client) DeleteSignalTemplate(ctx context.Context, id string) error {
	path := "/v1/companies/signals/templates/" + url.PathEscape(id)
	return c.Delete(ctx, path)
}
