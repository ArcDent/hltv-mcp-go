package config

import (
	"os"
	"strconv"
)

// DataSource controls the scraper data source strategy
type DataSource string

const (
	DataSourceAuto     DataSource = "auto"
	DataSourceDirect   DataSource = "direct"
	DataSourceChromedp DataSource = "chromedp"
)

// SummaryMode controls summary output style
type SummaryMode string

const (
	SummaryTemplate SummaryMode = "template"
	SummaryRaw      SummaryMode = "raw"
)

// Config holds all application configuration loaded from environment variables
type Config struct {
	MCPServerName    string
	MCPServerVersion string
	HTTPPort         int
	HTTPHost         string

	DataSource DataSource
	ChromePath string
	HTTPTimeoutMs int
	RetryCount int

	CacheTTLEntity       int
	CacheTTLTeam         int
	CacheTTLPlayer       int
	CacheTTLResults      int
	CacheTTLMatches      int
	CacheTTLNews         int
	CacheTTLRealtimeNews   int
	CacheTTLPlayerDetail   int
	CacheMaxEntries        int
	CacheStaleWindowSec  int

	DefaultResultLimit int
	SummaryMode        SummaryMode
	Timezone           string
}

// LoadConfig reads configuration from environment variables with defaults
func LoadConfig() (*Config, error) {
	return &Config{
		MCPServerName:    envStr("MCP_SERVER_NAME", "hltv-mcp-service"),
		MCPServerVersion: envStr("MCP_SERVER_VERSION", "1.0.0"),
		HTTPPort:         envInt("HTTP_PORT", 8082),
		HTTPHost:         envStr("HTTP_HOST", "0.0.0.0"),

		DataSource:    DataSource(envStr("HLTV_DATA_SOURCE", "auto")),
		ChromePath:    envStr("HLTV_CHROME_PATH", ""),
		HTTPTimeoutMs: envInt("HLTV_HTTP_TIMEOUT_MS", 8000),
		RetryCount:    envInt("HLTV_RETRY_COUNT", 2),

		CacheTTLEntity:       envInt("CACHE_TTL_ENTITY_SEC", 3600),
		CacheTTLTeam:         envInt("CACHE_TTL_TEAM_SEC", 300),
		CacheTTLPlayer:       envInt("CACHE_TTL_PLAYER_SEC", 300),
		CacheTTLResults:      envInt("CACHE_TTL_RESULTS_SEC", 120),
		CacheTTLMatches:      envInt("CACHE_TTL_MATCHES_SEC", 60),
		CacheTTLNews:         envInt("CACHE_TTL_NEWS_SEC", 180),
		CacheTTLRealtimeNews:   envInt("CACHE_TTL_REALTIME_NEWS_SEC", 60),
		CacheTTLPlayerDetail:   envInt("CACHE_TTL_PLAYER_DETAIL_SEC", 604800),
		CacheMaxEntries:        envInt("CACHE_MAX_ENTRIES", 500),
		CacheStaleWindowSec:  envInt("CACHE_STALE_WINDOW_SEC", 3600),

		DefaultResultLimit: envInt("DEFAULT_RESULT_LIMIT", 5),
		SummaryMode:        SummaryMode(envStr("SUMMARY_MODE", "template")),
		Timezone:           "Asia/Shanghai",
	}, nil
}

func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
