package localization

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "nicknames-test")
	if err != nil {
		os.Exit(1)
	}
	overridesFile = filepath.Join(dir, "nicknames.json")
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func TestInitOverrides_NoFile(t *testing.T) {
	if err := InitOverrides(); err != nil {
		t.Fatalf("InitOverrides: %v", err)
	}
	teams, _ := BuildFullDict()
	if teams["Vitality"] != "小蜜蜂" {
		t.Errorf("expected catalog default 小蜜蜂, got %q", teams["Vitality"])
	}
}

func TestSetAndGetTeamOverride(t *testing.T) {
	InitOverrides()

	if err := SetTeamOverride("Vitality", "蜜蜂"); err != nil {
		t.Fatalf("SetTeamOverride: %v", err)
	}
	teams, _ := BuildFullDict()
	if teams["Vitality"] != "蜜蜂" {
		t.Errorf("expected 蜜蜂, got %q", teams["Vitality"])
	}
}

func TestSetAndGetPlayerOverride(t *testing.T) {
	InitOverrides()

	if err := SetPlayerOverride("donk", "小驴"); err != nil {
		t.Fatalf("SetPlayerOverride: %v", err)
	}
	_, players := BuildFullDict()
	if players["donk"] != "小驴" {
		t.Errorf("expected 小驴, got %q", players["donk"])
	}
}

func TestDeleteOverride_EmptyNickname(t *testing.T) {
	InitOverrides()
	SetTeamOverride("Vitality", "蜜蜂")

	if err := SetTeamOverride("Vitality", ""); err != nil {
		t.Fatalf("delete: %v", err)
	}
	teams, _ := BuildFullDict()
	if teams["Vitality"] != "小蜜蜂" {
		t.Errorf("expected catalog default 小蜜蜂 after delete, got %q", teams["Vitality"])
	}
}

func TestOverridePersistence(t *testing.T) {
	InitOverrides()
	SetTeamOverride("Vitality", "蜜蜂")

	if err := InitOverrides(); err != nil {
		t.Fatalf("re-init: %v", err)
	}
	teams, _ := BuildFullDict()
	if teams["Vitality"] != "蜜蜂" {
		t.Errorf("persistence failed, got %q", teams["Vitality"])
	}
}

func TestConcurrentReadWrite(t *testing.T) {
	InitOverrides()
	SetTeamOverride("Vitality", "test")

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				BuildFullDict()
			}
			done <- true
		}()
	}
	go func() {
		for j := 0; j < 50; j++ {
			SetTeamOverride("Vitality", "test")
		}
		done <- true
	}()
	for i := 0; i < 11; i++ {
		<-done
	}

	teams, _ := BuildFullDict()
	if teams["Vitality"] != "test" {
		t.Errorf("expected 'test' after concurrent writes, got %q", teams["Vitality"])
	}
}
