package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestParseRange(t *testing.T) {
	tests := []struct {
		input   string
		min     float64
		max     float64
		points  float64
		wantErr bool
	}{
		{"0:500:5", 0, 500, 5, false},
		{"100.5:200.5:12.25", 100.5, 200.5, 12.25, false},
		{" 1 : 2 : 3 ", 1, 2, 3, false},
		{"0:500", 0, 0, 0, true},
		{"a:b:c", 0, 0, 0, true},
		{"", 0, 0, 0, true},
	}
	for _, tt := range tests {
		got, err := parseRange(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseRange(%q) expected error, got %+v", tt.input, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseRange(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if got.Min != tt.min || got.Max != tt.max || got.Points != tt.points {
			t.Errorf("parseRange(%q) = %+v, want min=%v max=%v points=%v",
				tt.input, got, tt.min, tt.max, tt.points)
		}
	}
}

func TestParseChoice(t *testing.T) {
	tests := []struct {
		input   string
		key     string
		points  float64
		wantErr bool
	}{
		{"Salesforce:10", "Salesforce", 10, false},
		{"None:-10", "None", -10, false},
		{"https://example.com:5", "https://example.com", 5, false},
		{"key:", "", 0, true},
		{":10", "", 0, true},
		{"no-colon", "", 0, true},
	}
	for _, tt := range tests {
		key, pts, err := parseChoice(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseChoice(%q) expected error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseChoice(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if key != tt.key || pts != tt.points {
			t.Errorf("parseChoice(%q) = (%q, %v), want (%q, %v)",
				tt.input, key, pts, tt.key, tt.points)
		}
	}
}

func TestBuildPointValues_RawJSON(t *testing.T) {
	cmd := newScoringRuleUpsertCmd()
	pv, err := buildPointValues(cmd, `{"ranges":[{"min":0,"max":500,"points":5}]}`, "", 0, 0, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pv.Ranges) != 1 || pv.Ranges[0].Max != 500 || pv.Ranges[0].Points != 5 {
		t.Errorf("unexpected pv: %+v", pv)
	}
}

func TestBuildPointValues_BoolFlags(t *testing.T) {
	cmd := newScoringRuleUpsertCmd()
	if err := cmd.Flags().Set("true", "20"); err != nil {
		t.Fatalf("set --true: %v", err)
	}
	if err := cmd.Flags().Set("false", "-5"); err != nil {
		t.Fatalf("set --false: %v", err)
	}
	pv, err := buildPointValues(cmd, "", "", 20, -5, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pv.True == nil || *pv.True != 20 {
		t.Errorf("True = %v, want 20", pv.True)
	}
	if pv.False == nil || *pv.False != -5 {
		t.Errorf("False = %v, want -5", pv.False)
	}
}

func TestBuildPointValues_Ranges(t *testing.T) {
	cmd := newScoringRuleUpsertCmd()
	pv, err := buildPointValues(cmd, "", "", 0, 0, []string{"0:100:2", "100:1000:10"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pv.Ranges) != 2 || pv.Ranges[1].Points != 10 {
		t.Errorf("unexpected ranges: %+v", pv.Ranges)
	}
}

func TestBuildPointValues_Choices(t *testing.T) {
	cmd := newScoringRuleUpsertCmd()
	pv, err := buildPointValues(cmd, "", "", 0, 0, nil, []string{"Salesforce:10", "HubSpot:8"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pv.Choices["Salesforce"] != 10 || pv.Choices["HubSpot"] != 8 {
		t.Errorf("unexpected choices: %+v", pv.Choices)
	}
}

func TestBuildPointValues_NoShape(t *testing.T) {
	cmd := newScoringRuleUpsertCmd()
	_, err := buildPointValues(cmd, "", "", 0, 0, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "must provide point values") {
		t.Errorf("expected missing-shape error, got %v", err)
	}
}

func TestBuildPointValues_MultipleShapes(t *testing.T) {
	cmd := newScoringRuleUpsertCmd()
	if err := cmd.Flags().Set("true", "20"); err != nil {
		t.Fatalf("set --true: %v", err)
	}
	_, err := buildPointValues(cmd, "", "", 20, 0, []string{"0:100:5"}, nil)
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("expected mutually-exclusive error, got %v", err)
	}
}

func TestBuildPointValues_JSONAndFile(t *testing.T) {
	cmd := newScoringRuleUpsertCmd()
	_, err := buildPointValues(cmd, `{"true":1}`, "/tmp/whatever.json", 0, 0, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("expected mutually-exclusive error, got %v", err)
	}
}

// Sanity check: scoring command tree wires up cleanly into root.
func TestRootCmd_ScoringCommand_HasSubcommands(t *testing.T) {
	root := NewRootCmd("1.0.0", "abc", "now")
	cmd, _, err := root.Find([]string{"scoring"})
	if err != nil {
		t.Fatalf("scoring command not found: %v", err)
	}
	wantSubs := []string{"profile", "rule", "assignment", "scores", "compute"}
	for _, name := range wantSubs {
		if !hasSubcommand(cmd, name) {
			t.Errorf("scoring is missing subcommand %q", name)
		}
	}
}

func hasSubcommand(parent *cobra.Command, name string) bool {
	for _, sub := range parent.Commands() {
		if sub.Name() == name {
			return true
		}
	}
	return false
}
