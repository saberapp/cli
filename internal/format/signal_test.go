package format

import (
	"testing"
)

func TestFormatAnswer(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  string
	}{
		{
			name:  "nil",
			input: nil,
			want:  "—",
		},
		{
			name:  "open_text",
			input: map[string]any{"type": "open_text", "openText": map[string]any{"value": "Yes"}},
			want:  "Yes",
		},
		{
			name:  "number integer",
			input: map[string]any{"type": "number", "number": map[string]any{"value": float64(14491)}},
			want:  "14491",
		},
		{
			name:  "number float",
			input: map[string]any{"type": "number", "number": map[string]any{"value": float64(3.14)}},
			want:  "3.14",
		},
		{
			name:  "boolean true",
			input: map[string]any{"type": "boolean", "boolean": map[string]any{"value": true}},
			want:  "true",
		},
		{
			name:  "boolean false",
			input: map[string]any{"type": "boolean", "boolean": map[string]any{"value": false}},
			want:  "false",
		},
		{
			name:  "list",
			input: map[string]any{"type": "list", "list": map[string]any{"value": []any{"Go", "TypeScript", "Python"}}},
			want:  "Go, TypeScript, Python",
		},
		{
			name:  "list empty",
			input: map[string]any{"type": "list", "list": map[string]any{"value": []any{}}},
			want:  "",
		},
		{
			name:  "percentage",
			input: map[string]any{"type": "percentage", "percentage": map[string]any{"value": float64(42)}},
			want:  "42",
		},
		{
			name:  "currency",
			input: map[string]any{"type": "currency", "currency": map[string]any{"value": "5000000"}},
			want:  "5000000",
		},
		{
			name:  "url",
			input: map[string]any{"type": "url", "url": map[string]any{"value": "https://stripe.com/jobs"}},
			want:  "https://stripe.com/jobs",
		},
		{
			name:  "unknown type falls back to json",
			input: map[string]any{"type": "future_type", "future_type": map[string]any{"value": "something"}},
			want:  "something",
		},
		{
			name:  "plain string scalar",
			input: "just a string",
			want:  `"just a string"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatAnswer(tt.input)
			if got != tt.want {
				t.Errorf("formatAnswer(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
