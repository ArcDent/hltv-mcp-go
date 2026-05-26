package handlers

import (
	"net/http"

	"github.com/arcdent/hltv-mcp/internal/types"
)

func (h *Handlers) GetRealtimeNews(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	resp := h.f.GetRealtimeNews(types.RealtimeNewsQuery{
		Limit:  atoi(q.Get("limit")),
		Offset: atoi(q.Get("offset")),
	})
	writeJSON(w, resp)
}

func (h *Handlers) GetNewsDigest(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	resp := h.f.GetNewsDigest(types.NewsDigestQuery{
		Tag:    q.Get("tag"),
		Month:  q.Get("month"),
		Year:   atoi(q.Get("year")),
		Limit:  atoi(q.Get("limit")),
		Offset: atoi(q.Get("offset")),
	})
	writeJSON(w, resp)
}
