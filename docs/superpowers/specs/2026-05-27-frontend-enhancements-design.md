# Frontend Enhancements Design

Date: 2026-05-27 | Status: confirmed

## Overview

Four enhancements to the HLTV MCP service:

1. README 标准 Stdio MCP 配置段
2. 队伍详情卡片（点击弹窗，含排名/积分/成就/10场战绩/队员列表）
3. 赛程按赛事名称分组（弹窗模式）
4. 新闻详情弹窗（chromedp 抓取正文 + 无限期缓存 + API 翻译）

Implementation: 方案 A 全栈串联，逐功能垂直打通。

---

## Feature 1: README Stdio MCP Config

**变更文件**: `README.md`

在 "用法示例 > MCP 工具" 部分，于现有 OpenCode 格式之前追加标准 Stdio 格式：

```jsonc
{
  "mcpServers": {
    "hltv": {
      "command": "/path/to/hltv-mcp",
      "args": []
    }
  }
}
```

说明文字：标准 MCP 客户端（Claude Desktop、VS Code Copilot、Gemini CLI）用此格式；OpenCode 用下方的 `"type": "local"` 格式。两段并存。

---

## Feature 2: Team Detail Card

### Backend

**New types** (`internal/types/types.go`):

```go
type TeamDetail struct {
    Profile       TeamDetailProfile  `json:"profile"`
    Ranking       TeamRanking        `json:"ranking"`
    Stats         TeamStats          `json:"stats"`
    Achievements  []TeamAchievement  `json:"achievements"`
    Roster        []TeamRosterPlayer `json:"roster"`
    RecentMatches []NormalizedMatch  `json:"recent_matches"`
}

type TeamDetailProfile struct {
    ID      int    `json:"id"`
    Name    string `json:"name"`
    Slug    string `json:"slug"`
    Country string `json:"country"`
    Region  string `json:"region,omitempty"`
}

type TeamRanking struct {
    WorldRank int `json:"world_rank"`
    Points    int `json:"points"`
}

type TeamStats struct {
    Wins       int    `json:"wins"`
    Losses     int    `json:"losses"`
    Draws      int    `json:"draws"`
    WinRate    string `json:"win_rate"`
    RecentForm string `json:"recent_form"`
}

type TeamAchievement struct {
    Label string `json:"label"`
    Count int    `json:"count"`
    Tier  string `json:"tier,omitempty"` // "major", "s", "a", "streak"
}

type TeamRosterPlayer struct {
    ID      int     `json:"id"`
    Name    string `json:"name"`
    Slug    string `json:"slug"`
    Rating  float64 `json:"rating"`
    Country string  `json:"country,omitempty"`
}
```

**Scraping** (`internal/scraper/team.go`):

Add method `GetTeamDetail` — fetches `/team/{id}/{slug}`, extracts:
- Ranking: `.world-rank` / `.profile-team-stat`
- Achievements: `.trophy-row` / `.achievement-*`
- Roster: `.player-card` / `.teammate` (name + rating)
- Stats: `.team-stat-row` (W/L/D/win rate)
- Recent matches: reuse existing normalize logic, take 10

**Normalizer** (`internal/normalizer/team.go`):

Add `NormalizeTeamDetail(doc) TeamDetail`.

**Facade** (`internal/facade/facade.go`):

Add `GetTeamDetailCached(ctx, id, slug) (TeamDetail, error)` — mirror of `GetPlayerDetailCached`:
- Cache key: `team_detail:<id>`
- TTL: 7 days
- chromedp fallback via `ScrapeTeamDetail`

**HTTP handler** (`internal/http/handlers/handlers.go`):

Replace `GetTeam` stub with real implementation:

```go
func (h *Handlers) GetTeam(w, r) {
    id := atoi(chi.URLParam(r, "id"))
    ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
    defer cancel()
    td, err := h.f.GetTeamDetailCached(ctx, id, "")
    // ... error handling mirrors GetPlayer
}
```

### Frontend

**New component** `frontend/src/components/TeamDetail.tsx`:

Layout (confirmed via design preview):
1. Header: avatar (first letter of name) + team name + country + CN nickname tag
2. Ranking section (centered): "World #N" badge + "积分 N pts"
3. Stats bar: Wins(green)/Losses(red)/Draws/win_rate + recent form streak
4. Achievements row (centered, same style as PlayerDetail honors): Major(N×) / S-Tier(N×) / A-Tier(N×) / win streak record
5. Two columns:
   - Left: 近期 10 场战绩 (W/L badge + opponent + score + event + date)
   - Right: 队员阵容 (index + name + CN nickname + Rating, hoverable, clickable → opens PlayerDetail)
6. BO1 note: 任一侧得分 ≥13 → 1:0/0:1, 与选手详情一致

**SearchableList changes**:

`type === 'team'` items become clickable, opening `TeamDetail` modal (currently only `type === 'player'` supports click-to-detail).

### BO1 Normalization

Consistent with player detail logic: if either side score >= 13, normalize to 1:0 or 0:1. Applied in the same `normalizeMatches` path.

### Data Source Notes

- Ranking and points: scraped from team profile page
- Achievements: scraped from team profile page (Major wins, S/A-tier trophies, streaks)
- Roster with ratings: scraped from team profile page player cards
- If a player in roster has a known ID (extractable from team page HTML), clicking opens existing PlayerDetail component
- **Fallback**: if player IDs cannot be extracted from team page, roster displays as static list (name + rating only, no click-to-detail). This is accepted behavior.

---

## Feature 3: Matches by Event Grouping

### Backend

**New API endpoint** `GET /api/events`:

Query params: `type` = `today` | `upcoming` | `results`, `limit` = default 150.

Response groups matches by event server-side:
```json
{
  "events": [
    {
      "name": "IEM Katowice",
      "date_start": "2026-05-24",
      "date_end": "2026-05-26",
      "match_count": 3,
      "matches": [...]
    }
  ],
  "other": [...matches with empty event...]
}
```

**Handler** (`internal/http/handlers/matches.go`):
- `GetEvents(w, r)` — reads `type` + `limit` from query, calls facade

**Facade** (`internal/facade/matches.go`):
- `GetEvents(type, limit)` — fetches matches via existing scraper, groups by `event` field, builds response

**Router**: `r.Get("/api/events", h.GetEvents)`

### Frontend

**Matches.tsx rewrite**:

Keep 3 tabs (今日赛程 / 即将开始 / 近期赛果).

Within each tab, call `/api/events?type=<tab>&limit=150`:
1. **Render event cards** in 2-column grid (same card style as current match cards)
   - Event name (Oswald font, original text, no Chinese translation)
   - Date range (earliest–latest match date within event)
   - Match count badge (N 场)
2. **Click event card** → modal with:
   - Event name title + date range + match count
   - Vertical list of all matches under that event
   - Each match row: Team1 / Score or Time / Team2 + BO1/BO3 + date/status
   - For results: show score with BO1 normalization
   - For upcoming: show HH:MM start time in gold
3. **ESC** or **click backdrop** closes modal

Empty event names (observed in ~30% of upcoming lower-tier matches, e.g. Young vs TNC, Galorys vs LP) are returned as `"other"` key in the response and rendered as "Other" card group.

---

## Feature 4: News Article Detail

### Backend

**New types** (`internal/types/types.go`):

```go
type NewsArticle struct {
    Title       string `json:"title"`
    PublishedAt string `json:"published_at"`
    Link        string `json:"link"`
    BodyText    string `json:"body_text"`
    Author      string `json:"author,omitempty"`
}
```

**New scraper** (`internal/scraper/news.go`):

Add method `GetArticle(ctx, url) (string, error)`:
- chromedp requests the article URL
- Wait for `.news-block` or `.news-body` or `article` selector
- Extract **text only** (`.textContent`), not HTML — avoids XSS risk from untrusted third-party HTML
- **PREREQUISITE**: Before implementing, must test that HLTV news article pages are accessible via chromedp (not CF-blocked like `/stats/matches/`). If CF-blocked, this feature needs alternative approach.

**New normalizer** (`internal/normalizer/news.go`):

`NormalizeNewsArticle(text, item) NewsArticle` — extracts title, date, body text, author.

**Facade** (`internal/facade/facade.go`):

Add `GetNewsArticleCached(ctx, url) (NewsArticle, error)`:
- Cache key: `news_article:<md5(url)>`
- TTL: 0 (infinite, never expires)
- chromedp fallback via scraper

**HTTP handler** (`internal/http/handlers/news.go`):

```go
func (h *Handlers) GetNewsArticle(w, r) {
    url := r.URL.Query().Get("url")
    if url == "" { writeError(w, 400, "url required"); return }
    ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
    defer cancel()
    article, err := h.f.GetNewsArticleCached(ctx, url)
    // ... error handling
    writeJSON(w, map[string]any{"data": article})
}
```

**Router** (`internal/http/router.go`):

Add route: `r.Get("/api/news/article", h.GetNewsArticle)`

### Frontend

**New component** `frontend/src/components/NewsDetail.tsx`:

Wide modal (~800px), triggered by clicking a news list item.

Layout:
1. Title (bold, larger)
2. Published time + author (if any)
3. Divider
4. Body content (plain text, preserved whitespace/line breaks via `white-space: pre-wrap`)
5. Divider
6. Translate button (reuses existing OpenAI translation API)
7. Translated text area below button (shown after translation completes)
8. "Read on HLTV" external link

**Translation flow**:
- User clicks translate button
- Entire body text sent to OpenAI API (same provider config as existing news title translation)
- System prompt: "将以下CS电竞新闻正文翻译为简体中文"
- Result cached in localStorage: key = `news_trans:<md5(url)>`, TTL = infinite
- Subsequent clicks serve cached translation immediately

**News.tsx changes**:
- Add `onClick` handler to each news item
- Track `selectedNewsUrl` state
- Render `NewsDetail` when URL is set

---

## Implementation Order

Following 方案 A (全栈串联):

1. **README Stdio config** — 1 file, no deps
2. **Team detail** — backend types → scraper/normalizer → facade → handler → frontend component
3. **Matches event grouping** — backend `/api/events` → frontend Matches.tsx rewrite
4. **News detail** — CF smoke test → (if pass) backend types → scraper/normalizer → facade → handler → frontend component

---

## Error Handling

- Team detail: if HLTV page fails chromedp, return `UPSTREAM_UNAVAILABLE` error (same as PlayerDetail)
- News article CF check: before implementation, run chromedp smoke test against a real HLTV article URL. If CF-blocked (like `/stats/matches/`), abort news detail feature and fall back to showing metadata + external link only
- Matches: if `event` field is empty string, group as "Other" bucket
- BO1 normalization: degenerate cases (both < 13, negative scores) pass through unchanged

## Testing & Verification

- All features verified via Chrome DevTools on WSL IP direct connection (`172.21.32.31:8082`)
- Team detail: verify with known team IDs (Vitality=9565, Spirit=7020)
- News article: verify article body renders and translation works
- Matches event grouping: verify events group correctly, modal opens/closes, BO1 scores normalized
