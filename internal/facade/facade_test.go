package facade

import "testing"

func TestIsPlaceholderText(t *testing.T) {
	for _, s := range []string{"x", "y", "?", "-", "n/a", "null", "undefined", "tbd", ""} {
		if !isPlaceholderText(s) {
			t.Errorf("%q should be placeholder", s)
		}
	}
	if isPlaceholderText("Vitality") {
		t.Error("Vitality should NOT be placeholder")
	}
}

func TestStripGenericFilter(t *testing.T) {
	tests := []struct{ in, want string }{
		{"today matches", ""},
		{"今日赛程", ""},
		{"Spirit", "Spirit"},
		{"today Vitality", "Vitality"},
		{"Vitality matches", "Vitality"},
	}
	for _, tt := range tests {
		if got := stripGenericFilter(tt.in); got != tt.want {
			t.Errorf("stripGenericFilter(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
