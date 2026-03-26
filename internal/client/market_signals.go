package client

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"time"
)

// MarketSignalSubscription represents a market signal subscription.
type MarketSignalSubscription struct {
	ID                  string         `json:"id"`
	OrganizationID      string         `json:"organizationId"`
	Type                string         `json:"type"`
	Name                *string        `json:"name"`
	Status              string         `json:"status"`
	Prompt              *string        `json:"prompt"`
	Filters             map[string]any `json:"filters"`
	WebhookURL          string         `json:"webhookUrl"`
	IntervalSignalLimit int            `json:"intervalSignalLimit"`
	Interval            string         `json:"interval"`
	CreatedAt           time.Time      `json:"createdAt"`
	UpdatedAt           time.Time      `json:"updatedAt"`
}

// MarketSignalSubscriptionListResponse wraps a paginated list of subscriptions.
type MarketSignalSubscriptionListResponse struct {
	Items  []MarketSignalSubscription `json:"items"`
	Total  int                        `json:"total"`
	Limit  int                        `json:"limit"`
	Offset int                        `json:"offset"`
}

// MarketSignal represents a single signal matched by a subscription.
type MarketSignal struct {
	ID               string         `json:"id"`
	SubscriptionID   string         `json:"subscriptionId"`
	JobPostingID     *string        `json:"jobPostingId"`
	ExternalSignalID *string        `json:"externalSignalId"`
	Payload          map[string]any `json:"payload"`
	ConfidenceScore  *float64       `json:"confidenceScore"`
	Status           string         `json:"status"`
	PublishedAt      *time.Time     `json:"publishedAt"`
	CreatedAt        time.Time      `json:"createdAt"`
	DeliveredAt      *time.Time     `json:"deliveredAt"`
}

// MarketSignalListResponse wraps a paginated list of market signals.
type MarketSignalListResponse struct {
	Items  []MarketSignal `json:"items"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// CreateMarketSignalSubscriptionRequest is the payload for creating a subscription.
type CreateMarketSignalSubscriptionRequest struct {
	Type                string         `json:"type"`
	Name                string         `json:"name,omitempty"`
	Prompt              string         `json:"prompt,omitempty"`
	Filters             map[string]any `json:"filters,omitempty"`
	WebhookURL          string         `json:"webhookUrl"`
	WebhookSecret       string         `json:"webhookSecret,omitempty"`
	IntervalSignalLimit int            `json:"intervalSignalLimit,omitempty"`
	Interval            string         `json:"interval,omitempty"`
}

// UpdateMarketSignalSubscriptionRequest is the payload for updating a subscription.
type UpdateMarketSignalSubscriptionRequest struct {
	Name                string         `json:"name,omitempty"`
	Prompt              string         `json:"prompt,omitempty"`
	Filters             map[string]any `json:"filters,omitempty"`
	WebhookURL          string         `json:"webhookUrl,omitempty"`
	WebhookSecret       string         `json:"webhookSecret,omitempty"`
	IntervalSignalLimit *int           `json:"intervalSignalLimit,omitempty"`
	Interval            string         `json:"interval,omitempty"`
}

const marketSignalsBasePath = "/v1/market-signals/subscriptions"

// CreateMarketSignalSubscription creates a new market signal subscription.
func (c *Client) CreateMarketSignalSubscription(ctx context.Context, req CreateMarketSignalSubscriptionRequest, rawDst io.Writer) (*MarketSignalSubscription, error) {
	if rawDst != nil {
		return nil, c.Post(ctx, marketSignalsBasePath, req, nil, rawDst)
	}
	var sub MarketSignalSubscription
	if err := c.Post(ctx, marketSignalsBasePath, req, &sub, nil); err != nil {
		return nil, err
	}
	return &sub, nil
}

// ListMarketSignalSubscriptions lists all market signal subscriptions.
func (c *Client) ListMarketSignalSubscriptions(ctx context.Context, limit, offset int, includeDeleted bool, rawDst io.Writer) (*MarketSignalSubscriptionListResponse, error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		q.Set("offset", strconv.Itoa(offset))
	}
	if includeDeleted {
		q.Set("includeDeleted", "true")
	}
	path := marketSignalsBasePath + "?" + q.Encode()
	if rawDst != nil {
		return nil, c.Get(ctx, path, nil, rawDst)
	}
	var resp MarketSignalSubscriptionListResponse
	if err := c.Get(ctx, path, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetMarketSignalSubscription retrieves a single subscription by ID.
func (c *Client) GetMarketSignalSubscription(ctx context.Context, id string, rawDst io.Writer) (*MarketSignalSubscription, error) {
	path := marketSignalsBasePath + "/" + url.PathEscape(id)
	if rawDst != nil {
		return nil, c.Get(ctx, path, nil, rawDst)
	}
	var sub MarketSignalSubscription
	if err := c.Get(ctx, path, &sub, nil); err != nil {
		return nil, err
	}
	return &sub, nil
}

// UpdateMarketSignalSubscription partially updates a subscription.
func (c *Client) UpdateMarketSignalSubscription(ctx context.Context, id string, req UpdateMarketSignalSubscriptionRequest, rawDst io.Writer) (*MarketSignalSubscription, error) {
	path := marketSignalsBasePath + "/" + url.PathEscape(id)
	if rawDst != nil {
		return nil, c.Patch(ctx, path, req, nil, rawDst)
	}
	var sub MarketSignalSubscription
	if err := c.Patch(ctx, path, req, &sub, nil); err != nil {
		return nil, err
	}
	return &sub, nil
}

// DeleteMarketSignalSubscription soft-deletes a subscription.
func (c *Client) DeleteMarketSignalSubscription(ctx context.Context, id string) error {
	path := marketSignalsBasePath + "/" + url.PathEscape(id)
	return c.Delete(ctx, path)
}

// PauseMarketSignalSubscription pauses an active subscription.
func (c *Client) PauseMarketSignalSubscription(ctx context.Context, id string, rawDst io.Writer) (*MarketSignalSubscription, error) {
	path := marketSignalsBasePath + "/" + url.PathEscape(id) + "/pause"
	if rawDst != nil {
		return nil, c.Post(ctx, path, nil, nil, rawDst)
	}
	var sub MarketSignalSubscription
	if err := c.Post(ctx, path, nil, &sub, nil); err != nil {
		return nil, err
	}
	return &sub, nil
}

// ResumeMarketSignalSubscription resumes a paused subscription.
func (c *Client) ResumeMarketSignalSubscription(ctx context.Context, id string, rawDst io.Writer) (*MarketSignalSubscription, error) {
	path := marketSignalsBasePath + "/" + url.PathEscape(id) + "/resume"
	if rawDst != nil {
		return nil, c.Post(ctx, path, nil, nil, rawDst)
	}
	var sub MarketSignalSubscription
	if err := c.Post(ctx, path, nil, &sub, nil); err != nil {
		return nil, err
	}
	return &sub, nil
}

// TriggerMarketSignalSubscription triggers an immediate run.
func (c *Client) TriggerMarketSignalSubscription(ctx context.Context, id string, rawDst io.Writer) (*MarketSignalSubscription, error) {
	path := marketSignalsBasePath + "/" + url.PathEscape(id) + "/trigger"
	if rawDst != nil {
		return nil, c.Post(ctx, path, nil, nil, rawDst)
	}
	var sub MarketSignalSubscription
	if err := c.Post(ctx, path, nil, &sub, nil); err != nil {
		return nil, err
	}
	return &sub, nil
}

// ListMarketSignals lists signals delivered by a subscription.
func (c *Client) ListMarketSignals(ctx context.Context, subscriptionID string, limit, offset int, rawDst io.Writer) (*MarketSignalListResponse, error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		q.Set("offset", strconv.Itoa(offset))
	}
	path := fmt.Sprintf("%s/%s/signals?%s", marketSignalsBasePath, url.PathEscape(subscriptionID), q.Encode())
	if rawDst != nil {
		return nil, c.Get(ctx, path, nil, rawDst)
	}
	var resp MarketSignalListResponse
	if err := c.Get(ctx, path, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}
