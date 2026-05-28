package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/arcdent/hltv-mcp/internal/localization"
)

type nicknameReq struct {
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
}

// GetNicknames returns the full nickname dictionaries.
func (h *Handlers) GetNicknames(w http.ResponseWriter, r *http.Request) {
	teams, players := localization.BuildFullDict()
	writeJSON(w, map[string]any{
		"teams":   teams,
		"players": players,
	})
}

// PutTeamNickname saves a team nickname override.
func (h *Handlers) PutTeamNickname(w http.ResponseWriter, r *http.Request) {
	var req nicknameReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	// Resolve to canonical via alias lookup
	e := localization.LookupTeam(req.Name)
	if e == nil {
		writeError(w, http.StatusBadRequest, "team not found in catalog")
		return
	}

	if err := localization.SetTeamOverride(e.Canonical, req.Nickname); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save")
		return
	}
	writeJSON(w, map[string]string{"status": "saved"})
}

// PutPlayerNickname saves a player nickname override.
func (h *Handlers) PutPlayerNickname(w http.ResponseWriter, r *http.Request) {
	var req nicknameReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	// Open mode: any player name is accepted
	if err := localization.SetPlayerOverride(req.Name, req.Nickname); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save")
		return
	}
	writeJSON(w, map[string]string{"status": "saved"})
}
