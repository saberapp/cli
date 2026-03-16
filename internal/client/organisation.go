package client

import (
	"context"
	"io"
)

// OrganisationDescription holds the description sub-fields of an organisation.
type OrganisationDescription struct {
	General   string `json:"general,omitempty"`
	Products  string `json:"products,omitempty"`
	UseCases  string `json:"useCases,omitempty"`
	ValueProp string `json:"valueProp,omitempty"`
}

// Organisation is returned by GET /v1/organisation.
type Organisation struct {
	Name        string                  `json:"name"`
	Website     string                  `json:"website"`
	Description OrganisationDescription `json:"description"`
}

// UpdateOrganisationRequest is the body for PUT /v1/organisation.
type UpdateOrganisationRequest struct {
	Name        string                   `json:"name,omitempty"`
	Website     string                   `json:"website,omitempty"`
	Description *OrganisationDescription `json:"description,omitempty"`
}

// GetOrganisation returns the organisation profile.
func (c *Client) GetOrganisation(ctx context.Context, rawDst io.Writer) (*Organisation, error) {
	if rawDst != nil {
		return nil, c.Get(ctx, "/v1/organisation", nil, rawDst)
	}
	var resp Organisation
	if err := c.Get(ctx, "/v1/organisation", &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateOrganisation updates the organisation profile.
func (c *Client) UpdateOrganisation(ctx context.Context, req UpdateOrganisationRequest, rawDst io.Writer) (*Organisation, error) {
	if rawDst != nil {
		return nil, c.Put(ctx, "/v1/organisation", req, nil, rawDst)
	}
	var resp Organisation
	if err := c.Put(ctx, "/v1/organisation", req, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}
