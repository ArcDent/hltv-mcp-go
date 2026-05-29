package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/arcdent/hltv-mcp/internal/cache"
	"github.com/arcdent/hltv-mcp/internal/client"
	"github.com/arcdent/hltv-mcp/internal/config"
	"github.com/arcdent/hltv-mcp/internal/crypto"
	"github.com/arcdent/hltv-mcp/internal/facade"
	httppkg "github.com/arcdent/hltv-mcp/internal/http"
	"github.com/arcdent/hltv-mcp/internal/http/handlers"
	"github.com/arcdent/hltv-mcp/internal/localization"
	"github.com/arcdent/hltv-mcp/internal/mcp"
	"github.com/arcdent/hltv-mcp/internal/renderer"
	"github.com/arcdent/hltv-mcp/internal/storage"
	"github.com/mark3labs/mcp-go/server"
)

//go:embed dist
var embeddedFrontend embed.FS

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("HLTV MCP starting...")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// Initialize encryption key (env → file → auto-generate)
	if err := crypto.InitKey(); err != nil {
		log.Fatalf("crypto init: %v", err)
	}

	// Migrate old config to encrypted data/ directory
	if err := handlers.MigrateConfig(); err != nil {
		log.Printf("config migration note: %v", err)
	}

	// Initialize nickname overrides
	if err := localization.InitOverrides(); err != nil {
		log.Printf("nickname overrides init note: %v", err)
	}

	c := cache.New(cfg.CacheMaxEntries, cfg.CacheStaleWindowSec)
	cli := client.NewHltvClient(cfg)

	// SSE hub for frontend live refresh
	sseHub := httppkg.NewSSEHub()
	notify := func(entity string, id int, name string) {
		sseHub.Broadcast(httppkg.SSEEvent{Entity: entity, ID: id, Name: name})
	}

	// SQLite persistent storage (Tier 2, optional)
	var store *storage.Store
	if cfg.DBPath != "" {
		s, err := storage.Open(cfg.DBPath, cfg.DBRetentionMatches, cfg.DBRetentionNews, cfg.DBRetentionRealtime)
		if err != nil {
			log.Printf("storage: %v (degraded: Cache-only)", err)
		} else {
			store = s
		}
	}

	f := facade.New(cfg, c, cli, store, notify)
	r := renderer.New()

	// MCP stdio goroutine
	mcpServer := mcp.CreateServer(cfg, f, r)
	go func() {
		log.Println("MCP stdio server starting")
		if err := server.ServeStdio(mcpServer); err != nil {
			log.Printf("MCP stdio error: %v", err)
		}
	}()

	// HTTP server goroutine
	var frontendFS fs.FS
	if cfg.HTTPPort > 0 {
		frontendFS = embeddedFrontend
	}
	router := httppkg.NewRouter(f, frontendFS, sseHub)
	httpAddr := fmt.Sprintf("%s:%d", cfg.HTTPHost, cfg.HTTPPort)
	httpServer := &http.Server{Addr: httpAddr, Handler: router}
	go func() {
		log.Printf("HTTP server listening on %s", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	log.Printf("Received %v, shutting down...", sig)
	httpServer.Shutdown(context.Background())
	log.Println("HLTV MCP stopped")
}
