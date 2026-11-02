package telemetry

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// stubPostHog returns a server that counts requests and keeps the last body.
func stubPostHog(t *testing.T) (*httptest.Server, *int32, *[]byte) {
	t.Helper()
	var calls int32
	var last []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		body, _ := io.ReadAll(r.Body)
		last = body
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	t.Cleanup(srv.Close)
	return srv, &calls, &last
}

// withKeyAndHost points the package at the stub server for one test.
func withKeyAndHost(t *testing.T, key, host string) {
	t.Helper()
	origKey, origHost := posthogKey, Host
	posthogKey, Host = key, host
	t.Cleanup(func() { posthogKey, Host = origKey, origHost })
}

func TestResolve(t *testing.T) {
	answered := &State{Enabled: true, ConsentedAt: time.Now()}
	declined := &State{Enabled: false, ConsentedAt: time.Now()}
	fresh := &State{}

	tests := []struct {
		name          string
		doNotTrack    string
		saberTelemetr string
		state         *State
		want          Decision
	}{
		{"fresh state asks", "", "", fresh, DecisionNeedsConsent},
		{"accepted state sends", "", "", answered, DecisionEnabled},
		{"declined state stays off", "", "", declined, DecisionDisabled},
		{"DO_NOT_TRACK beats consent", "1", "", answered, DecisionDisabled},
		{"DO_NOT_TRACK=0 is not a switch", "0", "", answered, DecisionEnabled},
		{"SABER_TELEMETRY=0 disables", "", "0", answered, DecisionDisabled},
		{"SABER_TELEMETRY=false disables", "", "false", answered, DecisionDisabled},
		{"SABER_TELEMETRY=off disables", "", "off", answered, DecisionDisabled},
		{"SABER_TELEMETRY=1 skips the prompt", "", "1", fresh, DecisionEnabled},
		{"DO_NOT_TRACK wins over SABER_TELEMETRY=1", "1", "1", fresh, DecisionDisabled},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DO_NOT_TRACK", tt.doNotTrack)
			t.Setenv("SABER_TELEMETRY", tt.saberTelemetr)
			if got := Resolve(tt.state); got != tt.want {
				t.Errorf("Resolve() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEventProperties(t *testing.T) {
	e := Event{
		Command:    "signal create",
		Flags:      []string{"json", "company"},
		CLIVersion: "1.2.3",
		Status:     StatusOK,
		ExitCode:   0,
		Duration:   1500 * time.Millisecond,
	}
	props := e.properties()

	if props["command"] != "signal create" {
		t.Errorf("command = %v", props["command"])
	}
	if props["source"] != Source {
		t.Errorf("source = %v, want %v", props["source"], Source)
	}
	if props["duration_ms"] != int64(1500) {
		t.Errorf("duration_ms = %v, want 1500", props["duration_ms"])
	}
	flags, ok := props["flags"].([]string)
	if !ok {
		t.Fatalf("flags is %T, want []string", props["flags"])
	}
	if strings.Join(flags, ",") != "company,json" {
		t.Errorf("flags = %v, want sorted [company json]", flags)
	}
	for _, key := range []string{"cli_version", "os", "arch", "status", "exit_code"} {
		if _, present := props[key]; !present {
			t.Errorf("property %q is missing", key)
		}
	}
}

// TestEventProperties_NoFlagValues pins the privacy contract: only flag names
// are sent, so a secret passed as a flag value can never reach PostHog.
func TestEventProperties_NoFlagValues(t *testing.T) {
	props := Event{Command: "auth login", Flags: []string{"key"}}.properties()
	encoded, err := json.Marshal(props)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "sk_live_") {
		t.Fatalf("payload carries a key-like value: %s", encoded)
	}
}

func TestSend_NoKeyIsSilent(t *testing.T) {
	srv, calls, _ := stubPostHog(t)
	withKeyAndHost(t, "", srv.URL)

	Send(Event{Command: "credits"}, "org_1", 500*time.Millisecond, nil)

	if got := atomic.LoadInt32(calls); got != 0 {
		t.Errorf("a build without a key sent %d requests, want 0", got)
	}
}

func TestSend_UnknownOrganizationIsSilent(t *testing.T) {
	srv, calls, _ := stubPostHog(t)
	withKeyAndHost(t, "phc_test", srv.URL)

	Send(Event{Command: "credits"}, "", 500*time.Millisecond, nil)

	if got := atomic.LoadInt32(calls); got != 0 {
		t.Errorf("an unknown organization sent %d requests, want 0", got)
	}
}

func TestSend_EnabledSendsOneEvent(t *testing.T) {
	srv, calls, last := stubPostHog(t)
	withKeyAndHost(t, "phc_test", srv.URL)

	Send(Event{
		Command:    "signal create",
		Flags:      []string{"json"},
		CLIVersion: "1.2.3",
		Status:     StatusError,
		ExitCode:   1,
		Duration:   2 * time.Second,
	}, "org_42", 2*time.Second, nil)

	if got := atomic.LoadInt32(calls); got != 1 {
		t.Fatalf("sent %d requests, want exactly 1", got)
	}

	body := string(*last)
	for _, want := range []string{EventCommandExecuted, "org_42", "signal create", "cli_version", "organization"} {
		if !strings.Contains(body, want) {
			t.Errorf("request body does not carry %q: %s", want, body)
		}
	}
}

// TestSend_SlowEndpointDoesNotHang pins the promise that telemetry never
// delays the command: an endpoint that never answers must not hold the process
// for longer than the flush timeout.
func TestSend_SlowEndpointDoesNotHang(t *testing.T) {
	blocked := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-blocked
	}))
	t.Cleanup(func() { close(blocked); srv.Close() })
	withKeyAndHost(t, "phc_test", srv.URL)

	timeout := 300 * time.Millisecond
	start := time.Now()
	Send(Event{Command: "credits"}, "org_1", timeout, nil)
	elapsed := time.Since(start)

	if elapsed > 4*timeout {
		t.Errorf("Send blocked for %s with a %s timeout", elapsed, timeout)
	}
}

func TestSend_UnreachableEndpointIsSilent(t *testing.T) {
	// Port 0 is never listening, so the transport fails immediately.
	withKeyAndHost(t, "phc_test", "http://127.0.0.1:0")

	done := make(chan struct{})
	go func() {
		defer close(done)
		Send(Event{Command: "credits"}, "org_1", 300*time.Millisecond, nil)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Send did not return after a connection failure")
	}
}

func TestConfigured(t *testing.T) {
	withKeyAndHost(t, "", "https://example.invalid")
	if Configured() {
		t.Error("an empty key must report not configured")
	}
	withKeyAndHost(t, "phc_test", "https://example.invalid")
	if !Configured() {
		t.Error("a key must report configured")
	}
}
