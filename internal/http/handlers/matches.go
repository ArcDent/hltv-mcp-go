package handlers

import (
	"net/http"

	"github.com/arcdent/hltv-mcp/internal/types"
)

func (h *Handlers) GetTodayMatches(w http.ResponseWriter, r *http.Request) {
	resp := h.f.GetTodayMatches()
	writeJSON(w, resp)
}

func (h *Handlers) GetUpcomingMatches(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	resp := h.f.GetUpcomingMatches(types.UpcomingMatchesQuery{
		Team:  q.Get("team"),
		Event: q.Get("event"),
		Limit: atoi(q.Get("limit")),
		Days:  atoi(q.Get("days")),
	})
	writeJSON(w, resp)
}

func (h *Handlers) GetResults(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	resp := h.f.GetResultsRecent(types.ResultsRecentQuery{
		Team:  q.Get("team"),
		Event: q.Get("event"),
		Limit: atoi(q.Get("limit")),
		Days:  atoi(q.Get("days")),
	})
	writeJSON(w, resp)
}
