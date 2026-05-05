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

// --- Signal-template extract (v1.5 stop-gap) ---
//
// Endpoints back the "extract templates from ad-hoc signals" flow: cluster
// historical ad-hoc signal_executions into reusable templates so they can be
// referenced by scoring rules.

// ExtractClusterKind values for ExtractCluster.Kind. Mirrors the platform-side
// `ClusterKind` enum so callers can switch on Kind without magic-string literals.
const (
	ExtractClusterKindNew      = "new"
	ExtractClusterKindExisting = "existing"
)

// ExtractCluster is one proposed action in the propose/apply flow. Discriminated
// by Kind (see ExtractClusterKind* constants):
//   - "new":      Name + Question + AnswerType describe a template to create
//   - "existing": TemplateID references an existing template to attach to
//
// The fields not relevant to a given Kind are zero/empty; omitempty keeps them
// out of the wire payload. SampleQuestions and Notes are propose-side only and
// stripped server-side on apply.
type ExtractCluster struct {
	Kind string `json:"kind"`

	Name       string `json:"name,omitempty"`
	Question   string `json:"question,omitempty"`
	AnswerType string `json:"answerType,omitempty"`

	TemplateID string `json:"templateId,omitempty"`

	ExecutionIDs    []string `json:"executionIds"`
	SampleQuestions []string `json:"sampleQuestions,omitempty"`
	Notes           string   `json:"notes,omitempty"`
}

// ExtractProposeRequest is the body of POST /v1/signal-templates/extract/propose.
type ExtractProposeRequest struct {
	SignalType    string `json:"signalType"`
	MaxCandidates int    `json:"maxCandidates,omitempty"`
}

// ExtractProposal is the response of /propose. Carries the clusters plus the
// pagination state so the caller knows whether more candidates remain.
type ExtractProposal struct {
	Clusters            []ExtractCluster `json:"clusters"`
	TotalCandidates     int              `json:"totalCandidates"`
	ProcessedCandidates int              `json:"processedCandidates"`
	HasMore             bool             `json:"hasMore"`
}

// ExtractAppliedTemplate is one row in the /apply response.
type ExtractAppliedTemplate struct {
	Kind       string `json:"kind"`
	TemplateID string `json:"templateId"`
	VersionID  string `json:"versionId"`
	Name       string `json:"name"`
	Attached   int    `json:"attached"`
}

// ExtractApplyResult is the response of /apply.
type ExtractApplyResult struct {
	Created []ExtractAppliedTemplate `json:"created"`
}

// ExtractApplyRequest is the body of POST /v1/signal-templates/extract/apply.
type ExtractApplyRequest struct {
	Clusters []ExtractCluster `json:"clusters"`
}

// ProposeExtractTemplates triggers LLM clustering of ad-hoc signal candidates
// for the org behind the API key. Returns clusters that can be edited locally
// and replayed via ApplyExtractTemplates.
func (c *Client) ProposeExtractTemplates(ctx context.Context, req ExtractProposeRequest, rawDst io.Writer) (*ExtractProposal, error) {
	if rawDst != nil {
		return nil, c.Post(ctx, "/v1/signal-templates/extract/propose", req, nil, rawDst)
	}
	var p ExtractProposal
	if err := c.Post(ctx, "/v1/signal-templates/extract/propose", req, &p, nil); err != nil {
		return nil, err
	}
	return &p, nil
}

// ApplyExtractTemplates creates new templates and/or attaches executions to
// existing ones in a single transaction. Re-running the same plan returns 409
// — drop the already-attached executionIds before retrying.
func (c *Client) ApplyExtractTemplates(ctx context.Context, req ExtractApplyRequest, rawDst io.Writer) (*ExtractApplyResult, error) {
	if rawDst != nil {
		return nil, c.Post(ctx, "/v1/signal-templates/extract/apply", req, nil, rawDst)
	}
	var r ExtractApplyResult
	if err := c.Post(ctx, "/v1/signal-templates/extract/apply", req, &r, nil); err != nil {
		return nil, err
	}
	return &r, nil
}
