package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetCreditsDecodesFractionalBalance(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/credits" {
			t.Fatalf("path = %q, want /v1/credits", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"remainingCredits":951332.7}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "test-key", "test", false, io.Discard)
	resp, err := c.GetCredits(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetCredits: %v", err)
	}
	if resp.RemainingCredits != 951332.7 {
		t.Fatalf("RemainingCredits = %v, want 951332.7", resp.RemainingCredits)
	}
}

func TestGetCreditsDecodesWholeNumberBalance(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"remainingCredits":58}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "test-key", "test", false, io.Discard)
	resp, err := c.GetCredits(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetCredits: %v", err)
	}
	if resp.RemainingCredits != 58 {
		t.Fatalf("RemainingCredits = %v, want 58", resp.RemainingCredits)
	}
}
