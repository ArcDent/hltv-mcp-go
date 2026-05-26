package handlers

import (
	"net/http"

	"github.com/arcdent/hltv-mcp/internal/types"
)

func (h *Handlers) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	t := r.URL.Query().Get("type")
	if t == "team" {
		resp := h.f.ResolveTeam(types.ResolveQuery{Name: q, Limit: 10})
		writeJSON(w, resp)
		return
	}
	resp := h.f.ResolvePlayer(types.ResolveQuery{Name: q, Limit: 10})
	writeJSON(w, resp)
}

func (h *Handlers) GetTeam(w http.ResponseWriter, r *http.Request) {
	// stub — will be wired after scraper integration
	writeJSON(w, map[string]string{"status": "not yet implemented"})
}

func (h *Handlers) GetPlayer(w http.ResponseWriter, r *http.Request) {
	// stub — will be wired after scraper integration
	writeJSON(w, map[string]string{"status": "not yet implemented"})
}
