# 比赛显示中文化、BO1 归一化、选手缓存与缓存统计修复

> **Goal:** 赛程占位符中文化、选手近期比赛 BO1 比分归一化、选手详情 chromedp 缓存、修复缓存统计始终为 0 的 bug

**Tech Stack:** Go 1.26, TypeScript/React, chromedp, goquery

---

## 需求 1：占位符中文化（含 winner/loser/TBD）

HLTV 锦标赛 bracket 阶段，未确定的队伍名在 HTML 中是 `winner`/`loser`/`tbd` 等英文占位文本，可能是 `"Winner of Group A"` 这种带后缀形式。需在所有数据出口做包含匹配 → 中文映射。

### 匹配策略

**包含匹配**（不区分大小写）：

| 包含文本 | 映射为 |
|---------|--------|
| `winner` | `胜者` |
| `loser`  | `败者` |
| `tbd`    | `待定` |

优先级：winner > loser > tbd（避免 `winner` 和 `loser` 同时包含对方的情况）

### Go: 共用映射函数

**文件：** `internal/normalizer/match.go`

新增导出函数：

```go
func TranslatePlaceholder(s string) string {
    lower := strings.ToLower(strings.TrimSpace(s))
    if lower == "" { return s }
    if strings.Contains(lower, "winner") { return "胜者" }
    if strings.Contains(lower, "loser")  { return "败者" }
    if strings.Contains(lower, "tbd")    { return "待定" }
    return s
}
```

### 调用点

**A. 赛程 normalizer** — `internal/normalizer/match.go` `NormalizeUpcomingMatches`：
- `m.Team1 = TranslatePlaceholder(m.Team1)`
- `m.Team2 = TranslatePlaceholder(m.Team2)`

**B. 选手 normalizer** — `internal/normalizer/player.go` `NormalizePlayerDetail`：
- `m.Opponent = TranslatePlaceholder(m.Opponent)`（第 205 行附近）
- `m.Team = TranslatePlaceholder(m.Team)`（第 208 行附近）

### 前端层

**文件：** `frontend/src/pages/Matches.tsx`

- 第 111 行：`{m.team1 ?? 'TBD'}` → `{m.team1 || '待定'}`
- 第 134 行：`{m.team2 ?? 'TBD'}` → `{m.team2 || '待定'}`

用 `||` 覆盖空字符串场景。Normalizer 已处理大部分情况，此处为兜底。

**文件：** `frontend/src/components/PlayerDetail.tsx`

- 第 130 行：`{m.team}` 和 `{m.opponent}` — 数据已由 normalizer 翻译，无需额外处理，但兜底显示空字符串时用 `|| '待定'`：
  ```
  {m.team || '待定'} vs {m.opponent || '待定'}
  ```

---

## 需求 2：BO1 比分归一化

选手近期比赛中，如 `13:5` 一方 >= 13 则判定为 BO1，归一化为 `1:0` / `0:1`。

### 归一化函数

**文件：** `internal/normalizer/player.go`

新增函数，在 `NormalizePlayerDetail` 中两处 `m.Score` 赋值后均调用：

- 第 204 行（主路径 `.playerpage-match-result`）：`m.Score = normalizeBO1Score(m.Score)`
- 第 221 行（fallback 路径 `.result-score`）：`m.Score = normalizeBO1Score(m.Score)`

两处均需调用，保持一致。fallback 路径实际比赛多为 `2:0` / `2:1` 不会触发归一化，但做防御性处理。

```go
func normalizeBO1Score(score string) string {
    parts := strings.SplitN(score, ":", 2)
    if len(parts) != 2 { return score }
    a, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
    b, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
    if a >= 13 || b >= 13 {
        if a > b { return "1:0" }
        if b > a { return "0:1" }
        return "平局" // 16:16 等，实际 CS 比赛不会发生
    }
    return score
}
```

- `2:0` / `2:1` → 保持原样（双方 < 13）
- `13:5` → `1:0`
- `5:13` → `0:1`
- `16:14` → `1:0`（加时，胜者得 1）
- `16:16` → `平局`（实际不会发生，防御性处理）

---

## 需求 3：选手详情缓存

`GetPlayer` HTTP 端点每次触发 chromedp 全量抓取，无缓存。需缓存 7 天。

### Config

**文件：** `internal/config/config.go`

新增字段：

```go
CacheTTLPlayerDetail int
```

默认值：`envInt("CACHE_TTL_PLAYER_DETAIL_SEC", 604800)` (7 天)

### Facade

**文件：** `internal/facade/facade.go`

新增方法，直接使用 `cache.Get`/`cache.Set`（不走 `withCache`，因为 `PlayerDetail` 不是 `*ToolResponse`）。

**注意：需要新增 import `"github.com/arcdent/hltv-mcp/internal/normalizer"` 和 `"fmt"`（fmt 已有）。**

```go
func (f *HltvFacade) GetPlayerDetailCached(ctx context.Context, id int, slug string) (types.PlayerDetail, error) {
    if slug == "" { slug = fmt.Sprintf("player-%d", id) }
    key := fmt.Sprintf("player_detail:%d", id)
    if cached, ok := f.cache.Get(key); ok {
        return cached.(types.PlayerDetail), nil
    }
    doc, err := f.ps.GetPlayer(ctx, id, slug)
    if err != nil { return types.PlayerDetail{}, err }
    pd := normalizer.NormalizePlayerDetail(doc)
    pd.Profile.ID = id
    f.cache.Set(key, pd, f.cfg.CacheTTLPlayerDetail)
    return pd, nil
}
```

### HTTP Handler

**文件：** `internal/http/handlers/search.go`

`GetPlayer` 改为调用 `f.GetPlayerDetailCached`（替换原来直接调 `ScrapePlayerDetail` + 手动 normalize 的模式）。

---

## Bug 修复：缓存统计始终为 0

### 根因

`internal/http/handlers/status.go:37-47` — `GetCacheStats` 和 `ClearCache` 是硬编码桩代码，返回值写死为 0。

### Cache 层：计数器

**文件：** `internal/cache/cache.go`

新增 `hits`/`misses` 字段（`sync/atomic.Int64`）：

- `Get()` 命中时 `hits.Add(1)`，未命中时 `misses.Add(1)`
- `GetStale()` **不计数**（stale 是降级兜底，既非纯 hit 也非纯 miss，避免与 `Get()` 的 miss 重复计数）
- `Clear()` 同时重置计数器：`hits.Store(0)`, `misses.Store(0)`
- 新增方法：`Hits() int64`、`Misses() int64`

### Facade 层：暴露缓存操作

**文件：** `internal/facade/facade.go`

新增方法：

```go
func (f *HltvFacade) CacheEntries() int    { return f.cache.Entries() }
func (f *HltvFacade) CacheHits() int64     { return f.cache.Hits() }
func (f *HltvFacade) CacheMisses() int64   { return f.cache.Misses() }
func (f *HltvFacade) ClearCache()          { f.cache.Clear() }
```

### HTTP Handler 层

**文件：** `internal/http/handlers/status.go`

`GetCacheStats` 和 `ClearCache` 改为调用 facade 方法获取真实数据。

---

## 非回归约束

- 不修改 `PlayerRecentMatch` 或 `NormalizedMatch` 的类型定义
- cache 包公开接口不变（仅内增 counters 和 `Hits()`/`Misses()` 方法）
- `Clear()` 必须重置计数器
- 前端 `/matches` 和 `/players` 页面行为不变
- 已有测试需保持通过
