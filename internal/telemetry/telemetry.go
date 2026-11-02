package telemetry

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	ph "github.com/posthog/posthog-go"
)

// EventCommandExecuted follows the domain.object.action convention in the
// platform repo (docs/analytics-conventions.md).
const EventCommandExecuted = "cli.command.executed"

// Source identifies the emitting system in every event.
const Source = "cli"

// Group type used for PostHog organization group analytics. It matches the
// group type the API and web-app already use.
const organizationGroupType = "organization"

var (
	// posthogKey is injected at release build time with
	// -X github.com/saberapp/cli/internal/telemetry.posthogKey=<key>.
	// Dev builds leave it empty, which disables telemetry completely.
	posthogKey string

	// Host is the PostHog ingestion endpoint. It matches the api and web-app
	// projects (EU cloud). Tests point it at a local server.
	Host = "https://eu.i.posthog.com"
)

// Configured reports whether this build can send telemetry at all.
func Configured() bool { return posthogKey != "" }

// Decision is the outcome of combining the environment and the saved state.
type Decision int

const (
	// DecisionDisabled means no event may be sent.
	DecisionDisabled Decision = iota
	// DecisionEnabled means events may be sent.
	DecisionEnabled
	// DecisionNeedsConsent means the user has never answered the question.
	DecisionNeedsConsent
)

// Resolve combines the environment switches and the saved state.
//
// Precedence:
//  1. DO_NOT_TRACK (any value other than "" / "0" / "false") disables.
//  2. SABER_TELEMETRY=0 disables; SABER_TELEMETRY=1 enables without a prompt.
//  3. The saved answer.
//  4. No saved answer: consent is needed.
func Resolve(s *State) Decision {
	if enabled, ok := envOverride(); ok {
		if enabled {
			return DecisionEnabled
		}
		return DecisionDisabled
	}
	if !s.Answered() {
		return DecisionNeedsConsent
	}
	if s.Enabled {
		return DecisionEnabled
	}
	return DecisionDisabled
}

// envOverride reports an environment switch and whether one was set.
func envOverride() (enabled, ok bool) {
	if truthy(os.Getenv("DO_NOT_TRACK")) {
		return false, true
	}
	switch strings.ToLower(strings.TrimSpace(os.Getenv("SABER_TELEMETRY"))) {
	case "":
		return false, false
	case "0", "false", "no", "off":
		return false, true
	default:
		return true, true
	}
}

// truthy reports whether an environment variable is set to a value that means
// "on". DO_NOT_TRACK is a cross-tool convention: presence means do not track,
// except for the explicit off values.
func truthy(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "", "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

// Event is one command run. Every field is a stable, low-cardinality scalar
// except Flags, which holds flag NAMES only.
type Event struct {
	Command    string
	Flags      []string
	CLIVersion string
	Status     string
	ExitCode   int
	Duration   time.Duration
}

// Status values.
const (
	StatusOK    = "ok"
	StatusError = "error"
)

// properties builds the PostHog payload. It contains no argument values, no
// flag values, no identifiers, and no API content.
func (e Event) properties() ph.Properties {
	flags := append([]string(nil), e.Flags...)
	sort.Strings(flags)
	return ph.NewProperties().
		Set("command", e.Command).
		Set("flags", flags).
		Set("cli_version", e.CLIVersion).
		Set("os", runtime.GOOS).
		Set("arch", runtime.GOARCH).
		Set("status", e.Status).
		Set("exit_code", e.ExitCode).
		Set("duration_ms", e.Duration.Milliseconds()).
		Set("source", Source)
}

// Send delivers one event and waits at most timeout for the flush. It is
// best-effort by design: a missing key, an unknown organization, a network
// failure, or a panic inside the PostHog client must never change the exit
// code, the output, or the timing of the command that the user ran.
//
// errOut receives a one-line diagnostic when it is non-nil (--verbose only).
func Send(e Event, organizationID string, timeout time.Duration, errOut io.Writer) {
	defer func() {
		if r := recover(); r != nil {
			logf(errOut, "telemetry: dropped event after panic: %v\n", r)
		}
	}()

	if !Configured() {
		logf(errOut, "telemetry: no project key in this build; nothing sent\n")
		return
	}
	if organizationID == "" {
		logf(errOut, "telemetry: organization not known yet; nothing sent\n")
		return
	}
	if timeout <= 0 {
		return
	}

	noRetries := 0
	client, err := ph.NewWithConfig(posthogKey, ph.Config{
		Endpoint: Host,
		// One event per run: send it immediately instead of batching.
		BatchSize: 1,
		Interval:  timeout,
		// Bound every stage, so a slow or unreachable endpoint cannot hold
		// the CLI open.
		ShutdownTimeout:    timeout,
		BatchUploadTimeout: timeout,
		MaxRetries:         &noRetries,
		// The PostHog client logs to stderr by default. The CLI owns stderr,
		// so keep the library silent unless the user asked for --verbose.
		Logger: newLogger(errOut),
	})
	if err != nil {
		logf(errOut, "telemetry: client not created: %v\n", err)
		return
	}

	if err := client.Enqueue(ph.Capture{
		DistinctId: organizationID,
		Event:      EventCommandExecuted,
		Properties: e.properties(),
		Groups:     ph.NewGroups().Set(organizationGroupType, organizationID),
	}); err != nil {
		logf(errOut, "telemetry: event not queued: %v\n", err)
		_ = client.Close()
		return
	}

	// Close flushes. It is bounded by ShutdownTimeout, and the select below is
	// a second guard so a hung transport cannot delay the process exit.
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() { _ = recover() }()
		_ = client.Close()
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		logf(errOut, "telemetry: flush timed out after %s\n", timeout)
	}
}

func logf(w io.Writer, format string, args ...any) {
	if w == nil {
		return
	}
	fmt.Fprintf(w, format, args...)
}

// logger adapts the PostHog logger interface. With a nil writer it discards
// everything, which is the default for the CLI.
type logger struct{ w io.Writer }

func newLogger(w io.Writer) ph.Logger { return logger{w: w} }

func (l logger) Debugf(format string, args ...any) { l.write(format, args...) }
func (l logger) Logf(format string, args ...any)   { l.write(format, args...) }
func (l logger) Warnf(format string, args ...any)  { l.write(format, args...) }
func (l logger) Errorf(format string, args ...any) { l.write(format, args...) }

func (l logger) write(format string, args ...any) {
	logf(l.w, "telemetry: "+strings.TrimRight(format, "\n")+"\n", args...)
}
