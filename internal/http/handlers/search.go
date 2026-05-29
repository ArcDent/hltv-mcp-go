package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/arcdent/hltv-mcp/internal/types"
)

func (h *Handlers) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	kind := r.URL.Query().Get("type")
	h.withTimeout(w, r, 30*time.Second, "搜索请求超时，请重试", func() *types.ToolResponse {
		if kind == "team" {
			return h.f.ResolveTeam(types.ResolveQuery{Name: query, Limit: 10})
		}
		return h.f.ResolvePlayer(types.ResolveQuery{Name: query, Limit: 10})
	})
}

func (h *Handlers) GetTeam(w http.ResponseWriter, r *http.Request) {
	id := atoi(chi.URLParam(r, "id"))
	if id == 0 {
		writeError(w, http.StatusBadRequest, "invalid team id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	td, err := h.f.GetTeamDetailCached(ctx, id, "")
	if err != nil {
		writeJSON(w, map[string]any{
			"error": map[string]any{"code": "UPSTREAM_UNAVAILABLE", "message": "详情暂时不可用"},
			"meta":  map[string]any{"partial": true},
		})
		return
	}
	writeJSON(w, map[string]any{"data": td, "meta": map[string]any{"partial": false}})
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
