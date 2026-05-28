package localization

import "testing"

func TestLookupTeam_Spirit(t *testing.T) {
	for _, q := range []string{"Spirit", "Team Spirit", "绿龙", "spirit", "TEAM SPIRIT"} {
		if e := LookupTeam(q); e == nil || e.Canonical != "Team Spirit" {
			t.Errorf("lookup(%q) failed", q)
		}
	}
}

func TestLookupTeam_Vitality(t *testing.T) {
	for _, q := range []string{"Vitality", "小蜜蜂", "蜜蜂"} {
		if e := LookupTeam(q); e == nil || e.Canonical != "Vitality" {
			t.Errorf("lookup(%q) failed", q)
		}
	}
}

func TestFormatTeamDisplay(t *testing.T) {
	result := FormatTeamDisplay("Spirit")
	if result == "" || result == "Spirit" {
		t.Errorf("expected formatted display, got %q", result)
	}
}

func TestFormatEventDisplay(t *testing.T) {
	result := FormatEventDisplay("IEM Rio")
	if result == "" {
		t.Errorf("expected non-empty display")
	}
}

