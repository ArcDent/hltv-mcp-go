package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/arcdent/hltv-mcp/internal/facade"
	"github.com/arcdent/hltv-mcp/internal/types"
)

type Handlers struct {
	f *facade.HltvFacade
}

func New(f *facade.HltvFacade) *Handlers {
	return &Handlers{f: f}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// withTimeout runs fn in a goroutine with a deadline. On timeout, writes a standard timeout error response.
func (h *Handlers) withTimeout(w http.ResponseWriter, r *http.Request, timeout time.Duration, errMsg string, fn func() *types.ToolResponse) {
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()
	ch := make(chan *types.ToolResponse, 1)
	go func() { ch <- fn() }()
	select {
	case resp := <-ch:
		writeJSON(w, resp)
	case <-ctx.Done():
		writeJSON(w, map[string]any{
			"error": map[string]any{"code": "TIMEOUT", "message": errMsg},
			"meta":  map[string]any{"partial": true},
		})
	}
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "ok"})
}
