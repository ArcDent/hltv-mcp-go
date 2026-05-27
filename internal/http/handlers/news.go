package handlers

import (
	"context"
	"net/http"
	"time"

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

func (h *Handlers) GetNewsArticle(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		writeError(w, http.StatusBadRequest, "url required")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	article, err := h.f.GetNewsArticleCached(ctx, url)
	if err != nil {
		writeJSON(w, map[string]any{
			"error": map[string]any{"code": "UPSTREAM_UNAVAILABLE", "message": "文章抓取失败，请在 HLTV 阅读原文"},
			"meta":  map[string]any{"partial": true},
		})
		return
	}
	writeJSON(w, map[string]any{"data": article, "meta": map[string]any{"partial": false}})
}
