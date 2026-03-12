package client

import (
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
