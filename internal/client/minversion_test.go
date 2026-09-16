package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func resetMinVersion(t *testing.T) {
	t.Helper()
	minVersion.Lock()
	minVersion.value = ""
	minVersion.Unlock()
	t.Cleanup(func() {
		minVersion.Lock()
		minVersion.value = ""
		minVersion.Unlock()
	})
}

func serveStatus(t *testing.T, status int, minHeader string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if minHeader != "" {
			w.Header().Set(MinVersionHeader, minHeader)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if status >= 400 {
			_, _ = w.Write([]byte(`{"error":{"message":"saber CLI v0.1.0 is no longer supported. Run \"saber update\" to upgrade."}}`))
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestMinSupportedVersion_RecordsHeader(t *testing.T) {
	resetMinVersion(t)
	srv := serveStatus(t, http.StatusOK, "0.2.0")

	c := New(srv.URL, "key", "0.1.0", false, io.Discard)
	if err := c.Get(context.Background(), "/v1/credits", nil, nil); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got := MinSupportedVersion(); got != "0.2.0" {
		t.Errorf("MinSupportedVersion() = %q, want %q", got, "0.2.0")
	}
}

func TestMinSupportedVersion_StripsVPrefix(t *testing.T) {
	resetMinVersion(t)
	srv := serveStatus(t, http.StatusOK, "v0.3.1")

	c := New(srv.URL, "key", "0.1.0", false, io.Discard)
	if err := c.Get(context.Background(), "/v1/credits", nil, nil); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got := MinSupportedVersion(); got != "0.3.1" {
		t.Errorf("MinSupportedVersion() = %q, want %q", got, "0.3.1")
	}
}

func TestMinSupportedVersion_HeaderAbsent(t *testing.T) {
	resetMinVersion(t)
	srv := serveStatus(t, http.StatusOK, "")

	c := New(srv.URL, "key", "0.1.0", false, io.Discard)
	if err := c.Get(context.Background(), "/v1/credits", nil, nil); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got := MinSupportedVersion(); got != "" {
		t.Errorf("MinSupportedVersion() = %q, want empty", got)
	}
}

func TestMinSupportedVersion_RecordedOn426(t *testing.T) {
	resetMinVersion(t)
	srv := serveStatus(t, http.StatusUpgradeRequired, "0.2.0")

	c := New(srv.URL, "key", "0.1.0", false, io.Discard)
	err := c.Get(context.Background(), "/v1/credits", nil, nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusUpgradeRequired {
		t.Fatalf("Get error = %v, want APIError 426", err)
	}
	if got := MinSupportedVersion(); got != "0.2.0" {
		t.Errorf("MinSupportedVersion() = %q, want %q", got, "0.2.0")
	}
}
