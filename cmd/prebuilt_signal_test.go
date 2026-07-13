package cmd

import "testing"

func TestValidateTechFlags(t *testing.T) {
	tests := []struct {
		name       string
		category   string
		technology string
		wantErr    bool
	}{
		{"neither set", "", "", true},
		{"both set", "crm", "salesforce", true},
		{"category only", "crm", "", false},
		{"technology only", "", "salesforce", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTechFlags(tt.category, tt.technology)
			if tt.wantErr && err == nil {
				t.Errorf("validateTechFlags(%q, %q) expected error, got nil", tt.category, tt.technology)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateTechFlags(%q, %q) unexpected error: %v", tt.category, tt.technology, err)
			}
		})
	}
}

func TestNewSignalCmdRegistersPrebuiltSubcommands(t *testing.T) {
	cmd := newSignalCmd()
	want := map[string]bool{
		"funding": false, "mna": false, "tech": false,
		"open-jobs": false, "firmographics": false,
	}
	for _, sub := range cmd.Commands() {
		if _, ok := want[sub.Name()]; ok {
			want[sub.Name()] = true
		}
	}
	for k, found := range want {
		if !found {
			t.Errorf("signal subcommand %q not registered", k)
		}
	}
}

func TestPrebuiltSignalCmdTechOnlyFlags(t *testing.T) {
	tech := newPrebuiltSignalCmd("tech")
	if tech.Flags().Lookup("category") == nil || tech.Flags().Lookup("technology") == nil {
		t.Errorf("tech subcommand missing --category/--technology flags")
	}
	funding := newPrebuiltSignalCmd("funding")
	if funding.Flags().Lookup("category") != nil || funding.Flags().Lookup("technology") != nil {
		t.Errorf("funding subcommand must not expose --category/--technology")
	}
	for _, name := range []string{"domain", "no-wait", "webhook", "force-refresh", "max-wait"} {
		if funding.Flags().Lookup(name) == nil {
			t.Errorf("funding subcommand missing --%s flag", name)
		}
	}
}
