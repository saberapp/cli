package client

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"
)

// Subscription represents a signal subscription resource.
type Subscription struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Question         string     `json:"question"`
	AnswerType       string     `json:"answerType"`
	Frequency        string     `json:"frequency"`
	CronExpression   string     `json:"cronExpression"`
	Timezone         string     `json:"timezone"`
	Status           string     `json:"status"` // active, stopped
	ListID           string     `json:"listId"`
	LastRunAt        *time.Time `json:"lastRunAt"`
	NextRunAt        *time.Time `json:"nextRunAt"`
	CreatedAt        time.Time  `json:"createdAt"`
}

// SubscriptionsResponse wraps a paginated list of subscriptions.
type SubscriptionsResponse struct {
	Items   []Subscription `json:"items"`
	Total   int            `json:"total"`
	HasMore bool           `json:"hasMore"`
}

// CreateSubscriptionRequest is the payload for POST /v1/companies/signals/subscriptions.
type CreateSubscriptionRequest struct {
	// Reference an existing template, or provide name+question to create inline.
	SignalTemplateID string `json:"signalTemplateId,omitempty"`
	Name             string `json:"name,omitempty"`
	Question         string `json:"question,omitempty"`
	AnswerType       string `json:"answerType,omitempty"`

	// Exactly one of Frequency or CronExpression must be set.
	Frequency      string `json:"frequency,omitempty"`
	CronExpression string `json:"cronExpression,omitempty"`
	Timezone       string `json:"timezone,omitempty"`

	ListID string `json:"listId"`
}

// CreateSubscription creates a signal subscription (initially stopped).
func (c *Client) CreateSubscription(ctx context.Context, req CreateSubscriptionRequest, rawDst io.Writer) (*Subscription, error) {
	if rawDst != nil {
		return nil, c.Post(ctx, "/v1/companies/signals/subscriptions", req, nil, rawDst)
	}
	var sub Subscription
	if err := c.Post(ctx, "/v1/companies/signals/subscriptions", req, &sub, nil); err != nil {
		return nil, err
	}
	return &sub, nil
}

// ListSubscriptions returns all subscriptions for the organisation.
func (c *Client) ListSubscriptions(ctx context.Context, limit, offset int, rawDst io.Writer) (*SubscriptionsResponse, error) {
	path := fmt.Sprintf("/v1/companies/signals/subscriptions?limit=%d&offset=%d", limit, offset)
	if rawDst != nil {
		return nil, c.Get(ctx, path, nil, rawDst)
	}
	var resp SubscriptionsResponse
	if err := c.Get(ctx, path, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetSubscription retrieves a single subscription by ID.
func (c *Client) GetSubscription(ctx context.Context, id string, rawDst io.Writer) (*Subscription, error) {
	path := "/v1/companies/signals/subscriptions/" + url.PathEscape(id)
	if rawDst != nil {
		return nil, c.Get(ctx, path, nil, rawDst)
	}
	var sub Subscription
	if err := c.Get(ctx, path, &sub, nil); err != nil {
		return nil, err
	}
	return &sub, nil
}

// StartSubscription activates the schedule for a subscription.
func (c *Client) StartSubscription(ctx context.Context, id string) (*Subscription, error) {
	path := "/v1/companies/signals/subscriptions/" + url.PathEscape(id) + "/start"
	var sub Subscription
	if err := c.Post(ctx, path, nil, &sub, nil); err != nil {
		return nil, err
	}
	return &sub, nil
}

// StopSubscription pauses the schedule for a subscription.
func (c *Client) StopSubscription(ctx context.Context, id string) (*Subscription, error) {
	path := "/v1/companies/signals/subscriptions/" + url.PathEscape(id) + "/stop"
	var sub Subscription
	if err := c.Post(ctx, path, nil, &sub, nil); err != nil {
		return nil, err
	}
	return &sub, nil
}

// TriggerSubscription runs a subscription immediately, regardless of its schedule.
func (c *Client) TriggerSubscription(ctx context.Context, id string) (*Subscription, error) {
	path := "/v1/companies/signals/subscriptions/" + url.PathEscape(id) + "/trigger"
	var sub Subscription
	if err := c.Post(ctx, path, nil, &sub, nil); err != nil {
		return nil, err
	}
	return &sub, nil
}
