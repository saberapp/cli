package cmd

import "testing"

func TestMinVersionWarning(t *testing.T) {
	below := `Warning: saber CLI v0.1.13 is below the minimum supported version v0.2.0 and will stop working. Run "saber update".`
	floor := `Warning: saber CLI v0.2.0 is the oldest version the API supports. Update to v0.2.1 before support ends. Run "saber update".`

	tests := []struct {
		name                     string
		current, minimum, latest string
		want                     string
	}{
		{"below minimum", "0.1.13", "0.2.0", "0.2.1", below},
		{"below minimum, no cached release", "0.1.13", "0.2.0", "", below},
		{"equal to minimum, newer release", "0.2.0", "0.2.0", "0.2.1", floor},
		{"equal to minimum, no newer release", "0.2.0", "0.2.0", "0.2.0", ""},
		{"equal to minimum, no cached release", "0.2.0", "0.2.0", "", ""},
		{"above minimum", "0.3.0", "0.2.0", "0.4.0", ""},
		{"dev build", "dev", "0.2.0", "0.2.1", ""},
		{"header absent", "0.1.13", "", "0.2.1", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := minVersionWarning(tt.current, tt.minimum, tt.latest); got != tt.want {
				t.Errorf("minVersionWarning(%q, %q, %q) = %q, want %q", tt.current, tt.minimum, tt.latest, got, tt.want)
			}
		})
	}
}
