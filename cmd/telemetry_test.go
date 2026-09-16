package cmd

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/saberapp/cli/internal/config"
	"github.com/saberapp/cli/internal/telemetry"
	"github.com/spf13/cobra"
)

func TestExitCodeFor(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"success", nil, 0},
		{"not authenticated keeps code 2", &config.ErrNotAuthenticated{}, 2},
		{"wrapped not authenticated keeps code 2", errors.Join(errors.New("ctx"), &config.ErrNotAuthenticated{}), 2},
		{"any other failure is code 1", errors.New("boom"), 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exitCodeFor(tt.err); got != tt.want {
				t.Errorf("exitCodeFor(%v) = %d, want %d", tt.err, got, tt.want)
			}
		})
	}
}

func TestTelemetryCommandPath(t *testing.T) {
	root := NewRootCmd("1.0.0", "abc", "now")

	// A real two-level path from the live command tree.
	sub, _, err := root.Find([]string{"telemetry", "status"})
	if err != nil {
		t.Fatalf("find telemetry status: %v", err)
	}
	if got := telemetryCommandPath(sub); got != "telemetry status" {
		t.Errorf("command path = %q, want %q", got, "telemetry status")
	}

	// The bare binary reports itself, never an empty string.
	if got := telemetryCommandPath(root); got != "saber" {
		t.Errorf("root command path = %q, want %q", got, "saber")
	}

	// A deeper tree keeps every level, so the value stays readable in PostHog.
	parent := &cobra.Command{Use: "list"}
	child := &cobra.Command{Use: "company"}
	grandchild := &cobra.Command{Use: "create"}
	child.AddCommand(grandchild)
	parent.AddCommand(child)
	root.AddCommand(parent)
	if got := telemetryCommandPath(grandchild); got != "list company create" {
		t.Errorf("deep command path = %q, want %q", got, "list company create")
	}
}

func TestTelemetrySetFlags_NamesOnly(t *testing.T) {
	c := &cobra.Command{Use: "demo", Run: func(*cobra.Command, []string) {}}
	c.Flags().String("company", "", "")
	c.Flags().Bool("json", false, "")
	c.Flags().String("untouched", "", "")
	c.SetArgs([]string{"--company", "sk_live_secret_value", "--json"})
	if err := c.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	names := telemetrySetFlags(c)
	if strings.Join(names, ",") != "company,json" && strings.Join(names, ",") != "json,company" {
		t.Errorf("flag names = %v, want company and json only", names)
	}
	for _, n := range names {
		if strings.Contains(n, "sk_live_") {
			t.Fatalf("a flag value leaked into the payload: %v", names)
		}
		if n == "untouched" {
			t.Error("a flag that was never set must not be reported")
		}
	}
}

func TestShouldAskTelemetryConsent_MachineReadable(t *testing.T) {
	root := NewRootCmd("1.0.0", "abc", "now")
	sub, _, err := root.Find([]string{"credits"})
	if err != nil {
		t.Fatalf("find credits: %v", err)
	}

	jsonOutput = true
	defer func() { jsonOutput = false }()
	if shouldAskTelemetryConsent(sub) {
		t.Error("a --json run must never ask the question")
	}

	jsonOutput = false
	quiet = true
	defer func() { quiet = false }()
	if shouldAskTelemetryConsent(sub) {
		t.Error("a --quiet run must never ask the question")
	}
}

func TestShouldAskTelemetryConsent_SkippedCommands(t *testing.T) {
	root := NewRootCmd("1.0.0", "abc", "now")
	for _, name := range []string{"telemetry", "version", "update"} {
		sub, _, err := root.Find([]string{name})
		if err != nil {
			t.Fatalf("find %s: %v", name, err)
		}
		if shouldAskTelemetryConsent(sub) {
			t.Errorf("the %q command must not ask the question", name)
		}
	}
}

// TestRecordTelemetry_DisabledIsSafe pins the promise that analytics never
// breaks a command: with telemetry off, or with no resolved command, the call
// must do nothing and must not panic.
func TestRecordTelemetry_DisabledIsSafe(t *testing.T) {
	telemetryEnabled = false
	executedCmd = nil
	telemetryState = nil
	recordTelemetry(time.Now(), 0, nil)

	telemetryEnabled = true
	recordTelemetry(time.Now(), 1, errors.New("boom"))
}

// TestTelemetryTarget_RespectsAnOptOutDuringTheRun pins the fix for a real
// defect found in manual testing: the decision is made before the command
// runs, so "saber telemetry disable" reported itself as one last event.
func TestTelemetryTarget_RespectsAnOptOutDuringTheRun(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SABER_TELEMETRY", "")
	t.Setenv("DO_NOT_TRACK", "")

	telemetry.SaveConsent(telemetry.LoadState(), true, "1.0.0", telemetry.ConsentModePrompt)
	telemetry.RecordOrganizationID("org_live")

	orgID, allowed := telemetryTarget()
	if !allowed || orgID != "org_live" {
		t.Fatalf("with consent: allowed=%v org=%q, want true/org_live", allowed, orgID)
	}

	// The command itself turned telemetry off.
	telemetry.SaveConsent(telemetry.LoadState(), false, "1.0.0", telemetry.ConsentModeCommand)

	if _, allowed := telemetryTarget(); allowed {
		t.Error("an opt-out during the run must stop the event for that same run")
	}
}

func TestTelemetryTarget_UnknownOrganization(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SABER_TELEMETRY", "")
	t.Setenv("DO_NOT_TRACK", "")

	telemetry.SaveConsent(telemetry.LoadState(), true, "1.0.0", telemetry.ConsentModePrompt)

	orgID, allowed := telemetryTarget()
	if !allowed {
		t.Fatal("consent alone must allow the attempt")
	}
	if orgID != "" {
		t.Errorf("organization = %q, want empty until it is learned", orgID)
	}
}

func TestRootCommandName(t *testing.T) {
	root := NewRootCmd("1.0.0", "abc", "now")
	sub, _, err := root.Find([]string{"signal", "create"})
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got := rootCommandName(sub); got != "signal" {
		t.Errorf("rootCommandName = %q, want signal", got)
	}
	if got := rootCommandName(nil); got != "" {
		t.Errorf("rootCommandName(nil) = %q, want empty", got)
	}
}
