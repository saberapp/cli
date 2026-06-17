package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestSearchContactsRequestEncoding verifies the request body matches the
// OpenAPI ContactSearchRequest schema — in particular that seniority is sent
// as `seniorityLevels` (not `seniorities`) and that limit/offset are wired.
func TestSearchContactsRequestEncoding(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[],"total":0,"limit":25,"offset":0,"hasMore":false,"salesNavConnected":true}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "test-key", "test", false, io.Discard)
	req := ContactSearchRequest{
		Departments:     []string{"Sales"},
		SeniorityLevels: []string{"CXO", "Vice President"},
		Limit:           50,
		Offset:          10,
	}
	if _, err := c.SearchContacts(context.Background(), req, nil); err != nil {
		t.Fatalf("SearchContacts: %v", err)
	}

	if _, ok := got["seniorities"]; ok {
		t.Errorf("request used legacy key 'seniorities'; API expects 'seniorityLevels'")
	}
	if _, ok := got["seniorityLevels"]; !ok {
		t.Errorf("request missing 'seniorityLevels' key; got %v", got)
	}
	if got["limit"] != float64(50) {
		t.Errorf("limit = %v, want 50", got["limit"])
	}
	if got["offset"] != float64(10) {
		t.Errorf("offset = %v, want 10", got["offset"])
	}
}

// TestSearchContactsResponseDecoding verifies the paginated envelope
// (items/total/limit/offset/hasMore) is decoded correctly.
func TestSearchContactsResponseDecoding(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items":[{"fullName":"John Doe","role":"VP Engineering","companyName":"Acme","location":"SF"}],
			"total":150,"limit":25,"offset":0,"hasMore":true,"salesNavConnected":true
		}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "test-key", "test", false, io.Discard)
	resp, err := c.SearchContacts(context.Background(), ContactSearchRequest{FirstName: "John"}, nil)
	if err != nil {
		t.Fatalf("SearchContacts: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].FullName != "John Doe" {
		t.Errorf("items not decoded: %+v", resp.Items)
	}
	if resp.Total != 150 {
		t.Errorf("total = %d, want 150", resp.Total)
	}
	if !resp.HasMore {
		t.Errorf("hasMore = false, want true")
	}
	if !resp.SalesNavConnected {
		t.Errorf("salesNavConnected = false, want true")
	}
}
