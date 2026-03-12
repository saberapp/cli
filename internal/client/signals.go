package client

import (
	"context"
	"fmt"
	"io"
	"strconv"
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
	Domain       string `json:"domain"`
	Question     string `json:"question"`
	AnswerType   string `json:"answerType,omitempty"`
	ForceRefresh bool   `json:"forceRefresh,omitempty"`
	WebhookURL   string `json:"webhookUrl,omitempty"`
}

// CreateContactSignalRequest is the payload for POST /v1/contacts/signals.
type CreateContactSignalRequest struct {
	ContactProfileURL string `json:"contactProfileUrl"`
	Question          string `json:"question"`
	AnswerType        string `json:"answerType,omitempty"`
	ForceRefresh      bool   `json:"forceRefresh,omitempty"`
	WebhookURL        string `json:"webhookUrl,omitempty"`
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
