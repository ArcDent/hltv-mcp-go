package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
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
	writeJSON(w, map[string]string{"status": "not yet implemented"})
}

func (h *Handlers) GetPlayer(w http.ResponseWriter, r *http.Request) {
	id := atoi(chi.URLParam(r, "id"))
	if id == 0 {
		writeError(w, http.StatusBadRequest, "invalid player id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	pd, err := h.f.GetPlayerDetailCached(ctx, id, "")
	if err != nil {
		writeJSON(w, map[string]any{
			"error": map[string]any{"code": "UPSTREAM_UNAVAILABLE", "message": "详情暂时不可用"},
			"meta":  map[string]any{"partial": true},
		})
		return
	}
	writeJSON(w, map[string]any{"data": pd, "meta": map[string]any{"partial": false}})
}
