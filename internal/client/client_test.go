package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestAttemptTimeout(t *testing.T) {
	c := New("", "test-key", "test", false, io.Discard)
	if c.httpClient.Timeout != 0 {
		t.Fatal("http.Client must not impose a second timeout")
	}
	for _, tt := range []struct {
		header string
		want   time.Duration
	}{
		{"", 60 * time.Second},
		{"0", 60 * time.Second},
		{"-1", 60 * time.Second},
		{"invalid", 60 * time.Second},
		{"1.5", 60 * time.Second},
		{"1", 60 * time.Second},
		{"45", 60 * time.Second},
		{"60", 75 * time.Second},
		{"120", 135 * time.Second},
		{"900", 915 * time.Second},
	} {
		t.Run(tt.header, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("X-Sbr-Timeout-Sec", tt.header)
			if got := c.attemptTimeout(req); got != tt.want {
				t.Fatalf("timeout = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestRequestAttemptDeadline(t *testing.T) {
	for _, tt := range []struct {
		name     string
		header   string
		flush    bool
		wantFail bool
	}{
		{name: "server wait exceeds default", header: "1"},
		{name: "default timeout", wantFail: true},
		{name: "body read timeout", flush: true, wantFail: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var attempts atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				attempts.Add(1)
				if tt.flush {
					w.WriteHeader(http.StatusOK)
					w.(http.Flusher).Flush()
				}
				select {
				case <-r.Context().Done():
					return
				case <-time.After(150 * time.Millisecond):
					_, _ = w.Write([]byte(`{"ok":true}`))
				}
			}))
			defer srv.Close()
			c := New(srv.URL, "test-key", "test", false, io.Discard)
			c.timeout = 30 * time.Millisecond
			err := c.Do(context.Background(), http.MethodPost, "/", nil, map[string]string{"X-Sbr-Timeout-Sec": tt.header}, nil, nil)
			if tt.wantFail {
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("error = %v, want deadline exceeded", err)
				}
				if attempts.Load() != maxRetries {
					t.Fatalf("attempts = %d, want %d", attempts.Load(), maxRetries)
				}
			} else if err != nil || attempts.Load() != 1 {
				t.Fatalf("error = %v, attempts = %d", err, attempts.Load())
			}
		})
	}
}

func TestRateLimitWaitUsesCallerContext(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"statusCode":429,"message":"try later","retryAfter":1}`))
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()
	c := New(srv.URL, "test-key", "test", false, io.Discard)
	c.timeout = 100 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.Get(ctx, "/", nil, nil); err != nil {
		t.Fatal(err)
	}
	if attempts.Load() != 2 {
		t.Fatalf("attempts = %d, want 2", attempts.Load())
	}
}

func TestServerWaitRespectsCallerDeadline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()
	c := New(srv.URL, "test-key", "test", false, io.Discard)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err := c.Do(ctx, http.MethodPost, "/", nil, map[string]string{"X-Sbr-Timeout-Sec": "120"}, nil, nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want caller deadline exceeded", err)
	}
}

func TestMaskKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"sk_live_abc", "***********"}, // <= 12 chars, all masked
		{"sk_live_FvSmh2g5buIBIs382_blK169J2rxNedbyVq2OJgJm5o", "sk_live_***************************************Jm5o"},
		{"short", "*****"},
	}
	for _, tt := range tests {
		got := MaskKey(tt.input)
		if got != tt.want {
			t.Errorf("MaskKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestErrorBodyMessage(t *testing.T) {
	tests := []struct {
		name string
		body errorBody
		want string
	}{
		{
			name: "go platform format",
			body: errorBody{Error: &struct {
				Message string `json:"message"`
				Code    string `json:"code"`
			}{Message: "resource not found", Code: "NOT_FOUND"}},
			want: "resource not found",
		},
		{
			name: "rate limiter format",
			body: errorBody{StatusCode: 429, Message: "Rate limit exceeded"},
			want: "Rate limit exceeded",
		},
		{
			name: "prefers error.message over top-level message",
			body: errorBody{
				Message: "Bad Request",
				Error: &struct {
					Message string `json:"message"`
					Code    string `json:"code"`
				}{Message: "validation failed: name is required"},
			},
			want: "validation failed: name is required",
		},
		{
			name: "empty body",
			body: errorBody{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.body.message()
			if got != tt.want {
				t.Errorf("message() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestErrorBodyDetails(t *testing.T) {
	tests := []struct {
		name     string
		body     errorBody
		wantLen  int
		contains string
	}{
		{
			name:    "no alreadyAttached",
			body:    errorBody{Message: "ok"},
			wantLen: 0,
		},
		{
			name: "alreadyAttached surfaces ids",
			body: errorBody{
				AlreadyAttached: []string{"abc-1", "abc-2", "abc-3"},
			},
			wantLen:  1,
			contains: "alreadyAttached (3): abc-1, abc-2, abc-3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.body.details()
			if len(got) != tt.wantLen {
				t.Fatalf("details() len = %d, want %d (got %v)", len(got), tt.wantLen, got)
			}
			if tt.contains != "" && !strings.Contains(got[0], tt.contains) {
				t.Errorf("details()[0] = %q, expected to contain %q", got[0], tt.contains)
			}
		})
	}
}

func TestAPIErrorErrorIncludesDetails(t *testing.T) {
	e := &APIError{
		StatusCode: 409,
		Message:    "one or more executionIds are already attached",
		Details:    []string{"alreadyAttached (2): id-1, id-2"},
	}
	msg := e.Error()
	if !strings.Contains(msg, "API error 409") {
		t.Errorf("missing status: %q", msg)
	}
	if !strings.Contains(msg, "alreadyAttached (2): id-1, id-2") {
		t.Errorf("missing details: %q", msg)
	}
}
