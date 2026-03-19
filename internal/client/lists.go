package client

import (
	"context"
	"fmt"
	"io"
	"net/url"
)

// --- Company Lists ---

// CompanyListFilter is the filter object used in company list create/update/search.
type CompanyListFilter struct {
	Domains    []string             `json:"domains,omitempty"`
	Names      []string             `json:"names,omitempty"`
	Industries []string             `json:"industries,omitempty"`
	Sizes      []string             `json:"sizes,omitempty"`
	Types      []string             `json:"types,omitempty"`
	Location   *CompanyListLocation `json:"location,omitempty"`
	Exclude    *CompanyListExclude  `json:"exclude,omitempty"`
}

// CompanyListLocation filters by geography.
type CompanyListLocation struct {
	Cities       []string `json:"cities,omitempty"`
	States       []string `json:"states,omitempty"`
	CountryCodes []string `json:"countryCodes,omitempty"`
}

// CompanyListExclude exclusion filters.
type CompanyListExclude struct {
	Industries []string `json:"industries,omitempty"`
	Sizes      []string `json:"sizes,omitempty"`
	Domains    []string `json:"domains,omitempty"`
}

// CompanyList represents a company list resource.
type CompanyList struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Filter    CompanyListFilter `json:"filter"`
	CreatedAt string            `json:"createdAt"`
	UpdatedAt string            `json:"updatedAt"`
}

// CompanyListsResponse wraps a paginated list of company lists.
type CompanyListsResponse struct {
	Items   []CompanyList `json:"items"`
	Total   int           `json:"total"`
	Limit   int           `json:"limit"`
	Offset  int           `json:"offset"`
	HasMore bool          `json:"hasMore"`
}

// CompanyListCompany is a company entry within a list.
type CompanyListCompany struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	Handle      string `json:"handle"`
	Industry    string `json:"industry"`
	Size        string `json:"size"`
	Type        string `json:"type"`
	Founded     int    `json:"founded"`
	City        string `json:"city"`
	State       string `json:"state"`
	CountryCode string `json:"countryCode"`
}

// CompaniesInListResponse wraps paginated companies in a list.
type CompaniesInListResponse struct {
	Items   []CompanyListCompany `json:"items"`
	Total   int                  `json:"total"`
	Limit   int                  `json:"limit"`
	Offset  int                  `json:"offset"`
	HasMore bool                 `json:"hasMore"`
}

// CompanySearchResponse wraps company search results.
type CompanySearchResponse struct {
	Companies []CompanyListCompany `json:"companies"`
	Total     int                  `json:"total"`
}

// CreateCompanyListRequest is the payload for POST /v1/companies/lists.
type CreateCompanyListRequest struct {
	Name   string            `json:"name"`
	Filter CompanyListFilter `json:"filter"`
}

// UpdateCompanyListRequest is the payload for PUT /v1/companies/lists/{id}.
type UpdateCompanyListRequest struct {
	Name   string            `json:"name"`
	Filter CompanyListFilter `json:"filter"`
}

// CompanySearchRequest is the payload for POST /v1/companies/search.
type CompanySearchRequest struct {
	Filter CompanyListFilter `json:"filter"`
}

// HubSpotPropertyFilter is the filter for importing from HubSpot.
type HubSpotPropertyFilter struct {
	PropertyName string `json:"propertyName"`
	Operator     string `json:"operator"`
	Value        string `json:"value,omitempty"`
}

// ImportCompanyListSource specifies the import source.
type ImportCompanyListSource struct {
	Type   string                `json:"type"` // "hubspot"
	Filter HubSpotPropertyFilter `json:"filter"`
}

// ImportCompanyListRequest is the payload for POST /v1/companies/lists/import.
type ImportCompanyListRequest struct {
	Name   string                  `json:"name"`
	Source ImportCompanyListSource `json:"source"`
}

func (c *Client) CreateCompanyList(ctx context.Context, req CreateCompanyListRequest, rawDst io.Writer) (*CompanyList, error) {
	if rawDst != nil {
		return nil, c.Post(ctx, "/v1/companies/lists", req, nil, rawDst)
	}
	var list CompanyList
	if err := c.Post(ctx, "/v1/companies/lists", req, &list, nil); err != nil {
		return nil, err
	}
	return &list, nil
}

func (c *Client) ListCompanyLists(ctx context.Context, limit, offset int, rawDst io.Writer) (*CompanyListsResponse, error) {
	path := fmt.Sprintf("/v1/companies/lists?limit=%d&offset=%d", limit, offset)
	if rawDst != nil {
		return nil, c.Get(ctx, path, nil, rawDst)
	}
	var resp CompanyListsResponse
	if err := c.Get(ctx, path, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetCompanyList(ctx context.Context, id string, rawDst io.Writer) (*CompanyList, error) {
	path := "/v1/companies/lists/" + url.PathEscape(id)
	if rawDst != nil {
		return nil, c.Get(ctx, path, nil, rawDst)
	}
	var list CompanyList
	if err := c.Get(ctx, path, &list, nil); err != nil {
		return nil, err
	}
	return &list, nil
}

func (c *Client) UpdateCompanyList(ctx context.Context, id string, req UpdateCompanyListRequest, rawDst io.Writer) (*CompanyList, error) {
	path := "/v1/companies/lists/" + url.PathEscape(id)
	if rawDst != nil {
		return nil, c.Put(ctx, path, req, nil, rawDst)
	}
	var list CompanyList
	if err := c.Put(ctx, path, req, &list, nil); err != nil {
		return nil, err
	}
	return &list, nil
}

func (c *Client) DeleteCompanyList(ctx context.Context, id string) error {
	return c.Delete(ctx, "/v1/companies/lists/"+url.PathEscape(id))
}

func (c *Client) GetCompaniesInList(ctx context.Context, id string, limit, offset int, rawDst io.Writer) (*CompaniesInListResponse, error) {
	path := fmt.Sprintf("/v1/companies/lists/%s/companies?limit=%d&offset=%d", url.PathEscape(id), limit, offset)
	if rawDst != nil {
		return nil, c.Get(ctx, path, nil, rawDst)
	}
	var resp CompaniesInListResponse
	if err := c.Get(ctx, path, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) SearchCompanies(ctx context.Context, req CompanySearchRequest, rawDst io.Writer) (*CompanySearchResponse, error) {
	if rawDst != nil {
		return nil, c.Post(ctx, "/v1/companies/search", req, nil, rawDst)
	}
	var resp CompanySearchResponse
	if err := c.Post(ctx, "/v1/companies/search", req, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ImportCompanyList(ctx context.Context, req ImportCompanyListRequest, rawDst io.Writer) (*CompanyList, error) {
	if rawDst != nil {
		return nil, c.Post(ctx, "/v1/companies/lists/import", req, nil, rawDst)
	}
	var list CompanyList
	if err := c.Post(ctx, "/v1/companies/lists/import", req, &list, nil); err != nil {
		return nil, err
	}
	return &list, nil
}

// --- Contact Lists ---

// ContactListFilters is the filter object for contact list create.
type ContactListFilters struct {
	CompanyLinkedInURLs []string `json:"companyLinkedInUrls,omitempty"`
	JobTitles           []string `json:"jobTitles,omitempty"`
	Keywords            string   `json:"keywords,omitempty"`
	Countries           []string `json:"countries,omitempty"`
}

// ContactList represents a contact list resource.
type ContactList struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Filters      ContactListFilters `json:"filters"`
	ContactCount int                `json:"contactCount"`
	CreatedAt    string             `json:"createdAt"`
	UpdatedAt    string             `json:"updatedAt"`
}

// ContactListsResponse wraps a paginated list of contact lists.
type ContactListsResponse struct {
	Items   []ContactList `json:"items"`
	Total   int           `json:"total"`
	Limit   int           `json:"limit"`
	Offset  int           `json:"offset"`
	HasMore bool          `json:"hasMore"`
}

// ContactListItem is a contact entry within a list.
type ContactListItem struct {
	ID          string `json:"id"`
	ListID      string `json:"listId"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	FullName    string `json:"fullName"`
	Headline    string `json:"headline"`
	Role        string `json:"role"`
	CompanyName string `json:"companyName"`
	Location    string `json:"location"`
}

// ContactsInListResponse wraps paginated contacts in a list.
type ContactsInListResponse struct {
	Items   []ContactListItem `json:"items"`
	Total   int               `json:"total"`
	Limit   int               `json:"limit"`
	Offset  int               `json:"offset"`
	HasMore bool              `json:"hasMore"`
}

// CreateContactListRequest is the payload for POST /v1/contacts/lists.
type CreateContactListRequest struct {
	Name    string             `json:"name"`
	Filters ContactListFilters `json:"filters"`
}

// UpdateContactListRequest is the payload for PUT /v1/contacts/lists/{id}.
type UpdateContactListRequest struct {
	Name string `json:"name"`
}

func (c *Client) CreateContactList(ctx context.Context, req CreateContactListRequest, rawDst io.Writer) (*ContactList, error) {
	if rawDst != nil {
		return nil, c.Post(ctx, "/v1/contacts/lists", req, nil, rawDst)
	}
	var list ContactList
	if err := c.Post(ctx, "/v1/contacts/lists", req, &list, nil); err != nil {
		return nil, err
	}
	return &list, nil
}

func (c *Client) ListContactLists(ctx context.Context, limit, offset int, rawDst io.Writer) (*ContactListsResponse, error) {
	path := fmt.Sprintf("/v1/contacts/lists?limit=%d&offset=%d", limit, offset)
	if rawDst != nil {
		return nil, c.Get(ctx, path, nil, rawDst)
	}
	var resp ContactListsResponse
	if err := c.Get(ctx, path, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetContactList(ctx context.Context, id string, rawDst io.Writer) (*ContactList, error) {
	path := "/v1/contacts/lists/" + url.PathEscape(id)
	if rawDst != nil {
		return nil, c.Get(ctx, path, nil, rawDst)
	}
	var list ContactList
	if err := c.Get(ctx, path, &list, nil); err != nil {
		return nil, err
	}
	return &list, nil
}

func (c *Client) UpdateContactList(ctx context.Context, id string, req UpdateContactListRequest, rawDst io.Writer) (*ContactList, error) {
	path := "/v1/contacts/lists/" + url.PathEscape(id)
	if rawDst != nil {
		return nil, c.Put(ctx, path, req, nil, rawDst)
	}
	var list ContactList
	if err := c.Put(ctx, path, req, &list, nil); err != nil {
		return nil, err
	}
	return &list, nil
}

func (c *Client) DeleteContactList(ctx context.Context, id string) error {
	return c.Delete(ctx, "/v1/contacts/lists/"+url.PathEscape(id))
}

func (c *Client) GetContactsInList(ctx context.Context, id string, limit, offset int, rawDst io.Writer) (*ContactsInListResponse, error) {
	path := fmt.Sprintf("/v1/contacts/lists/%s/contacts?limit=%d&offset=%d",
		url.PathEscape(id), limit, offset)
	if rawDst != nil {
		return nil, c.Get(ctx, path, nil, rawDst)
	}
	var resp ContactsInListResponse
	if err := c.Get(ctx, path, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ContactSearchRequest is the payload for POST /v1/contacts/search.
type ContactSearchRequest struct {
	CompanyLinkedInURLs []string `json:"companyLinkedInUrls,omitempty"`
	FirstName           string   `json:"firstName,omitempty"`
	LastName            string   `json:"lastName,omitempty"`
	JobTitles           []string `json:"jobTitles,omitempty"`
	Keywords            string   `json:"keywords,omitempty"`
	Countries           []string `json:"countries,omitempty"`
}

// ContactSearchResult is a single contact returned from search.
type ContactSearchResult struct {
	FirstName                        string   `json:"firstName"`
	LastName                         string   `json:"lastName"`
	FullName                         string   `json:"fullName"`
	Headline                         string   `json:"headline"`
	Role                             string   `json:"role"`
	CompanyName                      string   `json:"companyName"`
	Location                         string   `json:"location"`
	Seniority                        []string `json:"seniority"`
	LinkedInProfileURL               string   `json:"linkedInProfileUrl"`
	LinkedInSalesNavigatorProfileURL string   `json:"linkedInSalesNavigatorProfileUrl"`
}

// ContactSearchResponse wraps contact search results.
type ContactSearchResponse struct {
	Contacts          []ContactSearchResult `json:"contacts"`
	Count             int                   `json:"count"`
	SalesNavConnected bool                  `json:"salesNavConnected"`
}

func (c *Client) SearchContacts(ctx context.Context, req ContactSearchRequest, rawDst io.Writer) (*ContactSearchResponse, error) {
	if rawDst != nil {
		return nil, c.Post(ctx, "/v1/contacts/search", req, nil, rawDst)
	}
	var resp ContactSearchResponse
	if err := c.Post(ctx, "/v1/contacts/search", req, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}
