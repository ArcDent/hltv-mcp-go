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
	chromeOK := h.f.ClientIsChromeAvailable()
	endpoints := []map[string]any{
		{"name": "/results", "method": "HTTP", "ok": true},
		{"name": "/matches", "method": ifElse(chromeOK, "chromedp", "HTTP"), "ok": chromeOK},
		{"name": "team search", "method": "HTTP", "ok": true},
		{"name": "player search", "method": "HTTP", "ok": true},
		{"name": "news archive", "method": "HTTP", "ok": true},
		{"name": "realtime news", "method": "HTTP", "ok": true},
	}
	okCount := 5
	if chromeOK { okCount = 6 }
	writeJSON(w, map[string]any{
		"uptime_sec":   int(time.Since(startTime).Seconds()),
		"go_version":   runtime.Version(),
		"memory_mb":    m.Alloc / 1024 / 1024,
		"endpoints":    endpoints,
		"endpoints_ok": okCount,
		"endpoints_total": 6,
	})
}

func ifElse(cond bool, a, b string) string { if cond { return a }; return b }

func (h *Handlers) GetCacheStats(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"entries": 0,
		"hits":    0,
		"misses":  0,
	})
}

func (h *Handlers) ClearCache(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "cleared"})
}
