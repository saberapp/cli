package config

import (
	"testing"
)

func TestValidateKeyFormat(t *testing.T) {
	// Build a valid key: "sk_live_" + 43 chars = 51 total
	validSuffix := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQ"
	validKey := KeyPrefix + validSuffix

	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{"valid key", validKey, false},
		{"wrong prefix", "sk_test_" + validSuffix, true},
		{"too short", "sk_live_short", true},
		{"too long", validKey + "x", true},
		{"empty", "", true},
		{"no prefix at all", validSuffix, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateKeyFormat(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateKeyFormat(%q) error = %v, wantErr %v", tt.key, err, tt.wantErr)
			}
		})
	}
}

func TestErrNotAuthenticated(t *testing.T) {
	err := &ErrNotAuthenticated{}
	if err.Error() == "" {
		t.Error("ErrNotAuthenticated.Error() should not be empty")
	}
}
