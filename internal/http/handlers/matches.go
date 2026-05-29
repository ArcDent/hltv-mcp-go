package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/arcdent/hltv-mcp/internal/types"
)

func (h *Handlers) GetTodayMatches(w http.ResponseWriter, r *http.Request) {
	resp := h.f.GetTodayMatches()
	writeJSON(w, resp)
}

func (h *Handlers) GetUpcomingMatches(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	q := r.URL.Query()
	team := q.Get("team")
	event := q.Get("event")
	limit := atoi(q.Get("limit"))
	days := atoi(q.Get("days"))

	resultCh := make(chan *types.ToolResponse, 1)
	go func() {
		resultCh <- h.f.GetUpcomingMatches(types.UpcomingMatchesQuery{
			Team: team, Event: event, Limit: limit, Days: days,
		})
	}()

	select {
	case resp := <-resultCh:
		writeJSON(w, resp)
	case <-ctx.Done():
		writeJSON(w, map[string]any{
			"error": map[string]any{"code": "TIMEOUT", "message": "赛程请求超时，请重试"},
			"meta":  map[string]any{"partial": true},
		})
	}
}

func (h *Handlers) GetResults(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
	defer cancel()

	q := r.URL.Query()
	team := q.Get("team")
	event := q.Get("event")
	limit := atoi(q.Get("limit"))
	days := atoi(q.Get("days"))

	resultCh := make(chan *types.ToolResponse, 1)
	go func() {
		resultCh <- h.f.GetResultsRecent(types.ResultsRecentQuery{
			Team: team, Event: event, Limit: limit, Days: days,
		})
	}()

	select {
	case resp := <-resultCh:
		writeJSON(w, resp)
	case <-ctx.Done():
		writeJSON(w, map[string]any{
			"error": map[string]any{"code": "TIMEOUT", "message": "赛果请求超时，请重试"},
			"meta":  map[string]any{"partial": true},
		})
	}
}

func (h *Handlers) GetEvents(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
	defer cancel()

	q := r.URL.Query()
	matchType := q.Get("type")
	limit := atoi(q.Get("limit"))
	if limit == 0 {
		limit = 150
	}

	resultCh := make(chan *types.ToolResponse, 1)
	go func() {
		resultCh <- h.f.GetEvents(matchType, limit)
	}()

	select {
	case resp := <-resultCh:
		writeJSON(w, resp)
	case <-ctx.Done():
		writeJSON(w, map[string]any{
			"error": map[string]any{"code": "TIMEOUT", "message": "赛事请求超时，请重试"},
			"meta":  map[string]any{"partial": true},
		})
	}
}
