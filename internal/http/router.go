package http

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/arcdent/hltv-mcp/internal/facade"
	"github.com/arcdent/hltv-mcp/internal/http/handlers"
	"github.com/arcdent/hltv-mcp/internal/storage"
)

func NewRouter(f *facade.HltvFacade, frontendFS fs.FS, sseHub *SSEHub, store *storage.Store) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	h := handlers.New(f, store)

	r.Get("/api/sse", SSEHandler(sseHub))

	r.Get("/api/health", h.Health)
	r.Get("/api/status", h.Status)
	r.Get("/api/cache", h.GetCacheStats)
	r.Delete("/api/cache", h.ClearCache)
	r.Get("/api/search", h.Search)
	r.Get("/api/teams/{id}", h.GetTeam)
	r.Get("/api/players/{id}", h.GetPlayer)
	r.Get("/api/matches/today", h.GetTodayMatches)
	r.Get("/api/matches", h.GetUpcomingMatches)
	r.Get("/api/results", h.GetResults)
	r.Get("/api/events", h.GetEvents)
	r.Get("/api/news/realtime", h.GetRealtimeNews)
	r.Get("/api/news", h.GetNewsDigest)
	r.Get("/api/news/article", h.GetNewsArticle)
	r.Get("/api/translate/config", h.GetTranslateConfig)
	r.Put("/api/translate/config", h.PutTranslateConfig)
	r.Post("/api/translate", h.PostTranslate)
		r.Get("/api/nicknames", h.GetNicknames)
		r.Put("/api/nicknames/team", h.PutTeamNickname)
		r.Put("/api/nicknames/player", h.PutPlayerNickname)

	// SPA fallback
	if frontendFS != nil {
		feFS, err := fs.Sub(frontendFS, "dist")
		if err == nil {
			fileServer := http.FileServer(http.FS(feFS))
			r.Get("/*", func(w http.ResponseWriter, req *http.Request) {
				fsPath := strings.TrimPrefix(req.URL.Path, "/")
				if _, err := feFS.Open(fsPath); err != nil {
					req.URL.Path = "/"
				}
				fileServer.ServeHTTP(w, req)
			})
		}
	}

	return r
}
