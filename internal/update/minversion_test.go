package update

import "testing"

func TestIsBelow(t *testing.T) {
	tests := []struct {
		current, other string
		want           bool
	}{
		{"0.1.13", "0.2.0", true},
		{"0.2.0", "0.2.0", false},
		{"0.2.1", "0.2.0", false},
		{"0.9.0", "0.10.0", true},
		{"1.0.0", "0.10.0", false},
		{"v0.1.13", "v0.2.0", true},
		{"0.2.0-dev", "0.2.0", false},
		{"dev", "0.2.0", false},
		{"0.1.0", "", false},
	}
	for _, tt := range tests {
		if got := IsBelow(tt.current, tt.other); got != tt.want {
			t.Errorf("IsBelow(%q, %q) = %v, want %v", tt.current, tt.other, got, tt.want)
		}
	}
}

func TestFormatMinVersionWarning(t *testing.T) {
	got := FormatMinVersionWarning("0.1.13", "0.2.0")
	want := `Warning: saber CLI v0.1.13 is below the minimum supported version v0.2.0 and will stop working. Run "saber update".`
	if got != want {
		t.Errorf("FormatMinVersionWarning = %q, want %q", got, want)
	}
}

func TestFormatMinVersionFloorWarning(t *testing.T) {
	got := FormatMinVersionFloorWarning("0.2.0", "0.2.1")
	want := `Warning: saber CLI v0.2.0 is the oldest version the API supports. Update to v0.2.1 before support ends. Run "saber update".`
	if got != want {
		t.Errorf("FormatMinVersionFloorWarning = %q, want %q", got, want)
	}
}
