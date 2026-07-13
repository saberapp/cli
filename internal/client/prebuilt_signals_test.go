package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestPrebuiltSignalKeys pins the hardcoded v1 key set (SEL-3372 catalogue
// endpoint is deferred; the CLI is the source of truth for now).
func TestPrebuiltSignalKeys(t *testing.T) {
	want := []string{"funding", "mna", "tech", "open-jobs", "firmographics"}
	if len(PrebuiltSignalKeys) != len(want) {
		t.Fatalf("PrebuiltSignalKeys = %v, want %v", PrebuiltSignalKeys, want)
	}
	for i, k := range want {
		if PrebuiltSignalKeys[i] != k {
			t.Errorf("PrebuiltSignalKeys[%d] = %q, want %q", i, PrebuiltSignalKeys[i], k)
		}
	}
}

// TestCreatePrebuiltSignalSyncRequestEncoding verifies the sync call hits
// POST /v1/signals/sync/<key>, carries domain/forceRefresh in the JSON body,
// and sets the X-Sbr-Timeout-Sec header.
func TestCreatePrebuiltSignalSyncRequestEncoding(t *testing.T) {
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
	req := CreatePrebuiltSignalRequest{Domain: "acme.com", ForceRefresh: true}
	if _, err := c.CreatePrebuiltSignalSync(context.Background(), "funding", req, 90, nil); err != nil {
		t.Fatalf("CreatePrebuiltSignalSync: %v", err)
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

// TestCreatePrebuiltSignalSyncResponseDecoding verifies the response decodes
// into the existing Signal struct (no new response type).
func TestCreatePrebuiltSignalSyncResponseDecoding(t *testing.T) {
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
	sig, err := c.CreatePrebuiltSignalSync(context.Background(), "funding", CreatePrebuiltSignalRequest{Domain: "acme.com"}, 120, nil)
	if err != nil {
		t.Fatalf("CreatePrebuiltSignalSync: %v", err)
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

// TestCreatePrebuiltSignalAsyncPath verifies the async call hits
// POST /v1/signals/<key> (no /sync segment, no timeout header).
func TestCreatePrebuiltSignalAsyncPath(t *testing.T) {
	var (
		gotMethod  string
		gotPath    string
		gotTimeout string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotTimeout = r.Header.Get("X-Sbr-Timeout-Sec")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"sig-7","status":"processing"}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "test-key", "test", false, io.Discard)
	sig, err := c.CreatePrebuiltSignal(context.Background(), "mna", CreatePrebuiltSignalRequest{Domain: "acme.com", WebhookURL: "https://example.com/hook"})
	if err != nil {
		t.Fatalf("CreatePrebuiltSignal: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %s, want POST", gotMethod)
	}
	if gotPath != "/v1/signals/mna" {
		t.Errorf("path = %s, want /v1/signals/mna", gotPath)
	}
	if gotTimeout != "" {
		t.Errorf("X-Sbr-Timeout-Sec = %q, want unset on async path", gotTimeout)
	}
	if sig.ID != "sig-7" || sig.Status != SignalStatusProcessing {
		t.Errorf("id/status = %s/%s, want sig-7/processing", sig.ID, sig.Status)
	}
}

// TestCreatePrebuiltSignalTechBodySerialization pins the tech key's wire
// contract: setting category omits technology from the JSON body entirely
// (omitempty), and vice-versa — the server rejects requests carrying both.
func TestCreatePrebuiltSignalTechBodySerialization(t *testing.T) {
	tests := []struct {
		name    string
		req     CreatePrebuiltSignalRequest
		wantKey string
		wantVal string
		omitKey string
	}{
		{
			name:    "category set omits technology",
			req:     CreatePrebuiltSignalRequest{Domain: "acme.com", Category: "crm"},
			wantKey: "category",
			wantVal: "crm",
			omitKey: "technology",
		},
		{
			name:    "technology set omits category",
			req:     CreatePrebuiltSignalRequest{Domain: "acme.com", Technology: "salesforce"},
			wantKey: "technology",
			wantVal: "salesforce",
			omitKey: "category",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotBody map[string]any
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &gotBody)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":"sig-9","status":"processing"}`))
			}))
			defer srv.Close()

			c := New(srv.URL, "test-key", "test", false, io.Discard)
			if _, err := c.CreatePrebuiltSignal(context.Background(), "tech", tt.req); err != nil {
				t.Fatalf("CreatePrebuiltSignal: %v", err)
			}
			if gotBody[tt.wantKey] != tt.wantVal {
				t.Errorf("body %s = %v, want %q", tt.wantKey, gotBody[tt.wantKey], tt.wantVal)
			}
			if _, ok := gotBody[tt.omitKey]; ok {
				t.Errorf("body contains %q; want it omitted (omitempty)", tt.omitKey)
			}
		})
	}
}
