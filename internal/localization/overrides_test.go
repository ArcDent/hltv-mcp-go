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
	if n := GetTeamOverride("Vitality"); n != "" {
		t.Errorf("expected empty, got %q", n)
	}
	if n := GetPlayerOverride("donk"); n != "" {
		t.Errorf("expected empty, got %q", n)
	}
}

func TestSetAndGetTeamOverride(t *testing.T) {
	InitOverrides()

	if err := SetTeamOverride("Vitality", "蜜蜂"); err != nil {
		t.Fatalf("SetTeamOverride: %v", err)
	}
	if n := GetTeamOverride("Vitality"); n != "蜜蜂" {
		t.Errorf("expected 蜜蜂, got %q", n)
	}
}

func TestSetAndGetPlayerOverride(t *testing.T) {
	InitOverrides()

	if err := SetPlayerOverride("donk", "小驴"); err != nil {
		t.Fatalf("SetPlayerOverride: %v", err)
	}
	if n := GetPlayerOverride("donk"); n != "小驴" {
		t.Errorf("expected 小驴, got %q", n)
	}
}

func TestDeleteOverride_EmptyNickname(t *testing.T) {
	InitOverrides()
	SetTeamOverride("Vitality", "蜜蜂")

	if err := SetTeamOverride("Vitality", ""); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if n := GetTeamOverride("Vitality"); n != "" {
		t.Errorf("expected empty after delete, got %q", n)
	}
}

func TestOverridePersistence(t *testing.T) {
	InitOverrides()
	SetTeamOverride("Vitality", "蜜蜂")

	if err := InitOverrides(); err != nil {
		t.Fatalf("re-init: %v", err)
	}
	if n := GetTeamOverride("Vitality"); n != "蜜蜂" {
		t.Errorf("persistence failed, got %q", n)
	}
}

func TestConcurrentReadWrite(t *testing.T) {
	InitOverrides()
	SetTeamOverride("Vitality", "test")

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				GetTeamOverride("Vitality")
				GetPlayerOverride("donk")
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

	if n := GetTeamOverride("Vitality"); n != "test" {
		t.Errorf("expected 'test' after concurrent writes, got %q", n)
	}
}
