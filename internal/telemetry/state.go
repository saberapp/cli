// Package telemetry sends anonymous CLI usage events to PostHog.
//
// The package is a no-op unless all of these are true:
//   - the binary was built with a PostHog project key (release builds only),
//   - the user has consented (or consent was implied on a non-interactive run),
//   - the organization id is known.
//
// It never sends flag values, positional arguments, API keys, or API responses.
// See README.md ("Usage data") for the user-facing contract.
package telemetry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Consent modes recorded in the state file, for support and debugging.
const (
	// ConsentModePrompt means the user answered the first-run question.
	ConsentModePrompt = "prompt"
	// ConsentModeImplied means the first run could not ask (no terminal), so
	// the default (on) applied.
	ConsentModeImplied = "implied"
	// ConsentModeCommand means the user ran "saber telemetry enable|disable".
	ConsentModeCommand = "command"
)

// State is persisted to ~/.saber/telemetry.json (mode 0600).
type State struct {
	Enabled        bool      `json:"enabled"`
	ConsentedAt    time.Time `json:"consentedAt"`
	ConsentVersion string    `json:"consentVersion"`
	ConsentMode    string    `json:"consentMode"`
	// OrganizationID is learned from API responses the user's own commands
	// already made. It is the PostHog distinct id and organization group key.
	OrganizationID string `json:"organizationId,omitempty"`
}

// Answered reports whether consent was recorded before this run.
func (s *State) Answered() bool {
	return s != nil && !s.ConsentedAt.IsZero()
}

// StatePath returns the path of the telemetry state file.
func StatePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".saber", "telemetry.json"), nil
}

// LoadState reads the state file. A missing or corrupt file returns a
// zero-value State rather than an error, so callers never handle "first run"
// as a failure.
func LoadState() *State {
	path, err := StatePath()
	if err != nil {
		return &State{}
	}
	data, err := os.ReadFile(path) //nolint:gosec // path is derived from the user's home directory
	if err != nil {
		return &State{}
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return &State{}
	}
	return &s
}

// SaveState writes the state file atomically (temp file + rename), mode 0600.
func SaveState(s *State) error {
	path, err := StatePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// SaveConsent records the user's answer and returns the updated state.
// mode is one of the ConsentMode* constants.
func SaveConsent(s *State, enabled bool, cliVersion, mode string) *State {
	if s == nil {
		s = &State{}
	}
	s.Enabled = enabled
	s.ConsentedAt = time.Now().UTC()
	s.ConsentVersion = cliVersion
	s.ConsentMode = mode
	// A write failure must not break the command; the question is asked again
	// on the next run instead.
	_ = SaveState(s)
	return s
}

// RecordOrganizationID caches the organization id when it changes. It is called
// from the API client on every successful response, so it returns early on the
// common path where nothing changed.
func RecordOrganizationID(id string) {
	if id == "" {
		return
	}
	s := LoadState()
	if s.OrganizationID == id {
		return
	}
	s.OrganizationID = id
	_ = SaveState(s)
}
