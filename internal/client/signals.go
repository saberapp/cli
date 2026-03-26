package client

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// SignalStatus represents the processing state of a signal.
type SignalStatus string

const (
	SignalStatusProcessing SignalStatus = "processing"
	SignalStatusCompleted  SignalStatus = "completed"
	SignalStatusFailed     SignalStatus = "failed"
)

// Source is a citation returned with a signal answer.
type Source struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

// Signal is a research signal returned by the API (company or contact).
type Signal struct {
	ID                string       `json:"id"`
	Status            SignalStatus `json:"status"`
	Domain            string       `json:"domain"`
	ContactProfileURL string       `json:"contactProfileUrl"`
	Question          string       `json:"question"`
	AnswerType        string       `json:"answerType"`
	Answer            any          `json:"answer"`
	Confidence        *float64     `json:"confidence"`
	Reasoning         string       `json:"reasoning"`
	Sources           []Source     `json:"sources"`
	CreatedAt         string       `json:"createdAt"`
	Error             string       `json:"error"`
}

// CompanySignal is an alias kept for backwards compatibility within the package.
type CompanySignal = Signal

// CreateCompanySignalRequest is the payload for POST /v1/companies/signals.
type CreateCompanySignalRequest struct {
	Domain           string `json:"domain"`
	Question         string `json:"question,omitempty"`
	AnswerType       string `json:"answerType,omitempty"`
	ForceRefresh     bool   `json:"forceRefresh,omitempty"`
	WebhookURL       string `json:"webhookUrl,omitempty"`
	SignalTemplateID string `json:"signalTemplateId,omitempty"`
	VerificationMode string `json:"verificationMode,omitempty"`
}

// CreateContactSignalRequest is the payload for POST /v1/contacts/signals.
type CreateContactSignalRequest struct {
	ContactProfileURL string `json:"contactProfileUrl"`
	Question          string `json:"question,omitempty"`
	AnswerType        string `json:"answerType,omitempty"`
	ForceRefresh      bool   `json:"forceRefresh,omitempty"`
	WebhookURL        string `json:"webhookUrl,omitempty"`
	SignalTemplateID  string `json:"signalTemplateId,omitempty"`
	VerificationMode  string `json:"verificationMode,omitempty"`
}

// CreateCompanySignal creates an async company signal.
func (c *Client) CreateCompanySignal(ctx context.Context, req CreateCompanySignalRequest) (*Signal, error) {
	var sig Signal
	if err := c.Post(ctx, "/v1/companies/signals", req, &sig, nil); err != nil {
		return nil, err
	}
	return &sig, nil
}

// CreateContactSignal creates an async contact signal.
func (c *Client) CreateContactSignal(ctx context.Context, req CreateContactSignalRequest) (*Signal, error) {
	var sig Signal
	if err := c.Post(ctx, "/v1/contacts/signals", req, &sig, nil); err != nil {
		return nil, err
	}
	return &sig, nil
}

// CreateContactSignalSync creates a contact signal and waits synchronously for the result.
func (c *Client) CreateContactSignalSync(ctx context.Context, req CreateContactSignalRequest, timeoutSec int, rawDst io.Writer) (*Signal, error) {
	headers := map[string]string{
		"X-Sbr-Timeout-Sec": strconv.Itoa(timeoutSec),
	}
	if rawDst != nil {
		return nil, c.PostWithHeaders(ctx, "/v1/contacts/signals/sync", headers, req, nil, rawDst)
	}
	var sig Signal
	if err := c.PostWithHeaders(ctx, "/v1/contacts/signals/sync", headers, req, &sig, nil); err != nil {
		return nil, err
	}
	return &sig, nil
}

// CreateCompanySignalSync creates a company signal and waits synchronously for the result.
// timeoutSec is passed as X-Sbr-Timeout-Sec (1–900). A 202 response means the server
// timed out before the signal completed; the returned signal will have Status == "processing".
func (c *Client) CreateCompanySignalSync(ctx context.Context, req CreateCompanySignalRequest, timeoutSec int, rawDst io.Writer) (*CompanySignal, error) {
	headers := map[string]string{
		"X-Sbr-Timeout-Sec": strconv.Itoa(timeoutSec),
	}
	if rawDst != nil {
		return nil, c.PostWithHeaders(ctx, "/v1/companies/signals/sync", headers, req, nil, rawDst)
	}
	var sig CompanySignal
	if err := c.PostWithHeaders(ctx, "/v1/companies/signals/sync", headers, req, &sig, nil); err != nil {
		return nil, err
	}
	return &sig, nil
}

// GetCompanySignal retrieves a company signal by ID.
func (c *Client) GetCompanySignal(ctx context.Context, id string, rawDst io.Writer) (*CompanySignal, error) {
	path := fmt.Sprintf("/v1/companies/signals/%s", id)
	if rawDst != nil {
		return nil, c.Get(ctx, path, nil, rawDst)
	}
	var sig CompanySignal
	if err := c.Get(ctx, path, &sig, nil); err != nil {
		return nil, err
	}
	return &sig, nil
}

// GetContactSignal retrieves a contact signal by ID.
func (c *Client) GetContactSignal(ctx context.Context, id string, rawDst io.Writer) (*Signal, error) {
	path := fmt.Sprintf("/v1/contacts/signals/%s", id)
	if rawDst != nil {
		return nil, c.Get(ctx, path, nil, rawDst)
	}
	var sig Signal
	if err := c.Get(ctx, path, &sig, nil); err != nil {
		return nil, err
	}
	return &sig, nil
}

// GetSignal retrieves a signal by ID, trying the company endpoint first and
// falling back to the contact endpoint on 404. This lets callers look up any
// signal without knowing its type.
func (c *Client) GetSignal(ctx context.Context, id string, rawDst io.Writer) (*Signal, error) {
	sig, err := c.GetCompanySignal(ctx, id, rawDst)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode == 404 {
			return c.GetContactSignal(ctx, id, rawDst)
		}
		return nil, err
	}
	return sig, nil
}

// --- Signal Listing ---

// SignalListItem is a signal entry in a list response.
type SignalListItem struct {
	ID                string       `json:"id"`
	SignalType        string       `json:"signalType"`
	Status            SignalStatus `json:"status"`
	Domain            string       `json:"domain"`
	ContactProfileURL string       `json:"contactProfileUrl"`
	Question          string       `json:"question"`
	AnswerType        string       `json:"answerType"`
	Answer            any          `json:"answer"`
	Confidence        *float64     `json:"confidence"`
	Reasoning         string       `json:"reasoning"`
	Sources           []Source     `json:"sources"`
	CreatedAt         time.Time    `json:"createdAt"`
	CompletedAt       *time.Time   `json:"completedAt"`
	Error             string       `json:"error"`
}

// SignalListResponse wraps a paginated list of signals.
type SignalListResponse struct {
	Results []SignalListItem `json:"results"`
	Total   int              `json:"total"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
	Count   int              `json:"count"`
}

// ListSignalsParams holds query parameters for GET /v1/companies/signals.
type ListSignalsParams struct {
	Domain         string
	CompanyID      string
	Status         []string
	FromDate       string
	ToDate         string
	SubscriptionID string
	Limit          int
	Offset         int
}

// ListSignals retrieves company signals with optional filters.
func (c *Client) ListSignals(ctx context.Context, params ListSignalsParams, rawDst io.Writer) (*SignalListResponse, error) {
	q := url.Values{}
	if params.Domain != "" {
		q.Set("domain", params.Domain)
	}
	if params.CompanyID != "" {
		q.Set("companyId", params.CompanyID)
	}
	for _, s := range params.Status {
		q.Add("status", s)
	}
	if params.FromDate != "" {
		q.Set("fromDate", params.FromDate)
	}
	if params.ToDate != "" {
		q.Set("toDate", params.ToDate)
	}
	if params.SubscriptionID != "" {
		q.Set("subscriptionId", params.SubscriptionID)
	}
	if params.Limit > 0 {
		q.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Offset > 0 {
		q.Set("offset", strconv.Itoa(params.Offset))
	}
	path := "/v1/companies/signals?" + q.Encode()
	if rawDst != nil {
		return nil, c.Get(ctx, path, nil, rawDst)
	}
	var resp SignalListResponse
	if err := c.Get(ctx, path, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

// --- Signal Batch ---

// BatchSignalItem is a single signal definition in a batch request.
type BatchSignalItem struct {
	Question     string `json:"question,omitempty"`
	TemplateID   string `json:"templateId,omitempty"`
	AnswerType   string `json:"answerType,omitempty"`
	Weight       string `json:"weight,omitempty"`
	WebhookURL   string `json:"webhookUrl,omitempty"`
	ForceRefresh bool   `json:"forceRefresh,omitempty"`
}

// CreateSignalBatchRequest is the payload for POST /v1/companies/signals/batch.
type CreateSignalBatchRequest struct {
	Signals                   []BatchSignalItem `json:"signals"`
	Domains                   []string          `json:"domains"`
	GenerateSummaryOnComplete bool              `json:"generateSummaryOnComplete,omitempty"`
	Async                     bool              `json:"async,omitempty"`
}

// BatchResultItem is a single signal result from a batch response.
type BatchResultItem struct {
	ID         string `json:"id,omitempty"`
	TemplateID string `json:"templateId,omitempty"`
	Status     string `json:"status"`
	Domain     string `json:"domain"`
	Question   string `json:"question"`
	Error      string `json:"error,omitempty"`
	CreatedAt  string `json:"createdAt,omitempty"`
}

// SignalBatchResponse is the response for sync batch creation.
type SignalBatchResponse struct {
	BatchID      string            `json:"batchId,omitempty"`
	SubmittedAt  string            `json:"submittedAt"`
	TotalSignals int               `json:"totalSignals"`
	Accepted     int               `json:"accepted"`
	Rejected     int               `json:"rejected"`
	Results      []BatchResultItem `json:"results,omitempty"`
	// Async fields
	Status  string `json:"status,omitempty"`
	IsAsync bool   `json:"async,omitempty"`
}

// CreateSignalBatch creates multiple signals in a batch.
func (c *Client) CreateSignalBatch(ctx context.Context, req CreateSignalBatchRequest, rawDst io.Writer) (*SignalBatchResponse, error) {
	if rawDst != nil {
		return nil, c.Post(ctx, "/v1/companies/signals/batch", req, nil, rawDst)
	}
	var resp SignalBatchResponse
	if err := c.Post(ctx, "/v1/companies/signals/batch", req, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

// --- Subscription Logs ---

// ListSubscriptionLogs retrieves signal executions for a subscription.
func (c *Client) ListSubscriptionLogs(ctx context.Context, subscriptionID string, domain string, limit, offset int, rawDst io.Writer) (*SignalListResponse, error) {
	q := url.Values{}
	if domain != "" {
		q.Set("domain", domain)
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		q.Set("offset", strconv.Itoa(offset))
	}
	basePath := "/v1/companies/signals/subscriptions/" + url.PathEscape(subscriptionID) + "/logs"
	path := basePath
	if encoded := q.Encode(); encoded != "" {
		path = basePath + "?" + encoded
	}
	if rawDst != nil {
		return nil, c.Get(ctx, path, nil, rawDst)
	}
	var resp SignalListResponse
	if err := c.Get(ctx, path, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

// signalBatchSummary builds a human-readable one-line summary of a batch result.
func SignalBatchSummary(resp *SignalBatchResponse) string {
	if resp.IsAsync {
		return fmt.Sprintf("Batch accepted (async): %d signals queued, batch ID: %s", resp.TotalSignals, resp.BatchID)
	}
	parts := []string{
		fmt.Sprintf("%d signals", resp.TotalSignals),
		fmt.Sprintf("%d accepted", resp.Accepted),
	}
	if resp.Rejected > 0 {
		parts = append(parts, fmt.Sprintf("%d rejected", resp.Rejected))
	}
	return strings.Join(parts, ", ")
}
