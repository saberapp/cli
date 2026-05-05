package client

import (
	"strings"
	"testing"
)

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
