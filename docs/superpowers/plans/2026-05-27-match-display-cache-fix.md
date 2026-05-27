# 比赛中文化、BO1归一化、选手缓存与统计修复 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 赛程占位符中文化、选手近期比赛 BO1 比分归一化、选手详情 7 天缓存、修复缓存统计硬编码为 0

**Architecture:** 纯函数（`TranslatePlaceholder`/`normalizeBO1Score`）→ normalizer 集成 → facade 缓存封装 → HTTP handler 接入 → 前端兜底。每层独立可测。

**Tech Stack:** Go 1.26, chromedp, goquery, React/TypeScript, sync/atomic

**Spec:** `docs/superpowers/specs/2026-05-27-match-display-cache-fix.md`

---

## 文件修改清单

| 文件 | 操作 | 任务 |
|------|------|------|
| `internal/cache/cache.go` | 修改 | Task 1 |
| `internal/cache/cache_test.go` | 修改 | Task 1 |
| `internal/normalizer/match.go` | 修改 | Task 2, 3 |
| `internal/normalizer/player.go` | 修改 | Task 2, 3 |
| `internal/normalizer/normalizer_test.go` | 修改 | Task 2 |
| `internal/config/config.go` | 修改 | Task 4 |
| `internal/facade/facade.go` | 修改 | Task 4 |
| `internal/http/handlers/search.go` | 修改 | Task 5 |
| `internal/http/handlers/status.go` | 修改 | Task 5 |
| `frontend/src/pages/Matches.tsx` | 修改 | Task 6 |
| `frontend/src/components/PlayerDetail.tsx` | 修改 | Task 6 |

---

### Task 1: Cache 统计计数器

**Files:**
- Modify: `internal/cache/cache.go`
- Modify: `internal/cache/cache_test.go`

**Target Spec Section:** Bug 修复：缓存统计始终为 0 — Cache 层：计数器

- [ ] **Step 1: 写失败测试**

在 `internal/cache/cache_test.go` 末尾追加：

```go
func TestHitsMisses(t *testing.T) {
	c := New(100, 3600)

	// Fresh miss on empty cache
	c.Get("missing")
	if c.Misses() != 1 {
		t.Errorf("expected 1 miss, got %d", c.Misses())
	}
	if c.Hits() != 0 {
		t.Errorf("expected 0 hits, got %d", c.Hits())
	}

	// Hit after set
	c.Set("k", "v", 10)
	c.Get("k")
	if c.Hits() != 1 {
		t.Errorf("expected 1 hit, got %d", c.Hits())
	}
	if c.Misses() != 1 {
		t.Errorf("expected 1 miss, got %d", c.Misses())
	}

	// GetStale does NOT increment hit or miss
	c.Set("stale", "sv", 0)
	time.Sleep(10 * time.Millisecond)
	c.GetStale("stale")
	if c.Hits() != 1 {
		t.Errorf("GetStale should not increment hits, got %d", c.Hits())
	}
	if c.Misses() != 1 {
		t.Errorf("GetStale should not increment misses, got %d", c.Misses())
	}
}

func TestClearResetsCounters(t *testing.T) {
	c := New(100, 3600)
	c.Set("k", "v", 10)
	c.Get("k")
	c.Get("missing")
	if c.Hits() != 1 || c.Misses() != 1 {
		t.Fatal("setup failed")
	}
	c.Clear()
	if c.Hits() != 0 {
		t.Errorf("expected 0 hits after clear, got %d", c.Hits())
	}
	if c.Misses() != 0 {
		t.Errorf("expected 0 misses after clear, got %d", c.Misses())
	}
	if c.Entries() != 0 {
		t.Errorf("expected 0 entries after clear, got %d", c.Entries())
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go test ./internal/cache/ -v -run "TestHitsMisses|TestClearResetsCounters"
```

预期：FAIL — `c.Hits undefined` / `c.Misses undefined`

- [ ] **Step 3: 实现计数器**

在 `internal/cache/cache.go` 中修改 `Cache` 结构体，添加字段和方法：

```go
import (
	"sync"
	"sync/atomic"
	"time"
)

type Cache struct {
	mu         sync.RWMutex
	store      map[string]*entry
	inFlight   map[string]*inflightEntry
	maxEntries int
	maxStale   time.Duration
	hits       atomic.Int64
	misses     atomic.Int64
}
```

修改 `Get` 方法（添加计数）：

```go
func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	e, ok := c.store[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(e.expiresAt) {
		c.misses.Add(1)
		return nil, false
	}
	c.hits.Add(1)
	return e.value, true
}
```

修改 `Clear` 方法（重置计数器）：

```go
func (c *Cache) Clear() {
	c.mu.Lock()
	c.store = make(map[string]*entry)
	c.mu.Unlock()
	c.hits.Store(0)
	c.misses.Store(0)
}
```

在文件末尾新增方法：

```go
func (c *Cache) Hits() int64   { return c.hits.Load() }
func (c *Cache) Misses() int64 { return c.misses.Load() }
```

**注意：`GetStale` 不添加计数**（stale 是降级兜底，不改变计数器）。

- [ ] **Step 4: 运行测试确认通过**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go test ./internal/cache/ -v -run "TestHitsMisses|TestClearResetsCounters|TestSetGet|TestGetExpired|TestGetStale|TestRunOnceDedup|TestEvictOverflow"
```

预期：全部 PASS

- [ ] **Step 5: Commit**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild
git add internal/cache/cache.go internal/cache/cache_test.go
git commit -m "feat: add hits/misses counters to cache, reset on Clear, GetStale not counted"
```

---

### Task 2: TranslatePlaceholder + normalizeBO1Score 纯函数

**Files:**
- Modify: `internal/normalizer/match.go` (add `TranslatePlaceholder`)
- Modify: `internal/normalizer/player.go` (add `normalizeBO1Score`)
- Modify: `internal/normalizer/normalizer_test.go` (add tests)

**Target Spec Sections:** 需求 1 匹配策略 + Go 共用映射函数、需求 2 归一化函数

- [ ] **Step 1: 写失败测试**

在 `internal/normalizer/normalizer_test.go` 末尾追加：

```go
func TestTranslatePlaceholder(t *testing.T) {
	tests := []struct{ in, want string }{
		{"winner", "胜者"},
		{"Winner", "胜者"},
		{"Winner of Group A", "胜者"},
		{"WINNER", "胜者"},
		{"loser", "败者"},
		{"Loser of match 3", "败者"},
		{"tbd", "待定"},
		{"TBD", "待定"},
		{"  tbd  ", "待定"},
		{"Vitality", "Vitality"},
		{"FaZe Clan", "FaZe Clan"},
	}
	for _, tt := range tests {
		if got := TranslatePlaceholder(tt.in); got != tt.want {
			t.Errorf("TranslatePlaceholder(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestNormalizeBO1Score(t *testing.T) {
	tests := []struct{ in, want string }{
		{"2:0", "2:0"},
		{"2:1", "2:1"},
		{"0:2", "0:2"},
		{"13:5", "1:0"},
		{"5:13", "0:1"},
		{"16:14", "1:0"},
		{"14:16", "0:1"},
		{"16:16", "平局"},
		{"13:11", "1:0"},
		{"11:13", "0:1"},
		{"", ""},
		{"invalid", "invalid"},
		{"13 : 5", "1:0"},
		{"5 : 13", "0:1"},
	}
	for _, tt := range tests {
		if got := normalizeBO1Score(tt.in); got != tt.want {
			t.Errorf("normalizeBO1Score(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go test ./internal/normalizer/ -v -run "TestTranslatePlaceholder|TestNormalizeBO1Score"
```

预期：FAIL — `undefined: TranslatePlaceholder` / `undefined: normalizeBO1Score`

- [ ] **Step 3: 实现纯函数**

在 `internal/normalizer/match.go` 末尾新增：

```go
// TranslatePlaceholder maps HLTV bracket placeholder team names to Chinese
func TranslatePlaceholder(s string) string {
	lower := strings.ToLower(strings.TrimSpace(s))
	if lower == "" {
		return s
	}
	if strings.Contains(lower, "winner") {
		return "胜者"
	}
	if strings.Contains(lower, "loser") {
		return "败者"
	}
	if strings.Contains(lower, "tbd") {
		return "待定"
	}
	return s
}
```

在 `internal/normalizer/player.go` `import` 块中添加 `"strconv"`（已存在则跳过），在文件末尾新增：

```go
// normalizeBO1Score converts BO1 match scores (e.g. "13:5") to "1:0"/"0:1"
// If both sides < 13, returns the original score unchanged.
func normalizeBO1Score(score string) string {
	parts := strings.SplitN(score, ":", 2)
	if len(parts) != 2 {
		return score
	}
	a, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	b, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
	if a >= 13 || b >= 13 {
		if a > b {
			return "1:0"
		}
		if b > a {
			return "0:1"
		}
		return "平局"
	}
	return score
}
```

- [ ] **Step 4: 运行测试确认通过**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go test ./internal/normalizer/ -v -run "TestTranslatePlaceholder|TestNormalizeBO1Score"
```

预期：全部 PASS

- [ ] **Step 5: 运行全部 normalizer 测试确认无回归**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go test ./internal/normalizer/ -v
```

预期：全部 PASS

- [ ] **Step 6: Commit**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild
git add internal/normalizer/match.go internal/normalizer/player.go internal/normalizer/normalizer_test.go
git commit -m "feat: add TranslatePlaceholder and normalizeBO1Score pure functions"
```

---

### Task 3: Normalizer 集成调用

**Files:**
- Modify: `internal/normalizer/match.go` (NormalizeUpcomingMatches)
- Modify: `internal/normalizer/player.go` (NormalizePlayerDetail)

**Target Spec Sections:** 需求 1 调用点 A/B、需求 2 两处归一化调用

- [ ] **Step 1: 集成调用**

在 `internal/normalizer/match.go` 的 `NormalizeUpcomingMatches` 中，`m.Team1` / `m.Team2` 赋值之后、`if m.Team1 != "" && m.Team2 != ""` 判断之前（约第 112 行后、第 131 行前），增加：

```go
			m.Team1 = TranslatePlaceholder(m.Team1)
			m.Team2 = TranslatePlaceholder(m.Team2)
```

完整上下文（修改后）：

```go
			if perspective != "" {
				if m.Team1 == perspective {
					m.Opponent = m.Team2
				} else if m.Team2 == perspective {
					m.Opponent = m.Team1
				}
			}

			m.Team1 = TranslatePlaceholder(m.Team1)
			m.Team2 = TranslatePlaceholder(m.Team2)

			if m.Team1 != "" && m.Team2 != "" {
				matches = append(matches, m)
			}
```

在 `internal/normalizer/player.go` 的 `NormalizePlayerDetail` 中：

**位置 1** — 第 204 行，`m.Score = strings.ReplaceAll(...)` 之后增加：

```go
			m.Score = normalizeBO1Score(m.Score)
```

**位置 2** — 约第 205 行，`m.Opponent` 赋值后增加：

```go
			m.Opponent = TranslatePlaceholder(m.Opponent)
```

**位置 3** — 约第 208 行，`m.Team = pd.Profile.Team` 后增加：

```go
			m.Team = TranslatePlaceholder(m.Team)
```

**位置 4** — 第 221 行（fallback 路径），`m.Score = strings.ReplaceAll(...)` 之后增加：

```go
				m.Score = normalizeBO1Score(m.Score)
```

- [ ] **Step 2: 运行测试确认无回归**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go test ./internal/normalizer/ -v
```

预期：全部 PASS

- [ ] **Step 3: 编译检查**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build github.com/arcdent/hltv-mcp
```

预期：编译成功

- [ ] **Step 4: Commit**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild
git add internal/normalizer/match.go internal/normalizer/player.go
git commit -m "feat: integrate TranslatePlaceholder and normalizeBO1Score into normalizers"
```

---

### Task 4: Config + Facade 选手详情缓存 + 缓存暴露

**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/facade/facade.go`

**Target Spec Sections:** 需求 3 Config + Facade、Bug 修复 Facade 层

- [ ] **Step 1: 添加 Config 字段**

在 `internal/config/config.go` 的 `Config` 结构体中，紧跟 `CacheTTLRealtimeNews` 后添加：

```go
	CacheTTLPlayerDetail int
```

在 `LoadConfig()` 返回的 Config 字面量中，紧跟 `CacheTTLRealtimeNews` 行后添加：

```go
			CacheTTLPlayerDetail: envInt("CACHE_TTL_PLAYER_DETAIL_SEC", 604800),
```

- [ ] **Step 2: Facade 新增方法**

在 `internal/facade/facade.go` 的 import 块中添加：

```go
	"github.com/arcdent/hltv-mcp/internal/normalizer"
```

在 `ScrapePlayerDetail` 方法后新增 `GetPlayerDetailCached`：

```go
// GetPlayerDetailCached returns cached player detail, or scrapes via chromedp and caches for 7 days
func (f *HltvFacade) GetPlayerDetailCached(ctx context.Context, id int, slug string) (types.PlayerDetail, error) {
	if slug == "" {
		slug = fmt.Sprintf("player-%d", id)
	}
	key := fmt.Sprintf("player_detail:%d", id)
	if cached, ok := f.cache.Get(key); ok {
		return cached.(types.PlayerDetail), nil
	}
	doc, err := f.ps.GetPlayer(ctx, id, slug)
	if err != nil {
		return types.PlayerDetail{}, err
	}
	pd := normalizer.NormalizePlayerDetail(doc)
	pd.Profile.ID = id
	f.cache.Set(key, pd, f.cfg.CacheTTLPlayerDetail)
	return pd, nil
}
```

在文件末尾新增缓存暴露方法：

```go
// CacheEntries returns the number of entries currently in the cache
func (f *HltvFacade) CacheEntries() int { return f.cache.Entries() }

// CacheHits returns the cumulative cache hit count
func (f *HltvFacade) CacheHits() int64 { return f.cache.Hits() }

// CacheMisses returns the cumulative cache miss count
func (f *HltvFacade) CacheMisses() int64 { return f.cache.Misses() }

// ClearCache clears all cached entries and resets counters
func (f *HltvFacade) ClearCache() { f.cache.Clear() }
```

- [ ] **Step 3: 编译检查**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build github.com/arcdent/hltv-mcp
```

预期：编译成功

- [ ] **Step 4: Commit**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild
git add internal/config/config.go internal/facade/facade.go
git commit -m "feat: add player detail 7d cache and cache stats exposure methods"
```

---

### Task 5: HTTP Handlers 接入

**Files:**
- Modify: `internal/http/handlers/search.go` (GetPlayer)
- Modify: `internal/http/handlers/status.go` (GetCacheStats, ClearCache)

**Target Spec Sections:** 需求 3 HTTP Handler、Bug 修复 HTTP Handler 层

- [ ] **Step 1: 修改 GetPlayer 使用缓存**

将 `internal/http/handlers/search.go` 中 `GetPlayer` 方法替换为：

```go
func (h *Handlers) GetPlayer(w http.ResponseWriter, r *http.Request) {
	id := atoi(chi.URLParam(r, "id"))
	if id == 0 {
		writeError(w, http.StatusBadRequest, "invalid player id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	pd, err := h.f.GetPlayerDetailCached(ctx, id, "")
	if err != nil {
		writeJSON(w, map[string]any{
			"error": map[string]any{"code": "UPSTREAM_UNAVAILABLE", "message": "详情暂时不可用"},
			"meta":  map[string]any{"partial": true},
		})
		return
	}
	writeJSON(w, map[string]any{"data": pd, "meta": map[string]any{"partial": false}})
}
```

同时移除不再需要的 import：
- 删除 `"github.com/arcdent/hltv-mcp/internal/normalizer"`（如果存在）
- 删除 `"github.com/PuerkitoBio/goquery"`（如果存在且仅被 `GetPlayer` 使用）

**实际检查后**：原 `GetPlayer` 使用了 `normalizer.NormalizePlayerDetail(doc)` 且最终调用 `pd.Profile.ID = id`，现在 facade 层已封装。确认 import 中是否需要移除 normalizer。

保留现有 imports（`"context"`, `"net/http"`, `"time"`, `"github.com/go-chi/chi/v5"`），其余按实际引用保留/删除。

- [ ] **Step 2: 修改 GetCacheStats 和 ClearCache 使用真实数据**

将 `internal/http/handlers/status.go` 中的 `GetCacheStats` 替换为：

```go
func (h *Handlers) GetCacheStats(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"entries": h.f.CacheEntries(),
		"hits":    h.f.CacheHits(),
		"misses":  h.f.CacheMisses(),
	})
}
```

将 `ClearCache` 替换为：

```go
func (h *Handlers) ClearCache(w http.ResponseWriter, r *http.Request) {
	h.f.ClearCache()
	writeJSON(w, map[string]string{"status": "cleared"})
}
```

- [ ] **Step 3: 编译检查**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build github.com/arcdent/hltv-mcp
```

预期：编译成功

- [ ] **Step 4: Commit**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild
git add internal/http/handlers/search.go internal/http/handlers/status.go
git commit -m "feat: wire player detail cache and real cache stats into HTTP handlers"
```

---

### Task 6: Frontend 兜底处理

**Files:**
- Modify: `frontend/src/pages/Matches.tsx:111,134`
- Modify: `frontend/src/components/PlayerDetail.tsx:130`

**Target Spec Sections:** 需求 1 前端层

- [ ] **Step 1: Matches.tsx — TBD → 待定兜底**

修改 `frontend/src/pages/Matches.tsx`：

第 111 行，将：
```tsx
                    {m.team1 ?? 'TBD'}
```
改为：
```tsx
                    {m.team1 || '待定'}
```

第 134 行，将：
```tsx
                    {m.team2 ?? 'TBD'}
```
改为：
```tsx
                    {m.team2 || '待定'}
```

- [ ] **Step 2: PlayerDetail.tsx — team/opponent 待定兜底**

修改 `frontend/src/components/PlayerDetail.tsx`，第 130 行，将：
```tsx
                      <span style={{fontWeight:600}}>{m.team}</span> <span style={{color:'var(--text-muted)'}}>vs</span> {m.opponent}
```
改为：
```tsx
                      <span style={{fontWeight:600}}>{m.team || '待定'}</span> <span style={{color:'var(--text-muted)'}}>vs</span> {m.opponent || '待定'}
```

- [ ] **Step 3: 前端构建检查**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild/frontend && npm run build
```

预期：构建成功，无 TS 错误

- [ ] **Step 4: Commit**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild
git add frontend/src/pages/Matches.tsx frontend/src/components/PlayerDetail.tsx
git commit -m "feat: add Chinese fallback for TBD/empty team names in frontend"
```

---

### Task 7: 全量验证

**Files:** 无新建

- [ ] **Step 1: 运行全部 Go 测试**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go test ./internal/... -v
```

预期：全部 PASS

- [ ] **Step 2: Go 编译**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build github.com/arcdent/hltv-mcp
```

预期：编译成功

- [ ] **Step 3: 前端编译**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild/frontend && npm run build
```

预期：构建成功

- [ ] **Step 4: Commit（如有遗漏文件）**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild
git status
# 仅当有未提交变更时执行：
git add <files> && git commit -m "chore: final verification adjustments"
```
