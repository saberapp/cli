package format

import (
	"testing"
)

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input string
		n     int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 8, "hello..."},
		{"", 5, ""},
		{"abc", 3, "abc"},
		{"abcd", 3, "..."},
	}
	for _, tt := range tests {
		got := TruncateString(tt.input, tt.n)
		if got != tt.want {
			t.Errorf("TruncateString(%q, %d) = %q, want %q", tt.input, tt.n, got, tt.want)
		}
	}
}

func TestJoinStrings(t *testing.T) {
	tests := []struct {
		input    []string
		fallback string
		want     string
	}{
		{[]string{"a", "b", "c"}, "—", "a, b, c"},
		{[]string{"only"}, "—", "only"},
		{[]string{}, "—", "—"},
		{nil, "none", "none"},
	}
	for _, tt := range tests {
		got := JoinStrings(tt.input, tt.fallback)
		if got != tt.want {
			t.Errorf("JoinStrings(%v, %q) = %q, want %q", tt.input, tt.fallback, got, tt.want)
		}
	}
}
