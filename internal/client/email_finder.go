package client

import (
	"context"
	"io"
)

// FindEmailRequest is the payload for POST /v1/contacts/find-email.
//
// Snake-case JSON tags match the API surface defined in go-platform's
// OpenAPI spec (distinct from the camelCase used by the rest of the
// v1 contacts endpoints).
type FindEmailRequest struct {
	FullName string `json:"full_name"`
	Domain   string `json:"domain"`
}

// VerificationResult is the verification metadata for a found email.
// Nil when the response carries no match.
type VerificationResult struct {
	// State is the upstream verifier's verdict. The API only surfaces
	// `deliverable` or `risky` — undeliverable / unknown results fold
	// into a not-found response (email = nil).
	State string `json:"state"`
	// Score is the verifier's 0-100 confidence.
	Score int `json:"score"`
	// AcceptAll is true when the domain is a catch-all. The returned
	// email is the modal real-world pattern under catch-all, but cannot
	// be distinguished from a non-existent mailbox.
	AcceptAll bool `json:"accept_all"`
}

// FindEmailResponse is returned by POST /v1/contacts/find-email.
//
// Both fields are nullable in the API contract. Not-found is a 200
// response with `{"email": null, "verification": null}` — distinct from
// a server error so the customer can treat it definitively.
type FindEmailResponse struct {
	Email        *string             `json:"email"`
	Verification *VerificationResult `json:"verification"`
}

// FindEmail resolves a verified email for a (full name, domain) pair.
//
// The endpoint applies an internal 15-second budget; warm-path repeat
// lookups at the same domain typically finish under 200ms, cold starts
// run a bounded parallel sweep. Rate-limit responses (429, either from
// the per-key middleware or pass-through from the upstream verifier)
// surface as *APIError with RetryAfter populated; the shared Do() retry
// loop handles backoff automatically before giving up.
func (c *Client) FindEmail(ctx context.Context, req FindEmailRequest, rawDst io.Writer) (*FindEmailResponse, error) {
	if rawDst != nil {
		return nil, c.Post(ctx, "/v1/contacts/find-email", req, nil, rawDst)
	}
	var resp FindEmailResponse
	if err := c.Post(ctx, "/v1/contacts/find-email", req, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}
