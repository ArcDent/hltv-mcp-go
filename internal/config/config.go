package config

import (
	"os"
	"strconv"
)

// Config holds all application configuration loaded from environment variables
type Config struct {
	MCPServerName    string
	MCPServerVersion string
	HTTPPort         int
	HTTPHost         string

	FirecrawlKey  string
	HTTPTimeoutMs int
	RetryCount    int

	CacheTTLEntity       int
	CacheTTLTeam         int
	CacheTTLPlayer       int
	CacheTTLResults      int
	CacheTTLMatches      int
	CacheTTLNews         int
	CacheTTLRealtimeNews   int
	CacheTTLPlayerDetail   int
	CacheTTLNewsArticle    int
	CacheMaxEntries        int
	CacheStaleWindowSec  int

	DBPath              string
	DBRetentionMatches  int
	DBRetentionNews     int
	DBRetentionRealtime int

	DefaultResultLimit int
}

// LoadConfig reads configuration from environment variables with defaults
func LoadConfig() (*Config, error) {
	return &Config{
		MCPServerName:    envStr("MCP_SERVER_NAME", "hltv-mcp-service"),
		MCPServerVersion: envStr("MCP_SERVER_VERSION", "1.0.0"),
		HTTPPort:         envInt("HTTP_PORT", 8082),
		HTTPHost:         envStr("HTTP_HOST", "0.0.0.0"),

		FirecrawlKey:  envStr("FIRECRAWL_API_KEY", ""),
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
		CacheTTLNewsArticle:    envInt("CACHE_TTL_NEWS_ARTICLE_SEC", 100*365*24*3600), // ~100 years = infinite
		CacheMaxEntries:        envInt("CACHE_MAX_ENTRIES", 500),
		CacheStaleWindowSec:  envInt("CACHE_STALE_WINDOW_SEC", 3600),

		DBPath:              envStr("HLTV_DB_PATH", "data/hltv.db"),
		DBRetentionMatches:  envInt("HLTV_DB_RETENTION_MATCHES", 90),
		DBRetentionNews:     envInt("HLTV_DB_RETENTION_NEWS", 30),
		DBRetentionRealtime: envInt("HLTV_DB_RETENTION_REALTIME_NEWS", 7),

		DefaultResultLimit: envInt("DEFAULT_RESULT_LIMIT", 5),
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
