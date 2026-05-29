package handlers

import (
	"net/http"
	"time"

	"github.com/arcdent/hltv-mcp/internal/types"
)

func (h *Handlers) GetTodayMatches(w http.ResponseWriter, r *http.Request) {
	resp := h.f.GetTodayMatches()
	writeJSON(w, resp)
}

func (h *Handlers) GetUpcomingMatches(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	team, event := q.Get("team"), q.Get("event")
	limit, days := atoi(q.Get("limit")), atoi(q.Get("days"))
	h.withTimeout(w, r, 120*time.Second, "赛程请求超时，请重试", func() *types.ToolResponse {
		return h.f.GetUpcomingMatches(types.UpcomingMatchesQuery{Team: team, Event: event, Limit: limit, Days: days})
	})
}

func (h *Handlers) GetResults(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	team, event := q.Get("team"), q.Get("event")
	limit, days := atoi(q.Get("limit")), atoi(q.Get("days"))
	h.withTimeout(w, r, 45*time.Second, "赛果请求超时，请重试", func() *types.ToolResponse {
		return h.f.GetResultsRecent(types.ResultsRecentQuery{Team: team, Event: event, Limit: limit, Days: days})
	})
}

func (h *Handlers) GetEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	matchType := q.Get("type")
	limit := atoi(q.Get("limit"))
	if limit == 0 {
		limit = 150
	}
	h.withTimeout(w, r, 45*time.Second, "赛事请求超时，请重试", func() *types.ToolResponse {
		return h.f.GetEvents(matchType, limit)
	})
}
