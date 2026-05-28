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

func TestPlayerNickname_Builtin(t *testing.T) {
	if n := PlayerNickname("ZywOo"); n != "载物" {
		t.Errorf("expected 载物, got %q", n)
	}
	if n := PlayerNickname("donk"); n != "小驴" {
		t.Errorf("expected 小驴, got %q", n)
	}
}

func TestPlayerNickname_Unknown(t *testing.T) {
	if n := PlayerNickname("UnknownPlayerXYZ"); n != "" {
		t.Errorf("expected empty for unknown player, got %q", n)
	}
}

func TestTeamNickname(t *testing.T) {
	if n := TeamNickname("Spirit"); n != "绿龙" {
		t.Errorf("expected 绿龙 via alias, got %q", n)
	}
	if n := TeamNickname("Vitality"); n != "小蜜蜂" {
		t.Errorf("expected 小蜜蜂, got %q", n)
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

func TestTeamNickname_WithOverride(t *testing.T) {
	InitOverrides()
	SetTeamOverride("Vitality", "蜜蜂战队")

	if n := TeamNickname("Vitality"); n != "蜜蜂战队" {
		t.Errorf("expected 蜜蜂战队 override, got %q", n)
	}
	SetTeamOverride("Vitality", "")
	InitOverrides()
}

func TestPlayerNickname_WithOverride(t *testing.T) {
	InitOverrides()
	SetPlayerOverride("donk", "小驴子")

	if n := PlayerNickname("donk"); n != "小驴子" {
		t.Errorf("expected 小驴子 override, got %q", n)
	}
	SetPlayerOverride("donk", "")
	InitOverrides()
}

