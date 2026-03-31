// Package update implements a non-blocking background version check for the
// Saber CLI. It caches the result in ~/.saber/update-check.json and only hits
// the GitHub releases API at most once per check interval (default 24h).
package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ReleasesURL is the GitHub API endpoint queried for new releases. It is a
// variable (not a constant) so that tests can point it at an httptest server.
var ReleasesURL = "https://api.github.com/repos/saberapp/cli/releases?per_page=20"

// State is persisted to ~/.saber/update-check.json between runs.
type State struct {
	LatestVersion string    `json:"latestVersion"`
	CheckedAt     time.Time `json:"checkedAt"`
}

// CacheFilePath returns the path to the update-check cache file.
func CacheFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".saber", "update-check.json"), nil
}

// LoadState reads the cached update state. Returns a zero-value State (not an
// error) when the file is missing or corrupt, so callers never need to handle
// "first run" specially.
func LoadState() *State {
	path, err := CacheFilePath()
	if err != nil {
		return &State{}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return &State{}
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return &State{}
	}
	return &s
}

// SaveState atomically writes the cache file (write tmp + rename).
func SaveState(s *State) error {
	path, err := CacheFilePath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// ShouldCheck returns true when the cached state is stale (or empty).
func ShouldCheck(s *State, interval time.Duration) bool {
	if s.CheckedAt.IsZero() {
		return true
	}
	return time.Since(s.CheckedAt) >= interval
}

// FetchLatestVersion queries the GitHub releases API and returns the newest
// stable version string (without a leading "v"). Drafts and prereleases are
// skipped.
func FetchLatestVersion() (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(ReleasesURL)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var releases []struct {
		TagName string `json:"tag_name"`
		Draft   bool   `json:"draft"`
		Prerel  bool   `json:"prerelease"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", fmt.Errorf("parse releases: %w", err)
	}

	for _, r := range releases {
		if r.Draft || r.Prerel {
			continue
		}
		if strings.HasPrefix(r.TagName, "v") {
			return strings.TrimPrefix(r.TagName, "v"), nil
		}
	}
	return "", fmt.Errorf("no stable release found")
}

// FormatNotice returns a single-line update notice suitable for printing to
// stderr. Kept short to minimise noise for AI agents and scripts.
func FormatNotice(current, latest string) string {
	return fmt.Sprintf("Notice: Update available v%s -> v%s. Run \"saber update\" to upgrade.", current, latest)
}

// RunBackgroundCheck starts a non-blocking version check. It returns a channel
// that will receive at most one message (the update notice) if a newer version
// is available. The channel is closed when the check completes or is skipped.
//
// The caller should do a non-blocking read on the returned channel after the
// main command finishes:
//
//	select {
//	case msg := <-ch:
//	    if msg != "" { fmt.Fprintln(os.Stderr, msg) }
//	default:
//	}
func RunBackgroundCheck(currentVersion string, interval time.Duration) <-chan string {
	ch := make(chan string, 1)

	state := LoadState()

	// If we checked recently and have a cached result, use it immediately
	// without spawning a goroutine.
	if !ShouldCheck(state, interval) {
		if state.LatestVersion != "" && state.LatestVersion != currentVersion && isNewer(state.LatestVersion, currentVersion) {
			ch <- FormatNotice(currentVersion, state.LatestVersion)
		}
		close(ch)
		return ch
	}

	// Stale or missing cache: check in the background.
	go func() {
		defer close(ch)

		latest, err := FetchLatestVersion()
		if err != nil {
			// Silently swallow network errors; do not update cache so we
			// retry next time.
			return
		}

		// Always persist, even if versions match, to reset the timer.
		_ = SaveState(&State{
			LatestVersion: latest,
			CheckedAt:     time.Now().UTC(),
		})

		if latest != currentVersion && isNewer(latest, currentVersion) {
			ch <- FormatNotice(currentVersion, latest)
		}
	}()

	return ch
}

// isNewer does a simple semver comparison to determine if "latest" is newer
// than "current". Both strings are expected without a leading "v".
// Falls back to string inequality if parsing fails.
func isNewer(latest, current string) bool {
	lMajor, lMinor, lPatch, lok := parseSemver(latest)
	cMajor, cMinor, cPatch, cok := parseSemver(current)
	if !lok || !cok {
		return latest != current
	}
	if lMajor != cMajor {
		return lMajor > cMajor
	}
	if lMinor != cMinor {
		return lMinor > cMinor
	}
	return lPatch > cPatch
}

// parseSemver extracts major.minor.patch from a version string like "1.2.3".
func parseSemver(v string) (major, minor, patch int, ok bool) {
	n, err := fmt.Sscanf(v, "%d.%d.%d", &major, &minor, &patch)
	return major, minor, patch, err == nil && n == 3
}
