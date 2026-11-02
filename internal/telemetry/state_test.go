package telemetry

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// useTmpHome points HOME at a temp directory so tests never touch the real
// ~/.saber directory.
func useTmpHome(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	return tmp
}

func TestLoadState_Missing(t *testing.T) {
	useTmpHome(t)

	s := LoadState()
	if s.Answered() {
		t.Error("a missing file must not count as an answer")
	}
	if s.Enabled || s.OrganizationID != "" {
		t.Errorf("expected a zero state, got %+v", s)
	}
}

func TestLoadState_Corrupt(t *testing.T) {
	tmp := useTmpHome(t)
	dir := filepath.Join(tmp, ".saber")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "telemetry.json"), []byte("{not json"), 0600); err != nil {
		t.Fatal(err)
	}

	if LoadState().Answered() {
		t.Error("a corrupt file must fall back to the zero state")
	}
}

func TestSaveAndLoadState(t *testing.T) {
	useTmpHome(t)

	want := &State{
		Enabled:        true,
		ConsentedAt:    time.Now().UTC().Truncate(time.Second),
		ConsentVersion: "1.2.3",
		ConsentMode:    ConsentModePrompt,
		OrganizationID: "org_123",
	}
	if err := SaveState(want); err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	got := LoadState()
	if !got.Enabled || got.OrganizationID != "org_123" || got.ConsentMode != ConsentModePrompt {
		t.Errorf("round trip lost data: %+v", got)
	}
	if !got.ConsentedAt.Equal(want.ConsentedAt) {
		t.Errorf("ConsentedAt = %v, want %v", got.ConsentedAt, want.ConsentedAt)
	}
	if !got.Answered() {
		t.Error("a saved consent time must count as an answer")
	}
}

func TestSaveState_FileModeIsPrivate(t *testing.T) {
	tmp := useTmpHome(t)

	if err := SaveState(&State{Enabled: true}); err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	info, err := os.Stat(filepath.Join(tmp, ".saber", "telemetry.json"))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("file mode = %o, want 600", perm)
	}
}

func TestSaveConsent(t *testing.T) {
	useTmpHome(t)

	got := SaveConsent(LoadState(), true, "9.9.9", ConsentModeCommand)
	if !got.Enabled || got.ConsentVersion != "9.9.9" || got.ConsentMode != ConsentModeCommand {
		t.Errorf("unexpected state: %+v", got)
	}
	if !LoadState().Answered() {
		t.Error("consent must persist to disk")
	}
}

func TestSaveConsent_DeclinePersists(t *testing.T) {
	useTmpHome(t)

	SaveConsent(LoadState(), false, "1.0.0", ConsentModePrompt)

	reloaded := LoadState()
	if reloaded.Enabled {
		t.Error("a declined answer must stay declined")
	}
	if !reloaded.Answered() {
		t.Error("a declined answer is still an answer, so the prompt must not repeat")
	}
}

func TestRecordOrganizationID(t *testing.T) {
	useTmpHome(t)

	RecordOrganizationID("org_abc")
	if got := LoadState().OrganizationID; got != "org_abc" {
		t.Errorf("OrganizationID = %q, want org_abc", got)
	}

	// An empty value must never clear a known id.
	RecordOrganizationID("")
	if got := LoadState().OrganizationID; got != "org_abc" {
		t.Errorf("an empty id changed the state to %q", got)
	}

	// A consent answer must survive an id update.
	SaveConsent(LoadState(), true, "1.0.0", ConsentModePrompt)
	RecordOrganizationID("org_def")
	after := LoadState()
	if after.OrganizationID != "org_def" {
		t.Errorf("OrganizationID = %q, want org_def", after.OrganizationID)
	}
	if !after.Enabled || !after.Answered() {
		t.Errorf("updating the id lost the consent answer: %+v", after)
	}
}
