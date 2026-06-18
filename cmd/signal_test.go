package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseOutputSchema(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]any
		wantErr bool
	}{
		{
			name:    "empty string returns nil",
			input:   "",
			want:    nil,
			wantErr: false,
		},
		{
			name:  "inline JSON object",
			input: `{"type":"object","properties":{"name":{"type":"string"}}}`,
			want: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type": "string",
					},
				},
			},
			wantErr: false,
		},
		{
			name:  "inline JSON with nested structure",
			input: `{"type":"object","properties":{"revenue":{"type":"number"},"employees":{"type":"number"},"funding_stage":{"type":"string"}}}`,
			want: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"revenue": map[string]any{
						"type": "number",
					},
					"employees": map[string]any{
						"type": "number",
					},
					"funding_stage": map[string]any{
						"type": "string",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   `{not valid json}`,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "malformed JSON with missing quotes",
			input:   `{type: string}`,
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOutputSchema(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseOutputSchema(%q) expected error, got nil", tt.input)
				}
				if !strings.Contains(err.Error(), "parse output schema JSON") {
					t.Errorf("parseOutputSchema(%q) error should mention 'parse output schema JSON', got: %v", tt.input, err)
				}
				return
			}
			if err != nil {
				t.Errorf("parseOutputSchema(%q) unexpected error: %v", tt.input, err)
				return
			}
			if !schemasEqual(got, tt.want) {
				gotJSON, _ := json.Marshal(got)
				wantJSON, _ := json.Marshal(tt.want)
				t.Errorf("parseOutputSchema(%q) = %s, want %s", tt.input, gotJSON, wantJSON)
			}
		})
	}
}

func TestParseOutputSchema_FileInput(t *testing.T) {
	tests := []struct {
		name       string
		fileContent string
		wantErr    bool
		wantKey    string
	}{
		{
			name:        "valid schema from file",
			fileContent: `{"type":"object","properties":{"id":{"type":"integer"}}}`,
			wantErr:     false,
			wantKey:     "type",
		},
		{
			name:        "invalid JSON in file",
			fileContent: `{invalid}`,
			wantErr:     true,
			wantKey:     "",
		},
		{
			name:        "empty file",
			fileContent: ``,
			wantErr:     true,
			wantKey:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTempJSON(t, tt.fileContent)
			got, err := parseOutputSchema("@" + path)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseOutputSchema(@file) expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("parseOutputSchema(@file) unexpected error: %v", err)
				return
			}

			if tt.wantKey != "" && got[tt.wantKey] == nil {
				t.Errorf("parseOutputSchema(@file) result missing key %q", tt.wantKey)
			}
		})
	}
}

func TestParseOutputSchema_FileNotFound(t *testing.T) {
	got, err := parseOutputSchema("@/nonexistent/path/schema.json")
	if err == nil {
		t.Errorf("parseOutputSchema(@nonexistent) expected error, got nil: %v", got)
	}
	if !strings.Contains(err.Error(), "read output schema file") {
		t.Errorf("parseOutputSchema(@nonexistent) error should mention 'read output schema file', got: %v", err)
	}
}

// schemasEqual compares two schema maps recursively
func schemasEqual(a, b map[string]any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !mapValueEqual(a[k], b[k]) {
			return false
		}
	}
	return true
}

// mapValueEqual compares two values which may be nested maps or other types
func mapValueEqual(a, b any) bool {
	aMap, aIsMap := a.(map[string]any)
	bMap, bIsMap := b.(map[string]any)

	if aIsMap && bIsMap {
		return schemasEqual(aMap, bMap)
	}
	if aIsMap != bIsMap {
		return false
	}

	// For non-map values, use simple equality (works for strings, numbers, bools)
	return a == b
}
