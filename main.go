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
	"github.com/arcdent/hltv-mcp/internal/facade"
	httppkg "github.com/arcdent/hltv-mcp/internal/http"
	"github.com/arcdent/hltv-mcp/internal/mcp"
	"github.com/arcdent/hltv-mcp/internal/renderer"
	"github.com/arcdent/hltv-mcp/internal/summary"
)

//go:embed dist/*
var embeddedFrontend embed.FS

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("HLTV MCP starting...")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// Chrome detection (spec: warn and degrade to direct if not available)
	chromePath, chromeAvailable := client.CheckChromeAvailable(cfg)
	if !chromeAvailable && cfg.DataSource != config.DataSourceDirect {
		log.Printf("WARNING: Chrome/Chromium not found — degrading to direct HTTP mode only")
	}
	if chromeAvailable {
		log.Printf("Chrome found at: %s", chromePath)
	}

	c := cache.New(cfg.CacheMaxEntries, cfg.CacheStaleWindowSec)
	cli := client.NewHltvClient(cfg, chromeAvailable)
	f := facade.New(cfg, c, cli)
	r := renderer.New(summary.New(cfg.SummaryMode))

	// MCP stdio goroutine
	mcpServer := mcp.CreateServer(cfg, f, r)
	go func() {
		log.Println("MCP stdio server starting")
		if err := mcp.StartStdio(mcpServer); err != nil {
			log.Printf("MCP stdio error: %v", err)
		}
	}()

	// HTTP server goroutine
	var frontendFS fs.FS
	if cfg.HTTPPort > 0 {
		frontendFS = embeddedFrontend
	}
	router := httppkg.NewRouter(cfg, f, frontendFS)
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
