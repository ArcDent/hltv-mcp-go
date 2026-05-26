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
	writeJSON(w, map[string]any{
		"uptime_sec": int(time.Since(startTime).Seconds()),
		"go_version": runtime.Version(),
		"memory_mb":  m.Alloc / 1024 / 1024,
	})
}

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
