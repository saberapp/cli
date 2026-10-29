package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
)

type FindPhoneRequest struct {
	ContactProfileURL string `json:"contactProfileUrl,omitempty"`
	FirstName         string `json:"firstName,omitempty"`
	LastName          string `json:"lastName,omitempty"`
	CompanyName       string `json:"companyName,omitempty"`
	CompanyDomain     string `json:"companyDomain,omitempty"`
}

type PhoneRun struct {
	ID                  string       `json:"id"`
	Object              string       `json:"object"`
	EntityType          string       `json:"entityType"`
	Type                string       `json:"type"`
	Status              string       `json:"status"`
	Entity              PhoneEntity  `json:"entity"`
	Result              *PhoneResult `json:"result"`
	ResultSchemaVersion string       `json:"resultSchemaVersion"`
	Credits             *RunCredits  `json:"credits"`
	CreatedAt           string       `json:"createdAt"`
	CompletedAt         string       `json:"completedAt"`
	Error               string       `json:"error"`
	ErrorCode           string       `json:"errorCode"`
	Retryable           bool         `json:"retryable"`
}

type PhoneEntity struct {
	ContactProfileURL string `json:"contactProfileUrl,omitempty"`
	Domain            string `json:"domain,omitempty"`
}

type PhoneResult struct {
	Found bool   `json:"found"`
	Phone *Phone `json:"phone"`
}

type Phone struct {
	E164       string `json:"e164"`
	Display    string `json:"display"`
	LineType   string `json:"lineType"`
	LineStatus string `json:"lineStatus"`
	Country    string `json:"country"`
}

type RunCredits struct {
	State        string   `json:"state"`
	Amount       *float64 `json:"amount,omitempty"`
	PriceVersion string   `json:"priceVersion,omitempty"`
}

func (c *Client) FindPhone(ctx context.Context, req FindPhoneRequest, waitSec int, rawDst io.Writer) (*PhoneRun, error) {
	headers := map[string]string{}
	if waitSec > 0 {
		headers["X-Sbr-Timeout-Sec"] = strconv.Itoa(waitSec)
	}
	if rawDst != nil {
		return nil, c.PostWithHeaders(ctx, "/v1/contacts/enrich/phone", headers, req, nil, rawDst)
	}
	var run PhoneRun
	if err := c.PostWithHeaders(ctx, "/v1/contacts/enrich/phone", headers, req, &run, nil); err != nil {
		return nil, err
	}
	return &run, nil
}

func (c *Client) GetPhoneRun(ctx context.Context, id string, rawDst io.Writer) (*PhoneRun, error) {
	var body bytes.Buffer
	if err := c.Get(ctx, "/v1/enrichment-runs/"+url.PathEscape(id), nil, &body); err != nil {
		return nil, err
	}
	var run PhoneRun
	if err := json.Unmarshal(body.Bytes(), &run); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if run.Type != "phone" {
		return nil, fmt.Errorf("run %s is not a phone lookup", id)
	}
	if rawDst != nil {
		if _, err := io.Copy(rawDst, &body); err != nil {
			return nil, err
		}
	}
	return &run, nil
}
