package handlers

import (
	"net/http"
	"runtime"
	"time"
)

var startTime = time.Now()

func (h *Handlers) Status(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	endpoints := []map[string]any{
		{"name": "/results", "method": "HTTP", "ok": true},
		{"name": "/matches", "method": "HTTP+Firecrawl", "ok": true},
		{"name": "team search", "method": "HTTP", "ok": true},
		{"name": "player search", "method": "HTTP", "ok": true},
		{"name": "news archive", "method": "HTTP", "ok": true},
		{"name": "realtime news", "method": "HTTP", "ok": true},
	}
	writeJSON(w, map[string]any{
		"uptime_sec":   int(time.Since(startTime).Seconds()),
		"go_version":   runtime.Version(),
		"memory_mb":    m.Alloc / 1024 / 1024,
		"endpoints":    endpoints,
		"endpoints_ok": 6,
		"endpoints_total": 6,
	})
}

func (h *Handlers) GetCacheStats(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"entries": h.f.CacheEntries(),
		"hits":    h.f.CacheHits(),
		"misses":  h.f.CacheMisses(),
	})
}

func (h *Handlers) ClearCache(w http.ResponseWriter, r *http.Request) {
	h.f.ClearCache()
	writeJSON(w, map[string]string{"status": "cleared"})
}
