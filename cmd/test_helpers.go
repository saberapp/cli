package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// writeTempJSON writes a temporary file with the given content and returns its path.
// It's used by tests to create temporary schema/plan files for testing file-based input.
func writeTempJSON(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "temp.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}
