package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// captureOrgID installs a recorder for the observed organization id and
// restores the previous hook afterwards.
func captureOrgID(t *testing.T) *string {
	t.Helper()
	var seen string
	orig := OnOrganizationID
	OnOrganizationID = func(id string) { seen = id }
	t.Cleanup(func() { OnOrganizationID = orig })
	return &seen
}

func TestObserveOrganizationID(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"template response", `{"id":"t1","organizationId":"org_abc","name":"x"}`, "org_abc"},
		{"spaced json", `{"organizationId" : "org_spaced"}`, "org_spaced"},
		{"list response uses the first match", `{"items":[{"organizationId":"org_1"},{"organizationId":"org_2"}]}`, "org_1"},
		{"no organization field", `{"remainingCredits":10}`, ""},
		{"empty body", ``, ""},
		{"empty value is ignored", `{"organizationId":""}`, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seen := captureOrgID(t)
			observeOrganizationID([]byte(tt.body))
			if *seen != tt.want {
				t.Errorf("observed %q, want %q", *seen, tt.want)
			}
		})
	}
}

func TestObserveOrganizationID_OversizedBodyIsSkipped(t *testing.T) {
	seen := captureOrgID(t)

	big := strings.Repeat("x", maxOrgScanBytes+1) + `{"organizationId":"org_big"}`
	observeOrganizationID([]byte(big))

	if *seen != "" {
		t.Errorf("a very large body was scanned: %q", *seen)
	}
}

func TestObserveOrganizationID_NoHookIsSafe(t *testing.T) {
	orig := OnOrganizationID
	OnOrganizationID = nil
	t.Cleanup(func() { OnOrganizationID = orig })

	observeOrganizationID([]byte(`{"organizationId":"org_abc"}`))
}

// TestDo_LearnsOrganizationID proves the id is picked up from a real response
// the CLI already receives, with no extra API call.
func TestDo_LearnsOrganizationID(t *testing.T) {
	seen := captureOrgID(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"tmpl_1","organizationId":"org_from_response"}`))
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, "sk_live_test", "1.2.3", false, io.Discard)
	var dst map[string]any
	if err := c.Get(context.Background(), "/v1/companies/signals/templates", &dst, nil); err != nil {
		t.Fatalf("Get: %v", err)
	}

	if *seen != "org_from_response" {
		t.Errorf("organization id = %q, want org_from_response", *seen)
	}
}

// TestDo_ErrorResponseDoesNotLearn keeps a 4xx body from seeding the cache.
func TestDo_ErrorResponseDoesNotLearn(t *testing.T) {
	seen := captureOrgID(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"organizationId":"org_forbidden","error":{"message":"no"}}`))
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, "sk_live_test", "1.2.3", false, io.Discard)
	if err := c.Get(context.Background(), "/v1/credits", &map[string]any{}, nil); err == nil {
		t.Fatal("expected an API error")
	}

	if *seen != "" {
		t.Errorf("an error response seeded the organization id: %q", *seen)
	}
}
