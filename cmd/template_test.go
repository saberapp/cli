package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saberapp/cli/internal/client"
)

func TestParseExtractSignalType(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"company", "COMPANY", false},
		{"CONTACT", "CONTACT", false},
		{"  Company ", "COMPANY", false},
		{"market", "", true},
		{"", "", true},
	}
	for _, tt := range tests {
		got, err := parseExtractSignalType(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseExtractSignalType(%q) expected error, got %q", tt.input, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseExtractSignalType(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("parseExtractSignalType(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestLoadExtractClusters_BareArray(t *testing.T) {
	path := writeTempJSON(t, `[
		{"kind":"new","name":"Hiring","question":"Are they hiring?","answerType":"boolean","executionIds":["00000000-0000-0000-0000-000000000001"]},
		{"kind":"existing","templateId":"00000000-0000-0000-0000-0000000000aa","executionIds":["00000000-0000-0000-0000-000000000002"]}
	]`)
	clusters, err := loadExtractClusters(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(clusters) != 2 {
		t.Fatalf("expected 2 clusters, got %d", len(clusters))
	}
	if clusters[0].Kind != client.ExtractClusterKindNew || clusters[0].Name != "Hiring" {
		t.Errorf("first cluster wrong: %+v", clusters[0])
	}
	if clusters[1].Kind != client.ExtractClusterKindExisting || clusters[1].TemplateID == "" {
		t.Errorf("second cluster wrong: %+v", clusters[1])
	}
}

func TestLoadExtractClusters_WrappedProposeResponse(t *testing.T) {
	// Mirror the propose response shape — apply must accept it as-is so users
	// can `propose --json > plan.json` and feed it back unchanged.
	path := writeTempJSON(t, `{
		"clusters":[
			{"kind":"new","name":"X","question":"Q?","answerType":"number","executionIds":["00000000-0000-0000-0000-000000000003"]}
		],
		"totalCandidates":1,
		"processedCandidates":1,
		"hasMore":false
	}`)
	clusters, err := loadExtractClusters(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(clusters) != 1 || clusters[0].Name != "X" {
		t.Errorf("unexpected clusters: %+v", clusters)
	}
}

func TestLoadExtractClusters_EmptyFile(t *testing.T) {
	path := writeTempJSON(t, "")
	_, err := loadExtractClusters(path)
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected 'empty' error, got %v", err)
	}
}

func TestLoadExtractClusters_Malformed(t *testing.T) {
	path := writeTempJSON(t, `{not valid json`)
	if _, err := loadExtractClusters(path); err == nil {
		t.Errorf("expected parse error, got nil")
	}
}

func TestLoadExtractClusters_NoClustersKey(t *testing.T) {
	// Wrapped object without a `clusters` field — the loader returns an empty
	// slice (nil), and the caller's "no clusters to apply" check kicks in.
	path := writeTempJSON(t, `{"unrelated":42}`)
	clusters, err := loadExtractClusters(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(clusters) != 0 {
		t.Errorf("expected zero clusters, got %d", len(clusters))
	}
}

func TestLoadExtractClusters_MissingFile(t *testing.T) {
	_, err := loadExtractClusters("/nonexistent/path/plan.json")
	if err == nil || !strings.Contains(err.Error(), "read plan") {
		t.Errorf("expected read error, got %v", err)
	}
}

func writeTempJSON(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "plan.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}
