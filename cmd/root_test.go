package cmd

import (
	"bytes"
	"testing"
	"time"

	"github.com/saberapp/cli/internal/update"
)

// ---------- shouldSkipUpdateCheck ----------

func TestShouldSkipUpdateCheck_DevBuild(t *testing.T) {
	root := NewRootCmd("dev", "abc", "now")
	cliVersion = "dev"
	if !shouldSkipUpdateCheck(root) {
		t.Error("expected skip for dev build")
	}
}

func TestShouldSkipUpdateCheck_JSONFlag(t *testing.T) {
	NewRootCmd("1.0.0", "abc", "now")
	jsonOutput = true
	defer func() { jsonOutput = false }()

	root := NewRootCmd("1.0.0", "abc", "now")
	if !shouldSkipUpdateCheck(root) {
		t.Error("expected skip when --json is set")
	}
}

func TestShouldSkipUpdateCheck_QuietFlag(t *testing.T) {
	root := NewRootCmd("1.0.0", "abc", "now")
	quiet = true
	defer func() { quiet = false }()
	if !shouldSkipUpdateCheck(root) {
		t.Error("expected skip when --quiet is set")
	}
}

func TestShouldSkipUpdateCheck_EnvVar(t *testing.T) {
	root := NewRootCmd("1.0.0", "abc", "now")
	t.Setenv("SABER_NO_UPDATE_CHECK", "1")
	if !shouldSkipUpdateCheck(root) {
		t.Error("expected skip when SABER_NO_UPDATE_CHECK is set")
	}
}

func TestShouldSkipUpdateCheck_UpdateCommand(t *testing.T) {
	root := NewRootCmd("1.0.0", "abc", "now")
	updateCmd, _, err := root.Find([]string{"update"})
	if err != nil {
		t.Fatalf("could not find update command: %v", err)
	}
	if !shouldSkipUpdateCheck(updateCmd) {
		t.Error("expected skip for the update command itself")
	}
}

func TestShouldSkipUpdateCheck_NonTTY(t *testing.T) {
	root := NewRootCmd("1.0.0", "abc", "now")
	// In test context stderr is never a TTY, so this should skip.
	if !shouldSkipUpdateCheck(root) {
		t.Error("expected skip when stderr is not a TTY")
	}
}

// ---------- Update notice integration (lifecycle simulation) ----------

func TestUpdateLifecycle_NoticeShown(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// Pre-seed cache with a newer version, checked recently (synchronous path).
	if err := update.SaveState(&update.State{
		LatestVersion: "9.9.9",
		CheckedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	// Simulate PersistentPreRun:
	ch := update.RunBackgroundCheck("1.0.0", 24*time.Hour)

	// Simulate PersistentPostRun (non-blocking drain):
	var stderr bytes.Buffer
	select {
	case msg, ok := <-ch:
		if ok && msg != "" {
			stderr.WriteString(msg + "\n")
		}
	default:
	}

	if stderr.Len() == 0 {
		t.Error("expected update notice, got nothing")
	}
	expected := `Notice: Update available v1.0.0 -> v9.9.9. Run "saber update" to upgrade.`
	if !bytes.Contains(stderr.Bytes(), []byte(expected)) {
		t.Errorf("stderr = %q, want it to contain %q", stderr.String(), expected)
	}
}

func TestUpdateLifecycle_NoNoticeWhenCurrent(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	if err := update.SaveState(&update.State{
		LatestVersion: "1.0.0",
		CheckedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	ch := update.RunBackgroundCheck("1.0.0", 24*time.Hour)

	var stderr bytes.Buffer
	select {
	case msg, ok := <-ch:
		if ok && msg != "" {
			stderr.WriteString(msg + "\n")
		}
	default:
	}

	if stderr.Len() != 0 {
		t.Errorf("expected no notice when up to date, got %q", stderr.String())
	}
}

func TestUpdateLifecycle_NoNoticeWhenNewer(t *testing.T) {
	// User is on a version newer than latest release (e.g. dev or pre-release).
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	if err := update.SaveState(&update.State{
		LatestVersion: "0.9.0",
		CheckedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	ch := update.RunBackgroundCheck("1.0.0", 24*time.Hour)

	var stderr bytes.Buffer
	select {
	case msg, ok := <-ch:
		if ok && msg != "" {
			stderr.WriteString(msg + "\n")
		}
	default:
	}

	if stderr.Len() != 0 {
		t.Errorf("expected no notice when on newer version, got %q", stderr.String())
	}
}

// ---------- Backwards compat: commands don't error ----------

func TestRootCmd_VersionCommand_NoError(t *testing.T) {
	root := NewRootCmd("1.2.3", "abc123", "2025-01-01")
	root.SetArgs([]string{"version"})
	if err := root.Execute(); err != nil {
		t.Fatalf("version command failed: %v", err)
	}
}

func TestRootCmd_UpdateCommand_DevBuild_NoError(t *testing.T) {
	root := NewRootCmd("dev", "none", "unknown")
	root.SetArgs([]string{"update"})
	if err := root.Execute(); err != nil {
		t.Fatalf("update command with dev build failed: %v", err)
	}
}

func TestRootCmd_HelpCommand_NoError(t *testing.T) {
	root := NewRootCmd("1.0.0", "abc", "now")
	root.SetArgs([]string{"help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("help command failed: %v", err)
	}
}
