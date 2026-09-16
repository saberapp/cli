package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/saberapp/cli/internal/client"
	"github.com/saberapp/cli/internal/config"
	"github.com/saberapp/cli/internal/telemetry"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/term"
)

// telemetryFlushTimeout bounds how long the process may wait for the usage
// event after the command has finished its work.
const telemetryFlushTimeout = 2 * time.Second

var (
	// executedCmd is the command cobra resolved, set in PersistentPreRun.
	// It stays nil when argument parsing fails, and then no event is sent.
	executedCmd *cobra.Command

	// telemetryState is loaded once per run, before the command executes.
	telemetryState *telemetry.State
	// telemetryEnabled is the decision for this run.
	telemetryEnabled bool
)

// Execute builds the root command, runs it, and returns the process exit code.
// It owns the exit-code mapping and the usage event, so a failed command is
// still recorded with its real status.
func Execute(version, commit, date string) int {
	root := NewRootCmd(version, commit, date)

	start := time.Now()
	err := root.Execute()
	code := exitCodeFor(err)

	recordTelemetry(start, code, err)

	if err != nil && !errors.As(err, new(*config.ErrNotAuthenticated)) {
		fmt.Fprintln(os.Stderr, "error:", err)
	}
	return code
}

// exitCodeFor maps a command error to a process exit code. It keeps the
// historical contract: 2 for "not authenticated", 1 for anything else.
func exitCodeFor(err error) int {
	if err == nil {
		return 0
	}
	if errors.As(err, new(*config.ErrNotAuthenticated)) {
		return 2
	}
	return 1
}

// startTelemetry resolves consent before the command runs. On a first run in a
// terminal it asks the question; a run that cannot ask keeps the default (on)
// and records that consent was implied.
func startTelemetry(cmd *cobra.Command) {
	defer func() { _ = recover() }() // analytics must never break a command

	telemetryState = telemetry.LoadState()

	switch telemetry.Resolve(telemetryState) {
	case telemetry.DecisionEnabled:
		telemetryEnabled = true
	case telemetry.DecisionDisabled:
		telemetryEnabled = false
	case telemetry.DecisionNeedsConsent:
		if shouldAskTelemetryConsent(cmd) {
			answer := telemetry.Ask(os.Stdin, os.Stderr)
			telemetryState = telemetry.SaveConsent(telemetryState, answer, cliVersion, telemetry.ConsentModePrompt)
			telemetryEnabled = answer
		} else {
			telemetryState = telemetry.SaveConsent(telemetryState, true, cliVersion, telemetry.ConsentModeImplied)
			telemetryEnabled = true
		}
	}

	if telemetryEnabled {
		// Learn the organization id from the responses this command receives.
		client.OnOrganizationID = telemetry.RecordOrganizationID
	}
}

// shouldAskTelemetryConsent reports whether this run may ask the question.
// Asking needs a real terminal on both sides: stdin to read the answer, and
// stderr to show it. Machine-readable runs and the telemetry command itself
// never ask.
func shouldAskTelemetryConsent(cmd *cobra.Command) bool {
	if jsonOutput || quiet {
		return false
	}
	switch rootCommandName(cmd) {
	case "telemetry", "help", "version", "completion", "update":
		return false
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return false
	}
	return term.IsTerminal(int(os.Stderr.Fd()))
}

// rootCommandName returns the top-level command name, for example "signal" for
// "saber signal create".
func rootCommandName(cmd *cobra.Command) string {
	if cmd == nil {
		return ""
	}
	parts := strings.Fields(cmd.CommandPath())
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

// recordTelemetry sends the usage event. It is best-effort: every failure is
// swallowed, and the caller's exit code is never changed.
func recordTelemetry(start time.Time, exitCode int, runErr error) {
	defer func() { _ = recover() }()

	if !telemetryEnabled || executedCmd == nil || telemetryState == nil {
		return
	}

	status := telemetry.StatusOK
	if runErr != nil {
		status = telemetry.StatusError
	}

	// Keep this an io.Writer, never a typed-nil *os.File: a typed nil would be
	// a non-nil interface and would print library noise on every run.
	var errOut io.Writer
	if verbose {
		errOut = os.Stderr
	}

	organizationID, allowed := telemetryTarget()
	if !allowed {
		return
	}

	telemetry.Send(telemetry.Event{
		Command:    telemetryCommandPath(executedCmd),
		Flags:      telemetrySetFlags(executedCmd),
		CLIVersion: cliVersion,
		Status:     status,
		ExitCode:   exitCode,
		Duration:   time.Since(start),
	}, organizationID, telemetryFlushTimeout, errOut)
}

// telemetryTarget reports the organization to attribute the event to, and
// whether the event may be sent at all.
//
// The state is read again after the command has finished, for two reasons:
// the organization id may have been learned from a response during this run,
// and the user may have just turned telemetry off. "saber telemetry disable"
// must not report itself, so a decision made before the command ran is not
// good enough.
func telemetryTarget() (string, bool) {
	state := telemetry.LoadState()
	if telemetry.Resolve(state) != telemetry.DecisionEnabled {
		return "", false
	}
	return state.OrganizationID, true
}

// telemetryCommandPath returns the command path without the binary name, for
// example "signal create". It is a fixed set of values, so it stays
// low-cardinality.
func telemetryCommandPath(cmd *cobra.Command) string {
	parts := strings.Fields(cmd.CommandPath())
	if len(parts) < 2 {
		return "saber"
	}
	return strings.Join(parts[1:], " ")
}

// telemetrySetFlags returns the NAMES of the flags the user set. Values are
// never read, so no argument, identifier, or secret can reach PostHog.
func telemetrySetFlags(cmd *cobra.Command) []string {
	seen := map[string]bool{}
	var names []string
	cmd.Flags().Visit(func(f *pflag.Flag) {
		if seen[f.Name] {
			return
		}
		seen[f.Name] = true
		names = append(names, f.Name)
	})
	return names
}
