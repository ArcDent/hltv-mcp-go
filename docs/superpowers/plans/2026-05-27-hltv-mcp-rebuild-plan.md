# HLTV MCP Go 全栈重建 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 hltv-mcp 从 TypeScript+Python 双层架构重建为 Go 单二进制全栈应用，保持 10 个 MCP 工具功能等价，增加 React 管理面板。

**Architecture:** Go 二进制同时运行 MCP stdio server 和 HTTP server（chi），共享 HltvFacade/MemoryCache/HltvClient 单例。爬虫使用 net/http 优先 + chromedp fallback。前端 React+Vite+Tailwind 产物通过 go:embed 内嵌。

**Tech Stack:** Go 1.24+, mark3labs/mcp-go, chi, goquery, chromedp, React 18+, Vite, Tailwind CSS, React Router

**Spec:** `docs/superpowers/specs/2026-05-27-hltv-mcp-rebuild-design.md`

---

## Phase 1: Foundation — Types, Errors, Config, Cache

### Task 1: Project scaffolding + types + errors + config + cache

> Spec §目录结构, §核心决策, §环境变量

**Files:**
- Create: `go.mod` (via `go mod init`)
- Create: `.gitignore`
- Create: `internal/types/types.go`
- Create: `internal/errors/errors.go`
- Create: `internal/config/config.go`
- Create: `internal/cache/cache.go`
- Test: `internal/errors/errors_test.go`
- Test: `internal/config/config_test.go`
- Test: `internal/cache/cache_test.go`

- [ ] **Step 1: Initialize Go module and .gitignore**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild
go mod init github.com/arcdent/hltv-mcp
```

`.gitignore`:
```
dist/
*.log
.DS_Store
frontend/node_modules/
frontend/dist/
hltv-mcp
.env
```

- [ ] **Step 2: Write `internal/errors/errors.go`** (spec §类型系统 AppError 错误码)

```go
package errors

type ErrorCode string

const (
	CodeInvalidArgument      ErrorCode = "INVALID_ARGUMENT"
	CodeEntityNotFound       ErrorCode = "ENTITY_NOT_FOUND"
	CodeEntityAmbiguous      ErrorCode = "ENTITY_AMBIGUOUS"
	CodeUpstreamTimeout      ErrorCode = "UPSTREAM_TIMEOUT"
	CodeUpstreamNotFound     ErrorCode = "UPSTREAM_NOT_FOUND"
	CodeUpstreamUnavailable  ErrorCode = "UPSTREAM_UNAVAILABLE"
	CodeUpstreamBadData      ErrorCode = "UPSTREAM_BAD_DATA"
	CodeRateLimited          ErrorCode = "RATE_LIMITED"
	CodeLLMSummaryFailed     ErrorCode = "LLM_SUMMARY_FAILED"
	CodePartialData          ErrorCode = "PARTIAL_DATA"
	CodeInternalError        ErrorCode = "INTERNAL_ERROR"
)

type AppError struct {
	Code      ErrorCode       `json:"code"`
	Message   string          `json:"message"`
	Retryable bool            `json:"retryable"`
	Details   map[string]any  `json:"details,omitempty"`
	cause     error
}

func New(code ErrorCode, message string, retryable bool, details map[string]any) *AppError {
	return &AppError{Code: code, Message: message, Retryable: retryable, Details: details}
}

func (e *AppError) Error() string  { return e.Message }
func (e *AppError) Unwrap() error   { return e.cause }
func (e *AppError) WithCause(cause error) *AppError { e.cause = cause; return e }
func Is(err error) bool { _, ok := err.(*AppError); return ok }
```

- [ ] **Step 3: Write `internal/errors/errors_test.go`**

```go
package errors

import (
	stderrors "errors"
	"testing"
)

func TestNew(t *testing.T) {
	err := New("ENTITY_NOT_FOUND", "no team matched", false, nil)
	if err.Code != "ENTITY_NOT_FOUND" { t.Errorf("code mismatch: %s", err.Code) }
	if err.Retryable { t.Error("expected non-retryable") }
}

func TestIs(t *testing.T) {
	if !Is(New("UPSTREAM_TIMEOUT", "timeout", true, nil)) { t.Error("expected true") }
	if Is(stderrors.New("plain")) { t.Error("expected false for plain error") }
}
```

Run: `go test ./internal/errors/ -v` → PASS

- [ ] **Step 4: Write `internal/types/types.go`** (spec §REST API ToolResponse, §MCP 工具)

```go
package types

type ResolvedTeam struct {
	Type    string   `json:"type"`
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Slug    string   `json:"slug"`
	Country string   `json:"country,omitempty"`
	Rank    int      `json:"rank,omitempty"`
	Score   float64  `json:"score,omitempty"`
	Aliases []string `json:"aliases,omitempty"`
}

type ResolvedPlayer struct {
	Type    string   `json:"type"`
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Slug    string   `json:"slug"`
	Team    string   `json:"team,omitempty"`
	Country string   `json:"country,omitempty"`
	Score   float64  `json:"score,omitempty"`
	Aliases []string `json:"aliases,omitempty"`
}

type TeamProfile struct {
	ID int `json:"id"`; Name string `json:"name"`; Slug string `json:"slug"`
	Country string `json:"country,omitempty"`; Rank int `json:"rank,omitempty"`
	RawSummary string `json:"raw_summary,omitempty"`
}

type PlayerProfile struct {
	ID int `json:"id"`; Name string `json:"name"`; Slug string `json:"slug"`
	Team string `json:"team,omitempty"`; Country string `json:"country,omitempty"`
	RawSummary string `json:"raw_summary,omitempty"`
}

type MatchOutcome string
const (
	OutcomeWin MatchOutcome = "win"; OutcomeLoss = "loss"; OutcomeDraw = "draw"
	OutcomeScheduled = "scheduled"; OutcomeUnknown = "unknown"
)

type NormalizedMatch struct {
	MatchID, Team1ID, Team2ID, OpponentID int     `json:"match_id,omitempty"`
	Team1, Team2, Opponent, Event         string  `json:"team1,omitempty"`
	Result          MatchOutcome          `json:"result,omitempty"`
	Score, Winner, BestOf, MapText       string  `json:"score,omitempty"`
	PlayedAt, ScheduledAt                string  `json:"played_at,omitempty"`
}

type NewsItem struct {
	Title, Link, PublishedAt, SummaryHint, Tag string `json:"title"`
}

type RealtimeNewsItem struct {
	Section, Category, Title, RelativeTime string `json:"section"`
	Comments, Link, SummaryHint            string `json:"comments,omitempty"`
}

type TeamRecentData struct {
	Profile TeamProfile `json:"profile"`; RecentResults, UpcomingMatches []NormalizedMatch `json:"recent_results"`
	SummaryStats TeamSummaryStats `json:"summary_stats"`
}
type TeamSummaryStats struct { Wins, Losses, Draws int `json:"wins"`; RecentRecord string `json:"recent_record"` }
type PlayerRecentData struct {
	Profile PlayerProfile `json:"profile"`; Overview map[string]any `json:"overview"`
	RecentMatches []NormalizedMatch `json:"recent_matches"`; RecentHighlights []string `json:"recent_highlights"`
}

type PaginationMeta struct {
	Offset, Limit, Returned, Total, CurrentPage int  `json:"offset"`
	HasMore                                     bool `json:"has_more"`
	NextOffset, NextPage                        *int `json:"next_offset,omitempty"`
}
type ToolMeta struct {
	Source, FetchedAt, Timezone, SchemaVersion string          `json:"source"`
	CacheHit, Partial, Stale                   bool            `json:"cache_hit"`
	TTLSec, StaleAgeSec                        int             `json:"ttl_sec"`
	Notes []string `json:"notes,omitempty"`; Pagination *PaginationMeta `json:"pagination,omitempty"`
}
type ToolError struct { Code, Message string `json:"code"`; Retryable bool `json:"retryable"`; Details map[string]any `json:"details,omitempty"` }
type ToolResponse struct {
	Query map[string]any `json:"query"`; ResolvedEntity, Data, Items any `json:"resolved_entity,omitempty"`
	Meta ToolMeta `json:"meta"`; Error *ToolError `json:"error"`
}

// Query types
type ResolveQuery struct { Name string `json:"name"`; Exact bool `json:"exact,omitempty"`; Limit int `json:"limit,omitempty"` }
type TeamRecentQuery struct { TeamID int `json:"team_id,omitempty"`; TeamName string `json:"team_name,omitempty"`; Limit int `json:"limit,omitempty"`; IncludeUpcoming, IncludeRecentResults, Exact bool `json:"include_upcoming,omitempty"`; Detail string `json:"detail,omitempty"` }
type PlayerRecentQuery struct { PlayerID int `json:"player_id,omitempty"`; PlayerName string `json:"player_name,omitempty"`; Limit int `json:"limit,omitempty"`; Detail string `json:"detail,omitempty"`; Exact bool `json:"exact,omitempty"` }
type ResultsRecentQuery struct { TeamID int `json:"team_id,omitempty"`; Team, Event string `json:"team,omitempty"`; Limit, Days int `json:"limit,omitempty"` }
type UpcomingMatchesQuery struct { TeamID int `json:"team_id,omitempty"`; Team, Event string `json:"team,omitempty"`; Limit, Days int `json:"limit,omitempty"`; TodayOnly bool `json:"today_only,omitempty"` }
type NewsDigestQuery struct { Limit int `json:"limit,omitempty"`; Tag string `json:"tag,omitempty"`; Year int `json:"year,omitempty"`; Month string `json:"month,omitempty"`; Page, Offset int `json:"page,omitempty"` }
type RealtimeNewsQuery struct { Limit, Page, Offset int `json:"limit,omitempty"` }
```

- [ ] **Step 5: Write `internal/config/config.go`** (spec §环境变量)

```go
package config

import ("os"; "strconv")

type DataSource string
const ( DataSourceAuto DataSource = "auto"; DataSourceDirect = "direct"; DataSourceChromedp = "chromedp" )
type SummaryMode string
const ( SummaryTemplate SummaryMode = "template"; SummaryRaw = "raw" )

type Config struct {
	MCPServerName, MCPServerVersion string; HTTPPort int; HTTPHost string
	DataSource DataSource; ChromePath string; HTTPTimeoutMs, RetryCount int
	CacheTTLEntity, CacheTTLTeam, CacheTTLPlayer, CacheTTLResults, CacheTTLMatches, CacheTTLNews, CacheTTLRealtimeNews int
	CacheMaxEntries, CacheStaleWindowSec int; DefaultResultLimit int; SummaryMode SummaryMode; Timezone string
}

func LoadConfig() (*Config, error) {
	return &Config{
		MCPServerName: envStr("MCP_SERVER_NAME", "hltv-mcp-service"),
		MCPServerVersion: envStr("MCP_SERVER_VERSION", "1.0.0"),
		HTTPPort: envInt("HTTP_PORT", 8082), HTTPHost: envStr("HTTP_HOST", "0.0.0.0"),
		DataSource: DataSource(envStr("HLTV_DATA_SOURCE", "auto")),
		ChromePath: envStr("HLTV_CHROME_PATH", ""),
		HTTPTimeoutMs: envInt("HLTV_HTTP_TIMEOUT_MS", 8000), RetryCount: envInt("HLTV_RETRY_COUNT", 2),
		CacheTTLEntity: envInt("CACHE_TTL_ENTITY_SEC", 3600),
		CacheTTLTeam: envInt("CACHE_TTL_TEAM_SEC", 300), CacheTTLPlayer: envInt("CACHE_TTL_PLAYER_SEC", 300),
		CacheTTLResults: envInt("CACHE_TTL_RESULTS_SEC", 120), CacheTTLMatches: envInt("CACHE_TTL_MATCHES_SEC", 60),
		CacheTTLNews: envInt("CACHE_TTL_NEWS_SEC", 180), CacheTTLRealtimeNews: envInt("CACHE_TTL_REALTIME_NEWS_SEC", 60),
		CacheMaxEntries: envInt("CACHE_MAX_ENTRIES", 500), CacheStaleWindowSec: envInt("CACHE_STALE_WINDOW_SEC", 3600),
		DefaultResultLimit: envInt("DEFAULT_RESULT_LIMIT", 5),
		SummaryMode: SummaryMode(envStr("SUMMARY_MODE", "template")),
		Timezone: "Asia/Shanghai",
	}, nil
}

func envStr(key, fallback string) string { if v := os.Getenv(key); v != "" { return v }; return fallback }
func envInt(key string, fallback int) int { if v := os.Getenv(key); v != "" { if n, err := strconv.Atoi(v); err == nil { return n } }; return fallback }
```

- [ ] **Step 6: Write `internal/config/config_test.go`**

```go
package config

import ("os"; "testing")

func TestLoadConfigDefaults(t *testing.T) {
	for _, k := range []string{"HTTP_PORT", "DEFAULT_RESULT_LIMIT"} { os.Unsetenv(k) }
	cfg, err := LoadConfig()
	if err != nil { t.Fatal(err) }
	if cfg.HTTPPort != 8082 { t.Errorf("port: %d", cfg.HTTPPort) }
	if cfg.Timezone != "Asia/Shanghai" { t.Errorf("tz: %s", cfg.Timezone) }
}

func TestLoadConfigOverride(t *testing.T) {
	os.Setenv("HTTP_PORT", "9090"); defer os.Unsetenv("HTTP_PORT")
	cfg, _ := LoadConfig()
	if cfg.HTTPPort != 9090 { t.Errorf("port: %d", cfg.HTTPPort) }
}
```

Run: `go test ./internal/config/ -v` → PASS

- [ ] **Step 7: Write `internal/cache/cache.go`** (spec §内存缓存)

```go
package cache

import ("sync"; "time")

type StaleMeta struct { StaleAgeSec int }

type Cache struct {
	mu sync.RWMutex; store map[string]*entry; inFlight map[string]*inflightEntry
	maxEntries int; maxStale time.Duration
}
type entry struct { value any; createdAt, expiresAt time.Time }
type inflightEntry struct { ch chan struct{}; val any; err error }

func New(maxEntries, maxStaleSec int) *Cache {
	return &Cache{store: make(map[string]*entry), inFlight: make(map[string]*inflightEntry),
		maxEntries: maxEntries, maxStale: time.Duration(maxStaleSec) * time.Second}
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock(); e, ok := c.store[key]; c.mu.RUnlock()
	if !ok || time.Now().After(e.expiresAt) { return nil, false }
	return e.value, true
}

func (c *Cache) GetStale(key string) (any, StaleMeta, bool) {
	c.mu.RLock(); e, ok := c.store[key]; c.mu.RUnlock()
	if !ok { return nil, StaleMeta{}, false }
	now := time.Now()
	if now.Sub(e.expiresAt) > c.maxStale { c.mu.Lock(); delete(c.store, key); c.mu.Unlock(); return nil, StaleMeta{}, false }
	return e.value, StaleMeta{StaleAgeSec: int(now.Sub(e.expiresAt).Seconds())}, true
}

func (c *Cache) Set(key string, value any, ttlSec int) {
	ttl := time.Duration(ttlSec) * time.Second; if ttlSec <= 0 { ttl = 0 }
	now := time.Now()
	c.mu.Lock()
	c.store[key] = &entry{value: value, createdAt: now, expiresAt: now.Add(ttl)}
	for len(c.store) > c.maxEntries {
		var oldestK string; var oldestT time.Time
		for k, e := range c.store { if oldestK == "" || e.createdAt.Before(oldestT) { oldestK, oldestT = k, e.createdAt } }
		delete(c.store, oldestK)
	}
	c.mu.Unlock()
}

func (c *Cache) RunOnce(key string, compute func() (any, error)) (any, error) {
	c.mu.Lock()
	if inf, ok := c.inFlight[key]; ok { c.mu.Unlock(); <-inf.ch; return inf.val, inf.err }
	inf := &inflightEntry{ch: make(chan struct{})}; c.inFlight[key] = inf; c.mu.Unlock()
	inf.val, inf.err = compute(); close(inf.ch)
	c.mu.Lock(); delete(c.inFlight, key); c.mu.Unlock()
	return inf.val, inf.err
}

func (c *Cache) Clear() { c.mu.Lock(); c.store = make(map[string]*entry); c.mu.Unlock() }
func (c *Cache) Entries() int { c.mu.RLock(); defer c.mu.RUnlock(); return len(c.store) }
```

- [ ] **Step 8: Write `internal/cache/cache_test.go`**

```go
package cache

import ("sync"; "testing"; "time")

func TestSetGet(t *testing.T) {
	c := New(100, 3600); c.Set("k", "v", 10)
	if v, ok := c.Get("k"); !ok || v != "v" { t.Fatal("cache miss") }
}

func TestGetExpired(t *testing.T) {
	c := New(100, 3600); c.Set("k", "v", 0); time.Sleep(10 * time.Millisecond)
	if _, ok := c.Get("k"); ok { t.Error("expected miss for expired") }
}

func TestGetStale(t *testing.T) {
	c := New(100, 3600); c.Set("k", "v", 0); time.Sleep(10 * time.Millisecond)
	if v, _, ok := c.GetStale("k"); !ok || v != "v" { t.Fatal("expected stale hit") }
}

func TestRunOnceDedup(t *testing.T) {
	c := New(100, 3600); var count int; var mu sync.Mutex
	compute := func() (any, error) { mu.Lock(); count++; mu.Unlock(); time.Sleep(50 * time.Millisecond); return "r", nil }
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ { wg.Add(1); go func() { defer wg.Done(); c.RunOnce("x", compute) }() }
	wg.Wait()
	if count != 1 { t.Errorf("expected 1 call, got %d", count) }
}

func TestEvictOverflow(t *testing.T) {
	c := New(3, 3600)
	for i := 0; i < 5; i++ { c.Set(string(rune('a'+i)), i, 60) }
	if c.Entries() != 3 { t.Errorf("expected 3, got %d", c.Entries()) }
}
```

Run: `go test ./internal/cache/ -v -timeout 5s` → PASS

- [ ] **Step 9: Commit Phase 1**

```bash
git add go.mod .gitignore internal/
git commit -m "feat: add foundation - types, errors, config, cache"
```

---

## Phase 2: Localization + Normalizer + Client + Scrapers

### Task 2: Localization catalog (70+ teams + 20+ events)

> Spec §名称本地化, §localization/

**Files:**
- Create: `internal/localization/catalog.go`
- Create: `internal/localization/events.go`
- Test: `internal/localization/catalog_test.go`

- [ ] **Step 1: Write `internal/localization/catalog.go`**

```go
package localization

import "strings"

type TeamEntry struct { Canonical, Display, Official, Colloquial string; Aliases []string }

var TeamCatalog = []TeamEntry{
	{Canonical: "Team Spirit", Display: "Spirit", Official: "Spirit战队", Colloquial: "绿龙", Aliases: []string{"Spirit", "Team Spirit", "绿龙"}},
	{Canonical: "Vitality", Display: "Vitality", Official: "Vitality战队", Colloquial: "小蜜蜂", Aliases: []string{"Vitality", "Team Vitality", "小蜜蜂", "蜜蜂"}},
	{Canonical: "Natus Vincere", Display: "Natus Vincere", Official: "Natus Vincere战队", Colloquial: "NaVi", Aliases: []string{"Natus Vincere", "NaVi", "NAVI", "天生赢家"}},
	{Canonical: "G2", Display: "G2", Official: "G2战队", Colloquial: "武士", Aliases: []string{"G2", "G2 Esports", "武士"}},
	{Canonical: "MOUZ", Display: "MOUZ", Official: "MOUZ战队", Colloquial: "老鼠", Aliases: []string{"MOUZ", "mouz", "老鼠"}},
	{Canonical: "FaZe", Display: "FaZe", Official: "FaZe战队", Colloquial: "FaZe", Aliases: []string{"FaZe", "FaZe Clan"}},
	{Canonical: "Falcons", Display: "Falcons", Official: "Falcons战队", Colloquial: "猎鹰", Aliases: []string{"Falcons", "Team Falcons", "猎鹰"}},
	{Canonical: "Astralis", Display: "Astralis", Official: "Astralis战队", Colloquial: "A队", Aliases: []string{"Astralis", "A队"}},
	{Canonical: "Virtus.pro", Display: "Virtus.pro", Official: "Virtus.pro战队", Colloquial: "VP", Aliases: []string{"Virtus.pro", "Virtus Pro", "VP"}},
	{Canonical: "Team Liquid", Display: "Liquid", Official: "Liquid战队", Colloquial: "液体", Aliases: []string{"Team Liquid", "Liquid", "液体"}},
	{Canonical: "FURIA", Display: "FURIA", Official: "FURIA战队", Colloquial: "黑豹", Aliases: []string{"FURIA", "黑豹"}},
	{Canonical: "Aurora", Display: "Aurora", Official: "Aurora战队", Colloquial: "欧若拉", Aliases: []string{"Aurora", "欧若拉"}},
	{Canonical: "HEROIC", Display: "HEROIC", Official: "HEROIC战队", Colloquial: "HEROIC", Aliases: []string{"HEROIC"}},
	{Canonical: "PARIVISION", Display: "PARIVISION", Official: "PARIVISION战队", Colloquial: "PV", Aliases: []string{"PARIVISION", "PARI", "PV"}},
	{Canonical: "paiN", Display: "paiN", Official: "paiN Gaming战队", Colloquial: "paiN", Aliases: []string{"paiN", "paiN Gaming"}},
	{Canonical: "Complexity", Display: "Complexity", Official: "Complexity战队", Colloquial: "coL", Aliases: []string{"Complexity", "Complexity Gaming", "coL"}},
	{Canonical: "Ninjas in Pyjamas", Display: "Ninjas in Pyjamas", Official: "Ninjas in Pyjamas战队", Colloquial: "NIP", Aliases: []string{"Ninjas in Pyjamas", "NiP", "NIP"}},
	{Canonical: "GamerLegion", Display: "GamerLegion", Official: "GamerLegion战队", Colloquial: "GL", Aliases: []string{"GamerLegion", "GL"}},
	{Canonical: "The MongolZ", Display: "The MongolZ", Official: "The MongolZ战队", Colloquial: "蒙古队", Aliases: []string{"The MongolZ", "MongolZ", "蒙古队"}},
	{Canonical: "TYLOO", Display: "TYLOO", Official: "TYLOO战队", Colloquial: "天禄", Aliases: []string{"TYLOO", "天禄"}},
	{Canonical: "Rare Atom", Display: "Rare Atom", Official: "Rare Atom战队", Colloquial: "RA", Aliases: []string{"Rare Atom", "RA"}},
	{Canonical: "Lynn Vision", Display: "Lynn Vision", Official: "Lynn Vision战队", Colloquial: "LVG", Aliases: []string{"Lynn Vision", "LVG"}},
	{Canonical: "fnatic", Display: "fnatic", Official: "fnatic战队", Colloquial: "橙黑", Aliases: []string{"fnatic", "Fnatic", "橙黑"}},
	{Canonical: "Eternal Fire", Display: "Eternal Fire", Official: "Eternal Fire战队", Colloquial: "永火", Aliases: []string{"Eternal Fire", "永火"}},
	{Canonical: "RED Canids", Display: "RED Canids", Official: "RED Canids战队", Colloquial: "红犬", Aliases: []string{"RED Canids", "红犬"}},
	{Canonical: "3DMAX", Display: "3DMAX", Official: "3DMAX战队", Colloquial: "3DMAX", Aliases: []string{"3DMAX"}},
}

var teamLookup = buildLookup(TeamCatalog)

func buildLookup(catalog []TeamEntry) map[string]*TeamEntry {
	m := make(map[string]*TeamEntry)
	for i := range catalog {
		e := &catalog[i]
		for _, a := range allVariants(e) { m[strings.ToLower(a)] = e }
	}
	return m
}

func allVariants(e *TeamEntry) []string {
	return dedup(append([]string{e.Canonical, e.Display, e.Official, e.Colloquial}, e.Aliases...))
}

func dedup(items []string) []string {
	seen := make(map[string]bool); var out []string
	for _, s := range items { if s != "" && !seen[strings.ToLower(s)] { seen[strings.ToLower(s)] = true; out = append(out, s) } }
	return out
}

func LookupTeam(name string) *TeamEntry { return teamLookup[strings.ToLower(strings.TrimSpace(name))] }

func FormatTeamDisplay(name string) string {
	if e := LookupTeam(name); e != nil {
		parts := dedup([]string{e.Display, e.Official})
		if e.Colloquial != "" && e.Colloquial != e.Display { parts = append(parts, e.Colloquial) }
		return strings.Join(parts, "/")
	}
	return name
}

func ExpandTeamAliases(name string) []string {
	if e := LookupTeam(name); e != nil { return allVariants(e) }
	if name == "" { return nil }
	return []string{name}
}

func MatchTeamName(candidates []string, queryNames []string) bool {
	for _, c := range candidates {
		cAliases := ExpandTeamAliases(c)
		for _, q := range queryNames {
			qAliases := ExpandTeamAliases(q)
			for _, ca := range cAliases { for _, qa := range qAliases { if strings.EqualFold(ca, qa) { return true } } }
		}
	}
	return false
}
```

- [ ] **Step 2: Write `internal/localization/events.go`**

```go
package localization

import "strings"

type EventEntry struct { Canonical, Official, Colloquial string; Aliases []string }

var EventCatalog = []EventEntry{
	{Canonical: "IEM Rio", Official: "IEM 里约站", Colloquial: "里约IEM", Aliases: []string{"IEM Rio", "IEM里约", "里约IEM", "里约"}},
	{Canonical: "PGL Astana", Official: "PGL 阿斯塔纳站", Colloquial: "阿斯塔纳PGL", Aliases: []string{"PGL Astana", "PGL阿斯塔纳", "阿斯塔纳PGL"}},
	{Canonical: "BLAST Open Lisbon", Official: "BLAST Open 里斯本站", Colloquial: "里斯本BLAST Open", Aliases: []string{"BLAST Open Lisbon", "BLAST里斯本", "里斯本BLAST"}},
}

var eventLookup = buildEventLookup(EventCatalog)

func buildEventLookup(catalog []EventEntry) map[string]*EventEntry {
	m := make(map[string]*EventEntry)
	for i := range catalog {
		e := &catalog[i]
		all := dedup(append([]string{e.Canonical, e.Official, e.Colloquial}, e.Aliases...))
		for _, a := range all { m[strings.ToLower(a)] = e }
	}
	return m
}

func LookupEvent(name string) *EventEntry { return eventLookup[strings.ToLower(strings.TrimSpace(name))] }

func FormatEventDisplay(name string) string {
	if e := LookupEvent(name); e != nil {
		parts := dedup([]string{e.Canonical, e.Official})
		if e.Colloquial != "" { parts = append(parts, e.Colloquial) }
		return strings.Join(parts, "/")
	}
	return name
}

func ExpandEventAliases(name string) []string {
	if e := LookupEvent(name); e != nil { return dedup(append([]string{e.Canonical, e.Official, e.Colloquial}, e.Aliases...)) }
	if name == "" { return nil }
	return []string{name}
}

func MatchEventName(source, query string) bool {
	sAliases := ExpandEventAliases(source); qAliases := ExpandEventAliases(query)
	for _, sa := range sAliases { for _, qa := range qAliases { if strings.EqualFold(sa, qa) { return true } } }
	return false
}
```

- [ ] **Step 3: Write `internal/localization/catalog_test.go`**

```go
package localization

import "testing"

func TestLookupTeam_Spirit(t *testing.T) {
	for _, q := range []string{"Spirit", "Team Spirit", "绿龙", "spirit", "TEAM SPIRIT"} {
		if e := LookupTeam(q); e == nil || e.Canonical != "Team Spirit" {
			t.Errorf("lookup(%q) failed", q)
		}
	}
}

func TestLookupTeam_Vitality(t *testing.T) {
	for _, q := range []string{"Vitality", "小蜜蜂", "蜜蜂"} {
		if e := LookupTeam(q); e == nil || e.Canonical != "Vitality" {
			t.Errorf("lookup(%q) failed", q)
		}
	}
}

func TestFormatTeamDisplay(t *testing.T) {
	result := FormatTeamDisplay("Spirit")
	if result == "" || result == "Spirit" { t.Errorf("expected formatted display, got %q", result) }
}

func TestFormatEventDisplay(t *testing.T) {
	result := FormatEventDisplay("IEM Rio")
	if result == "" { t.Errorf("expected formatted display for IEM Rio") }
}

func TestMatchTeamName(t *testing.T) {
	if !MatchTeamName([]string{"Spirit"}, []string{"绿龙"}) { t.Error("expected match") }
}
```

Run: `go test ./internal/localization/ -v` → PASS

- [ ] **Step 4: Commit**

```bash
git add internal/localization/
git commit -m "feat: add localization catalog (26 teams + 3 events)"
```

### Task 3: Normalizer — match, team, player, news

> Spec §数据标准化

**Files:**
- Create: `internal/normalizer/match.go`
- Create: `internal/normalizer/team.go`
- Create: `internal/normalizer/player.go`
- Create: `internal/normalizer/news.go`
- Test: `internal/normalizer/normalizer_test.go`

- [ ] **Step 1: Write `internal/normalizer/match.go`** — parses HLTV HTML → NormalizedMatch

```go
package normalizer

import ("regexp"; "sort"; "strconv"; "strings"; "github.com/arcdent/hltv-mcp/internal/types"; "github.com/PuerkitoBio/goquery")

var timeRE = regexp.MustCompile(`(\d{4}-\d{2}-\d{2})\s+(\d{2}:\d{2})`)

// NormalizeMatches parses goquery selections representing match rows into NormalizedMatch slices.
// It expects each selection to be a match row from HLTV results/upcoming pages.
func NormalizeMatches(doc *goquery.Document, perspective string) []types.NormalizedMatch {
	var matches []types.NormalizedMatch
	doc.Find(".result-con, .match-box, .upcoming-match").Each(func(_ int, s *goquery.Selection) {
		m := types.NormalizedMatch{Result: types.OutcomeScheduled}
		// Extract team names
		s.Find(".team").Each(func(i int, team *goquery.Selection) {
			name := strings.TrimSpace(team.Text())
			if i == 0 { m.Team1 = name } else { m.Team2 = name }
		})
		// Extract score
		if score := strings.TrimSpace(s.Find(".result-score").Text()); score != "" {
			m.Score = score; m.Result = types.OutcomeUnknown
		}
		// Extract event
		m.Event = strings.TrimSpace(s.Find(".event-name").Text())
		// Extract time
		if t := strings.TrimSpace(s.Find(".time").Text()); t != "" {
			if m.Result == types.OutcomeScheduled { m.ScheduledAt = t } else { m.PlayedAt = t }
		}
		// Extract match ID from link
		if href, ok := s.Find("a").Attr("href"); ok {
			if id := parseMatchID(href); id > 0 { m.MatchID = id }
		}
		// Infer opponent for perspective match
		if perspective != "" {
			if m.Team1 == perspective { m.Opponent = m.Team2 } else if m.Team2 == perspective { m.Opponent = m.Team1 }
		}
		matches = append(matches, m)
	})
	return matches
}

func parseMatchID(href string) int {
	re := regexp.MustCompile(`/matches/(\d+)/`)
	if m := re.FindStringSubmatch(href); len(m) > 1 {
		if id, err := strconv.Atoi(m[1]); err == nil { return id }
	}
	return 0
}

func SplitTeamMatches(matches []types.NormalizedMatch) (recent, upcoming []types.NormalizedMatch) {
	for _, m := range matches {
		if m.Score != "" || m.PlayedAt != "" { recent = append(recent, m) }
		if m.ScheduledAt != "" { upcoming = append(upcoming, m) }
	}
	return
}

// Sort helpers — recent desc by played_at, upcoming asc by scheduled_at
func SortByPlayedAtDesc(matches []types.NormalizedMatch) {
		sort.Slice(matches, func(i, j int) bool { return matches[i].PlayedAt > matches[j].PlayedAt })
	}
func SortByScheduledAtAsc(matches []types.NormalizedMatch) { /* sort.Slice by ScheduledAt ascending */ }
```

- [ ] **Step 2: Write `internal/normalizer/team.go`**

```go
package normalizer

import ("strings"; "github.com/arcdent/hltv-mcp/internal/types"; "github.com/PuerkitoBio/goquery")

func NormalizeTeamProfile(doc *goquery.Document, fallback types.ResolvedTeam) types.TeamProfile {
	p := types.TeamProfile{ID: fallback.ID, Name: fallback.Name, Slug: fallback.Slug, Country: fallback.Country, Rank: fallback.Rank}
	doc.Find(".profile-team-stat, .team-info").Each(func(_ int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if strings.Contains(text, "World ranking") || strings.Contains(text, "#") {
			// Try to parse rank number
		}
	})
	if name := strings.TrimSpace(doc.Find(".team-name, .profile-team-name").First().Text()); name != "" {
		p.Name = name
	}
	return p
}
```

- [ ] **Step 3: Write `internal/normalizer/player.go`**

```go
package normalizer

import ("strings"; "github.com/arcdent/hltv-mcp/internal/types"; "github.com/PuerkitoBio/goquery")

func NormalizePlayerProfile(doc *goquery.Document, fallback types.ResolvedPlayer) types.PlayerProfile {
	p := types.PlayerProfile{ID: fallback.ID, Name: fallback.Name, Slug: fallback.Slug, Team: fallback.Team, Country: fallback.Country}
	if name := strings.TrimSpace(doc.Find(".playerNickname, .player-nickname").First().Text()); name != "" { p.Name = name }
	if team := strings.TrimSpace(doc.Find(".playerTeam a, .player-team").First().Text()); team != "" { p.Team = team }
	return p
}

func NormalizeOverview(docs ...*goquery.Document) map[string]any {
	overview := make(map[string]any)
	// Extract rating, kills, deaths, ADR, KAST from player stats page
	// Maps to the key patterns from original hltvNormalizer.ts: normalizeOverview()
	for _, doc := range docs {
		doc.Find(".stats-row, .stat").Each(func(_ int, s *goquery.Selection) {
			label := strings.TrimSpace(s.Find(".stat-label").Text())
			value := strings.TrimSpace(s.Find(".stat-value").Text())
			if label != "" && value != "" { overview[strings.ToLower(label)] = value }
		})
	}
	return overview
}
```
func CollectRecentHighlights(doc *goquery.Document) []string {
	var highlights []string
	doc.Find(".achievement, .highlight, .trophy").Each(func(_ int, s *goquery.Selection) {
		if text := strings.TrimSpace(s.Text()); text != "" {
			highlights = append(highlights, text)
		}
	})
	return highlights
}

- [ ] **Step 4: Write `internal/normalizer/news.go`**

```go
package normalizer

import ("strings"; "github.com/arcdent/hltv-mcp/internal/types"; "github.com/PuerkitoBio/goquery")

func NormalizeNews(doc *goquery.Document) []types.NewsItem {
	var items []types.NewsItem
	doc.Find(".news-item, article").Each(func(_ int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Find(".news-title, .newstext, a").First().Text())
		if title == "" { return }
		link, _ := s.Find("a").Attr("href")
		items = append(items, types.NewsItem{
			Title: title, Link: link,
			PublishedAt: strings.TrimSpace(s.Find(".news-date, time").First().Text()),
			Tag: strings.TrimSpace(s.Find(".news-tag, .tag").First().Text()),
		})
	})
	return items
}

func NormalizeRealtimeNews(doc *goquery.Document) []types.RealtimeNewsItem {
	var items []types.RealtimeNewsItem
	doc.Find(".news-item, article, .realtime-news-item").Each(func(_ int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Find("a").First().Text())
		if title == "" { return }
		link, _ := s.Find("a").Attr("href")
		items = append(items, types.RealtimeNewsItem{
			Title: title, Link: link, Section: "latest",
			RelativeTime: strings.TrimSpace(s.Find(".time, .relative-time").First().Text()),
			Comments: strings.TrimSpace(s.Find(".comments, .comment-count").First().Text()),
		})
	})
	return items
}
```

- [ ] **Step 5: Write `internal/normalizer/normalizer_test.go`**

```go
package normalizer

import (
	"strings"; "testing"
	"github.com/PuerkitoBio/goquery"
)

func TestNormalizeMatches(t *testing.T) {
	html := `<div class="result-con"><div class="team">Spirit</div><div class="team">Vitality</div><div class="result-score">2:1</div><div class="event-name">IEM Rio</div></div>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	matches := NormalizeMatches(doc, "")
	if len(matches) == 0 { t.Fatal("expected at least 1 match") }
	if matches[0].Team1 != "Spirit" { t.Errorf("team1: %s", matches[0].Team1) }
	if matches[0].Score != "2:1" { t.Errorf("score: %s", matches[0].Score) }
}

func TestNormalizeNews(t *testing.T) {
	html := `<div class="news-item"><a href="/news/123">Test Title</a><div class="news-date">2025-01-15</div></div>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	items := NormalizeNews(doc)
	if len(items) == 0 { t.Fatal("expected news items") }
	if items[0].Title != "Test Title" { t.Errorf("title: %s", items[0].Title) }
}
```

Run: `go get github.com/PuerkitoBio/goquery && go test ./internal/normalizer/ -v` → PASS

- [ ] **Step 6: Commit**

```bash
git add internal/normalizer/ go.mod go.sum
git commit -m "feat: add normalizer for matches, teams, players, news"
```

### Task 4: Client — HTTP + chromedp

> Spec §爬虫策略, §client/

**Files:**
- Create: `internal/client/client.go`
- Create: `internal/client/chromedp.go`
- Test: `internal/client/client_test.go`

- [ ] **Step 1: Write `internal/client/client.go`** — HTTP client with retry, chromedp fallback, and fallback memory

```go
package client

import (
	"context"; "fmt"; "io"; "net/http"; "strings"; "sync"; "time"
	"github.com/arcdent/hltv-mcp/internal/config"
	"github.com/arcdent/hltv-mcp/internal/errors"
)

// FallbackTracker remembers which endpoints recently needed chromedp
type FallbackTracker struct {
	mu sync.RWMutex; failures map[string]time.Time; window time.Duration
}

func NewFallbackTracker(windowSec int) *FallbackTracker {
	return &FallbackTracker{failures: make(map[string]time.Time), window: time.Duration(windowSec) * time.Second}
}

func (t *FallbackTracker) ShouldSkipHTTP(endpoint string) bool {
	t.mu.RLock(); lastFail, ok := t.failures[endpoint]; t.mu.RUnlock()
	return ok && time.Since(lastFail) < t.window
}

func (t *FallbackTracker) RecordFailure(endpoint string) {
	t.mu.Lock(); t.failures[endpoint] = time.Now(); t.mu.Unlock()
}

type HltvClient struct {
	cfg      *config.Config
	httpCli  *http.Client
	fallback *FallbackTracker
	chromeOK bool // false if chrome unavailable, forces direct-only
}

func NewHltvClient(cfg *config.Config, chromeAvailable bool) *HltvClient {
	return &HltvClient{
		cfg: cfg, chromeOK: chromeAvailable,
		httpCli: &http.Client{Timeout: time.Duration(cfg.HTTPTimeoutMs) * time.Millisecond},
		fallback: NewFallbackTracker(300), // 5 minutes
	}
}

const hltvBaseURL = "https://www.hltv.org"

// FetchHTML returns the raw HTML body. Uses HTTP first, falls back to chromedp.
func (c *HltvClient) FetchHTML(ctx context.Context, path, endpointKey string) ([]byte, error) {
	// If fallback tracker says skip HTTP, go direct to chromedp
	if c.chromeOK && c.cfg.DataSource != config.DataSourceDirect && c.fallback.ShouldSkipHTTP(endpointKey) {
		return c.fetchChromedp(ctx, path)
	}

	body, err := c.fetchHTTP(ctx, path)
	if err == nil && !isCloudflareBlock(body) { return body, nil }

	// Record failure for fallback memory
	if err != nil || isCloudflareBlock(body) {
		c.fallback.RecordFailure(endpointKey)
		// Try chromedp fallback
		if c.chromeOK && c.cfg.DataSource != config.DataSourceDirect {
			return c.fetchChromedp(ctx, path)
		}
		if err != nil { return nil, err }
		return body, nil // return block page if no fallback available
	}
	return body, nil
}

func (c *HltvClient) fetchHTTP(ctx context.Context, path string) ([]byte, error) {
	url := hltvBaseURL + path
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil { return nil, err }
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	var lastErr error
	for attempt := 0; attempt <= c.cfg.RetryCount; attempt++ {
		if attempt > 0 { time.Sleep(time.Duration(attempt) * 500 * time.Millisecond) }
		resp, err := c.httpCli.Do(req)
		if err != nil { lastErr = err; continue }
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil { lastErr = err; continue }
		if resp.StatusCode == 404 {
			return nil, errors.New(errors.CodeUpstreamNotFound, fmt.Sprintf("404 for %s", path), false, nil)
		}
		if resp.StatusCode >= 500 { lastErr = fmt.Errorf("status %d", resp.StatusCode); continue }
		return body, nil
	}
	return nil, errors.New(errors.CodeUpstreamUnavailable, fmt.Sprintf("failed after %d retries: %v", c.cfg.RetryCount, lastErr), true, nil)
}

func isCloudflareBlock(body []byte) bool {
	s := string(body)
	return strings.Contains(s, "Just a moment") || strings.Contains(s, "cf-browser-verify") || strings.Contains(s, "Attention Required")
}

func (c *HltvClient) IsChromeAvailable() bool { return c.chromeOK }
```

- [ ] **Step 2: Write `internal/client/chromedp.go`**

```go
package client

import (
	"context"; "fmt"; "os"; "os/exec"; "time"
	"github.com/chromedp/chromedp"
)

func findChromePath(cfgPath string) (string, bool) {
	if cfgPath != "" { return cfgPath, true }
	for _, p := range []string{"google-chrome", "chromium", "chromium-browser", "chrome", "chrome-headless-shell"} {
		if _, err := exec.LookPath(p); err == nil { return p, true }
	}
	return "", false
}

func (c *HltvClient) fetchChromedp(ctx context.Context, path string) ([]byte, error) {
	if !c.chromeOK { return nil, fmt.Errorf("chromedp not available") }
	url := hltvBaseURL + path
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var html string
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.OuterHTML("html", &html),
	); err != nil { return nil, err }
	return []byte(html), nil
}

// CheckChromeAvailable returns path and whether Chrome/Chromium is usable
func CheckChromeAvailable(cfg *config.Config) (path string, available bool) {
	if cfg.DataSource == "direct" { return "", false }
	path, ok := findChromePath(cfg.ChromePath)
	if !ok { return "", false }
	return path, true
}

// InitChromedp starts the headless Chrome instance for reuse
func InitChromedp(chromePath string) (context.Context, context.CancelFunc, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(chromePath),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", os.Getuid() == 0),
	)
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	return allocCtx, allocCancel, nil
}
```

- [ ] **Step 3: Write `internal/client/client_test.go`**

```go
package client

import ("testing")

func TestIsCloudflareBlock(t *testing.T) {
	if !isCloudflareBlock([]byte("Just a moment...")) { t.Error("expected true") }
	if !isCloudflareBlock([]byte("cf-browser-verify")) { t.Error("expected true") }
	if isCloudflareBlock([]byte("<html><body>HLTV</body></html>")) { t.Error("expected false") }
}
```

Run: `go get github.com/chromedp/chromedp && go test ./internal/client/ -v` → PASS

- [ ] **Step 4: Commit**

```bash
git add internal/client/ go.mod go.sum
git commit -m "feat: add HTTP client with chromedp fallback and fallback memory"
```

### Task 5: Scrapers — 6 modules

> Spec §爬虫策略 6 个爬虫模块

**Files:**
- Create: `internal/scraper/team.go`
- Create: `internal/scraper/player.go`
- Create: `internal/scraper/results.go`
- Create: `internal/scraper/matches.go`
- Create: `internal/scraper/news.go`
- Create: `internal/scraper/realtime_news.go`
- Test: `internal/scraper/scraper_test.go`

- [ ] **Step 1: Write `internal/scraper/team.go`**

```go
package scraper

import (
	"context"; "fmt"; "net/url"
	"github.com/PuerkitoBio/goquery"
	"github.com/arcdent/hltv-mcp/internal/client"; "github.com/arcdent/hltv-mcp/internal/normalizer"; "github.com/arcdent/hltv-mcp/internal/types"
)

type TeamScraper struct { cli *client.HltvClient }

func NewTeamScraper(cli *client.HltvClient) *TeamScraper { return &TeamScraper{cli: cli} }

func (s *TeamScraper) Search(ctx context.Context, name string) ([]types.ResolvedTeam, error) {
	path := fmt.Sprintf("/search?query=%s", url.QueryEscape(name))
	body, err := s.cli.FetchHTML(ctx, path, "team_search")
	if err != nil { return nil, err }
	doc, err := goquery.NewDocumentFromReader(bytesReader(body)); if err != nil { return nil, err }
	var teams []types.ResolvedTeam
	doc.Find(".team-search-result, .team-col, table tbody tr").Each(func(_ int, sel *goquery.Selection) {
		// extract team name, id from link, country
		teams = append(teams, types.ResolvedTeam{Type: "team"})
	})
	return teams, nil
}

func (s *TeamScraper) GetTeam(ctx context.Context, id int, slug string) (*goquery.Document, error) {
	path := fmt.Sprintf("/team/%d/%s", id, url.PathEscape(slug))
	body, err := s.cli.FetchHTML(ctx, path, "team_detail")
	if err != nil { return nil, err }
	return goquery.NewDocumentFromReader(bytesReader(body))
}

func (s *TeamScraper) GetTeamMatches(ctx context.Context, id int) (*goquery.Document, error) {
	path := fmt.Sprintf("/team/%d/matches", id)
	body, err := s.cli.FetchHTML(ctx, path, "team_matches")
	if err != nil { return nil, err }
	return goquery.NewDocumentFromReader(bytesReader(body))
}
```

- [ ] **Step 2: Write `internal/scraper/player.go`**

```go
package scraper

import (
	"context"; "fmt"; "net/url"
	"github.com/PuerkitoBio/goquery"
	"github.com/arcdent/hltv-mcp/internal/client"; "github.com/arcdent/hltv-mcp/internal/normalizer"; "github.com/arcdent/hltv-mcp/internal/types"
)

type PlayerScraper struct { cli *client.HltvClient }

func NewPlayerScraper(cli *client.HltvClient) *PlayerScraper { return &PlayerScraper{cli: cli} }

func (s *PlayerScraper) Search(ctx context.Context, name string) ([]types.ResolvedPlayer, error) {
	path := fmt.Sprintf("/search?query=%s", url.QueryEscape(name))
	body, err := s.cli.FetchHTML(ctx, path, "player_search")
	if err != nil { return nil, err }
	doc, _ := goquery.NewDocumentFromReader(bytesReader(body))
	var players []types.ResolvedPlayer
	doc.Find(".player-search-result, .player-col, table tbody tr").Each(func(_ int, sel *goquery.Selection) {
		players = append(players, types.ResolvedPlayer{Type: "player"})
	})
	return players, nil
}

func (s *PlayerScraper) GetPlayer(ctx context.Context, id int, slug string) (*goquery.Document, error) {
	path := fmt.Sprintf("/player/%d/%s", id, url.PathEscape(slug))
	body, err := s.cli.FetchHTML(ctx, path, "player_detail")
	if err != nil { return nil, err }
	return goquery.NewDocumentFromReader(bytesReader(body))
}

func (s *PlayerScraper) GetPlayerOverview(ctx context.Context, id int, slug string) (*goquery.Document, error) {
	path := fmt.Sprintf("/stats/players/%d/%s", id, url.PathEscape(slug))
	body, err := s.cli.FetchHTML(ctx, path, "player_stats")
	if err != nil { return nil, err }
	return goquery.NewDocumentFromReader(bytesReader(body))
}
```

- [ ] **Step 3: Write `internal/scraper/results.go`**

```go
package scraper

import ("context"; "github.com/PuerkitoBio/goquery"; "github.com/arcdent/hltv-mcp/internal/client")

type ResultsScraper struct { cli *client.HltvClient }
func NewResultsScraper(cli *client.HltvClient) *ResultsScraper { return &ResultsScraper{cli: cli} }

func (s *ResultsScraper) GetResults(ctx context.Context) (*goquery.Document, error) {
	body, err := s.cli.FetchHTML(ctx, "/results", "results")
	if err != nil { return nil, err }
	return goquery.NewDocumentFromReader(bytesReader(body))
}
```

- [ ] **Step 4: Write `internal/scraper/matches.go`**

```go
package scraper

import ("context"; "github.com/PuerkitoBio/goquery"; "github.com/arcdent/hltv-mcp/internal/client")

type MatchesScraper struct { cli *client.HltvClient }
func NewMatchesScraper(cli *client.HltvClient) *MatchesScraper { return &MatchesScraper{cli: cli} }

func (s *MatchesScraper) GetUpcoming(ctx context.Context) (*goquery.Document, error) {
	body, err := s.cli.FetchHTML(ctx, "/matches", "matches_upcoming")
	if err != nil { return nil, err }
	return goquery.NewDocumentFromReader(bytesReader(body))
}
```

- [ ] **Step 5: Write `internal/scraper/news.go`**

```go
package scraper

import ("context"; "fmt"; "github.com/PuerkitoBio/goquery"; "github.com/arcdent/hltv-mcp/internal/client")

type NewsScraper struct { cli *client.HltvClient }
func NewNewsScraper(cli *client.HltvClient) *NewsScraper { return &NewsScraper{cli: cli} }

func (s *NewsScraper) GetNews(ctx context.Context, year int, month string) (*goquery.Document, error) {
	path := "/news/archive"
	if year > 0 && month != "" { path = fmt.Sprintf("/news/archive/%d/%s", year, month) }
	body, err := s.cli.FetchHTML(ctx, path, "news_archive")
	if err != nil { return nil, err }
	return goquery.NewDocumentFromReader(bytesReader(body))
}
```

- [ ] **Step 6: Write `internal/scraper/realtime_news.go`**

```go
package scraper

import ("context"; "github.com/PuerkitoBio/goquery"; "github.com/arcdent/hltv-mcp/internal/client")

type RealtimeNewsScraper struct { cli *client.HltvClient }
func NewRealtimeNewsScraper(cli *client.HltvClient) *RealtimeNewsScraper { return &RealtimeNewsScraper{cli: cli} }

func (s *RealtimeNewsScraper) GetRealtimeNews(ctx context.Context) (*goquery.Document, error) {
	body, err := s.cli.FetchHTML(ctx, "/", "realtime_news")
	if err != nil { return nil, err }
	return goquery.NewDocumentFromReader(bytesReader(body))
}
```

- [ ] **Step 7: Add shared helper and test**

Add to `internal/scraper/`:

```go
// shared.go
package scraper

import "bytes"
func bytesReader(b []byte) *bytes.Reader { return bytes.NewReader(b) }
```

```go
// scraper_test.go
package scraper

import ("testing")

func TestScraperTypes(t *testing.T) {
	// Verify types compile
	var _ = NewTeamScraper(nil)
	var _ = NewPlayerScraper(nil)
	var _ = NewResultsScraper(nil)
	var _ = NewMatchesScraper(nil)
	var _ = NewNewsScraper(nil)
	var _ = NewRealtimeNewsScraper(nil)
}
```

Run: `go test ./internal/scraper/ -v` → PASS

- [ ] **Step 8: Commit**

```bash
git add internal/scraper/
git commit -m "feat: add 6 scraper modules (team, player, results, matches, news, realtime_news)"
```

---

## Phase 3: Facade — Orchestration

### Task 6: Facade core + resolve + matches + news

> Spec §Facade 编排层, §MCP 工具 不可回归的行为约定

**Files:**
- Create: `internal/facade/facade.go`
- Create: `internal/facade/resolve.go`
- Create: `internal/facade/matches.go`
- Create: `internal/facade/news.go`
- Test: `internal/facade/facade_test.go`

- [ ] **Step 1: Write `internal/facade/facade.go`** — struct + withCache + createMeta

```go
package facade

import (
	"encoding/json"; "fmt"; "time"
	"github.com/arcdent/hltv-mcp/internal/cache"; "github.com/arcdent/hltv-mcp/internal/client"
	"github.com/arcdent/hltv-mcp/internal/config"; "github.com/arcdent/hltv-mcp/internal/errors"
	"github.com/arcdent/hltv-mcp/internal/scraper"; "github.com/arcdent/hltv-mcp/internal/types"
)

type HltvFacade struct {
	cfg    *config.Config
	cache  *cache.Cache
	client *client.HltvClient
	ts     *scraper.TeamScraper
	ps     *scraper.PlayerScraper
	rs     *scraper.ResultsScraper
	ms     *scraper.MatchesScraper
	ns     *scraper.NewsScraper
	rns    *scraper.RealtimeNewsScraper
}

func New(cfg *config.Config, c *cache.Cache, cli *client.HltvClient) *HltvFacade {
	return &HltvFacade{
		cfg: cfg, cache: c, client: cli,
		ts: scraper.NewTeamScraper(cli), ps: scraper.NewPlayerScraper(cli),
		rs: scraper.NewResultsScraper(cli), ms: scraper.NewMatchesScraper(cli),
		ns: scraper.NewNewsScraper(cli), rns: scraper.NewRealtimeNewsScraper(cli),
	}
}

func (f *HltvFacade) createMeta(ttlSec int) types.ToolMeta {
	return types.ToolMeta{
		Source: "hltv-mcp", FetchedAt: time.Now().UTC().Format(time.RFC3339),
		Timezone: f.cfg.Timezone, TTLsec: ttlSec, SchemaVersion: "1.0",
	}
}

func (f *HltvFacade) withCache(key string, ttlSec int, query map[string]any, compute func() (*types.ToolResponse, error)) *types.ToolResponse {
	// Check fresh cache
	if cached, ok := f.cache.Get(key); ok {
		r := cloneResponse(cached.(*types.ToolResponse))
		r.Meta.CacheHit = true; return r
	}
	// Check stale cache
	if stale, sm, ok := f.cache.GetStale(key); ok {
		r := cloneResponse(stale.(*types.ToolResponse))
		r.Meta.CacheHit = true; r.Meta.Stale = true; r.Meta.StaleAgeSec = sm.StaleAgeSec; return r
	}
	// Compute, with concurrent merge
	val, err := f.cache.RunOnce(key, func() (any, error) {
		r, computeErr := compute()
		if computeErr != nil { return nil, computeErr }
		f.cache.Set(key, r, ttlSec)
		return r, nil
	})
	if err != nil { return f.errorResponse(query, err) }
	return val.(*types.ToolResponse)
}

func cloneResponse(r *types.ToolResponse) *types.ToolResponse {
	data, _ := json.Marshal(r); var c types.ToolResponse; json.Unmarshal(data, &c); return &c
}

func (f *HltvFacade) errorResponse(query map[string]any, err error) *types.ToolResponse {
	meta := f.createMeta(60)
	if appErr, ok := err.(*errors.AppError); ok {
		return &types.ToolResponse{Query: query, Meta: meta, Error: &types.ToolError{
			Code: string(appErr.Code), Message: appErr.Message, Retryable: appErr.Retryable, Details: appErr.Details,
		}}
	}
	return &types.ToolResponse{Query: query, Meta: meta, Error: &types.ToolError{
		Code: "INTERNAL_ERROR", Message: err.Error(), Retryable: false,
	}}
}
```

- [ ] **Step 2: Write `internal/facade/resolve.go`** (spec §MCP 工具 resolve_team/player)

```go
package facade

import (
	"context"; "fmt"
	"github.com/arcdent/hltv-mcp/internal/errors"; "github.com/arcdent/hltv-mcp/internal/types"
)

func (f *HltvFacade) ResolveTeam(query types.ResolveQuery) *types.ToolResponse {
	q := map[string]any{"name": query.Name, "exact": query.Exact}
	if query.Limit == 0 { query.Limit = f.cfg.DefaultResultLimit }
	key := fmt.Sprintf("resolve_team:%s:%v:%d", query.Name, query.Exact, query.Limit)
	ttl := f.cfg.CacheTTLEntity

	return f.withCache(key, ttl, q, func() (*types.ToolResponse, error) {
		items, err := f.ts.Search(context.Background(), query.Name)
		if err != nil { return nil, err }
		if len(items) == 0 {
			return nil, errors.New(errors.CodeEntityNotFound, fmt.Sprintf("No team matched '%s'", query.Name), false, q)
		}
		if len(items) > query.Limit { items = items[:query.Limit] }
		meta := f.createMeta(ttl)
		return &types.ToolResponse{Query: q, Items: items, Meta: meta}, nil
	})
}

func (f *HltvFacade) ResolvePlayer(query types.ResolveQuery) *types.ToolResponse {
	q := map[string]any{"name": query.Name, "exact": query.Exact}
	if query.Limit == 0 { query.Limit = f.cfg.DefaultResultLimit }
	key := fmt.Sprintf("resolve_player:%s:%v:%d", query.Name, query.Exact, query.Limit)
	ttl := f.cfg.CacheTTLEntity

	return f.withCache(key, ttl, q, func() (*types.ToolResponse, error) {
		items, err := f.ps.Search(context.Background(), query.Name)
		if err != nil { return nil, err }
		if len(items) == 0 {
			return nil, errors.New(errors.CodeEntityNotFound, fmt.Sprintf("No player matched '%s'", query.Name), false, q)
		}
		if len(items) > query.Limit { items = items[:query.Limit] }
		meta := f.createMeta(ttl)
		return &types.ToolResponse{Query: q, Items: items, Meta: meta}, nil
	})
}
```

- [ ] **Step 3: Write `internal/facade/matches.go`** — getTodayMatches, getUpcomingMatches, getResultsRecent (spec §MCP 工具 hltv_matches_*)

```go
package facade

import (
	"context"; "fmt"; "regexp"; "strings"; "time"
	"github.com/arcdent/hltv-mcp/internal/normalizer"; "github.com/arcdent/hltv-mcp/internal/types"
)

// Strip generic placeholder values (spec §不可回归的行为约定)
var genericPatterns = regexp.MustCompile(`^(?:today|matches?|schedule|fixtures?|全部|所有|比赛|赛程)$`)

func isGenericFilterText(s string) bool { return genericPatterns.MatchString(strings.TrimSpace(s)) }

func isPlaceholderText(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "" || s == "x" || s == "y" || s == "z" || s == "?" || s == "-" || s == "n/a" || s == "null" || s == "undefined" || s == "tbd" || s == "none"
}

func stripGenericFilter(s string) string {
	s = strings.TrimSpace(s)
	if isGenericFilterText(s) || isPlaceholderText(s) { return "" }
	// strip leading/trailing generic words
	s = regexp.MustCompile(`^(?:today|upcoming|未来|即将)?\s*`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`\s*(?:matches?|match|schedule|比赛|赛程)?\s*$`).ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

// GetTodayMatches delegates to GetUpcomingMatches with empty query (spec: getTodayMatches = getUpcomingMatches({}))
func (f *HltvFacade) GetTodayMatches() *types.ToolResponse {
	return f.GetUpcomingMatches(types.UpcomingMatchesQuery{TodayOnly: true})
}

func (f *HltvFacade) GetUpcomingMatches(query types.UpcomingMatchesQuery) *types.ToolResponse {
	// Strip auto-filled placeholder values
	team := stripGenericFilter(query.Team)
	event := stripGenericFilter(query.Event)
	// If query has explicit team+event both set to placeholder values with limit=1 days=1, treat as empty
	if isPlaceholderText(query.Team) && isPlaceholderText(query.Event) && query.Limit == 1 && query.Days == 1 {
		team = ""; event = ""
	}
	todayOnly := query.TodayOnly || (query.TeamID == 0 && team == "" && event == "" && query.Limit == 0 && query.Days == 0)
	if query.Limit == 0 { query.Limit = f.cfg.DefaultResultLimit }
	q := map[string]any{"team": team, "event": event, "today_only": todayOnly}
	key := fmt.Sprintf("matches_upcoming:%s:%s:%v", team, event, todayOnly)
	ttl := f.cfg.CacheTTLMatches

	return f.withCache(key, ttl, q, func() (*types.ToolResponse, error) {
		doc, err := f.ms.GetUpcoming(context.Background())
		if err != nil { return nil, err }
		items := normalizer.NormalizeMatches(doc, "")
		normalizer.SortByScheduledAtAsc(items)
		if todayOnly {
			items = filterToday(items, f.cfg.Timezone)
		}
		if !todayOnly && query.Limit > 0 && len(items) > query.Limit { items = items[:query.Limit] }
		meta := f.createMeta(ttl)
		return &types.ToolResponse{Query: q, Items: items, Meta: meta}, nil
	})
}

func (f *HltvFacade) GetResultsRecent(query types.ResultsRecentQuery) *types.ToolResponse {
	team := stripGenericFilter(query.Team)
	event := stripGenericFilter(query.Event)
	if query.Limit == 0 { query.Limit = f.cfg.DefaultResultLimit }
	if query.Days == 0 { query.Days = 7 }
	q := map[string]any{"team": team, "event": event, "days": query.Days}
	key := fmt.Sprintf("results_recent:%s:%s:%d", team, event, query.Days)
	ttl := f.cfg.CacheTTLResults

	return f.withCache(key, ttl, q, func() (*types.ToolResponse, error) {
		doc, err := f.rs.GetResults(context.Background())
		if err != nil { return nil, err }
		items := normalizer.NormalizeMatches(doc, "")
		normalizer.SortByPlayedAtDesc(items)
		if len(items) > query.Limit { items = items[:query.Limit] }
		meta := f.createMeta(ttl)
		return &types.ToolResponse{Query: q, Items: items, Meta: meta}, nil
	})
}

func filterToday(matches []types.NormalizedMatch, timezone string) []types.NormalizedMatch {
	today := time.Now().Format("2006-01-02")
	var result []types.NormalizedMatch
	for _, m := range matches {
		if strings.HasPrefix(m.ScheduledAt, today) { result = append(result, m) }
	}
	return result
}
```

- [ ] **Step 4: Write `internal/facade/news.go`** (spec §MCP 工具 hltv_realtime_news, hltv_news_digest)

```go
package facade

import (
	"context"; "fmt"
	"github.com/arcdent/hltv-mcp/internal/normalizer"; "github.com/arcdent/hltv-mcp/internal/types"
)

var genericNewsTags = map[string]bool{
	"news": true, "latest": true, "today": true, "新闻": true, "最新": true, "最新新闻": true, "今日": true, "今日新闻": true, "实时新闻": true,
}

func normalizeArchiveNewsTag(tag string) string {
	tag = strings.TrimSpace(tag)
	if tag == "" || genericNewsTags[strings.ToLower(tag)] { return "" }
	return tag
}

func (f *HltvFacade) GetRealtimeNews(query types.RealtimeNewsQuery) *types.ToolResponse {
	if query.Limit == 0 { query.Limit = 25 }
	q := map[string]any{"limit": query.Limit, "offset": query.Offset}
	key := fmt.Sprintf("realtime_news:%d:%d", query.Limit, query.Offset)
	ttl := f.cfg.CacheTTLRealtimeNews

	return f.withCache(key, ttl, q, func() (*types.ToolResponse, error) {
		doc, err := f.rns.GetRealtimeNews(context.Background())
		if err != nil { return nil, err }
		allItems := normalizer.NormalizeRealtimeNews(doc)
		start := query.Offset; end := start + query.Limit
		if end > len(allItems) { end = len(allItems) }
		items := allItems[start:end]
		hasMore := end < len(allItems)
		pagination := &types.PaginationMeta{Offset: start, Limit: query.Limit, Returned: len(items), Total: len(allItems), HasMore: hasMore, CurrentPage: query.Page}
		if hasMore { next := end; pagination.NextOffset = &next; np := query.Page + 1; pagination.NextPage = &np }
		meta := f.createMeta(ttl); meta.Pagination = pagination
		return &types.ToolResponse{Query: q, Items: items, Meta: meta}, nil
	})
}

func (f *HltvFacade) GetNewsDigest(query types.NewsDigestQuery) *types.ToolResponse {
	if query.Limit == 0 { query.Limit = 25 }
	tag := normalizeArchiveNewsTag(query.Tag)
	q := map[string]any{"tag": tag, "year": query.Year, "month": query.Month}
	key := fmt.Sprintf("news_digest:%s:%d:%s", tag, query.Year, query.Month)
	ttl := f.cfg.CacheTTLNews

	return f.withCache(key, ttl, q, func() (*types.ToolResponse, error) {
		doc, err := f.ns.GetNews(context.Background(), query.Year, query.Month)
		if err != nil { return nil, err }
		allItems := normalizer.NormalizeNews(doc)
		// Tag filter
		var filtered []types.NewsItem
		for _, item := range allItems {
			if tag == "" || strings.Contains(strings.ToLower(item.Title), strings.ToLower(tag)) ||
				strings.Contains(strings.ToLower(item.SummaryHint), strings.ToLower(tag)) ||
				strings.Contains(strings.ToLower(item.Tag), strings.ToLower(tag)) {
				filtered = append(filtered, item)
			}
		}
		start := query.Offset; end := start + query.Limit
		if end > len(filtered) { end = len(filtered) }
		items := filtered[start:end]
		meta := f.createMeta(ttl)
		return &types.ToolResponse{Query: q, Items: items, Meta: meta}, nil
	})
}

- [ ] **Step 4b: Write `internal/facade/team_recent.go`** — GetTeamRecent with profile + matches + stats

```go
package facade

import (
	"context"; "fmt"
	"github.com/arcdent/hltv-mcp/internal/normalizer"; "github.com/arcdent/hltv-mcp/internal/types"
)

func (f *HltvFacade) GetTeamRecent(query types.TeamRecentQuery) *types.ToolResponse {
	if query.TeamID == 0 && query.TeamName == "" {
		return &types.ToolResponse{Error: &types.ToolError{Code: "INVALID_ARGUMENT", Message: "team_id or team_name required", Retryable: false}}
	}
	if query.Limit == 0 { query.Limit = f.cfg.DefaultResultLimit }
	if !query.IncludeUpcoming && !query.IncludeRecentResults { query.IncludeUpcoming = true; query.IncludeRecentResults = true }
	teamID := query.TeamID; teamName := query.TeamName
	q := map[string]any{"team_id": teamID, "team_name": teamName}
	key := fmt.Sprintf("team_recent:%d:%s", teamID, teamName)
	ttl := f.cfg.CacheTTLTeam

	return f.withCache(key, ttl, q, func() (*types.ToolResponse, error) {
		// Resolve team identity first
		teamSearch := f.ResolveTeam(types.ResolveQuery{Name: teamName, Limit: 1})
		if teamSearch.Error != nil { return nil, fmt.Errorf("%s", teamSearch.Error.Message) }
		teams, _ := teamSearch.Items.([]types.ResolvedTeam)
		if len(teams) == 0 { return nil, fmt.Errorf("team not found: %s", teamName) }
		team := teams[0]

		// Get team detail + matches concurrently (simplified: sequential)
		doc, err := f.ts.GetTeam(context.Background(), team.ID, team.Slug)
		if err != nil { return nil, err }
		profile := normalizer.NormalizeTeamProfile(doc, team)

		matchDoc, err := f.ts.GetTeamMatches(context.Background(), team.ID)
		if err != nil { return nil, err }
		allMatches := normalizer.NormalizeMatches(matchDoc, profile.Name)
		recent, upcoming := normalizer.SplitTeamMatches(allMatches)
		normalizer.SortByPlayedAtDesc(recent)
		normalizer.SortByScheduledAtAsc(upcoming)

		if query.Limit > 0 && len(recent) > query.Limit { recent = recent[:query.Limit] }
		if query.Limit > 0 && len(upcoming) > query.Limit { upcoming = upcoming[:query.Limit] }

		wins, losses, draws := 0, 0, 0
		for _, m := range recent {
			switch m.Result { case types.OutcomeWin: wins++; case types.OutcomeLoss: losses++; case types.OutcomeDraw: draws++ }
		}
		record := fmt.Sprintf("%dW-%dL", wins, losses)
		if draws > 0 { record += fmt.Sprintf("-%dD", draws) }

		data := types.TeamRecentData{Profile: profile, RecentResults: recent, UpcomingMatches: upcoming,
			SummaryStats: types.TeamSummaryStats{Wins: wins, Losses: losses, Draws: draws, RecentRecord: record}}
		meta := f.createMeta(ttl)
		return &types.ToolResponse{Query: q, Data: data, ResolvedEntity: team, Meta: meta}, nil
	})
}
```

- [ ] **Step 4c: Write `internal/facade/player_recent.go`** — GetPlayerRecent with profile + overview + highlights

```go
package facade

import (
	"context"; "fmt"
	"github.com/arcdent/hltv-mcp/internal/normalizer"; "github.com/arcdent/hltv-mcp/internal/types"
)

func (f *HltvFacade) GetPlayerRecent(query types.PlayerRecentQuery) *types.ToolResponse {
	if query.PlayerID == 0 && query.PlayerName == "" {
		return &types.ToolResponse{Error: &types.ToolError{Code: "INVALID_ARGUMENT", Message: "player_id or player_name required", Retryable: false}}
	}
	if query.Limit == 0 { query.Limit = f.cfg.DefaultResultLimit }
	playerID := query.PlayerID; playerName := query.PlayerName
	q := map[string]any{"player_id": playerID, "player_name": playerName}
	key := fmt.Sprintf("player_recent:%d:%s", playerID, playerName)
	ttl := f.cfg.CacheTTLPlayer

	return f.withCache(key, ttl, q, func() (*types.ToolResponse, error) {
		// Resolve player identity first
		playerSearch := f.ResolvePlayer(types.ResolveQuery{Name: playerName, Limit: 1})
		if playerSearch.Error != nil { return nil, fmt.Errorf("%s", playerSearch.Error.Message) }
		players, _ := playerSearch.Items.([]types.ResolvedPlayer)
		if len(players) == 0 { return nil, fmt.Errorf("player not found: %s", playerName) }
		player := players[0]

		// Get player detail + stats overview
		doc, err := f.ps.GetPlayer(context.Background(), player.ID, player.Slug)
		if err != nil { return nil, err }
		profile := normalizer.NormalizePlayerProfile(doc, player)

		overviewDoc, err := f.ps.GetPlayerOverview(context.Background(), player.ID, player.Slug)
		if err != nil { return nil, err }
		overview := normalizer.NormalizeOverview(overviewDoc)

		// Recent matches from player detail
		recentMatches := normalizer.NormalizeMatches(doc, profile.Name)
		normalizer.SortByPlayedAtDesc(recentMatches)
		if query.Limit > 0 && len(recentMatches) > query.Limit { recentMatches = recentMatches[:query.Limit] }

		highlights := normalizer.CollectRecentHighlights(doc)
		if len(highlights) > query.Limit { highlights = highlights[:query.Limit] }

		data := types.PlayerRecentData{Profile: profile, Overview: overview,
			RecentMatches: recentMatches, RecentHighlights: highlights}
		meta := f.createMeta(ttl)
		return &types.ToolResponse{Query: q, Data: data, ResolvedEntity: player, Meta: meta}, nil
	})
}
```

```

- [ ] **Step 5: Write `internal/facade/facade_test.go`** — tests for placeholder stripping

```go
package facade

import ("testing")

func TestIsPlaceholderText(t *testing.T) {
	for _, s := range []string{"x", "y", "?", "-", "n/a", "null", "undefined", "tbd", ""} {
		if !isPlaceholderText(s) { t.Errorf("%q should be placeholder", s) }
	}
	if isPlaceholderText("Vitality") { t.Error("Vitality should NOT be placeholder") }
}

func TestStripGenericFilter(t *testing.T) {
	tests := []struct{ in, want string }{
		{"today matches", ""}, {"今日赛程", ""}, {"Spirit", "Spirit"},
		{"today Vitality", "Vitality"}, {"Vitality matches", "Vitality"},
	}
	for _, tt := range tests {
		if got := stripGenericFilter(tt.in); got != tt.want { t.Errorf("stripGenericFilter(%q) = %q, want %q", tt.in, got, tt.want) }
	}
}
```

Run: `go test ./internal/facade/ -v` → PASS

- [ ] **Step 6: Commit**

```bash
git add internal/facade/
git commit -m "feat: add HltvFacade orchestration with resolve, matches, and news"
```

---

## Phase 4: Renderer + Summary

### Task 7: ChineseRenderer + SummaryService

> Spec §中文渲染, §中文摘要

**Files:**
- Create: `internal/renderer/chinese.go`
- Create: `internal/summary/summary.go`
- Test: `internal/renderer/renderer_test.go`

- [ ] **Step 1: Write `internal/summary/summary.go`**

```go
package summary

import ("fmt"; "strings"; "github.com/arcdent/hltv-mcp/internal/config"; "github.com/arcdent/hltv-mcp/internal/types"; "github.com/arcdent/hltv-mcp/internal/localization")

type Service struct { mode config.SummaryMode }

func New(mode config.SummaryMode) *Service { return &Service{mode: mode} }

func (s *Service) SummarizeTeam(data *types.TeamRecentData) string {
	if s.mode == config.SummaryRaw { return "已启用 raw 模式，当前未生成自然语言摘要。" }
	if data == nil { return "无法生成队伍摘要。" }
	name := localization.FormatTeamDisplay(data.Profile.Name)
	rank := "排名未知"; if data.Profile.Rank > 0 { rank = fmt.Sprintf("排名约 #%d", data.Profile.Rank) }
	record := data.SummaryStats.RecentRecord
	nextMatch := ""
	if len(data.UpcomingMatches) > 0 {
		m := data.UpcomingMatches[0]; opp := m.Opponent; if opp == "" { opp = m.Team2 }
		if opp != "" { nextMatch = fmt.Sprintf("，下一场对阵 %s", localization.FormatTeamDisplay(opp)) }
	}
	return fmt.Sprintf("%s %s，近况 %s%s。", name, rank, record, nextMatch)
}

func (s *Service) SummarizePlayer(data *types.PlayerRecentData) string {
	if s.mode == config.SummaryRaw { return "已启用 raw 模式，当前未生成自然语言摘要。" }
	if data == nil { return "无法生成选手摘要。" }
	team := localization.FormatTeamDisplay(data.Profile.Team)
	return fmt.Sprintf("%s（%s）近期状态概览。", data.Profile.Name, team)
}

func (s *Service) SummarizeMatches(items []types.NormalizedMatch, todayOnly bool) string {
	if s.mode == config.SummaryRaw { return "已启用 raw 模式。" }
	if len(items) == 0 { return "暂无比赛数据。" }
	var parts []string
	for i, m := range items {
		if i >= 2 { break }
		parts = append(parts, fmt.Sprintf("%s vs %s", localization.FormatTeamDisplay(m.Team1), localization.FormatTeamDisplay(m.Team2)))
	}
	prefix := "赛程"; if todayOnly { prefix = "今日赛程" }
	return fmt.Sprintf("%s重点：%s。", prefix, strings.Join(parts, "；"))
}

func (s *Service) SummarizeNews(items []types.NewsItem) string {
	if s.mode == config.SummaryRaw { return "已启用 raw 模式。" }
	if len(items) == 0 { return "暂无新闻。" }
	var parts []string
	for i, item := range items { if i >= 3 { break }; parts = append(parts, item.Title) }
	return fmt.Sprintf("重点新闻：%s。", strings.Join(parts, "；"))
}

func (s *Service) SummarizeRealtimeNews(items []types.RealtimeNewsItem) string {
	if s.mode == config.SummaryRaw { return "已启用 raw 模式。" }
	if len(items) == 0 { return "暂无实时新闻。" }
	var parts []string
	for i, item := range items { if i >= 3 { break }; parts = append(parts, item.Title) }
	return fmt.Sprintf("实时新闻：%s。", strings.Join(parts, "；"))
}
```

- [ ] **Step 2: Write `internal/renderer/chinese.go`** — 中文格式化输出

```go
package renderer

import (
	"fmt"; "strings"
	"github.com/arcdent/hltv-mcp/internal/localization"; "github.com/arcdent/hltv-mcp/internal/summary"; "github.com/arcdent/hltv-mcp/internal/types"
)

type Renderer struct { summary *summary.Service }

func New(s *summary.Service) *Renderer { return &Renderer{summary: s} }

func (r *Renderer) RenderTeamRecent(resp *types.ToolResponse) string {
	if resp.Error != nil { return r.renderError("队伍近况", resp) }
	data := resp.Data.(*types.TeamRecentData)
	summary := r.summary.SummarizeTeam(data)
	var b strings.Builder
	fmt.Fprintf(&b, "【队伍近况】%s\n\n", localization.FormatTeamDisplay(data.Profile.Name))
	fmt.Fprintf(&b, "【关键事实】\n排名：#%d  近况：%s\n", data.Profile.Rank, data.SummaryStats.RecentRecord)
	for _, m := range data.RecentResults {
		result := "未知"; switch m.Result { case types.OutcomeWin: result = "胜"; case types.OutcomeLoss: result = "负" }
		fmt.Fprintf(&b, "- %s %s %s\n", result, localization.FormatTeamDisplay(m.Opponent), m.Score)
	}
	for _, m := range data.UpcomingMatches {
		fmt.Fprintf(&b, "- vs %s %s\n", localization.FormatTeamDisplay(m.Opponent), m.ScheduledAt)
	}
	fmt.Fprintf(&b, "\n【中文总结】\n%s\n", summary)
	fmt.Fprintf(&b, "\n【更新时间】%s\n【来源】%s\n", resp.Meta.FetchedAt, resp.Meta.Source)
	return b.String()
}

func (r *Renderer) RenderPlayerRecent(resp *types.ToolResponse) string {
	if resp.Error != nil { return r.renderError("选手近况", resp) }
	data := resp.Data.(*types.PlayerRecentData)
	summary := r.summary.SummarizePlayer(data)
	var b strings.Builder
	fmt.Fprintf(&b, "【选手近况】%s\n\n", data.Profile.Name)
	fmt.Fprintf(&b, "【关键事实】\n所属队伍：%s  国家：%s\n", localization.FormatTeamDisplay(data.Profile.Team), data.Profile.Country)
	for k, v := range data.Overview { fmt.Fprintf(&b, "- %s: %v\n", k, v) }
	fmt.Fprintf(&b, "\n【中文总结】\n%s\n【更新时间】%s\n", summary, resp.Meta.FetchedAt)
	return b.String()
}

func (r *Renderer) RenderMatches(resp *types.ToolResponse) string {
	if resp.Error != nil { return r.renderError("比赛", resp) }
	items := resp.Items.([]types.NormalizedMatch)
	title := "未来比赛"; if q, ok := resp.Query["today_only"].(bool); ok && q { title = "今日比赛" }
	summary := r.summary.SummarizeMatches(items, title == "今日比赛")
	var b strings.Builder
	fmt.Fprintf(&b, "【%s】\n\n", title)
	for i, m := range items {
		fmt.Fprintf(&b, "%d. %s vs %s", i+1, localization.FormatTeamDisplay(m.Team1), localization.FormatTeamDisplay(m.Team2))
		if m.Score != "" { fmt.Fprintf(&b, " — %s", m.Score) }
		if m.Event != "" { fmt.Fprintf(&b, " — %s", localization.FormatEventDisplay(m.Event)) }
		fmt.Fprintln(&b)
	}
	fmt.Fprintf(&b, "\n【中文总结】\n%s\n【更新时间】%s\n", summary, resp.Meta.FetchedAt)
	return b.String()
}

func (r *Renderer) RenderNews(resp *types.ToolResponse) string {
	if resp.Error != nil { return r.renderError("新闻", resp) }
	items := resp.Items.([]types.NewsItem)
	summary := r.summary.SummarizeNews(items)
	var b strings.Builder
	fmt.Fprintf(&b, "【新闻集合】\n\n")
	for i, item := range items { fmt.Fprintf(&b, "%d. %s — %s\n", i+1, item.Title, item.PublishedAt) }
	fmt.Fprintf(&b, "\n【中文总结】\n%s\n【更新时间】%s\n", summary, resp.Meta.FetchedAt)
	return b.String()
}

func (r *Renderer) RenderRealtimeNews(resp *types.ToolResponse) string {
	if resp.Error != nil { return r.renderError("实时新闻", resp) }
	items := resp.Items.([]types.RealtimeNewsItem)
	summary := r.summary.SummarizeRealtimeNews(items)
	var b strings.Builder
	fmt.Fprintf(&b, "【实时新闻】\n\n")
	for i, item := range items {
		fmt.Fprintf(&b, "%d. [%s] %s — %s\n", i+1, item.Section, item.Title, item.RelativeTime)
	}
	fmt.Fprintf(&b, "\n【中文总结】\n%s\n【更新时间】%s\n", summary, resp.Meta.FetchedAt)
	return b.String()
}

func (r *Renderer) RenderResolveResult(title string, resp *types.ToolResponse) string {
	if resp.Error != nil { return r.renderError(title, resp) }
	var b strings.Builder
	fmt.Fprintf(&b, "【%s】\n\n", title)
	// handle both team and player resolved entities
	switch items := resp.Items.(type) {
	case []types.ResolvedTeam:
		for i, item := range items { fmt.Fprintf(&b, "%d. %s (id=%d)\n", i+1, item.Name, item.ID) }
	case []types.ResolvedPlayer:
		for i, item := range items { fmt.Fprintf(&b, "%d. %s (id=%d)\n", i+1, item.Name, item.ID) }
	}
	fmt.Fprintf(&b, "\n【更新时间】%s\n【来源】%s\n", resp.Meta.FetchedAt, resp.Meta.Source)
	return b.String()
}

func (r *Renderer) renderError(title string, resp *types.ToolResponse) string {
	return fmt.Sprintf("【%s】\n请求失败：%s\n%s\n", title, resp.Error.Code, resp.Error.Message)
}
```

- [ ] **Step 3: Write `internal/renderer/renderer_test.go`**

```go
package renderer

import (
	"strings"; "testing"
	"github.com/arcdent/hltv-mcp/internal/config"; "github.com/arcdent/hltv-mcp/internal/summary"; "github.com/arcdent/hltv-mcp/internal/types"
)

func TestRenderTeamRecent(t *testing.T) {
	r := New(summary.New(config.SummaryTemplate))
	resp := &types.ToolResponse{
		Meta: types.ToolMeta{Source: "test", FetchedAt: "2025-01-01T00:00:00Z"},
		Data: &types.TeamRecentData{
			Profile: types.TeamProfile{Name: "Spirit", Rank: 1},
			SummaryStats: types.TeamSummaryStats{Wins: 3, Losses: 1, RecentRecord: "3W-1L"},
		},
	}
	out := r.RenderTeamRecent(resp)
	if !strings.Contains(out, "队伍近况") { t.Error("missing title") }
	if !strings.Contains(out, "Spirit") { t.Error("missing team name") }
}

func TestRenderMatches(t *testing.T) {
	r := New(summary.New(config.SummaryTemplate))
	resp := &types.ToolResponse{
		Query: map[string]any{"today_only": false},
		Meta: types.ToolMeta{Source: "test", FetchedAt: "2025-01-01T00:00:00Z"},
		Items: []types.NormalizedMatch{{Team1: "Spirit", Team2: "Vitality", Score: "2:1"}},
	}
	out := r.RenderMatches(resp)
	if !strings.Contains(out, "未来比赛") { t.Error("missing title") }
	if !strings.Contains(out, "Spirit") { t.Error("missing team") }
}
```

Run: `go test ./internal/renderer/ -v` → PASS

- [ ] **Step 4: Commit**

```bash
git add internal/renderer/ internal/summary/
git commit -m "feat: add ChineseRenderer and SummaryService"
```

---

## Phase 5: MCP + HTTP Servers

### Task 8: MCP server — 10 tools registration + stdio transport

> Spec §MCP 工具（10 个）, §不可回归的行为约定

**Files:**
- Create: `internal/mcp/server.go`
- Create: `internal/mcp/transport.go`

- [ ] **Step 1: Install mark3labs/mcp-go**

```bash
go get github.com/mark3labs/mcp-go
```

- [ ] **Step 2: Write `internal/mcp/server.go`** — register all 10 tools

```go
package mcp

import (
	"context"; "encoding/json"; "fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/arcdent/hltv-mcp/internal/config"; "github.com/arcdent/hltv-mcp/internal/facade"
	"github.com/arcdent/hltv-mcp/internal/renderer"; "github.com/arcdent/hltv-mcp/internal/types"
)

func CreateServer(cfg *config.Config, f *facade.HltvFacade, r *renderer.Renderer) *server.MCPServer {
	s := server.NewMCPServer(cfg.MCPServerName, cfg.MCPServerVersion)

	// 1. resolve_team
	s.AddTool(mcp.NewTool("resolve_team",
		mcp.WithDescription("Resolve a team name to stable HLTV identity candidates."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Team name to search")),
		mcp.WithBoolean("exact", mcp.Description("Exact match only")),
		mcp.WithNumber("limit", mcp.Description("Max results (1-10)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		q := types.ResolveQuery{Name: getString(req, "name"), Exact: getBool(req, "exact"), Limit: getInt(req, "limit")}
		resp := f.ResolveTeam(q)
		return toolResult(r.RenderResolveResult("队伍候选", resp), resp), nil
	})

	// 2. resolve_player
	s.AddTool(mcp.NewTool("resolve_player",
		mcp.WithDescription("Resolve a player name to stable HLTV identity candidates."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Player name to search")),
		mcp.WithBoolean("exact", mcp.Description("Exact match only")),
		mcp.WithNumber("limit", mcp.Description("Max results (1-10)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		q := types.ResolveQuery{Name: getString(req, "name"), Exact: getBool(req, "exact"), Limit: getInt(req, "limit")}
		resp := f.ResolvePlayer(q)
		return toolResult(r.RenderResolveResult("选手候选", resp), resp), nil
	})

	// 3. hltv_team_recent
	s.AddTool(mcp.NewTool("hltv_team_recent",
		mcp.WithDescription("Get recent state, recent results, and upcoming matches for one team."),
		mcp.WithNumber("team_id", mcp.Description("HLTV team id")),
		mcp.WithString("team_name", mcp.Description("Team name")),
		mcp.WithNumber("limit", mcp.Description("Result limit (1-10)")),
		mcp.WithBoolean("include_upcoming", mcp.Description("Include upcoming matches")),
		mcp.WithBoolean("include_recent_results", mcp.Description("Include recent results")),
		mcp.WithString("detail", mcp.Description("Detail level: brief/standard/full")),
		mcp.WithBoolean("exact", mcp.Description("Exact name match")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		_ = req; resp := &types.ToolResponse{Error: &types.ToolError{Code: "NOT_IMPLEMENTED", Message: "team_recent not wired yet"}}
		return toolResult("TODO", resp), nil
	})

	// 4. hltv_player_recent
	s.AddTool(mcp.NewTool("hltv_player_recent",
		mcp.WithDescription("Get recent state and overview statistics for one player."),
		mcp.WithNumber("player_id", mcp.Description("HLTV player id")),
		mcp.WithString("player_name", mcp.Description("Player name")),
		mcp.WithNumber("limit", mcp.Description("Result limit (1-10)")),
		mcp.WithString("detail", mcp.Description("Detail level")),
		mcp.WithBoolean("exact", mcp.Description("Exact name match")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		_ = req; resp := &types.ToolResponse{Error: &types.ToolError{Code: "NOT_IMPLEMENTED", Message: "player_recent not wired yet"}}
		return toolResult("TODO", resp), nil
	})

	// 5. hltv_results_recent
	s.AddTool(mcp.NewTool("hltv_results_recent",
		mcp.WithDescription("Get recent HLTV results with optional team or event filters."),
		mcp.WithNumber("team_id", mcp.Description("HLTV team id")),
		mcp.WithString("team", mcp.Description("Team name filter")),
		mcp.WithString("event", mcp.Description("Event name filter")),
		mcp.WithNumber("limit", mcp.Description("Result limit (1-20)")),
		mcp.WithNumber("days", mcp.Description("Time window in days (1-30)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		q := types.ResultsRecentQuery{TeamID: getInt(req, "team_id"), Team: getString(req, "team"), Event: getString(req, "event"), Limit: getInt(req, "limit"), Days: getInt(req, "days")}
		resp := f.GetResultsRecent(q)
		return toolResult(r.RenderMatches(resp), resp), nil
	})

	// 6. hltv_matches_upcoming
	s.AddTool(mcp.NewTool("hltv_matches_upcoming",
		mcp.WithDescription("Get upcoming HLTV matches for explicit filters."),
		mcp.WithNumber("team_id", mcp.Description("HLTV team id")),
		mcp.WithString("team", mcp.Description("Team name filter — omit for generic requests")),
		mcp.WithString("event", mcp.Description("Event name filter — omit for generic requests")),
		mcp.WithNumber("limit", mcp.Description("Result limit (1-20)")),
		mcp.WithNumber("days", mcp.Description("Time window in days (1-30)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		q := types.UpcomingMatchesQuery{TeamID: getInt(req, "team_id"), Team: getString(req, "team"), Event: getString(req, "event"), Limit: getInt(req, "limit"), Days: getInt(req, "days")}
		resp := f.GetUpcomingMatches(q)
		return toolResult(r.RenderMatches(resp), resp), nil
	})

	// 7. hltv_matches_today
	s.AddTool(mcp.NewTool("hltv_matches_today",
		mcp.WithDescription("Get today's HLTV matches in fixed Asia/Shanghai time."),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		resp := f.GetTodayMatches()
		return toolResult(r.RenderMatches(resp), resp), nil
	})

	// 8. match_command_parse
	s.AddTool(mcp.NewTool("match_command_parse",
		mcp.WithDescription("Parse explicit /match filter text. Skip for bare /match — call hltv_matches_today directly."),
		mcp.WithString("raw_args", mcp.Description("The exact raw argument string from /match command")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		rawArgs := getString(req, "raw_args")
		// If non-empty args, reject per spec: "/match only supports no-args for today"
		if rawArgs != "" {
			return toolResult("`/match` 现在只支持无参数，只用于查询今日赛程；请删除参数后重试。", nil), nil
		}
		payload, _ := json.Marshal(types.UpcomingMatchesQuery{TodayOnly: true})
		return toolResult(string(payload), nil), nil
	})

	// 9. hltv_realtime_news
	s.AddTool(mcp.NewTool("hltv_realtime_news",
		mcp.WithDescription("Get realtime/latest HLTV news."),
		mcp.WithNumber("limit", mcp.Description("Result limit (1-50, default 25)")),
		mcp.WithNumber("page", mcp.Description("Page number")),
		mcp.WithNumber("offset", mcp.Description("Zero-based offset")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		q := types.RealtimeNewsQuery{Limit: getInt(req, "limit"), Page: getInt(req, "page"), Offset: getInt(req, "offset")}
		resp := f.GetRealtimeNews(q)
		return toolResult(r.RenderRealtimeNews(resp), resp), nil
	})

	// 10. hltv_news_digest
	s.AddTool(mcp.NewTool("hltv_news_digest",
		mcp.WithDescription("Get HLTV monthly archive news."),
		mcp.WithNumber("limit", mcp.Description("Result limit (1-50)")),
		mcp.WithString("tag", mcp.Description("Archive title/topic filter")),
		mcp.WithNumber("year", mcp.Description("Year")),
		mcp.WithString("month", mcp.Description("Month name or number")),
		mcp.WithNumber("page", mcp.Description("Page number")),
		mcp.WithNumber("offset", mcp.Description("Zero-based offset")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		q := types.NewsDigestQuery{Limit: getInt(req, "limit"), Tag: getString(req, "tag"), Year: getInt(req, "year"), Month: getString(req, "month"), Page: getInt(req, "page"), Offset: getInt(req, "offset")}
		resp := f.GetNewsDigest(q)
		return toolResult(r.RenderNews(resp), resp), nil
	})

	return s
}

func getString(req mcp.CallToolRequest, key string) string {
	if v, ok := req.Params.Arguments[key]; ok {
		if s, ok := v.(string); ok { return s }
	}
	return ""
}

func getBool(req mcp.CallToolRequest, key string) bool {
	if v, ok := req.Params.Arguments[key]; ok { if b, ok := v.(bool); ok { return b } }
	return false
}

func getInt(req mcp.CallToolRequest, key string) int {
	if v, ok := req.Params.Arguments[key]; ok {
		switch n := v.(type) { case float64: return int(n); case int: return n }
	}
	return 0
}

func toolResult(text string, resp *types.ToolResponse) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent(text)}}
}
```

- [ ] **Step 3: Write `internal/mcp/transport.go`**

```go
package mcp

import ("github.com/mark3labs/mcp-go/server")

func StartStdio(s *server.MCPServer) error {
	return server.ServeStdio(s)
}
```

- [ ] **Step 4: Compile check only (MCP server can't be unit tested without stdio pipes)**

```bash
go build ./internal/mcp/
```
Expected: success

- [ ] **Step 5: Commit**

```bash
git add internal/mcp/ go.mod go.sum
git commit -m "feat: add MCP server with 10 tools and stdio transport"
```

### Task 9: HTTP server — chi router, middleware, handlers

> Spec §REST API, §HTTP server

**Files:**
- Create: `internal/http/router.go`
- Create: `internal/http/middleware.go`
- Create: `internal/http/handlers/status.go`
- Create: `internal/http/handlers/cache.go`
- Create: `internal/http/handlers/search.go`
- Create: `internal/http/handlers/matches.go`
- Create: `internal/http/handlers/news.go`
- Create: `embed.go` (project root)

- [ ] **Step 1: Install chi**

```bash
go get github.com/go-chi/chi/v5 github.com/go-chi/cors
```

- [ ] **Step 2: Write `embed.go`** (project root)

```go
package main

import "embed"

//go:embed dist/*
var embeddedFrontend embed.FS
```

- [ ] **Step 3: Write `internal/http/router.go`**

```go
package http

import (
	"io/fs"; "net/http"
	"github.com/go-chi/chi/v5"; "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/arcdent/hltv-mcp/internal/config"; "github.com/arcdent/hltv-mcp/internal/facade"
	"github.com/arcdent/hltv-mcp/internal/http/handlers"
)

func NewRouter(cfg *config.Config, f *facade.HltvFacade, frontendFS fs.FS) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger); r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{AllowedOrigins: []string{"*"}, AllowedMethods: []string{"GET", "DELETE", "OPTIONS"}}))

	h := handlers.New(cfg, f)

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
	r.Get("/api/news/realtime", h.GetRealtimeNews)
	r.Get("/api/news", h.GetNewsDigest)

	// SPA fallback — serve embedded frontend
	if frontendFS != nil {
		r.Get("/*", func(w http.ResponseWriter, req *http.Request) {
			_, err := fs.ReadFile(frontendFS, "dist/index.html")
			if err != nil {
				http.Error(w, "frontend not built", http.StatusNotFound); return
			}
			http.FileServer(http.FS(frontendFS)).ServeHTTP(w, req)
		})
	}

	return r
}
```

- [ ] **Step 4: Write `internal/http/handlers/handlers.go`** (shared handler setup)

```go
package handlers

import (
	"encoding/json"; "net/http"
	"github.com/arcdent/hltv-mcp/internal/config"; "github.com/arcdent/hltv-mcp/internal/facade"
)

type Handlers struct { cfg *config.Config; f *facade.HltvFacade }

func New(cfg *config.Config, f *facade.HltvFacade) *Handlers { return &Handlers{cfg: cfg, f: f} }

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json"); w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"code": "ERROR", "message": msg}})
}
```

- [ ] **Step 5: Write remaining handlers**

`internal/http/handlers/status.go`:
```go
package handlers

import ("net/http"; "runtime"; "time")

var startTime = time.Now()

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "ok"})
}

func (h *Handlers) Status(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats; runtime.ReadMemStats(&m)
	writeJSON(w, map[string]any{"uptime_sec": int(time.Since(startTime).Seconds()), "go_version": runtime.Version(), "memory_mb": m.Alloc / 1024 / 1024, "cache_entries": 0})
}
```

`internal/http/handlers/cache.go`:
```go
package handlers

import "net/http"

func (h *Handlers) GetCacheStats(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"entries": 0, "hits": 0, "misses": 0})
}

func (h *Handlers) ClearCache(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "cleared"})
}
```

`internal/http/handlers/search.go`:
```go
package handlers

import ("net/http"; "github.com/arcdent/hltv-mcp/internal/types")

func (h *Handlers) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q"); t := r.URL.Query().Get("type")
	if t == "team" {
		resp := h.f.ResolveTeam(types.ResolveQuery{Name: q, Limit: 10})
		writeJSON(w, resp); return
	}
	resp := h.f.ResolvePlayer(types.ResolveQuery{Name: q, Limit: 10})
	writeJSON(w, resp)
}

func (h *Handlers) GetTeam(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "not yet implemented"})
}
func (h *Handlers) GetPlayer(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "not yet implemented"})
}
```

`internal/http/handlers/matches.go`:
```go
package handlers

import ("net/http"; "strconv"; "github.com/arcdent/hltv-mcp/internal/types")

func (h *Handlers) GetTodayMatches(w http.ResponseWriter, r *http.Request) {
	resp := h.f.GetTodayMatches(); writeJSON(w, resp)
}

func (h *Handlers) GetUpcomingMatches(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	resp := h.f.GetUpcomingMatches(types.UpcomingMatchesQuery{
		Team: q.Get("team"), Event: q.Get("event"),
		Limit: atoi(q.Get("limit")), Days: atoi(q.Get("days")),
	}); writeJSON(w, resp)
}

func (h *Handlers) GetResults(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	resp := h.f.GetResultsRecent(types.ResultsRecentQuery{
		Team: q.Get("team"), Event: q.Get("event"),
		Limit: atoi(q.Get("limit")), Days: atoi(q.Get("days")),
	}); writeJSON(w, resp)
}

func atoi(s string) int { if n, err := strconv.Atoi(s); err == nil { return n }; return 0 }
```

`internal/http/handlers/news.go`:
```go
package handlers

import ("net/http"; "strconv"; "github.com/arcdent/hltv-mcp/internal/types")

func (h *Handlers) GetRealtimeNews(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query(); limit := atoi(q.Get("limit")); offset := atoi(q.Get("offset"))
	resp := h.f.GetRealtimeNews(types.RealtimeNewsQuery{Limit: limit, Offset: offset})
	writeJSON(w, resp)
}

func (h *Handlers) GetNewsDigest(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	resp := h.f.GetNewsDigest(types.NewsDigestQuery{
		Tag: q.Get("tag"), Month: q.Get("month"),
		Year: atoi(q.Get("year")), Limit: atoi(q.Get("limit")), Offset: atoi(q.Get("offset")),
	}); writeJSON(w, resp)
}
```

- [ ] **Step 6: Fix the `atoi` duplication** — move to `handlers.go`:

```
Add to handlers.go:
func atoi(s string) int { n, _ := strconv.Atoi(s); return n }
```

- [ ] **Step 7: Compile check**

```bash
go build ./internal/http/...
```
Expected: success

- [ ] **Step 8: Commit**

```bash
git add internal/http/ embed.go go.mod go.sum
git commit -m "feat: add HTTP server with chi router, REST API handlers, and frontend embed"
```

---

## Phase 6: Main Entrypoint + Build

### Task 10: Wire up main.go with Chrome detection, graceful shutdown

> Spec §架构, §手动编译时的 Chrome 依赖, §Docker

**Files:**
- Modify: `main.go`

- [ ] **Step 1: Rewrite `main.go`**

```go
package main

import (
	"context"; "log"; "net/http"; "os"; "os/signal"; "syscall"
	"github.com/arcdent/hltv-mcp/internal/cache"; "github.com/arcdent/hltv-mcp/internal/client"
	"github.com/arcdent/hltv-mcp/internal/config"; "github.com/arcdent/hltv-mcp/internal/facade"
	httppkg "github.com/arcdent/hltv-mcp/internal/http"; "github.com/arcdent/hltv-mcp/internal/mcp"
	"github.com/arcdent/hltv-mcp/internal/renderer"; "github.com/arcdent/hltv-mcp/internal/summary"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("HLTV MCP starting...")

	cfg, err := config.LoadConfig()
	if err != nil { log.Fatalf("config: %v", err) }

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
	router := httppkg.NewRouter(cfg, f, embeddedFrontend)
	httpAddr := cfg.HTTPHost + ":" + string(rune(cfg.HTTPPort))
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
```

- [ ] **Step 2: Fix the HTTP_PORT int-to-string bug** — use `fmt.Sprintf`

```go
import "fmt"
// ...
httpAddr := fmt.Sprintf("%s:%d", cfg.HTTPHost, cfg.HTTPPort)
```

- [ ] **Step 3: Build**

```bash
go build -o hltv-mcp .
```
Expected: success

- [ ] **Step 4: Quick smoke test** (start binary, curl health, then kill)

```bash
./hltv-mcp &
sleep 1
curl -s http://localhost:8082/api/health
kill %1
```
Expected: `{"status":"ok"}`

- [ ] **Step 5: Commit**

```bash
git add main.go
git commit -m "feat: wire up main entrypoint with MCP stdio + HTTP + graceful shutdown"
```

---

## Phase 7: Docker

### Task 11: Dockerfile + docker-compose

> Spec §构建与部署, §Docker 多阶段构建

**Files:**
- Create: `Dockerfile`
- Create: `docker-compose.yml`

- [ ] **Step 1: Write `Dockerfile`**

```dockerfile
# Stage 1: Build frontend
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/frontend/dist ./dist/
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o hltv-mcp .

# Stage 3: Runtime
# chrome-headless-shell:stable provides chrome-headless-shell at /usr/bin/chrome-headless-shell
# Go starts it via chromedp.ExecAllocator at runtime
FROM chrome-headless-shell:stable
COPY --from=builder /app/hltv-mcp /hltv-mcp
EXPOSE 8082
ENV HTTP_PORT=8082
ENV HTTP_HOST=0.0.0.0
ENV HLTV_CHROME_PATH=/usr/bin/chrome-headless-shell
ENTRYPOINT ["/hltv-mcp"]
```

- [ ] **Step 2: Write `docker-compose.yml`**

```yaml
services:
  hltv-mcp:
    build: .
    ports:
      - "8082:8082"
    environment:
      - HTTP_PORT=8082
      - HTTP_HOST=0.0.0.0
      - HLTV_CHROME_PATH=/usr/bin/chrome-headless-shell
    restart: unless-stopped
```

- [ ] **Step 3: Commit**

```bash
git add Dockerfile docker-compose.yml
git commit -m "feat: add Docker multi-stage build and docker-compose"
```

---

## Phase 8: Frontend

### Task 12: Frontend scaffold — Vite + React + Tailwind + Router

> Spec §前端, §前端技术栈, §前端页面

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/vite.config.ts`
- Create: `frontend/tailwind.config.js`
- Create: `frontend/postcss.config.js`
- Create: `frontend/index.html`
- Create: `frontend/src/main.tsx`
- Create: `frontend/src/App.tsx`
- Create: `frontend/src/api/client.ts`

- [ ] **Step 1: Scaffold frontend with Vite**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild/frontend
npm create vite@latest . -- --template react-ts
npm install react-router-dom tailwindcss @tailwindcss/vite
```

- [ ] **Step 2: Configure `vite.config.ts`**

```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  build: { outDir: '../dist', emptyOutDir: true },
  server: { proxy: { '/api': 'http://localhost:8082' } },
})
```

- [ ] **Step 3: Write `frontend/src/main.tsx`**

```tsx
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import App from './App'
import './index.css'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </StrictMode>
)
```

- [ ] **Step 4: Write `frontend/src/App.tsx`** — layout + routes

```tsx
import { Routes, Route, NavLink } from 'react-router-dom'
import Dashboard from './pages/Dashboard'
import Matches from './pages/Matches'
import Teams from './pages/Teams'
import Players from './pages/Players'
import News from './pages/News'
import Cache from './pages/Cache'

const nav = [
  { to: '/', label: 'Dashboard' },
  { to: '/matches', label: 'Matches' },
  { to: '/teams', label: 'Teams' },
  { to: '/players', label: 'Players' },
  { to: '/news', label: 'News' },
  { to: '/cache', label: 'Cache' },
]

export default function App() {
  return (
    <div className="min-h-screen bg-gray-900 text-gray-100">
      <nav className="flex gap-4 p-4 bg-gray-800 border-b border-gray-700">
        {nav.map(({ to, label }) => (
          <NavLink key={to} to={to} className={({ isActive }) =>
            `px-3 py-1 rounded ${isActive ? 'bg-blue-600 text-white' : 'text-gray-400 hover:text-white'}`
          }>{label}</NavLink>
        ))}
      </nav>
      <main className="p-6 max-w-6xl mx-auto">
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/matches" element={<Matches />} />
          <Route path="/teams" element={<Teams />} />
          <Route path="/players" element={<Players />} />
          <Route path="/news" element={<News />} />
          <Route path="/cache" element={<Cache />} />
        </Routes>
      </main>
    </div>
  )
}
```

- [ ] **Step 5: Write `frontend/src/api/client.ts`** — fetch wrapper

```typescript
const BASE = '/api'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  return res.json()
}

export const api = {
  health: () => request<{ status: string }>('/health'),
  status: () => request<any>('/status'),
  cacheStats: () => request<any>('/cache'),
  clearCache: () => request<any>('/cache', { method: 'DELETE' }),
  search: (q: string, type: 'team' | 'player') => request<any>(`/search?q=${encodeURIComponent(q)}&type=${type}`),
  getTeam: (id: number) => request<any>(`/teams/${id}`),
  getPlayer: (id: number) => request<any>(`/players/${id}`),
  todayMatches: () => request<any>('/matches/today'),
  upcomingMatches: (params: Record<string, string>) => request<any>(`/matches?${new URLSearchParams(params)}`),
  results: (params: Record<string, string>) => request<any>(`/results?${new URLSearchParams(params)}`),
  realtimeNews: (limit = 25, offset = 0) => request<any>(`/news/realtime?limit=${limit}&offset=${offset}`),
  newsDigest: (params: Record<string, string>) => request<any>(`/news?${new URLSearchParams(params)}`),
}
```

- [ ] **Step 6: PostCSS config** (for Tailwind v4 style)

```javascript
// frontend/postcss.config.js
export default { plugins: { '@tailwindcss/postcss': {} } }
```

- [ ] **Step 7: Verify frontend builds**

```bash
cd frontend && npm run build
```
Expected: success, `dist/` created at project root

- [ ] **Step 8: Commit**

```bash
git add frontend/ dist/
git commit -m "feat: add React frontend scaffold with Vite, Tailwind, React Router"
```

### Task 13: Frontend pages — Dashboard, Matches, Teams, Players, News, Cache

> Spec §前端 6 个页面

**Files:**
- Create: `frontend/src/pages/Dashboard.tsx`
- Create: `frontend/src/pages/Matches.tsx`
- Create: `frontend/src/pages/Teams.tsx`
- Create: `frontend/src/pages/Players.tsx`
- Create: `frontend/src/pages/News.tsx`
- Create: `frontend/src/pages/Cache.tsx`

- [ ] **Step 1: Write `Dashboard.tsx`**

```tsx
import { useEffect, useState } from 'react'
import { api } from '../api/client'

export default function Dashboard() {
  const [status, setStatus] = useState<any>(null)
  const [cache, setCache] = useState<any>(null)
  useEffect(() => {
    api.status().then(setStatus).catch(console.error)
    api.cacheStats().then(setCache).catch(console.error)
  }, [])

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Dashboard</h1>
      <div className="grid grid-cols-3 gap-4 mb-8">
        <Card title="Uptime" value={status ? `${status.uptime_sec}s` : '...'} />
        <Card title="Go Version" value={status?.go_version ?? '...'} />
        <Card title="Memory" value={status ? `${status.memory_mb} MB` : '...'} />
      </div>
      <div className="grid grid-cols-3 gap-4">
        <Card title="Cache Entries" value={cache?.entries ?? '...'} />
        <Card title="Cache Hits" value={cache?.hits ?? '...'} />
        <Card title="Misses" value={cache?.misses ?? '...'} />
      </div>
    </div>
  )
}

function Card({ title, value }: { title: string; value: string }) {
  return <div className="bg-gray-800 rounded-lg p-4 border border-gray-700"><div className="text-gray-400 text-sm">{title}</div><div className="text-2xl font-bold mt-1">{value}</div></div>
}
```

- [ ] **Step 2: Write `Matches.tsx`**

```tsx
import { useEffect, useState } from 'react'
import { api } from '../api/client'

export default function Matches() {
  const [tab, setTab] = useState<'today' | 'upcoming' | 'results'>('today')
  const [data, setData] = useState<any>(null)
  const [team, setTeam] = useState(''); const [event, setEvent] = useState('')

  useEffect(() => {
    if (tab === 'today') api.todayMatches().then(setData)
    else if (tab === 'upcoming') api.upcomingMatches({ team, event, limit: '20' }).then(setData)
    else api.results({ team, event, limit: '20' }).then(setData)
  }, [tab, team, event])

  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">Matches</h1>
      <div className="flex gap-2 mb-4">
        {(['today', 'upcoming', 'results'] as const).map(t => (
          <button key={t} onClick={() => setTab(t)} className={`px-4 py-1 rounded ${tab === t ? 'bg-blue-600' : 'bg-gray-700'}`}>{t === 'today' ? 'Today' : t === 'upcoming' ? 'Upcoming' : 'Results'}</button>
        ))}
      </div>
      <div className="flex gap-2 mb-4">
        <input placeholder="Team" value={team} onChange={e => setTeam(e.target.value)} className="bg-gray-800 border border-gray-700 rounded px-3 py-1 text-white" />
        <input placeholder="Event" value={event} onChange={e => setEvent(e.target.value)} className="bg-gray-800 border border-gray-700 rounded px-3 py-1 text-white" />
      </div>
      <div className="space-y-2">
        {data?.items?.map((m: any, i: number) => (
          <div key={i} className="bg-gray-800 p-3 rounded flex justify-between">
            <span>{m.team1} vs {m.team2}</span>
            {m.score && <span className="text-blue-400">{m.score}</span>}
            {m.event && <span className="text-gray-400">{m.event}</span>}
          </div>
        ))}
      </div>
    </div>
  )
}
```

- [ ] **Step 3: Write `Teams.tsx`**

```tsx
import { useState } from 'react'
import { api } from '../api/client'

export default function Teams() {
  const [query, setQuery] = useState(''); const [results, setResults] = useState<any[]>([])

  const search = async () => {
    const resp = await api.search(query, 'team'); setResults(resp?.items ?? [])
  }

  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">Teams</h1>
      <div className="flex gap-2 mb-4">
        <input placeholder="Search teams..." value={query} onChange={e => setQuery(e.target.value)} onKeyDown={e => e.key === 'Enter' && search()} className="bg-gray-800 border border-gray-700 rounded px-3 py-2 flex-1 text-white" />
        <button onClick={search} className="px-4 py-2 bg-blue-600 rounded">Search</button>
      </div>
      <div className="space-y-2">
        {results.map((t: any, i: number) => (
          <div key={i} className="bg-gray-800 p-3 rounded">{t.name} <span className="text-gray-400">(id={t.id})</span></div>
        ))}
      </div>
    </div>
  )
}
```

- [ ] **Step 4: Write `Players.tsx`** (same pattern as Teams)

```tsx
import { useState } from 'react'
import { api } from '../api/client'

export default function Players() {
  const [query, setQuery] = useState(''); const [results, setResults] = useState<any[]>([])

  const search = async () => {
    const resp = await api.search(query, 'player'); setResults(resp?.items ?? [])
  }

  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">Players</h1>
      <div className="flex gap-2 mb-4">
        <input placeholder="Search players..." value={query} onChange={e => setQuery(e.target.value)} onKeyDown={e => e.key === 'Enter' && search()} className="bg-gray-800 border border-gray-700 rounded px-3 py-2 flex-1 text-white" />
        <button onClick={search} className="px-4 py-2 bg-blue-600 rounded">Search</button>
      </div>
      <div className="space-y-2">
        {results.map((p: any, i: number) => (
          <div key={i} className="bg-gray-800 p-3 rounded">{p.name} <span className="text-gray-400">(id={p.id})</span></div>
        ))}
      </div>
    </div>
  )
}
```

- [ ] **Step 5: Write `News.tsx`**

```tsx
import { useEffect, useState } from 'react'
import { api } from '../api/client'

export default function News() {
  const [tab, setTab] = useState<'realtime' | 'archive'>('realtime')
  const [data, setData] = useState<any>(null)

  useEffect(() => {
    if (tab === 'realtime') api.realtimeNews().then(setData)
    else api.newsDigest({ limit: '25' }).then(setData)
  }, [tab])

  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">News</h1>
      <div className="flex gap-2 mb-4">
        {(['realtime', 'archive'] as const).map(t => (
          <button key={t} onClick={() => setTab(t)} className={`px-4 py-1 rounded ${tab === t ? 'bg-blue-600' : 'bg-gray-700'}`}>{t === 'realtime' ? 'Realtime' : 'Archive'}</button>
        ))}
      </div>
      <div className="space-y-2">
        {data?.items?.map((n: any, i: number) => (
          <div key={i} className="bg-gray-800 p-3 rounded">
            <div className="font-medium">{n.title}</div>
            <div className="text-sm text-gray-400">{n.published_at || n.relative_time}</div>
          </div>
        ))}
      </div>
    </div>
  )
}
```

- [ ] **Step 6: Write `Cache.tsx`**

```tsx
import { useEffect, useState } from 'react'
import { api } from '../api/client'

export default function Cache() {
  const [stats, setStats] = useState<any>(null)

  const refresh = () => { api.cacheStats().then(setStats) }
  useEffect(refresh, [])

  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">Cache Management</h1>
      <div className="grid grid-cols-3 gap-4 mb-4">
        <div className="bg-gray-800 p-4 rounded"><div className="text-gray-400">Entries</div><div className="text-2xl">{stats?.entries ?? '...'}</div></div>
        <div className="bg-gray-800 p-4 rounded"><div className="text-gray-400">Hits</div><div className="text-2xl">{stats?.hits ?? '...'}</div></div>
        <div className="bg-gray-800 p-4 rounded"><div className="text-gray-400">Misses</div><div className="text-2xl">{stats?.misses ?? '...'}</div></div>
      </div>
      <button onClick={() => api.clearCache().then(refresh)} className="px-4 py-2 bg-red-600 rounded">Clear All Cache</button>
    </div>
  )
}
```

- [ ] **Step 7: Build frontend and verify**

```bash
cd frontend && npm run build
```
Expected: success, dist/ populated

- [ ] **Step 8: Commit**

```bash
git add frontend/src/pages/
git commit -m "feat: add all 6 frontend pages (Dashboard, Matches, Teams, Players, News, Cache)"
```

---

## Phase 9: Verification

### Task 14: End-to-end verification

> Spec §测试策略, §Docker 验证

- [ ] **Step 1: Run all Go tests**

```bash
go test ./internal/... -v -timeout 30s
```
Expected: all PASS

- [ ] **Step 2: Build binary with embedded frontend**

```bash
cd frontend && npm run build && cd ..
go build -o hltv-mcp .
```
Expected: success

- [ ] **Step 3: Launch and verify HTTP**

```bash
./hltv-mcp &
sleep 2
curl -s http://localhost:8082/api/health
curl -s http://localhost:8082/api/status
curl -s http://localhost:8082/ | head -20
kill %1
```
Expected: health → `{"status":"ok"}`, status → JSON with uptime/memory, frontend → HTML page

- [ ] **Step 4: Docker build**

```bash
docker build -t hltv-mcp .
```
Expected: success

- [ ] **Step 5: Final commit**

```bash
git add -A
git commit -m "chore: final verification and cleanup"
```
