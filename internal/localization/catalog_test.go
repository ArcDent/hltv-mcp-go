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

func TestBuildFullDict_TeamsHasVariants(t *testing.T) {
	teams, _ := BuildFullDict()
	if teams["Spirit"] != "绿龙" {
		t.Errorf("expected Spirit→绿龙, got %q", teams["Spirit"])
	}
	if teams["Team Spirit"] != "绿龙" {
		t.Errorf("expected Team Spirit→绿龙, got %q", teams["Team Spirit"])
	}
	if teams["NAVI"] != "NaVi" {
		t.Errorf("expected NAVI→NaVi, got %q", teams["NAVI"])
	}
}

func TestBuildFullDict_Players(t *testing.T) {
	_, players := BuildFullDict()
	if players["ZywOo"] != "载物" {
		t.Errorf("expected ZywOo→载物, got %q", players["ZywOo"])
	}
}

func TestOverrides(t *testing.T) {
	InitOverrides()
	SetTeamOverride("Vitality", "蜜蜂战队")
	SetPlayerOverride("donk", "小驴子")

	teams, players := BuildFullDict()
	if teams["Vitality"] != "蜜蜂战队" {
		t.Errorf("expected Vitality→蜜蜂战队 override, got %q", teams["Vitality"])
	}
	if players["donk"] != "小驴子" {
		t.Errorf("expected donk→小驴子 override, got %q", players["donk"])
	}

	SetTeamOverride("Vitality", "")
	SetPlayerOverride("donk", "")
	InitOverrides()
}
