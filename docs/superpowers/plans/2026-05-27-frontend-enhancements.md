# Frontend Enhancements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax.

**Goal:** Four frontend/backend enhancements — README Stdio config, team detail cards, event-grouped matches, news article detail with infinite cache and translation.

**Architecture:** Follows existing patterns: types → scraper → normalizer → facade (with cache) → HTTP handler → React frontend. Team detail mirrors PlayerDetail flow exactly. News article adds new chromedp scraper with infinite TTL caching. Events API groups existing match data server-side.

**Tech Stack:** Go 1.26, mark3labs/mcp-go, chi, goquery, chromedp, React 18, Vite, Tailwind CSS v4

---

## Feature 1: README Stdio MCP Config

### Task 1: Add standard stdio MCP config to README

**Files:**
- Modify: `README.md` (MCP 工具 section)

- [ ] **Step 1: Add stdio config before existing OpenCode format**

Edit `README.md`. In the "MCP 工具（OpenCode 注册）" section header, change it to "MCP 工具" and add the standard stdio format before the OpenCode one. Approximate location: after `### MCP 工具` heading.

```markdown
### MCP 工具

**标准 MCP 客户端**（Claude Desktop、VS Code Copilot、Gemini CLI 等）：

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

**OpenCode** 使用下方格式：

```jsonc
{
  "mcp": {
    "hltv_local": {
      "type": "local",
      "command": ["/path/to/hltv-mcp"],
      "enabled": true
    }
  }
}
```
```

- [ ] **Step 2: Commit**

```bash
git add README.md
git commit -m "docs: add standard stdio MCP config to README"
```

---

## Feature 2: Team Detail Card

### Task 2: Add TeamDetail, TeamRanking, TeamStats, TeamAchievement, TeamRosterPlayer types

**Files:**
- Modify: `internal/types/types.go` (append new types before end of file)

- [ ] **Step 1: Add new types**

Insert the following types after the existing `PlayerRecentMatch` type (end of `types.go`):

```go
// TeamDetail is the full team profile scraped from HLTV team page
type TeamDetail struct {
	Profile       TeamDetailProfile  `json:"profile"`
	Ranking       TeamRanking        `json:"ranking"`
	Stats         TeamStats          `json:"stats"`
	Achievements  []TeamAchievement  `json:"achievements"`
	Roster        []TeamRosterPlayer `json:"roster"`
	RecentMatches []NormalizedMatch  `json:"recent_matches"`
}

// TeamDetailProfile holds basic team identity
type TeamDetailProfile struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	Country string `json:"country"`
	Region  string `json:"region,omitempty"`
}

// TeamRanking holds world ranking and points
type TeamRanking struct {
	WorldRank int `json:"world_rank"`
	Points    int `json:"points"`
}

// TeamStats holds win/loss/draw and form
type TeamStats struct {
	Wins       int    `json:"wins"`
	Losses     int    `json:"losses"`
	Draws      int    `json:"draws"`
	WinRate    string `json:"win_rate"`
	RecentForm string `json:"recent_form"`
}

// TeamAchievement represents a trophy or record
type TeamAchievement struct {
	Label string `json:"label"`
	Count int    `json:"count"`
	Tier  string `json:"tier,omitempty"` // "major", "s", "a", "streak"
}

// TeamRosterPlayer is a player in the team roster
type TeamRosterPlayer struct {
	ID      int     `json:"id"`
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	Rating  float64 `json:"rating"`
	Country string  `json:"country,omitempty"`
}
```

- [ ] **Step 2: Verify compilation**

```bash
go build github.com/arcdent/hltv-mcp
```

Expected: compiles successfully.

- [ ] **Step 3: Commit**

```bash
git add internal/types/types.go
git commit -m "feat: add TeamDetail, TeamRanking, TeamStats, TeamRoster types"
```

### Task 3: Add TeamDetail normalizer

**Files:**
- Modify: `internal/normalizer/team.go`
- Check: `internal/normalizer/player.go` (NormalizePlayerDetail for pattern reference)

- [ ] **Step 1: Read the player normalizer for pattern reference**

```bash
# Review how NormalizePlayerDetail works
cat internal/normalizer/player.go
```

- [ ] **Step 2: Add NormalizeTeamDetail function to normalizer/team.go**

Append to `internal/normalizer/team.go`:

```go
import (
	"strconv"
	"strings"
)

// NormalizeTeamDetail extracts full team detail from the team page HTML
func NormalizeTeamDetail(doc *goquery.Document) types.TeamDetail {
	td := types.TeamDetail{}

	// Ranking
	rankEl := doc.Find(".profile-team-stat .value, .world-rank, .rank-value").First()
	if rankText := cleanText(rankEl.Text()); rankText != "" {
		td.Ranking.WorldRank, _ = strconv.Atoi(strings.TrimPrefix(rankText, "#"))
	}
	pointsEl := doc.Find(".profile-team-stat .description:contains('points'), .points-value").First()
	if pointsText := cleanText(pointsEl.Text()); pointsText != "" {
		parts := strings.Fields(pointsText)
		for _, p := range parts {
			if n, err := strconv.Atoi(strings.Trim(p, "()")); err == nil {
				td.Ranking.Points = n
				break
			}
		}
	}

	// Stats
	winsEl := doc.Find(".team-stat-row .stat:contains('Wins'), .win-count").First()
	if winsText := cleanText(winsEl.Text()); winsText != "" {
		td.Stats.Wins, _ = strconv.Atoi(strings.TrimPrefix(strings.Fields(winsText)[0], ""))
		// Fallback: extract from team-stat items
	}

	// Achievements
	doc.Find(".trophy, .achievement, .honor-item").Each(func(_ int, s *goquery.Selection) {
		label := cleanText(s.Find(".trophy-name, .achievement-label, .label").First().Text())
		countText := cleanText(s.Find(".trophy-count, .achievement-count, .count").First().Text())
		count, _ := strconv.Atoi(countText)
		if label == "" {
			return
		}
		tier := "a"
		lower := strings.ToLower(label)
		if strings.Contains(lower, "major") {
			tier = "major"
		} else if strings.Contains(lower, "s-tier") || strings.Contains(lower, "intel") || strings.Contains(lower, "esl pro league") || strings.Contains(lower, "blast") {
			tier = "s"
		}
		if strings.Contains(lower, "win streak") || strings.Contains(lower, "连胜") {
			tier = "streak"
		}
		td.Achievements = append(td.Achievements, types.TeamAchievement{
			Label: label, Count: count, Tier: tier,
		})
	})

	// Roster
	doc.Find(".player-card, .teammate, .player-holder").Each(func(_ int, s *goquery.Selection) {
		nameEl := s.Find(".player-name, .name, a[href*='/player/']").First()
		name := cleanText(nameEl.Text())
		if name == "" {
			return
		}
		p := types.TeamRosterPlayer{Name: name}
		href, exists := nameEl.Attr("href")
		if exists && strings.Contains(href, "/player/") {
			parts := strings.Split(strings.Trim(href, "/"), "/")
			for i, part := range parts {
				if part == "player" && i+1 < len(parts) {
					p.ID, _ = strconv.Atoi(parts[i+1])
				}
				if i+2 < len(parts) && part == "player" {
					p.Slug = parts[i+2]
				}
			}
		}
		ratingEl := s.Find(".rating, .player-rating, .stat-rating").First()
		p.Rating, _ = strconv.ParseFloat(cleanText(ratingEl.Text()), 64)
		countryEl := s.Find(".flag, .country, .player-country").First()
		if alt, ok := countryEl.Attr("alt"); ok {
			p.Country = alt
		} else {
			p.Country = cleanText(countryEl.Text())
		}
		td.Roster = append(td.Roster, p)
	})

	return td
}
```

- [ ] **Step 3: Verify compilation**

```bash
go build github.com/arcdent/hltv-mcp
```

Expected: compiles successfully.

- [ ] **Step 4: Commit**

```bash
git add internal/normalizer/team.go
git commit -m "feat: add NormalizeTeamDetail with ranking, achievements, roster extraction"
```

### Task 4: Add GetTeamDetail scraper and facade method

**Files:**
- Modify: `internal/scraper/team.go`
- Modify: `internal/facade/helpers.go`
- Modify: `internal/facade/facade.go`

- [ ] **Step 1: Add ScrapeTeamDetail to facade**

In `internal/facade/facade.go`, after `ScrapePlayerDetail`:

```go
func (f *HltvFacade) ScrapeTeamDetail(ctx context.Context, id int, slug string) (*goquery.Document, error) {
	if slug == "" { slug = fmt.Sprintf("team-%d", id) }
	return f.ts.GetTeam(ctx, id, slug)
}
```

- [ ] **Step 2: Add GetTeamDetailCached to facade**

In `internal/facade/facade.go`, after `GetPlayerDetailCached`:

```go
// GetTeamDetailCached returns cached team detail, or scrapes via chromedp and caches for 7 days
func (f *HltvFacade) GetTeamDetailCached(ctx context.Context, id int, slug string) (types.TeamDetail, error) {
	if slug == "" {
		slug = fmt.Sprintf("team-%d", id)
	}
	key := fmt.Sprintf("team_detail:%d", id)
	if cached, ok := f.cache.Get(key); ok {
		return cached.(types.TeamDetail), nil
	}
	doc, err := f.ts.GetTeam(ctx, id, slug)
	if err != nil {
		return types.TeamDetail{}, err
	}
	td := normalizer.NormalizeTeamDetail(doc)
	td.Profile.ID = id
	td.Profile.Slug = slug
	// Fetch recent 10 matches via team matches page
	matchDoc, err := f.ts.GetTeamMatches(ctx, id)
	if err == nil {
		matches := normalizer.NormalizeMatches(matchDoc, td.Profile.Name)
		normalizer.SortByPlayedAtDesc(matches)
		if len(matches) > 10 {
			matches = matches[:10]
		}
		td.RecentMatches = matches
		// Compute stats from recent matches
		for _, m := range matches {
			switch m.Result {
			case types.OutcomeWin:
				td.Stats.Wins++
			case types.OutcomeLoss:
				td.Stats.Losses++
			case types.OutcomeDraw:
				td.Stats.Draws++
			}
		}
		total := td.Stats.Wins + td.Stats.Losses + td.Stats.Draws
		if total > 0 {
			td.Stats.WinRate = fmt.Sprintf("%.0f%%", float64(td.Stats.Wins)/float64(total)*100)
		}
		// Recent form string
		for i, m := range matches {
			if i >= 5 { break }
			switch m.Result {
			case types.OutcomeWin:
				td.Stats.RecentForm += "W"
			case types.OutcomeLoss:
				td.Stats.RecentForm += "L"
			case types.OutcomeDraw:
				td.Stats.RecentForm += "D"
			}
		}
	}
	f.cache.Set(key, td, f.cfg.CacheTTLPlayerDetail) // reuse 7d TTL
	return td, nil
}
```

- [ ] **Step 3: Add normalizeTeamDetail wrapper to helpers.go**

In `internal/facade/helpers.go`, after `normalizeTeamProfile`:

```go
func normalizeTeamDetail(doc *goquery.Document) types.TeamDetail {
	return normalizer.NormalizeTeamDetail(doc)
}
```

- [ ] **Step 4: Verify compilation**

```bash
go build github.com/arcdent/hltv-mcp
```

Expected: compiles successfully.

- [ ] **Step 5: Commit**

```bash
git add internal/facade/facade.go internal/facade/helpers.go
git commit -m "feat: add GetTeamDetailCached with 7d cache and stats computation"
```

### Task 5: Replace GetTeam HTTP handler stub

**Files:**
- Modify: `internal/http/handlers/handlers.go` (GetTeam method, line ~24)

- [ ] **Step 1: Replace the GetTeam stub**

Replace:
```go
func (h *Handlers) GetTeam(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "not yet implemented"})
}
```

With:
```go
func (h *Handlers) GetTeam(w http.ResponseWriter, r *http.Request) {
	id := atoi(chi.URLParam(r, "id"))
	if id == 0 {
		writeError(w, http.StatusBadRequest, "invalid team id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	td, err := h.f.GetTeamDetailCached(ctx, id, "")
	if err != nil {
		writeJSON(w, map[string]any{
			"error": map[string]any{"code": "UPSTREAM_UNAVAILABLE", "message": "详情暂时不可用"},
			"meta":  map[string]any{"partial": true},
		})
		return
	}
	writeJSON(w, map[string]any{"data": td, "meta": map[string]any{"partial": false}})
}
```

- [ ] **Step 2: Add missing chi import if needed**

Ensure `github.com/go-chi/chi/v5` is imported in `handlers.go`. Check existing imports — if `chi` is already used (Search handler uses `chi.URLParam`), no change needed.

```bash
grep "chi" internal/http/handlers/handlers.go
```

If `chi` not found, add:
```go
import "github.com/go-chi/chi/v5"
```

Also ensure `context` and `time` are imported.

- [ ] **Step 3: Verify compilation**

```bash
go build github.com/arcdent/hltv-mcp
```

Expected: compiles successfully.

- [ ] **Step 4: Test with curl**

```bash
# Start the server if not running, then:
curl -s http://localhost:8082/api/teams/9565 | python3 -m json.tool | head -30
```

Expected: returns JSON with `data.profile.name`, `data.ranking`, `data.stats`, `data.achievements`, `data.roster`, `data.recent_matches`.

- [ ] **Step 5: Commit**

```bash
git add internal/http/handlers/handlers.go
git commit -m "feat: replace GetTeam stub with TeamDetail handler"
```

### Task 6: Build TeamDetail React component

**Files:**
- Create: `frontend/src/components/TeamDetail.tsx`
- Check: `frontend/src/components/PlayerDetail.tsx` (for pattern reference)

- [ ] **Step 1: Review PlayerDetail component structure**

```bash
cat frontend/src/components/PlayerDetail.tsx
```

- [ ] **Step 2: Read SearchableList for nicknames and team CN mappings**

```bash
grep "teamNicknames\|playerNicknames" frontend/src/components/SearchableList.tsx
```

- [ ] **Step 3: Create TeamDetail.tsx**

Create `frontend/src/components/TeamDetail.tsx`:

```tsx
import { useEffect, useState } from 'react'
import PlayerDetail from './PlayerDetail'

type TeamData = {
  profile: { id: number; name: string; slug: string; country?: string; region?: string }
  ranking: { world_rank: number; points: number }
  stats: { wins: number; losses: number; draws: number; win_rate: string; recent_form: string }
  achievements?: { label: string; count: number; tier: string }[]
  roster?: { id: number; name: string; slug: string; rating: number; country?: string }[]
  recent_matches?: { team1?: string; team2?: string; opponent?: string; score?: string; result: string; event?: string; played_at?: string; map_text?: string; best_of?: string }[]
}

const teamNicknames: Record<string, string> = {
  'Vitality':'小蜜蜂','Spirit':'绿龙','Team Spirit':'绿龙','Natus Vincere':'天生赢家',
  'NAVI':'天生赢家','FaZe':'FaZe Clan','G2':'武士','MOUZ':'老鼠','Falcons':'猎鹰',
  'Astralis':'A队','Virtus.pro':'VP','Team Liquid':'液体','FURIA':'黑豹',
  'The MongolZ':'蒙古队','TYLOO':'天禄','3DMAX':'3DMAX','paiN':'paiN',
  'HEROIC':'HEROIC','Complexity':'coL','Ninjas in Pyjamas':'NIP',
  'Eternal Fire':'永火','fnatic':'橙黑','Rare Atom':'RA','Lynn Vision':'LVG',
  'Aurora':'欧若拉','RED Canids':'红犬','GamerLegion':'GL','PARIVISION':'PV',
}

const playerNicknames: Record<string, string> = {
  'ZywOo': '载物', 's1mple': '森破', 'm0NESY': '小孩', 'donk': '洞克',
  'NiKo': '尼扣', 'dev1ce': '设备', 'ropz': '肉铺子', 'karrigan': '大表哥',
  'apEX': 'A队长', 'flameZ': '火焰', 'Spinx': '斯宾克斯', 'mezii': '梅子',
  'jL': '杰L', 'Aleksib': '阿列克西', 'b1t': '比特', 'iM': '爱慕',
  'w0nderful': '神奇', 'broky': '布洛基', 'frozen': '寒王', 'Twistzz': '总监',
  'huNter-': '猎人', 'jks': '杰克S', 'NAF': '纳夫', 'YEKINDAR': '叶金达',
  'cadiaN': '卡点', 'stavn': '斯塔文', 'jabbi': '贾比', 'TeSeS': '特塞斯',
  'EliGE': '一粒鸡', 'Magisk': '魔法男孩', 'dupreeh': '杜普瑞', 'Xyp9x': '九爷',
  'gla1ve': '格莱乌', 'electroNic': '电子哥', 'Perfecto': '完美', 'Boombl4': '胖球',
  'sh1ro': '细弱', 'Ax1Le': '阿列克斯', 'Hobbit': '霍比特', 'KSCERATO': '卡斯赛拉托',
  'yuurih': '优日', 'arT': '阿特', 'FalleN': '教父',
}

export default function TeamDetail({ id, onClose }: { id: number; onClose: () => void }) {
  const [data, setData] = useState<TeamData | null>(null)
  const [loading, setLoading] = useState(true)
  const [selectedPlayerId, setSelectedPlayerId] = useState<number | null>(null)

  useEffect(() => {
    setLoading(true)
    fetch(`/api/teams/${id}`).then(r => r.json()).then(d => {
      setData(d.data ?? null); setLoading(false)
    }).catch(() => setLoading(false))
  }, [id])

  const p = data?.profile
  const rank = data?.ranking
  const stats = data?.stats
  const achievements = data?.achievements ?? []
  const roster = data?.roster ?? []
  const matches = data?.recent_matches ?? []

  const cnName = teamNicknames[p?.name ?? '']

  return (
    <div onClick={onClose} style={{position:'fixed',inset:0,zIndex:100,background:'rgba(0,0,0,0.5)',backdropFilter:'blur(4px)',display:'flex',alignItems:'center',justifyContent:'center',animation:'fadeIn 0.2s ease'}}>
      <div onClick={e => e.stopPropagation()} style={{position:'relative',background:'var(--card)',border:'1px solid var(--border)',borderRadius:'var(--radius)',width:840,maxWidth:'95vw',maxHeight:'90vh',overflowY:'auto',padding:28,boxShadow:'0 20px 60px rgba(0,0,0,0.3)',animation:'slideUp 0.25s ease'}}>

        <button onClick={onClose} style={{position:'absolute',top:14,right:14,width:30,height:30,borderRadius:'50%',border:'1px solid var(--border)',background:'var(--card)',color:'var(--text-secondary)',fontSize:16,cursor:'pointer',display:'flex',alignItems:'center',justifyContent:'center'}}>✕</button>

        {loading && <div style={{textAlign:'center',padding:60,color:'var(--text-muted)'}}>加载中...</div>}
        {!loading && !p && <div style={{textAlign:'center',padding:60,color:'var(--text-muted)'}}>详情暂时不可用</div>}

        {!loading && p && (
          <>
            {/* Header */}
            <div style={{display:'flex',gap:14,marginBottom:18}}>
              <div style={{width:60,height:60,borderRadius:12,background:'linear-gradient(135deg,var(--gold),#8b6914)',display:'flex',alignItems:'center',justifyContent:'center',fontSize:26,color:'#fff',fontWeight:700,fontFamily:'var(--font-display)',flexShrink:0}}>
                {p.name.charAt(0)}
              </div>
              <div style={{flex:1}}>
                <div style={{fontSize:24,fontWeight:700,color:'var(--text)',lineHeight:1.2}}>{p.name}</div>
                <div style={{fontSize:12,color:'var(--text-muted)',marginTop:2}}>{p.country || '—'}{roster.length > 0 ? ` · 队员 ${roster.length} 人` : ''}{p.region ? ` · ${p.region}` : ''}</div>
                <div style={{display:'flex',flexWrap:'wrap',gap:6,marginTop:6}}>
                  {p.country ? <span style={{padding:'2px 10px',borderRadius:4,fontSize:11,background:'var(--input-bg)',color:'var(--text-secondary)',fontWeight:500}}>{p.country}</span> : null}
                  {cnName ? <span style={{padding:'2px 10px',borderRadius:4,fontSize:11,background:'var(--gold-dim)',color:'var(--gold)',fontWeight:600}}>{cnName}</span> : null}
                </div>
              </div>
            </div>

            {/* Ranking + Points */}
            {rank && rank.world_rank > 0 && (
              <div style={{display:'flex',justifyContent:'center',gap:10,marginBottom:16}}>
                <span style={{display:'flex',alignItems:'center',gap:8,padding:'6px 16px',borderRadius:20,background:'linear-gradient(135deg,#f0c040,#c48a0a)',color:'#1a1d29',fontFamily:'var(--font-display)',fontSize:14,fontWeight:700,letterSpacing:'0.04em'}}>
                  World #{rank.world_rank}
                </span>
                {rank.points > 0 && (
                  <span style={{padding:'6px 16px',borderRadius:20,background:'var(--input-bg)',color:'var(--text-secondary)',fontSize:12,fontWeight:500}}>
                    积分 <span style={{fontFamily:'var(--font-display)',fontSize:15,fontWeight:700,color:'var(--text)'}}>{rank.points}</span> pts
                  </span>
                )}
              </div>
            )}

            {/* Stats Bar */}
            {(stats?.wins !== undefined) && (
              <div style={{display:'flex',marginBottom:18,border:'1px solid var(--border)',borderRadius:'var(--radius-sm)',overflow:'hidden'}}>
                <div style={{flex:1,textAlign:'center',padding:'10px 6px',background:'var(--input-bg)'}}>
                  <div style={{fontFamily:'var(--font-display)',fontSize:22,fontWeight:700,color:'var(--green)',lineHeight:1}}>{stats!.wins}</div>
                  <div style={{fontSize:10,color:'var(--text-muted)',marginTop:3,textTransform:'uppercase',letterSpacing:'0.05em'}}>胜</div>
                </div>
                <div style={{flex:1,textAlign:'center',padding:'10px 6px',background:'var(--input-bg)',borderLeft:'1px solid var(--border)'}}>
                  <div style={{fontFamily:'var(--font-display)',fontSize:22,fontWeight:700,color:'var(--red)',lineHeight:1}}>{stats!.losses}</div>
                  <div style={{fontSize:10,color:'var(--text-muted)',marginTop:3,textTransform:'uppercase',letterSpacing:'0.05em'}}>负</div>
                </div>
                <div style={{flex:1,textAlign:'center',padding:'10px 6px',background:'var(--input-bg)',borderLeft:'1px solid var(--border)'}}>
                  <div style={{fontFamily:'var(--font-display)',fontSize:22,fontWeight:700,color:'var(--text)',lineHeight:1}}>{stats!.draws}</div>
                  <div style={{fontSize:10,color:'var(--text-muted)',marginTop:3,textTransform:'uppercase',letterSpacing:'0.05em'}}>平</div>
                </div>
                <div style={{flex:1,textAlign:'center',padding:'10px 6px',background:'var(--input-bg)',borderLeft:'1px solid var(--border)'}}>
                  <div style={{fontFamily:'var(--font-display)',fontSize:22,fontWeight:700,color:'var(--text)',lineHeight:1}}>{stats!.win_rate || '—'}</div>
                  <div style={{fontSize:10,color:'var(--text-muted)',marginTop:3,textTransform:'uppercase',letterSpacing:'0.05em'}}>胜率</div>
                  {stats!.recent_form && <div style={{fontFamily:'var(--font-mono)',fontSize:10,color:'var(--gold)',marginTop:2}}>近5场 {stats!.recent_form}</div>}
                </div>
              </div>
            )}

            {/* Achievements */}
            {achievements.length > 0 && (
              <div style={{display:'flex',gap:6,flexWrap:'wrap',justifyContent:'center',marginBottom:20}}>
                {achievements.map((a, i) => (
                  <span key={i} style={{
                    fontSize:11,padding:'3px 10px',borderRadius:10,fontWeight:a.tier==='major'?600:500,display:'flex',alignItems:'center',gap:3,
                    background: a.tier==='major'?'linear-gradient(135deg,rgba(240,192,64,0.15),rgba(196,138,10,0.1))':'rgba(196,138,10,0.06)',
                    color: a.tier==='major'?'#f0c040':'var(--gold)',
                  }}>
                    {a.tier==='major'?'\u{1f3c6} ':''}{a.label} <span style={{fontFamily:'var(--font-mono)',fontWeight:700,opacity:0.8}}>{a.count}&times;</span>
                  </span>
                ))}
              </div>
            )}

            {/* Two columns: recent matches + roster */}
            <div style={{display:'grid',gridTemplateColumns:'1fr 1fr',gap:24}}>

              {/* Left: Recent 10 matches */}
              <div>
                <div style={{fontFamily:'var(--font-display)',fontSize:14,fontWeight:600,color:'var(--gold)',letterSpacing:'0.05em',textTransform:'uppercase',marginBottom:10,paddingBottom:6,borderBottom:'1px solid var(--border)',display:'flex',justifyContent:'space-between'}}>
                  近期战绩
                  <span style={{fontSize:11,fontWeight:400,color:'var(--text-muted)',fontFamily:'var(--font-body)',textTransform:'none',letterSpacing:0}}>{matches.length} 场</span>
                </div>
                {matches.length === 0 && <div style={{fontSize:12,color:'var(--text-muted)',textAlign:'center',padding:'20px 0'}}>暂无数据</div>}
                {matches.map((m, i) => (
                  <div key={i} style={{display:'flex',alignItems:'center',gap:10,padding:'7px 0',borderBottom:i<matches.length-1?'1px solid rgba(128,128,128,0.06)':'none',fontSize:12}}>
                    <span style={{minWidth:26,textAlign:'center',fontSize:10,fontWeight:700,fontFamily:'var(--font-mono)',padding:'2px 0',borderRadius:3,
                      color:m.result==='win'?'var(--green)':m.result==='loss'?'var(--red)':'var(--text-muted)',
                      background:m.result==='win'?'rgba(0,200,83,0.1)':m.result==='loss'?'rgba(255,82,82,0.1)':'var(--input-bg)'}}>
                      {m.result==='win'?'W':m.result==='loss'?'L':'—'}
                    </span>
                    <span style={{flex:1,minWidth:0,whiteSpace:'nowrap',overflow:'hidden',textOverflow:'ellipsis'}}>
                      <b style={{fontWeight:600}}>{p.name}</b> vs {m.opponent || m.team2 || '待定'}
                    </span>
                    {m.score && <span style={{fontFamily:'var(--font-mono)',fontSize:11,color:'var(--text-secondary)',minWidth:30,textAlign:'center'}}>{m.score}</span>}
                    <span style={{fontSize:10,color:'var(--text-muted)',maxWidth:80,whiteSpace:'nowrap',overflow:'hidden',textOverflow:'ellipsis'}}>{m.event || ''}</span>
                    <span style={{fontSize:10,color:'var(--text-muted)',minWidth:48,textAlign:'right'}}>{(m.played_at || '').slice(5,10)}</span>
                  </div>
                ))}
              </div>

              {/* Right: Roster */}
              <div>
                <div style={{fontFamily:'var(--font-display)',fontSize:14,fontWeight:600,color:'var(--gold)',letterSpacing:'0.05em',textTransform:'uppercase',marginBottom:10,paddingBottom:6,borderBottom:'1px solid var(--border)',display:'flex',justifyContent:'space-between'}}>
                  队员阵容
                  <span style={{fontSize:11,fontWeight:400,color:'var(--text-muted)',fontFamily:'var(--font-body)',textTransform:'none',letterSpacing:0}}>{roster.length} 人</span>
                </div>
                {roster.length === 0 && <div style={{fontSize:12,color:'var(--text-muted)',textAlign:'center',padding:'20px 0'}}>暂无数据</div>}
                {roster.map((pl, i) => (
                  <div key={i} onClick={() => pl.id > 0 && setSelectedPlayerId(pl.id)}
                    style={{
                      display:'flex',alignItems:'center',gap:10,padding:'7px 4px',fontSize:13,
                      borderBottom:i<roster.length-1?'1px solid rgba(128,128,128,0.06)':'none',
                      cursor: pl.id > 0 ? 'pointer' : 'default', borderRadius:4,
                    }}
                    onMouseEnter={e => { if(pl.id>0) e.currentTarget.style.background='var(--gold-dim)' }}
                    onMouseLeave={e => { e.currentTarget.style.background='transparent' }}>
                    <span style={{fontFamily:'var(--font-mono)',fontSize:11,fontWeight:700,color:'var(--text-muted)',minWidth:18}}>
                      {String(i+1).padStart(2,'0')}
                    </span>
                    <span style={{fontWeight:600,flex:1}}>
                      {pl.name}
                      {(playerNicknames[pl.name]) && <span style={{fontSize:11,color:'var(--text-muted)',marginLeft:4,fontWeight:400}}>{playerNicknames[pl.name]}</span>}
                    </span>
                    {pl.rating > 0 && <span style={{fontFamily:'var(--font-mono)',fontSize:11,color:'var(--text-secondary)',background:'var(--input-bg)',padding:'2px 7px',borderRadius:4}}>Rating {pl.rating.toFixed(2)}</span>}
                    {pl.id > 0 && <span style={{fontSize:10,color:'var(--gold)',opacity:0.5}}>→</span>}
                  </div>
                ))}
              </div>
            </div>

            {/* BO1 note */}
            <div style={{marginTop:16,padding:'8px 14px',background:'var(--input-bg)',borderRadius:'var(--radius-sm)',fontSize:11,color:'var(--text-muted)',textAlign:'center',border:'1px dashed var(--border)'}}>
              BO1 比分归一化：任一侧得分 &ge;13 &rarr; <code style={{fontFamily:'var(--font-mono)',color:'var(--text-secondary)',background:'rgba(196,138,10,0.08)',padding:'1px 5px',borderRadius:3}}>1:0</code> / <code style={{fontFamily:'var(--font-mono)',color:'var(--text-secondary)',background:'rgba(196,138,10,0.08)',padding:'1px 5px',borderRadius:3}}>0:1</code>，与选手详情保持一致
            </div>

            <div style={{marginTop:14,textAlign:'center',fontSize:11,color:'var(--text-muted)'}}>点击队员可查看选手详情 · ESC 关闭</div>
          </>
        )}
      </div>
      {selectedPlayerId !== null && <PlayerDetail id={selectedPlayerId} onClose={() => setSelectedPlayerId(null)} />}
    </div>
  )
}
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/TeamDetail.tsx
git commit -m "feat: add TeamDetail modal component with rankings, achievements, roster"
```

### Task 7: Wire TeamDetail into SearchableList for team type

**Files:**
- Modify: `frontend/src/components/SearchableList.tsx`

- [ ] **Step 1: Import TeamDetail**

Add import at top:
```tsx
import TeamDetail from './TeamDetail'
```

- [ ] **Step 2: Add team click handler state**

After existing state:
```tsx
const [selectedTeamId, setSelectedTeamId] = useState<number | null>(null)
```

- [ ] **Step 3: Make team items clickable**

In the list item render, change:
```tsx
onClick={() => type === 'player' && item.id && setSelectedId(item.id)}
```

To:
```tsx
onClick={() => {
  if (type === 'player' && item.id) setSelectedId(item.id)
  if (type === 'team' && item.id) setSelectedTeamId(item.id)
}}
```

And change the cursor style from:
```tsx
cursor: type === 'player' ? 'pointer' : 'default'
```

To:
```tsx
cursor: (type === 'player' || type === 'team') ? 'pointer' : 'default'
```

- [ ] **Step 4: Render TeamDetail when selected**

After the existing `PlayerDetail` render line, add:
```tsx
{type === 'team' && selectedTeamId !== null && <TeamDetail id={selectedTeamId} onClose={() => setSelectedTeamId(null)} />}
```

- [ ] **Step 5: Rebuild frontend and verify**

```bash
cd frontend && npm run build && cd ..
go build github.com/arcdent/hltv-mcp
```

Expected: both compile successfully.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/SearchableList.tsx
git commit -m "feat: wire TeamDetail click-to-open for team search results"
```

---

## Feature 3: Matches by Event Grouping

### Task 8: Add GetEvents facade and HTTP endpoint

**Files:**
- Modify: `internal/facade/matches.go`
- Modify: `internal/http/handlers/matches.go`
- Modify: `internal/http/router.go`

- [ ] **Step 1: Add GetEvents facade method**

In `internal/facade/matches.go`, after `GetResultsRecent`:

```go
// EventGroup holds matches grouped under one event name
type EventGroup struct {
	Name      string                  `json:"name"`
	DateStart string                  `json:"date_start"`
	DateEnd   string                  `json:"date_end"`
	MatchCount int                    `json:"match_count"`
	Matches   []types.NormalizedMatch `json:"matches"`
}

// EventsResponse is the response for /api/events
type EventsResponse struct {
	Events []EventGroup           `json:"events"`
	Other  []types.NormalizedMatch `json:"other,omitempty"`
}

// GetEvents fetches matches and groups them by event name
func (f *HltvFacade) GetEvents(matchType string, limit int) *types.ToolResponse {
	q := map[string]any{"type": matchType, "limit": limit}
	key := fmt.Sprintf("events:%s:%d", matchType, limit)
	ttl := f.cfg.CacheTTLMatches

	return f.withCache(key, ttl, q, func() (*types.ToolResponse, error) {
		var items []types.NormalizedMatch
		var err error
		switch matchType {
		case "today", "upcoming":
			doc, err := f.ms.GetUpcoming(context.Background())
			if err != nil {
				return nil, err
			}
			items = normalizer.NormalizeUpcomingMatches(doc, "")
			if matchType == "today" {
				items = filterToday(items)
			}
		case "results":
			doc, err := f.rs.GetResults(context.Background())
			if err != nil {
				return nil, err
			}
			items = normalizer.NormalizeMatches(doc, "")
			normalizer.SortByPlayedAtDesc(items)
		default:
			return nil, fmt.Errorf("invalid type: %s", matchType)
		}
		_ = err
		if len(items) > limit {
			items = items[:limit]
		}

		resp := groupByEvent(items)
		meta := f.createMeta(ttl)
		r := &types.ToolResponse{Query: q, Meta: meta}
		r.Data = resp
		return r, nil
	})
}

func groupByEvent(matches []types.NormalizedMatch) EventsResponse {
	groups := make(map[string]*EventGroup)
	var other []types.NormalizedMatch

	for _, m := range matches {
		event := strings.TrimSpace(m.Event)
		if event == "" {
			other = append(other, m)
			continue
		}
		g, exists := groups[event]
		if !exists {
			g = &EventGroup{Name: event}
			groups[event] = g
		}
		g.Matches = append(g.Matches, m)
		g.MatchCount++

		date := m.PlayedAt
		if date == "" {
			date = m.ScheduledAt
		}
		if date == "" {
			continue
		}
		date = strings.SplitN(date, " ", 2)[0]
		if g.DateStart == "" || date < g.DateStart {
			g.DateStart = date
		}
		if g.DateEnd == "" || date > g.DateEnd {
			g.DateEnd = date
		}
	}

	var events []EventGroup
	for _, g := range groups {
		events = append(events, *g)
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].DateStart < events[j].DateStart
	})

	return EventsResponse{Events: events, Other: other}
}
```

- [ ] **Step 2: Add GetEvents handler**

In `internal/http/handlers/matches.go`, after `GetResults`:

```go
func (h *Handlers) GetEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	matchType := q.Get("type")
	limit := atoi(q.Get("limit"))
	if limit == 0 {
		limit = 150
	}
	resp := h.f.GetEvents(matchType, limit)
	writeJSON(w, resp)
}
```

- [ ] **Step 3: Add route**

In `internal/http/router.go`, after the existing match routes:

```go
r.Get("/api/events", h.GetEvents)
```

- [ ] **Step 4: Verify compilation**

```bash
go build github.com/arcdent/hltv-mcp
```

Expected: compiles successfully.

- [ ] **Step 5: Test with curl**

```bash
curl -s 'http://localhost:8082/api/events?type=upcoming&limit=20' | python3 -c "import json,sys;d=json.load(sys.stdin);print(f'Events: {len(d.get(\"data\",{}).get(\"events\",[]))}, Other: {len(d.get(\"data\",{}).get(\"other\",[]))}')"
```

Expected: number of events > 0.

- [ ] **Step 6: Commit**

```bash
git add internal/facade/matches.go internal/http/handlers/matches.go internal/http/router.go
git commit -m "feat: add /api/events endpoint for event-grouped matches"
```

### Task 9: Rewrite Matches.tsx for event grouping with modal

**Files:**
- Modify: `frontend/src/pages/Matches.tsx`
- Modify: `frontend/src/api/client.ts`

- [ ] **Step 1: Add getEvents to API client**

In `frontend/src/api/client.ts`, add to the `api` object:

```ts
getEvents: (type: string, limit = 150) =>
  request<any>(`/events?type=${encodeURIComponent(type)}&limit=${limit}`),
```

- [ ] **Step 2: Rewrite Matches.tsx**

Write the complete new `Matches.tsx`. The file replaces the current flat match list with two-level event grouping + modal detail.

```tsx
import { useEffect, useState } from 'react'
import { api } from '../api/client'

type Tab = 'today' | 'upcoming' | 'results'

const tabs: { key: Tab; label: string }[] = [
  { key: 'today',    label: '今日赛程' },
  { key: 'upcoming', label: '即将开始' },
  { key: 'results',  label: '近期赛果' },
]

const nicknames: Record<string, string> = {
  'Vitality':'小蜜蜂','Spirit':'绿龙','Team Spirit':'绿龙','Natus Vincere':'天生赢家',
  'NAVI':'天生赢家','FaZe':'FaZe Clan','G2':'武士','MOUZ':'老鼠','Falcons':'猎鹰',
  'Astralis':'A队','Virtus.pro':'VP','Team Liquid':'液体','FURIA':'黑豹',
  'The MongolZ':'蒙古队','TYLOO':'天禄','3DMAX':'3DMAX','paiN':'paiN',
  'HEROIC':'HEROIC','Complexity':'coL','Ninjas in Pyjamas':'NIP',
  'Eternal Fire':'永火','fnatic':'橙黑','Rare Atom':'RA','Lynn Vision':'LVG',
  'Aurora':'欧若拉','RED Canids':'红犬','GamerLegion':'GL','PARIVISION':'PV',
}

export default function Matches() {
  const [tab, setTab] = useState<Tab>('today')
  const [events, setEvents] = useState<any[]>([])
  const [other, setOther] = useState<any[]>([])
  const [loading, setLoading] = useState(false)
  const [selectedEvent, setSelectedEvent] = useState<any>(null)

  useEffect(() => {
    setLoading(true)
    api.getEvents(tab, 150).then(d => {
      setEvents(d?.data?.events ?? [])
      setOther(d?.data?.other ?? [])
      setLoading(false)
    }).catch(() => setLoading(false))
  }, [tab])

  const cardStyle: React.CSSProperties = {
    background: 'var(--card)', border: '1px solid var(--border)',
    borderRadius: 'var(--radius)', padding: '16px 20px',
    boxShadow: 'var(--card-shadow)',
  }

  const tabBtn = (active: boolean): React.CSSProperties => ({
    fontSize: 16, fontWeight: 600, fontFamily: 'var(--font-display)',
    letterSpacing: '0.04em', textTransform: 'uppercase' as const,
    color: active ? 'var(--gold)' : 'var(--text-muted)',
    borderBottom: active ? '2px solid var(--gold)' : '2px solid transparent',
    paddingBottom: 6, background: 'none', cursor: 'pointer',
  })

  const totalEvents = events.length + (other.length > 0 ? 1 : 0)

  return (
    <div className="anim-in" style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>

      {/* Tab bar */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 24 }}>
        {tabs.map(t => (
          <button key={t.key} onClick={() => setTab(t.key)} style={tabBtn(tab === t.key)}>
            {t.label}
          </button>
        ))}
        <div style={{ flex: 1 }} />
        {!loading && totalEvents > 0 && (
          <span style={{ fontSize: 14, color: 'var(--text-muted)' }}>{totalEvents} 个赛事</span>
        )}
      </div>

      {/* Event cards grid */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 12 }}>
        {!loading && totalEvents === 0 && (
          <div style={{ ...cardStyle, gridColumn: '1 / -1', textAlign: 'center', padding: '80px 0', color: 'var(--text-muted)', fontSize: 15 }}>
            暂无赛事数据
          </div>
        )}

        {loading && (
          <div style={{ ...cardStyle, gridColumn: '1 / -1', textAlign: 'center', padding: '80px 0', color: 'var(--text-muted)', fontSize: 15 }}>
            加载中...
          </div>
        )}

        {events.map((ev, i) => (
          <div key={i} className="anim-in" style={{ ...cardStyle, cursor: 'pointer', animationDelay: `${i * 30}ms` }}
            onClick={() => setSelectedEvent(ev)}
            onMouseEnter={e => { e.currentTarget.style.borderColor = 'var(--gold)' }}
            onMouseLeave={e => { e.currentTarget.style.borderColor = 'var(--border)' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
              <span style={{ flex: 1, fontSize: 16, fontWeight: 700, fontFamily: 'var(--font-display)', letterSpacing: '0.03em', color: 'var(--text)' }}>
                {ev.name}
              </span>
              <span style={{ background: 'var(--gold-dim)', color: 'var(--gold)', fontFamily: 'var(--font-mono)', fontSize: 12, fontWeight: 700, padding: '3px 10px', borderRadius: 20 }}>
                {ev.match_count}
              </span>
            </div>
            <div style={{ marginTop: 8, fontSize: 12, color: 'var(--text-muted)' }}>
              {ev.date_start || '?'} ~ {ev.date_end || '?'}
            </div>
          </div>
        ))}

        {/* Other bucket */}
        {other.length > 0 && (
          <div className="anim-in" style={{ ...cardStyle, cursor: 'pointer' }}
            onClick={() => setSelectedEvent({ name: 'Other', date_start: '—', date_end: '—', match_count: other.length, matches: other })}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
              <span style={{ flex: 1, fontSize: 16, fontWeight: 700, fontFamily: 'var(--font-display)', letterSpacing: '0.03em', color: 'var(--text-secondary)' }}>
                Other
              </span>
              <span style={{ background: 'var(--input-bg)', color: 'var(--text-muted)', fontFamily: 'var(--font-mono)', fontSize: 12, fontWeight: 700, padding: '3px 10px', borderRadius: 20 }}>
                {other.length}
              </span>
            </div>
            <div style={{ marginTop: 8, fontSize: 12, color: 'var(--text-muted)' }}>未分配赛事</div>
          </div>
        )}
      </div>

      {/* Event Detail Modal */}
      {selectedEvent && (
        <div onClick={() => setSelectedEvent(null)} style={{ position: 'fixed', inset: 0, zIndex: 100, background: 'rgba(0,0,0,0.5)', backdropFilter: 'blur(4px)', display: 'flex', alignItems: 'center', justifyContent: 'center', animation: 'fadeIn 0.2s ease' }}>
          <div onClick={e => e.stopPropagation()} style={{ position: 'relative', background: 'var(--card)', border: '1px solid var(--border)', borderRadius: 'var(--radius)', width: 700, maxWidth: '90vw', maxHeight: '85vh', overflowY: 'auto', padding: 28, boxShadow: '0 20px 60px rgba(0,0,0,0.3)', animation: 'slideUp 0.25s ease' }}>
            <button onClick={() => setSelectedEvent(null)} style={{ position: 'absolute', top: 14, right: 14, width: 30, height: 30, borderRadius: '50%', border: '1px solid var(--border)', background: 'var(--card)', color: 'var(--text-secondary)', fontSize: 16, cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>✕</button>

            <div style={{ fontFamily: 'var(--font-display)', fontSize: 20, fontWeight: 700, color: 'var(--gold)', letterSpacing: '0.04em', marginBottom: 6 }}>
              {selectedEvent.name}
            </div>
            <div style={{ fontSize: 13, color: 'var(--text-muted)', marginBottom: 18, paddingBottom: 14, borderBottom: '1px solid var(--border)' }}>
              {selectedEvent.date_start || '?'} ~ {selectedEvent.date_end || '?'} · {selectedEvent.match_count} 场比赛
            </div>

            {(selectedEvent.matches || []).map((m: any, i: number) => {
              const c1 = nicknames[m.team1 ?? ''] ?? ''
              const c2 = nicknames[m.team2 ?? ''] ?? ''
              const isUpcoming = !m.score
              return (
                <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 14, padding: '12px 0', borderTop: i > 0 ? '1px solid rgba(128,128,128,0.06)' : 'none', fontSize: 13 }}>
                  <div style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 2 }}>
                    <span style={{ fontSize: 15, fontWeight: 600, fontFamily: 'var(--font-display)', letterSpacing: '0.03em', color: 'var(--text)' }}>{m.team1 || '待定'}</span>
                    <span style={{ fontSize: 11, color: 'var(--text-muted)', height: 16 }}>{c1}</span>
                  </div>
                  {m.score ? (
                    <span style={{ fontFamily: 'var(--font-mono)', fontSize: 20, fontWeight: 700, color: 'var(--text)', minWidth: 50, textAlign: 'center' }}>{m.score}</span>
                  ) : (
                    <span style={{ fontFamily: 'var(--font-display)', fontSize: 20, fontWeight: 700, color: 'var(--gold)', minWidth: 50, textAlign: 'center' }}>
                      {m.scheduled_at ? m.scheduled_at.slice(11, 16) : '—:—'}
                    </span>
                  )}
                  <div style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 2 }}>
                    <span style={{ fontSize: 15, fontWeight: 600, fontFamily: 'var(--font-display)', letterSpacing: '0.03em', color: 'var(--text)' }}>{m.team2 || '待定'}</span>
                    <span style={{ fontSize: 11, color: 'var(--text-muted)', height: 16 }}>{c2}</span>
                  </div>
                  <span style={{ fontSize: 11, color: isUpcoming ? 'var(--gold)' : 'var(--text-muted)', minWidth: 60, textAlign: 'right' }}>
                    {m.best_of ? `${m.best_of.toUpperCase()}` : ''}{m.played_at ? ` · ${m.played_at.slice(5, 10)}` : ''}
                  </span>
                </div>
              )
            })}
          </div>
        </div>
      )}
    </div>
  )
}
```

Note: The Chinese strings in the code above use Unicode escapes for portability. In actual file write, use the actual characters (今日赛程, 即将开始, 近期赛果, 加载中, 暂无赛事数据, 未分配赛事).

- [ ] **Step 3: Rebuild frontend**

```bash
cd frontend && npm run build && cd ..
```

- [ ] **Step 4: Verify compilation**

```bash
go build github.com/arcdent/hltv-mcp
```

Expected: both compile successfully.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/Matches.tsx frontend/src/api/client.ts
git commit -m "feat: rewrite matches page with event grouping and modal detail"
```

---

## Feature 4: News Article Detail

### Task 10: CF Smoke Test — verify HLTV news articles are accessible

**Files:**
- Check: AGENTS.md (known CF block patterns)

- [ ] **Step 1: Run chromedp smoke test against a real HLTV article URL**

Find a news article URL from existing data:
```bash
# Get a real article link from the API
curl -s 'http://localhost:8082/api/news?limit=1' | python3 -c "import json,sys;d=json.load(sys.stdin);items=d.get('items',[]);print(items[0].get('link','') if items else 'no link')"
```

- [ ] **Step 2: Verify curl access (HTTP direct, non-chromedp)**

```bash
# Use the link from step 1
curl -sI "https://www.hltv.org/news/<id>/<slug>"
```

Expected: HTTP 200 (not 403/503 CF challenge).

- [ ] **Step 3: If curl returns 200, mark CF test as PASSED**

News article pages are accessible. Proceed with implementation.

- [ ] **Step 4: If CF-blocked, ABORT feature 4**

Instead, update News.tsx to show metadata-only modal (title, time, summary_hint, "Read on HLTV" link). Commit and skip Tasks 11-14.

```bash
# Only if CF test FAILS:
git commit -m "doc: news article CF smoke test — blocked, falling back to metadata-only modal"
```

### Task 11: Add NewsArticle type and config

**Files:**
- Modify: `internal/types/types.go`
- Modify: `internal/config/config.go`

- [ ] **Step 1: Add NewsArticle type**

After `TeamRosterPlayer` type in `internal/types/types.go`:

```go
// NewsArticle is the full text of a news article scraped from HLTV
type NewsArticle struct {
	Title       string `json:"title"`
	PublishedAt string `json:"published_at"`
	Link        string `json:"link"`
	BodyText    string `json:"body_text"`
	Author      string `json:"author,omitempty"`
}
```

- [ ] **Step 2: Add CacheTTLNewsArticle to config**

In `internal/config/config.go`, add to the `Config` struct:
```go
CacheTTLNewsArticle int
```

In `LoadConfig()`:
```go
CacheTTLNewsArticle: envInt("CACHE_TTL_NEWS_ARTICLE_SEC", 100*365*24*3600), // ~100 years = infinite
```

- [ ] **Step 3: Verify compilation**

```bash
go build github.com/arcdent/hltv-mcp
```

- [ ] **Step 4: Commit**

```bash
git add internal/types/types.go internal/config/config.go
git commit -m "feat: add NewsArticle type and infinite cache TTL config"
```

### Task 12: Add news article scraper, normalizer, and facade

**Files:**
- Create: `internal/scraper/news_article.go`
- Create: `internal/normalizer/news_article.go`
- Modify: `internal/facade/facade.go`
- Modify: `internal/facade/news.go`

- [ ] **Step 1: Create news article scraper**

Create `internal/scraper/news_article.go`:

```go
package scraper

import (
	"context"
	"github.com/PuerkitoBio/goquery"
)

type NewsArticleScraper struct{ cli *client.HltvClient }

func NewNewsArticleScraper(cli *client.HltvClient) *NewsArticleScraper {
	return &NewsArticleScraper{cli: cli}
}

// GetArticle fetches a news article page and returns the parsed document
func (s *NewsArticleScraper) GetArticle(ctx context.Context, url string) (*goquery.Document, error) {
	body, err := s.cli.FetchHTML(ctx, url, "news_article")
	if err != nil {
		return nil, err
	}
	return goquery.NewDocumentFromReader(bytesReader(body))
}
```

- [ ] **Step 2: Create news article normalizer**

Create `internal/normalizer/news_article.go`:

```go
package normalizer

import (
	"strings"
	"github.com/PuerkitoBio/goquery"
	"github.com/arcdent/hltv-mcp/internal/types"
)

// NormalizeNewsArticle extracts plain text from a news article page
func NormalizeNewsArticle(doc *goquery.Document, link string) types.NewsArticle {
	a := types.NewsArticle{Link: link}

	titleEl := doc.Find(".news-headline, .article-title, h1").First()
	a.Title = cleanText(titleEl.Text())

	dateEl := doc.Find(".news-date, .article-date, .date").First()
	a.PublishedAt = cleanText(dateEl.Text())

	authorEl := doc.Find(".news-author, .author-name").First()
	a.Author = cleanText(authorEl.Text())

	bodyEl := doc.Find(".news-block, .news-body, article, .body").First()
	if bodyEl.Length() == 0 {
		bodyEl = doc.Find(".content, .main-content, .article-content").First()
	}
	a.BodyText = strings.TrimSpace(bodyEl.Text())

	return a
}
```

- [ ] **Step 3: Add facade methods for news article**

Add to `internal/facade/facade.go`, after `GetPlayerDetailCached`:

```go
import "crypto/md5"

// ScrapeNewsArticle fetches a news article page via chromedp
func (f *HltvFacade) ScrapeNewsArticle(ctx context.Context, url string) (*goquery.Document, error) {
	return f.nas.GetArticle(ctx, url)
}

// GetNewsArticleCached returns cached article body, or scrapes and caches indefinitely
func (f *HltvFacade) GetNewsArticleCached(ctx context.Context, url string) (types.NewsArticle, error) {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(url)))
	key := fmt.Sprintf("news_article:%s", hash)
	if cached, ok := f.cache.Get(key); ok {
		return cached.(types.NewsArticle), nil
	}
	doc, err := f.nas.GetArticle(ctx, url)
	if err != nil {
		return types.NewsArticle{}, err
	}
	article := normalizer.NormalizeNewsArticle(doc, url)
	f.cache.Set(key, article, f.cfg.CacheTTLNewsArticle)
	return article, nil
}
```

Also add `nas *scraper.NewsArticleScraper` to the `HltvFacade` struct and initialize it in `New()`:
```go
nas: scraper.NewNewsArticleScraper(cli),
```

- [ ] **Step 4: Verify compilation**

```bash
go build github.com/arcdent/hltv-mcp
```

Expected: compiles successfully.

- [ ] **Step 5: Commit**

```bash
git add internal/scraper/news_article.go internal/normalizer/news_article.go internal/facade/facade.go
git commit -m "feat: add news article scraper, normalizer, and infinite-cache facade"
```

### Task 13: Add news article HTTP handler and route

**Files:**
- Modify: `internal/http/handlers/news.go`
- Modify: `internal/http/router.go`

- [ ] **Step 1: Add GetNewsArticle handler**

In `internal/http/handlers/news.go`, after existing news handlers:

```go
func (h *Handlers) GetNewsArticle(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		writeError(w, http.StatusBadRequest, "url required")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	article, err := h.f.GetNewsArticleCached(ctx, url)
	if err != nil {
		writeJSON(w, map[string]any{
			"error": map[string]any{"code": "UPSTREAM_UNAVAILABLE", "message": "文章抓取失败，请在 HLTV 阅读原文"},
			"meta":  map[string]any{"partial": true},
		})
		return
	}
	writeJSON(w, map[string]any{"data": article, "meta": map[string]any{"partial": false}})
}
```

Ensure imports include `"context"` and `"time"`.

- [ ] **Step 2: Add route**

In `internal/http/router.go`, after `r.Get("/api/news", ...)`:

```go
r.Get("/api/news/article", h.GetNewsArticle)
```

- [ ] **Step 3: Verify compilation**

```bash
go build github.com/arcdent/hltv-mcp
```

- [ ] **Step 4: Test with curl**

```bash
ARTICLE_URL=$(curl -s 'http://localhost:8082/api/news?limit=1' | python3 -c "import json,sys;d=json.load(sys.stdin);print(d['items'][0].get('link',''))")
curl -s "http://localhost:8082/api/news/article?url=$(python3 -c "import urllib.parse;print(urllib.parse.quote('$ARTICLE_URL'))")" | python3 -c "import json,sys;d=json.load(sys.stdin);a=d.get('data',{});print(f'Title: {a.get(\"title\",\"?\")}\\nBody length: {len(a.get(\"body_text\",\"\"))}')"
```

Expected: body_text length > 0.

- [ ] **Step 5: Commit**

```bash
git add internal/http/handlers/news.go internal/http/router.go
git commit -m "feat: add /api/news/article endpoint for full article text"
```

### Task 14: Build NewsDetail component with translation

**Files:**
- Create: `frontend/src/components/NewsDetail.tsx`
- Modify: `frontend/src/pages/News.tsx`
- Modify: `frontend/src/api/client.ts`

- [ ] **Step 1: Add getNewsArticle to API client**

In `frontend/src/api/client.ts`:

```ts
getNewsArticle: (url: string) =>
  request<any>(`/news/article?url=${encodeURIComponent(url)}`),
```

- [ ] **Step 2: Create NewsDetail.tsx**

Create `frontend/src/components/NewsDetail.tsx`:

```tsx
import { useEffect, useState } from 'react'
import { useTranslateConfig } from './TranslateProvider'

type ArticleData = {
  title: string; published_at: string; link: string; body_text: string; author?: string;
}

export default function NewsDetail({ url, onClose }: { url: string; onClose: () => void }) {
  const [data, setData] = useState<ArticleData | null>(null)
  const [loading, setLoading] = useState(true)
  const [translating, setTranslating] = useState(false)
  const [translated, setTranslated] = useState('')
  const { cfg, realKey } = useTranslateConfig()

  useEffect(() => {
    setLoading(true)
    fetch(`/api/news/article?url=${encodeURIComponent(url)}`).then(r => r.json()).then(d => {
      setData(d.data ?? null); setLoading(false)
    }).catch(() => setLoading(false))
  }, [url])

  // Check localStorage for cached translation
  useEffect(() => {
    if (!data?.body_text) return
    try {
      // Simple hash for URL-based localStorage key
      let hash = 0; for (let i = 0; i < url.length; i++) { hash = (hash * 31 + url.charCodeAt(i)) >>> 0 }
      const key = `news_trans:${hash.toString(16)}`
      const cached = localStorage.getItem(key)
      if (cached) {
        const { zh, ts } = JSON.parse(cached)
        setTranslated(zh)
      }
    } catch {}
  }, [data, url])

  const doTranslate = async () => {
    if (!data?.body_text || !cfg?.configured) return
    setTranslating(true)
    try {
      const res = await fetch(`${cfg!.provider_url}/chat/completions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${realKey}` },
        body: JSON.stringify({
          model: cfg!.model,
          messages: [
            { role: 'system', content: '将以下CS电竞新闻正文翻译为简体中文' },
            { role: 'user', content: data.body_text.slice(0, 8000) },
          ],
          temperature: 0.1,
        }),
      })
      const body = await res.text()
      if (!res.ok) throw new Error(body)
      const j = JSON.parse(body)
      const zh = (j?.choices?.[0]?.message?.content as string)?.trim() ?? ''
      if (zh) {
        setTranslated(zh)
        try {
          let hash = 0; for (let i = 0; i < url.length; i++) { hash = (hash * 31 + url.charCodeAt(i)) >>> 0 }
          localStorage.setItem(`news_trans:${hash.toString(16)}`, JSON.stringify({ zh, ts: Date.now() }))
        } catch {}
      }
    } catch (e) { console.error('translate article failed:', e) }
    setTranslating(false)
  }

  return (
    <div onClick={onClose} style={{position:'fixed',inset:0,zIndex:100,background:'rgba(0,0,0,0.5)',backdropFilter:'blur(4px)',display:'flex',alignItems:'center',justifyContent:'center',animation:'fadeIn 0.2s ease'}}>
      <div onClick={e => e.stopPropagation()} style={{position:'relative',background:'var(--card)',border:'1px solid var(--border)',borderRadius:'var(--radius)',width:800,maxWidth:'95vw',maxHeight:'90vh',overflowY:'auto',padding:32,boxShadow:'0 20px 60px rgba(0,0,0,0.3)',animation:'slideUp 0.25s ease'}}>

        <button onClick={onClose} style={{position:'absolute',top:14,right:14,width:30,height:30,borderRadius:'50%',border:'1px solid var(--border)',background:'var(--card)',color:'var(--text-secondary)',fontSize:16,cursor:'pointer',display:'flex',alignItems:'center',justifyContent:'center'}}>✕</button>

        {loading && <div style={{textAlign:'center',padding:60,color:'var(--text-muted)'}}>抓取中...</div>}
        {!loading && !data && <div style={{textAlign:'center',padding:60,color:'var(--text-muted)'}}>文章暂时不可用</div>}

        {!loading && data && (
          <>
            <div style={{fontSize:22,fontWeight:700,color:'var(--text)',lineHeight:1.3,marginBottom:8}}>{data.title}</div>
            <div style={{display:'flex',alignItems:'center',gap:12,fontSize:12,color:'var(--text-muted)',marginBottom:18,paddingBottom:14,borderBottom:'1px solid var(--border)'}}>
              {data.published_at && <span>{data.published_at}</span>}
              {data.author && <span>· {data.author}</span>}
            </div>

            <div style={{fontSize:14,lineHeight:1.8,color:'var(--text)',whiteSpace:'pre-wrap',marginBottom:20,maxHeight:translated?'50vh':'none',overflowY:translated?'auto':'visible'}}>
              {data.body_text}
            </div>

            {translated && (
              <>
                <div style={{fontFamily:'var(--font-display)',fontSize:14,fontWeight:600,color:'var(--gold)',letterSpacing:'0.05em',textTransform:'uppercase',marginBottom:10,paddingBottom:8,borderBottom:'1px solid var(--border)'}}>
                  中文翻译
                </div>
                <div style={{fontSize:14,lineHeight:1.8,color:'var(--text-secondary)',whiteSpace:'pre-wrap',marginBottom:20}}>
                  {translated}
                </div>
              </>
            )}

            <div style={{display:'flex',gap:12,justifyContent:'center',paddingTop:14,borderTop:'1px solid var(--border)'}}>
              {cfg?.configured && !translated && (
                <button onClick={doTranslate} disabled={translating} style={{
                  padding:'8px 20px',background:'var(--gold)',color:'#fff',border:'none',borderRadius:'var(--radius-sm)',
                  fontSize:14,fontWeight:600,fontFamily:'var(--font-display)',letterSpacing:'0.04em',textTransform:'uppercase',
                  cursor:translating?'not-allowed':'pointer',opacity:translating?0.5:1,
                }}>
                  {translating ? '翻译中...' : '翻译正文'}
                </button>
              )}
              {data.link && (
                <a href={data.link} target="_blank" rel="noopener noreferrer" style={{
                  padding:'8px 20px',background:'var(--input-bg)',color:'var(--text-secondary)',border:'1px solid var(--border)',
                  borderRadius:'var(--radius-sm)',fontSize:14,textDecoration:'none',
                }}>
                  在 HLTV 阅读原文 →
                </a>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  )
}
```

Note: The Chinese strings above use Unicode escapes. In actual file, use: 将以下CS电竞新闻正文翻译为简体中文, 抓取中..., 文章暂时不可用, 中文翻译, 翻译正文, 翻译中..., 在 HLTV 阅读原文 →.

- [ ] **Step 3: Wire NewsDetail into News.tsx**

In `frontend/src/pages/News.tsx`:
1. Add import: `import NewsDetail from '../components/NewsDetail'`
2. Add state: `const [selectedNewsUrl, setSelectedNewsUrl] = useState<string | null>(null)`
3. Add `onClick={() => setSelectedNewsUrl(n.link)}` to each news item div
4. Add `style={{cursor:'pointer'}}` to each news item div
5. At end of component (before closing `</div>`), add:
```tsx
{selectedNewsUrl && <NewsDetail url={selectedNewsUrl} onClose={() => setSelectedNewsUrl(null)} />}
```

- [ ] **Step 4: Rebuild frontend and verify**

```bash
cd frontend && npm run build && cd ..
go build github.com/arcdent/hltv-mcp
```

Expected: both compile successfully.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/NewsDetail.tsx frontend/src/pages/News.tsx frontend/src/api/client.ts
git commit -m "feat: add NewsDetail modal with infinite-cache article body and API translation"
```

---

## Final Verification

### Task 15: Rebuild, run, and verify all features

- [ ] **Step 1: Full rebuild**

```bash
cd frontend && npm run build && cd ..
go build -o hltv-mcp github.com/arcdent/hltv-mcp
```

- [ ] **Step 2: Restart server**

```bash
# Kill existing process and restart
pkill hltv-mcp || true
./hltv-mcp &
sleep 2
curl -s http://localhost:8082/api/health
```

- [ ] **Step 3: Verify README**

```bash
grep -A5 "mcpServers" README.md
```

Expected: finds the standard stdio config block.

- [ ] **Step 4: Verify team detail**

```bash
curl -s http://localhost:8082/api/teams/9565 | python3 -c "
import json,sys
d=json.load(sys.stdin).get('data',{})
p=d.get('profile',{})
r=d.get('ranking',{})
s=d.get('stats',{})
a=d.get('achievements',[])
roster=d.get('roster',[])
rm=d.get('recent_matches',[])
print(f'Team: {p.get(\"name\")}')
print(f'Rank: #{r.get(\"world_rank\")} ({r.get(\"points\")}pts)')
print(f'Stats: {s.get(\"wins\")}W/{s.get(\"losses\")}L/{s.get(\"draws\")}D (Form: {s.get(\"recent_form\")})')
print(f'Achievements: {len(a)}')
print(f'Roster: {len(roster)} players')
print(f'Recent matches: {len(rm)}')
"
```

Expected: All fields populated. Roster may be 0 (if HLTV page changed) — this is acceptable per fallback spec.

- [ ] **Step 5: Verify events API**

```bash
curl -s 'http://localhost:8082/api/events?type=upcoming&limit=20' | python3 -c "
import json,sys
d=json.load(sys.stdin).get('data',{})
events=d.get('events',[])
other=d.get('other',[])
print(f'Events: {len(events)}, Other: {len(other)}')
for e in events[:3]:
    print(f'  {e[\"name\"]}: {e[\"match_count\"]} matches ({e[\"date_start\"]}~{e[\"date_end\"]})')
"
```

Expected: events grouped by name, "Other" bucket for empty-event matches.

- [ ] **Step 6: Verify news article API (if CF test passed)**

```bash
ARTICLE_URL=$(curl -s 'http://localhost:8082/api/news?limit=1' | python3 -c "import json,sys;print(json.load(sys.stdin)['items'][0].get('link',''))")
if [ -n "$ARTICLE_URL" ]; then
  ENCODED=$(python3 -c "import urllib.parse;print(urllib.parse.quote('$ARTICLE_URL'))")
  curl -s "http://localhost:8082/api/news/article?url=$ENCODED" | python3 -c "
import json,sys
d=json.load(sys.stdin).get('data',{})
print(f'Title: {d.get(\"title\",\"?\")}')
print(f'Body chars: {len(d.get(\"body_text\",\"\"))}')
"
fi
```

Expected: body_text > 100 characters.

- [ ] **Step 7: Final commit**

```bash
git add -A
git commit -m "chore: final rebuild and verification after all 4 features"
```
