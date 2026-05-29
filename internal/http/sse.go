package http

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// SSEEvent is a refresh notification for frontend EventSource consumers.
type SSEEvent struct {
	Entity string `json:"entity"`
	ID     int    `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
}

// SSEHub manages SSE client connections. Nil Broadcast is safe (no-op).
type SSEHub struct {
	mu      sync.RWMutex
	clients map[chan SSEEvent]struct{}
}

func NewSSEHub() *SSEHub {
	return &SSEHub{clients: make(map[chan SSEEvent]struct{})}
}

// Broadcast sends event to all connected clients. No-op if hub is nil.
func (h *SSEHub) Broadcast(evt SSEEvent) {
	if h == nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- evt:
		default: // drop if client buffer full
		}
	}
}

func (h *SSEHub) register(ch chan SSEEvent) {
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
}

func (h *SSEHub) unregister(ch chan SSEEvent) {
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
}

// SSEHandler returns an http.HandlerFunc for GET /api/sse.
func SSEHandler(hub *SSEHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		ch := make(chan SSEEvent, 32)
		hub.register(ch)
		defer hub.unregister(ch)

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case evt := <-ch:
				data, _ := json.Marshal(evt)
				if _, err := w.Write([]byte("event: refreshed\ndata: " + string(data) + "\n\n")); err != nil {
					return
				}
				flusher.Flush()
			case <-ticker.C:
				if _, err := w.Write([]byte(": keepalive\n\n")); err != nil {
					return
				}
				flusher.Flush()
			case <-r.Context().Done():
				return
			}
		}
	}
}
