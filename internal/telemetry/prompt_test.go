package telemetry

import (
	"bytes"
	"strings"
	"testing"
)

func TestAsk(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"enter accepts the recommended answer", "\n", true},
		{"y accepts", "y\n", true},
		{"yes accepts", "yes\n", true},
		{"uppercase Y accepts", "Y\n", true},
		{"n declines", "n\n", false},
		{"no declines", "no\n", false},
		{"any other word declines", "maybe\n", false},
		{"closed input declines", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if got := Ask(strings.NewReader(tt.input), &out); got != tt.want {
				t.Errorf("Ask(%q) = %v, want %v", tt.input, got, tt.want)
			}
			if !strings.Contains(out.String(), PromptQuestion) {
				t.Errorf("the question was not printed: %q", out.String())
			}
		})
	}
}

// TestAsk_ExplainsTheContract keeps the prompt honest: it must say what is
// never sent and how to turn the data off.
func TestAsk_ExplainsTheContract(t *testing.T) {
	var out bytes.Buffer
	Ask(strings.NewReader("n\n"), &out)

	text := out.String()
	for _, want := range []string{"never sends", "API key", "saber telemetry disable"} {
		if !strings.Contains(text, want) {
			t.Errorf("the prompt does not mention %q:\n%s", want, text)
		}
	}
}
