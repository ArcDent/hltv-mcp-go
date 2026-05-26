package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/arcdent/hltv-mcp/internal/config"
	"github.com/arcdent/hltv-mcp/internal/facade"
)

type Handlers struct {
	cfg *config.Config
	f   *facade.HltvFacade
}

func New(cfg *config.Config, f *facade.HltvFacade) *Handlers {
	return &Handlers{cfg: cfg, f: f}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "ok"})
}
