package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestLibrarySignalKeys pins the hardcoded v1 key set (SEL-3372 catalogue
// endpoint is deferred; the CLI is the source of truth for now).
func TestLibrarySignalKeys(t *testing.T) {
	want := []string{"funding", "mna", "tech", "open-jobs", "firmographics"}
	if len(LibrarySignalKeys) != len(want) {
		t.Fatalf("LibrarySignalKeys = %v, want %v", LibrarySignalKeys, want)
	}
	for i, k := range want {
		if LibrarySignalKeys[i] != k {
			t.Errorf("LibrarySignalKeys[%d] = %q, want %q", i, LibrarySignalKeys[i], k)
		}
	}
}

// TestCreateLibrarySignalSyncRequestEncoding verifies the sync call hits
// POST /v1/signals/sync/<key>, carries domain/forceRefresh in the JSON body,
// and sets the X-Sbr-Timeout-Sec header.
func TestCreateLibrarySignalSyncRequestEncoding(t *testing.T) {
	var (
		gotMethod  string
		gotPath    string
		gotTimeout string
		gotBody    map[string]any
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotTimeout = r.Header.Get("X-Sbr-Timeout-Sec")
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"sig-1","status":"completed"}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "test-key", "test", false, io.Discard)
	req := CreateLibrarySignalRequest{Domain: "acme.com", ForceRefresh: true}
	if _, err := c.CreateLibrarySignalSync(context.Background(), "funding", req, 90, nil); err != nil {
		t.Fatalf("CreateLibrarySignalSync: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %s, want POST", gotMethod)
	}
	if gotPath != "/v1/signals/sync/funding" {
		t.Errorf("path = %s, want /v1/signals/sync/funding", gotPath)
	}
	if gotTimeout != "90" {
		t.Errorf("X-Sbr-Timeout-Sec = %q, want %q", gotTimeout, "90")
	}
	if gotBody["domain"] != "acme.com" {
		t.Errorf("body domain = %v, want acme.com", gotBody["domain"])
	}
	if gotBody["forceRefresh"] != true {
		t.Errorf("body forceRefresh = %v, want true", gotBody["forceRefresh"])
	}
}

// TestCreateLibrarySignalSyncResponseDecoding verifies the response decodes
// into the existing Signal struct (no new response type).
func TestCreateLibrarySignalSyncResponseDecoding(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"sig-42","status":"completed","domain":"acme.com",
			"answer":true,"confidence":0.91,"reasoning":"Raised a Series B in May.",
			"sources":[{"title":"TechCrunch","url":"https://techcrunch.com/x"}]
		}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "test-key", "test", false, io.Discard)
	sig, err := c.CreateLibrarySignalSync(context.Background(), "funding", CreateLibrarySignalRequest{Domain: "acme.com"}, 120, nil)
	if err != nil {
		t.Fatalf("CreateLibrarySignalSync: %v", err)
	}
	if sig.ID != "sig-42" || sig.Status != SignalStatusCompleted {
		t.Errorf("id/status = %s/%s, want sig-42/completed", sig.ID, sig.Status)
	}
	if sig.Confidence == nil || *sig.Confidence != 0.91 {
		t.Errorf("confidence = %v, want 0.91", sig.Confidence)
	}
	if len(sig.Sources) != 1 || sig.Sources[0].URL != "https://techcrunch.com/x" {
		t.Errorf("sources not decoded: %+v", sig.Sources)
	}
}
