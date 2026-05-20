package format

import (
	"bytes"
	"strings"
	"testing"

	"github.com/saberapp/cli/internal/client"
)

// ptr is a tiny helper so tests can take addresses of string literals.
func ptr(s string) *string { return &s }

func TestPrintFindEmailResult(t *testing.T) {
	tests := []struct {
		name       string
		fullName   string
		domain     string
		resp       *client.FindEmailResponse
		wantSubstr []string
		notSubstr  []string
	}{
		{
			name:     "not-found prints one-liner",
			fullName: "Jane Doe",
			domain:   "ghost.example",
			resp:     &client.FindEmailResponse{Email: nil, Verification: nil},
			wantSubstr: []string{
				"no email found",
				"Jane Doe",
				"ghost.example",
			},
			notSubstr: []string{"Email:", "State:", "Score:", "Accept-All:"},
		},
		{
			name:       "nil response prints one-liner",
			fullName:   "Jane Doe",
			domain:     "ghost.example",
			resp:       nil,
			wantSubstr: []string{"no email found", "Jane Doe", "ghost.example"},
		},
		{
			name:     "normal hit prints KV block without catch-all note",
			fullName: "Joey van Ommen",
			domain:   "saber.app",
			resp: &client.FindEmailResponse{
				Email: ptr("joey.vanommen@saber.app"),
				Verification: &client.VerificationResult{
					State:     "deliverable",
					Score:     95,
					AcceptAll: false,
				},
			},
			wantSubstr: []string{
				"Email:", "joey.vanommen@saber.app",
				"State:", "deliverable",
				"Score:", "95",
				"Accept-All:", "false",
			},
			notSubstr: []string{"catch-all", "no email found"},
		},
		{
			name:     "catch-all hit appends the lower-confidence note",
			fullName: "Joey van Ommen",
			domain:   "catchall.example",
			resp: &client.FindEmailResponse{
				Email: ptr("joey.vanommen@catchall.example"),
				Verification: &client.VerificationResult{
					State:     "deliverable",
					Score:     90,
					AcceptAll: true,
				},
			},
			wantSubstr: []string{
				"Email:", "joey.vanommen@catchall.example",
				"Accept-All:", "true",
				"catch-all",
				"modal real-world pattern",
			},
			notSubstr: []string{"no email found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			PrintFindEmailResult(&buf, tt.fullName, tt.domain, tt.resp)
			got := buf.String()

			for _, want := range tt.wantSubstr {
				if !strings.Contains(got, want) {
					t.Errorf("output missing %q\n--- output ---\n%s", want, got)
				}
			}
			for _, notWant := range tt.notSubstr {
				if strings.Contains(got, notWant) {
					t.Errorf("output unexpectedly contains %q\n--- output ---\n%s", notWant, got)
				}
			}
		})
	}
}
