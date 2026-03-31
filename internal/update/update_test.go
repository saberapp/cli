package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ---------- helpers ----------

// mockReleasesServer returns an httptest.Server that serves the given releases
// payload. The caller must defer srv.Close().
func mockReleasesServer(t *testing.T, releases []mockRelease) *httptest.Server {
	t.Helper()
	data, err := json.Marshal(releases)
	if err != nil {
		t.Fatalf("marshal mock releases: %v", err)
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
}

type mockRelease struct {
	TagName    string `json:"tag_name"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

// overrideReleasesURL points ReleasesURL at the mock server for the duration
// of the test, restoring the original value on cleanup.
func overrideReleasesURL(t *testing.T, url string) {
	t.Helper()
	orig := ReleasesURL
	ReleasesURL = url
	t.Cleanup(func() { ReleasesURL = orig })
}

// useTmpHome sets HOME to a temp directory so tests don't touch the real
// ~/.saber/ directory. Returns the temp dir path.
func useTmpHome(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	return tmp
}

// ---------- ShouldCheck ----------

func TestShouldCheck_ZeroTime(t *testing.T) {
	s := &State{}
	if !ShouldCheck(s, 24*time.Hour) {
		t.Error("expected true for zero-value CheckedAt")
	}
}

func TestShouldCheck_Stale(t *testing.T) {
	s := &State{CheckedAt: time.Now().Add(-25 * time.Hour)}
	if !ShouldCheck(s, 24*time.Hour) {
		t.Error("expected true for stale state")
	}
}

func TestShouldCheck_Recent(t *testing.T) {
	s := &State{CheckedAt: time.Now().Add(-1 * time.Hour)}
	if ShouldCheck(s, 24*time.Hour) {
		t.Error("expected false for recent state")
	}
}

func TestShouldCheck_ExactBoundary(t *testing.T) {
	// Exactly at the boundary should trigger a check.
	s := &State{CheckedAt: time.Now().Add(-24 * time.Hour)}
	if !ShouldCheck(s, 24*time.Hour) {
		t.Error("expected true at exact boundary")
	}
}

// ---------- LoadState / SaveState ----------

func TestLoadState_MissingFile(t *testing.T) {
	useTmpHome(t)
	s := LoadState()
	if !s.CheckedAt.IsZero() {
		t.Error("expected zero CheckedAt for missing file")
	}
	if s.LatestVersion != "" {
		t.Error("expected empty LatestVersion for missing file")
	}
}

func TestSaveAndLoadState(t *testing.T) {
	useTmpHome(t)

	original := &State{
		LatestVersion: "1.2.3",
		CheckedAt:     time.Now().UTC().Truncate(time.Second),
	}

	if err := SaveState(original); err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	loaded := LoadState()
	if loaded.LatestVersion != original.LatestVersion {
		t.Errorf("LatestVersion = %q, want %q", loaded.LatestVersion, original.LatestVersion)
	}
	if !loaded.CheckedAt.Equal(original.CheckedAt) {
		t.Errorf("CheckedAt = %v, want %v", loaded.CheckedAt, original.CheckedAt)
	}
}

func TestSaveState_CreatesDirectory(t *testing.T) {
	tmp := useTmpHome(t)

	// ~/.saber/ does not exist yet.
	saberDir := filepath.Join(tmp, ".saber")
	if _, err := os.Stat(saberDir); !os.IsNotExist(err) {
		t.Fatal("expected .saber dir to not exist before SaveState")
	}

	if err := SaveState(&State{LatestVersion: "1.0.0", CheckedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	if _, err := os.Stat(saberDir); err != nil {
		t.Errorf("expected .saber dir to exist after SaveState: %v", err)
	}
}

func TestLoadState_CorruptFile(t *testing.T) {
	tmp := useTmpHome(t)

	dir := filepath.Join(tmp, ".saber")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "update-check.json"), []byte("{bad json"), 0600); err != nil {
		t.Fatal(err)
	}

	s := LoadState()
	if !s.CheckedAt.IsZero() {
		t.Error("expected zero State for corrupt file")
	}
}

func TestLoadState_EmptyFile(t *testing.T) {
	tmp := useTmpHome(t)

	dir := filepath.Join(tmp, ".saber")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "update-check.json"), []byte(""), 0600); err != nil {
		t.Fatal(err)
	}

	s := LoadState()
	if !s.CheckedAt.IsZero() {
		t.Error("expected zero State for empty file")
	}
}

// ---------- FormatNotice ----------

func TestFormatNotice(t *testing.T) {
	msg := FormatNotice("0.1.5", "0.1.7")
	expected := `Notice: Update available v0.1.5 -> v0.1.7. Run "saber update" to upgrade.`
	if msg != expected {
		t.Errorf("FormatNotice = %q, want %q", msg, expected)
	}
}

// ---------- isNewer ----------

func TestIsNewer(t *testing.T) {
	tests := []struct {
		latest, current string
		want            bool
	}{
		{"1.0.0", "0.9.0", true},
		{"0.2.0", "0.1.9", true},
		{"0.1.8", "0.1.7", true},
		{"0.1.7", "0.1.7", false},
		{"0.1.6", "0.1.7", false},
		{"1.0.0", "1.0.0", false},
		{"2.0.0", "1.9.9", true},
		{"0.0.1", "0.0.2", false},
		{"10.0.0", "9.9.9", true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_vs_%s", tt.latest, tt.current), func(t *testing.T) {
			got := isNewer(tt.latest, tt.current)
			if got != tt.want {
				t.Errorf("isNewer(%q, %q) = %v, want %v", tt.latest, tt.current, got, tt.want)
			}
		})
	}
}

func TestIsNewer_NonSemver(t *testing.T) {
	// When either side is not valid semver, falls back to string inequality.
	if !isNewer("abc", "def") {
		t.Error("expected true for non-semver strings that differ")
	}
	if isNewer("same", "same") {
		t.Error("expected false for identical non-semver strings")
	}
}

// ---------- parseSemver ----------

func TestParseSemver(t *testing.T) {
	major, minor, patch, ok := parseSemver("1.2.3")
	if !ok || major != 1 || minor != 2 || patch != 3 {
		t.Errorf("parseSemver(1.2.3) = %d.%d.%d ok=%v", major, minor, patch, ok)
	}

	_, _, _, ok = parseSemver("bad")
	if ok {
		t.Error("expected parseSemver to fail on 'bad'")
	}

	_, _, _, ok = parseSemver("1.2")
	if ok {
		t.Error("expected parseSemver to fail on '1.2' (missing patch)")
	}

	_, _, _, ok = parseSemver("")
	if ok {
		t.Error("expected parseSemver to fail on empty string")
	}
}

// ---------- FetchLatestVersion ----------

func TestFetchLatestVersion_SkipsDraftAndPrerelease(t *testing.T) {
	srv := mockReleasesServer(t, []mockRelease{
		{"v2.0.0", true, false},  // draft
		{"v1.5.0", false, true},  // prerelease
		{"v1.0.0", false, false}, // first stable
		{"v0.9.0", false, false},
	})
	defer srv.Close()
	overrideReleasesURL(t, srv.URL)

	got, err := FetchLatestVersion()
	if err != nil {
		t.Fatalf("FetchLatestVersion: %v", err)
	}
	if got != "1.0.0" {
		t.Errorf("got %q, want %q", got, "1.0.0")
	}
}

func TestFetchLatestVersion_NoStableRelease(t *testing.T) {
	srv := mockReleasesServer(t, []mockRelease{
		{"v2.0.0-beta", false, true},
		{"v1.0.0-rc1", true, false},
	})
	defer srv.Close()
	overrideReleasesURL(t, srv.URL)

	_, err := FetchLatestVersion()
	if err == nil {
		t.Error("expected error when no stable release exists")
	}
}

func TestFetchLatestVersion_EmptyReleases(t *testing.T) {
	srv := mockReleasesServer(t, []mockRelease{})
	defer srv.Close()
	overrideReleasesURL(t, srv.URL)

	_, err := FetchLatestVersion()
	if err == nil {
		t.Error("expected error for empty releases list")
	}
}

func TestFetchLatestVersion_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	overrideReleasesURL(t, srv.URL)

	_, err := FetchLatestVersion()
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestFetchLatestVersion_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()
	overrideReleasesURL(t, srv.URL)

	_, err := FetchLatestVersion()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestFetchLatestVersion_NetworkError(t *testing.T) {
	// Point at a server that's already closed.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()
	overrideReleasesURL(t, srv.URL)

	_, err := FetchLatestVersion()
	if err == nil {
		t.Error("expected error for unreachable server")
	}
}

// ---------- RunBackgroundCheck ----------

func TestRunBackgroundCheck_CachedRecent_NoUpdate(t *testing.T) {
	useTmpHome(t)

	_ = SaveState(&State{
		LatestVersion: "1.0.0",
		CheckedAt:     time.Now().UTC(),
	})

	ch := RunBackgroundCheck("1.0.0", 24*time.Hour)
	// Channel should be closed synchronously (no goroutine, cached path).
	select {
	case msg, ok := <-ch:
		if ok && msg != "" {
			t.Errorf("expected no message, got %q", msg)
		}
	case <-time.After(time.Second):
		t.Error("channel was not closed promptly")
	}
}

func TestRunBackgroundCheck_CachedRecent_UpdateAvailable(t *testing.T) {
	useTmpHome(t)

	_ = SaveState(&State{
		LatestVersion: "1.1.0",
		CheckedAt:     time.Now().UTC(),
	})

	ch := RunBackgroundCheck("1.0.0", 24*time.Hour)
	select {
	case msg, ok := <-ch:
		if !ok || msg == "" {
			t.Error("expected an update notice from cached state")
		}
	case <-time.After(time.Second):
		t.Error("channel was not closed promptly")
	}
}

func TestRunBackgroundCheck_CachedRecent_OlderVersion(t *testing.T) {
	// Cache has an older version than current (user downgraded the cache
	// somehow, or was on a newer build). Should not show a notice.
	useTmpHome(t)

	_ = SaveState(&State{
		LatestVersion: "0.9.0",
		CheckedAt:     time.Now().UTC(),
	})

	ch := RunBackgroundCheck("1.0.0", 24*time.Hour)
	select {
	case msg, ok := <-ch:
		if ok && msg != "" {
			t.Errorf("expected no message for older cached version, got %q", msg)
		}
	case <-time.After(time.Second):
		t.Error("channel was not closed promptly")
	}
}

func TestRunBackgroundCheck_StaleCache_FetchesAndNotifies(t *testing.T) {
	useTmpHome(t)

	// Mock server returns a newer version.
	srv := mockReleasesServer(t, []mockRelease{
		{"v2.0.0", false, false},
	})
	defer srv.Close()
	overrideReleasesURL(t, srv.URL)

	// Seed a stale cache (checked 25h ago).
	_ = SaveState(&State{
		LatestVersion: "1.0.0",
		CheckedAt:     time.Now().Add(-25 * time.Hour),
	})

	ch := RunBackgroundCheck("1.0.0", 24*time.Hour)
	select {
	case msg := <-ch:
		if msg == "" {
			t.Error("expected an update notice after fetching from server")
		}
		// Verify the notice mentions both versions.
		if msg != FormatNotice("1.0.0", "2.0.0") {
			t.Errorf("unexpected notice: %q", msg)
		}
	case <-time.After(5 * time.Second):
		t.Error("timed out waiting for background check")
	}

	// Verify the cache was updated.
	state := LoadState()
	if state.LatestVersion != "2.0.0" {
		t.Errorf("cache LatestVersion = %q, want %q", state.LatestVersion, "2.0.0")
	}
	if time.Since(state.CheckedAt) > 5*time.Second {
		t.Error("cache CheckedAt was not updated recently")
	}
}

func TestRunBackgroundCheck_StaleCache_AlreadyUpToDate(t *testing.T) {
	useTmpHome(t)

	srv := mockReleasesServer(t, []mockRelease{
		{"v1.0.0", false, false},
	})
	defer srv.Close()
	overrideReleasesURL(t, srv.URL)

	// Stale cache, but server says we're current.
	_ = SaveState(&State{
		LatestVersion: "0.9.0",
		CheckedAt:     time.Now().Add(-25 * time.Hour),
	})

	ch := RunBackgroundCheck("1.0.0", 24*time.Hour)
	select {
	case msg := <-ch:
		if msg != "" {
			t.Errorf("expected no notice when up to date, got %q", msg)
		}
	case <-time.After(5 * time.Second):
		t.Error("timed out waiting for background check")
	}

	// Cache should still be updated (timer reset).
	state := LoadState()
	if state.LatestVersion != "1.0.0" {
		t.Errorf("cache LatestVersion = %q, want %q", state.LatestVersion, "1.0.0")
	}
}

func TestRunBackgroundCheck_StaleCache_NetworkError(t *testing.T) {
	useTmpHome(t)

	// Closed server = network error.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()
	overrideReleasesURL(t, srv.URL)

	oldState := &State{
		LatestVersion: "0.9.0",
		CheckedAt:     time.Now().Add(-25 * time.Hour),
	}
	_ = SaveState(oldState)

	ch := RunBackgroundCheck("1.0.0", 24*time.Hour)
	select {
	case msg := <-ch:
		if msg != "" {
			t.Errorf("expected no notice on network error, got %q", msg)
		}
	case <-time.After(5 * time.Second):
		t.Error("timed out waiting for background check")
	}

	// Cache should NOT be updated on error (so we retry next time).
	state := LoadState()
	if state.LatestVersion != "0.9.0" {
		t.Errorf("cache should be unchanged on error, got LatestVersion = %q", state.LatestVersion)
	}
}

func TestRunBackgroundCheck_FreshInstall_NoCache(t *testing.T) {
	useTmpHome(t)
	// No cache file at all.

	srv := mockReleasesServer(t, []mockRelease{
		{"v1.5.0", false, false},
	})
	defer srv.Close()
	overrideReleasesURL(t, srv.URL)

	ch := RunBackgroundCheck("1.0.0", 24*time.Hour)
	select {
	case msg := <-ch:
		if msg == "" {
			t.Error("expected update notice on fresh install with older version")
		}
	case <-time.After(5 * time.Second):
		t.Error("timed out waiting for background check")
	}
}
