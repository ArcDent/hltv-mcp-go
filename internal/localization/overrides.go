package localization

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type overridesStore struct {
	mu      sync.RWMutex
	teams   map[string]string
	players map[string]string
}

var ov = &overridesStore{}

var overridesFile = "data/nicknames.json"

type overridesData struct {
	Teams   map[string]string `json:"teams"`
	Players map[string]string `json:"players"`
}

func InitOverrides() error {
	data, err := os.ReadFile(overridesFile)
	if err != nil {
		if os.IsNotExist(err) {
			ov.mu.Lock()
			ov.teams = make(map[string]string)
			ov.players = make(map[string]string)
			ov.mu.Unlock()
			return nil
		}
		return err
	}
	var d overridesData
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}
	ov.mu.Lock()
	defer ov.mu.Unlock()
	ov.teams = d.Teams
	ov.players = d.Players
	if ov.teams == nil {
		ov.teams = make(map[string]string)
	}
	if ov.players == nil {
		ov.players = make(map[string]string)
	}
	return nil
}

func saveOverrides() error {
	d := overridesData{Teams: ov.teams, Players: ov.players}
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(overridesFile), 0700); err != nil {
		return err
	}
	return os.WriteFile(overridesFile, data, 0600)
}

func SetTeamOverride(name, nickname string) error {
	ov.mu.Lock()
	defer ov.mu.Unlock()
	if nickname == "" {
		delete(ov.teams, name)
	} else {
		ov.teams[name] = nickname
	}
	return saveOverrides()
}

func SetPlayerOverride(name, nickname string) error {
	ov.mu.Lock()
	defer ov.mu.Unlock()
	if nickname == "" {
		delete(ov.players, name)
	} else {
		ov.players[name] = nickname
	}
	return saveOverrides()
}
