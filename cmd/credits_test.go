package cmd

import "testing"

func TestFormatCredits(t *testing.T) {
	tests := []struct {
		name string
		in   float64
		want string
	}{
		{name: "whole number", in: 4200, want: "4,200"},
		{name: "single decimal", in: 951332.7, want: "951,332.7"},
		{name: "two decimals below one", in: 0.05, want: "0.05"},
		{name: "four decimals", in: 1234.5678, want: "1,234.5678"},
		{name: "trims trailing decimals", in: 1000.1, want: "1,000.1"},
		{name: "negative fractional", in: -1234.5, want: "-1,234.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatCredits(tt.in); got != tt.want {
				t.Fatalf("formatCredits(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
