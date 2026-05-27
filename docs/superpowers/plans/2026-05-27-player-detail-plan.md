# 选手详情卡片 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 选手搜索结果点击弹出详情卡片，展示 chromedp 抓取的完整资料/能力评分/Top20/生涯/荣誉/近期比赛，无数据时布局自适应收缩。

**Architecture:** Go 后端用 chromedp 抓取 HLTV 选手详情页 → normalizer 标准化 → handler 返回 JSON。前端 SearchableList 点击触发 → PlayerDetail 居中模态卡片。

**Tech Stack:** Go 1.26 + chromedp + goquery, React 18 + TypeScript, Tailwind CSS v4

**Spec:** `docs/superpowers/specs/2026-05-27-player-detail.md`

---

### Task 1: Go 后端 — chromedp 抓取 + normalizer + handler

**Files:**
- Modify: `internal/scraper/player.go` — 追加 `GetPlayerDetail`
- Modify: `internal/normalizer/player.go` — 追加 `NormalizePlayerDetail`
- Modify: `internal/http/handlers/search.go` — 替换桩实现

**New Go types** (add to `internal/types/types.go`):
```go
type PlayerDetail struct {
	Profile       PlayerDetailProfile  `json:"profile"`
	Rating        PlayerRating         `json:"rating"`
	Abilities     []PlayerAbility      `json:"abilities"`
	Career        PlayerCareer         `json:"career"`
	Top20Ranks    map[string]int       `json:"top20_ranks"`
	Honors        []PlayerHonor        `json:"honors,omitempty"`
	RecentMatches []PlayerRecentMatch  `json:"recent_matches,omitempty"`
}
type PlayerDetailProfile struct {
	ID int `json:"id"`; Name string `json:"name"`; RealName string `json:"real_name,omitempty"`
	Slug string `json:"slug"`; Country string `json:"country,omitempty"`
	Age int `json:"age,omitempty"`; Team string `json:"team,omitempty"`; PrizeMoney string `json:"prize_money,omitempty"`
}
type PlayerRating struct { Value float64 `json:"value"`; Maps int `json:"maps"` }
type PlayerAbility struct { Key, LabelEn, LabelZh string; Value float64 `json:"value"`; Max int `json:"max"`; Format string `json:"format,omitempty"` }
type PlayerCareer struct {
	Rating float64 `json:"rating"`; Matches int `json:"matches"`; WinRate string `json:"win_rate"`
	KD float64 `json:"kd"`; HeadshotPct string `json:"headshot_pct"`; WinStreak int `json:"win_streak"`
}
type PlayerHonor struct { Label string `json:"label"`; Value int `json:"value"` }
type PlayerRecentMatch struct {
	Date, Team, Opponent, Score, Result, Event string
	Rating float64 `json:"rating"`; Kills, Deaths int `json:"kills"`
}
```

- [ ] **Step 1: Add GetPlayerDetail to scraper**

In `internal/scraper/player.go`, add after `GetPlayerOverview`:

```go
func (s *PlayerScraper) GetPlayerDetail(ctx context.Context, id int, slug string) (*goquery.Document, error) {
	path := fmt.Sprintf("/player/%d/%s", id, url.PathEscape(slug))
	body, err := s.cli.FetchHTML(ctx, path, "player_detail")
	if err != nil { return nil, err }
	return goquery.NewDocumentFromReader(bytesReader(body))
}
```

- [ ] **Step 2: Add NormalizePlayerDetail to normalizer**

In `internal/normalizer/player.go`, add:

```go
func NormalizePlayerDetail(doc *goquery.Document) types.PlayerDetail {
	pd := types.PlayerDetail{}

	// Profile
	pd.Profile.Name = cleanText(doc.Find(".playerNickname").First().Text())
	if pd.Profile.Name == "" { return pd }
	pd.Profile.RealName = cleanText(doc.Find(".playerRealname").First().Text())
	pd.Profile.Team = cleanText(doc.Find(".playerTeam a").First().Text())
	pd.Profile.Country, _ = doc.Find("img.flag").First().Attr("title")
	// Age: find span containing "Age"
	doc.Find(".playerAge, .player-info span, .listRight").Each(func(_ int, s *goquery.Selection) {
		t := cleanText(s.Text())
		if strings.Contains(t, "years") && pd.Profile.Age == 0 {
			fmt.Sscanf(t, "%d", &pd.Profile.Age)
		}
		if strings.HasPrefix(t, "$") && pd.Profile.PrizeMoney == "" {
			pd.Profile.PrizeMoney = t
		}
	})

	// Rating
	doc.Find(".player-stat").Each(func(_ int, s *goquery.Selection) {
		label := cleanText(s.Find(".player-stat-top, .stat-label").Text())
		if strings.Contains(label, "Rating") && pd.Rating.Value == 0 {
			v, _ := strconv.ParseFloat(cleanText(s.Find(".statsVal").Text()), 64)
			pd.Rating.Value = v
		}
	})
	// Maps count
	if t := cleanText(doc.Find(".stats-window").Text()); t != "" {
		fmt.Sscanf(t, "Past 3 months • %d maps", &pd.Rating.Maps)
	}

	// Abilities (8 axes)
	abilityKeys := []string{"rating","firepower","opening","clutching","sniping","entrying","trading","utility"}
	abilityLabels := map[string][2]string{
		"rating":{"Rating","综合"},"firepower":{"Firepower","火力"},"opening":{"Opening","突破"},
		"clutching":{"Clutching","残局"},"sniping":{"Sniping","狙击"},"entrying":{"Entrying","进点"},
		"trading":{"Trading","补枪"},"utility":{"Utility","道具"},
	}
	for _, key := range abilityKeys {
		ab := types.PlayerAbility{Key:key, LabelEn:abilityLabels[key][0], LabelZh:abilityLabels[key][1], Max:100}
		if key == "rating" { ab.Value = pd.Rating.Value; ab.Max = 0; ab.Format = "decimal" }
		doc.Find(".player-stat").Each(func(_ int, s *goquery.Selection) {
			if strings.Contains(strings.ToLower(cleanText(s.Text())), strings.ToLower(ab.LabelEn)) && ab.Value == 0 {
				v, _ := strconv.ParseFloat(cleanText(s.Find(".statsVal").Text()), 64)
				ab.Value = v
			}
		})
		pd.Abilities = append(pd.Abilities, ab)
	}

	// Career stats from .all-time-stat
	doc.Find(".all-time-stat").Each(func(_ int, s *goquery.Selection) {
		t := cleanText(s.Text())
		if strings.Contains(t, "KDR") { v,_:=strconv.ParseFloat(cleanText(s.Find(".stat").Text()),64); pd.Career.KD=v }
		if strings.Contains(t, "Win rate") { pd.Career.WinRate=cleanText(s.Find(".stat").Text()) }
		if strings.Contains(t, "Headshots") { pd.Career.HeadshotPct=cleanText(s.Find(".stat").Text()) }
		if strings.Contains(t, "Win streak") { v,_:=strconv.Atoi(cleanText(s.Find(".stat").Text())); pd.Career.WinStreak=v }
	})
	doc.Find(".highlighted-stat").Each(func(_ int, s *goquery.Selection) {
		t := cleanText(s.Text())
		if strings.Contains(t, "Matches") && pd.Career.Matches==0 { v,_:=strconv.Atoi(cleanText(s.Find(".stat").Text())); pd.Career.Matches=v }
	})

	// Top 20 from profile text
	profileText := cleanText(doc.Find(".playerInfo, .player-profile, .playerSummaryContainer, .profile-summary").Text())
	re := regexp.MustCompile(`#(\d)\s*\('(\d{2})\)`)
	if matches := re.FindAllStringSubmatch(profileText, -1); len(matches) > 0 {
		pd.Top20Ranks = make(map[string]int)
		for _, m := range matches {
			year := "20" + m[2]; rank, _ := strconv.Atoi(m[1])
			pd.Top20Ranks[year] = rank
		}
	}

	// Honors
	honorMap := map[string]*struct{label string; count int}{}
	doc.Find(".highlighted-stat").Each(func(_ int, s *goquery.Selection) {
		t := cleanText(s.Text())
		if strings.Contains(t, "Majors won") { v,_:=strconv.Atoi(cleanText(s.Find(".stat").Text())); pd.Honors=append(pd.Honors,types.PlayerHonor{Label:"Major 冠军",Value:v}) }
		if strings.Contains(t, "Total MVPs") { v,_:=strconv.Atoi(cleanText(s.Find(".stat").Text())); pd.Honors=append(pd.Honors,types.PlayerHonor{Label:"总 MVP",Value:v}) }
		if strings.Contains(t, "Major MVPs") || strings.Contains(t, "Major MVP") { v,_:=strconv.Atoi(cleanText(s.Find(".stat").Text())); pd.Honors=append(pd.Honors,types.PlayerHonor{Label:"Major MVP",Value:v}) }
	})
	_ = honorMap

	// Recent matches
	doc.Find(".recent-matches tbody tr, .matches-table tbody tr, .result-row").Each(func(i int, s *goquery.Selection) {
		if i >= 7 { return }
		cells := s.Find("td")
		if cells.Length() < 5 { return }
		m := types.PlayerRecentMatch{
			Date: cleanText(cells.Eq(0).Text()), Team: cleanText(cells.Eq(1).Text()),
			Opponent: cleanText(cells.Eq(2).Text()), Score: cleanText(cells.Eq(3).Text()),
			Event: cleanText(cells.Eq(4).Text()),
		}
		if r, err := strconv.ParseFloat(cleanText(cells.Eq(5).Text()), 64); err == nil { m.Rating = r }
		pd.RecentMatches = append(pd.RecentMatches, m)
	})

	return pd
}
```

- [ ] **Step 3: Replace search.go handler stub**

Replace `GetPlayer` stub in `internal/http/handlers/search.go`:

```go
func (h *Handlers) GetPlayer(w http.ResponseWriter, r *http.Request) {
	id := atoi(chi.URLParam(r, "id"))
	if id == 0 { writeError(w, http.StatusBadRequest, "invalid player id"); return }
	
	// Resolve slug via player search
	resolveResp := h.f.ResolvePlayer(types.ResolveQuery{Name: "", Limit: 1})
	// Use the ID directly with a fallback slug
	slug := fmt.Sprintf("player-%d", id)
	
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	
	doc, err := h.f.ScrapePlayerDetail(ctx, id, slug)
	if err != nil {
		writeJSON(w, map[string]any{
			"error": map[string]any{"code":"UPSTREAM_UNAVAILABLE","message":"detail unavailable"},
			"meta": map[string]any{"partial": true},
		})
		return
	}
	pd := normalizer.NormalizePlayerDetail(doc)
	pd.Profile.ID = id; pd.Profile.Slug = slug
	writeJSON(w, map[string]any{"data": pd, "meta": map[string]any{"partial": false}})
}
```

Note: Add to chi mux import `"github.com/go-chi/chi/v5"` and `context`, `"time"`, `"strconv"`, `"regexp"`, `"strings"`, `"fmt"` imports as needed in the normalizer file.

- [ ] **Step 4: Expose ScrapePlayerDetail on facade**

Add to `internal/facade/facade.go`:

```go
func (f *HltvFacade) ScrapePlayerDetail(ctx context.Context, id int, slug string) (*goquery.Document, error) {
	return f.ps.GetPlayerDetail(ctx, id, slug)
}
```

- [ ] **Step 5: Build and verify**

```bash
go build github.com/arcdent/hltv-mcp/...
```
Expected: success

- [ ] **Step 6: Commit**

```bash
git add internal/scraper/player.go internal/normalizer/player.go internal/http/handlers/search.go internal/facade/facade.go internal/types/types.go
git commit -m "feat: add player detail chromedp scraping + normalizer + API handler"
```

---

### Task 2: 前端 — PlayerDetail 模态卡片

**Files:**
- Create: `frontend/src/components/PlayerDetail.tsx`

- [ ] **Step 1: Write PlayerDetail.tsx**

This is a large component. Write exact code:

```tsx
import { useEffect, useState } from 'react'

type PlayerData = {
  profile: { id: number; name: string; real_name?: string; slug: string; country?: string; age?: number; team?: string; prize_money?: string }
  rating: { value: number; maps: number }
  abilities: { key: string; label_en: string; label_zh: string; value: number; max: number; format?: string }[]
  career: { rating?: number; matches?: number; win_rate?: string; kd?: number; headshot_pct?: string; win_streak?: number }
  top20_ranks?: Record<string, number>
  honors?: { label: string; value: number }[]
  recent_matches?: { date: string; team: string; opponent: string; score: string; result: string; rating: number; kills: number; deaths: number; event: string }[]
}

export default function PlayerDetail({ id, onClose }: { id: number; onClose: () => void }) {
  const [data, setData] = useState<PlayerData | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    fetch(`/api/players/${id}`).then(r => r.json()).then(d => {
      setData(d.data ?? null); setLoading(false)
    }).catch(() => setLoading(false))
  }, [id])

  const p = data?.profile

  // Abilities sorted: rating first, then by value desc for the rest
  const abilities = data?.abilities ?? []
  const radarPoints = abilities.slice(0, 8)

  // Top 20 ranks sorted by year desc
  const top20 = data?.top20_ranks ? Object.entries(data.top20_ranks).sort((a,b) => Number(b[0])-Number(a[0])) : []

  const rankClass = (r: number) => r === 1 ? 'rank-1' : r === 2 ? 'rank-2' : r === 3 ? 'rank-3' : 'rank-other'

  return (
    <div onClick={onClose} style={{position:'fixed',inset:0,zIndex:100,background:'rgba(0,0,0,0.5)',backdropFilter:'blur(4px)',display:'flex',alignItems:'center',justifyContent:'center',animation:'fadeIn 0.2s ease'}}>
      <div onClick={e => e.stopPropagation()} style={{background:'var(--card)',border:'1px solid var(--border)',borderRadius:'var(--radius)',width:580,maxWidth:'90vw',maxHeight:'90vh',overflowY:'auto',padding:28,boxShadow:'0 20px 60px rgba(0,0,0,0.3)',animation:'slideUp 0.25s ease'}}>
        
        {/* Close button */}
        <button onClick={onClose} style={{position:'absolute',top:16,right:16,width:30,height:30,borderRadius:'50%',border:'1px solid var(--border)',background:'var(--card)',color:'var(--text-secondary)',fontSize:16,cursor:'pointer',display:'flex',alignItems:'center',justifyContent:'center'}}>✕</button>

        {loading && <div style={{textAlign:'center',padding:60,color:'var(--text-muted)'}}>加载中...</div>}
        {!loading && !p && <div style={{textAlign:'center',padding:60,color:'var(--text-muted)'}}>详情暂时不可用</div>}
        
        {!loading && p && (
          <>
            {/* Profile header */}
            <div style={{display:'flex',gap:14,marginBottom:14}}>
              <div style={{width:56,height:56,borderRadius:'50%',background:'linear-gradient(135deg,var(--gold),#c48a0a)',display:'flex',alignItems:'center',justifyContent:'center',fontSize:24,color:'#fff',fontWeight:700,fontFamily:'var(--font-display)',flexShrink:0}}>
                {p.name.charAt(0)}
              </div>
              <div style={{flex:1}}>
                <div style={{fontSize:22,fontWeight:700,color:'var(--text)',lineHeight:1.2}}>{p.name}</div>
                {p.real_name && <div style={{fontSize:13,color:'var(--text-muted)'}}>{p.real_name}</div>}
                <div style={{display:'flex',flexWrap:'wrap',gap:6,marginTop:6}}>
                  {p.country && <span style={{padding:'2px 8px',background:'var(--input-bg)',borderRadius:4,fontSize:11,color:'var(--text-secondary)'}}>{p.country}</span>}
                  {p.age && <span style={{padding:'2px 8px',background:'var(--input-bg)',borderRadius:4,fontSize:11,color:'var(--text-secondary)'}}>Age {p.age}</span>}
                  {p.team && <span style={{padding:'2px 8px',background:'var(--gold-dim)',borderRadius:4,fontSize:11,color:'var(--gold)',fontWeight:600}}>{p.team}</span>}
                  {p.prize_money && <span style={{padding:'2px 8px',background:'var(--input-bg)',borderRadius:4,fontSize:11,color:'var(--text-secondary)'}}>{p.prize_money}</span>}
                </div>
              </div>
            </div>

            {/* Top 20 pills */}
            {top20.length > 0 && (
              <div style={{display:'flex',gap:4,flexWrap:'wrap',justifyContent:'center',marginBottom:16}}>
                {top20.map(([year, rank]) => (
                  <span key={year} className={rankClass(rank)} style={{fontSize:11,fontWeight:700,padding:'2px 7px',borderRadius:10,display:'flex',alignItems:'center',gap:2,
                    background: rank===1?'linear-gradient(135deg,#f0c040,#c48a0a)':rank===2?'#e0e0d8':rank===3?'rgba(196,138,10,0.15)':'#f0f0e8',
                    color: rank===1?'#fff':rank===2?'#6b7280':rank===3?'#c48a0a':'#9ca3af'}}>{year} #{rank}</span>
                ))}
              </div>
            )}

            {/* Abilities — radar + labels centered */}
            <h2 style={{fontFamily:'var(--font-display)',fontSize:16,fontWeight:600,color:'var(--gold)',letterSpacing:'0.06em',textTransform:'uppercase',marginBottom:12,paddingBottom:8,borderBottom:'1px solid var(--border)'}}>
              能力评分 &nbsp;<span style={{fontSize:13,fontWeight:400,color:'var(--text-muted)'}}>近 3 月 · {data.rating.maps} maps</span>
            </h2>
            <div style={{display:'flex',justifyContent:'center',alignItems:'center',gap:24,marginBottom:16}}>
              {/* Radar SVG (8 axes) */}
              <svg width="140" height="140" viewBox="0 0 140 140">
                {[66,48,30,12].map(r => <circle key={r} cx="70" cy="70" r={r} fill="none" stroke="var(--border)" strokeWidth="1"/>)}
                {[0,45,90,135].map(a => (
                  <line key={a} x1={70+66*Math.cos(a*Math.PI/180)} y1={70+66*Math.sin(a*Math.PI/180)} x2={70-66*Math.cos(a*Math.PI/180)} y2={70-66*Math.sin(a*Math.PI/180)} stroke="var(--border)" strokeWidth="0.5"/>
                ))}
                {radarPoints.length === 8 && (
                  <polygon
                    points={radarPoints.map((ab,i) => {
                      const angle = (i*45-90)*Math.PI/180
                      const v = ab.format === 'decimal' ? ab.value/2*66 : ab.value/100*66
                      return `${(70+v*Math.cos(angle)).toFixed(0)},${(70+v*Math.sin(angle)).toFixed(0)}`
                    }).join(' ')}
                    fill="rgba(196,138,10,0.12)" stroke="var(--gold)" strokeWidth="1.5"
                  />
                )}
              </svg>
              <div style={{display:'flex',flexDirection:'column',gap:3,fontSize:11,color:'var(--text-secondary)'}}>
                {abilities.map(ab => (
                  <div key={ab.key} style={{display:'flex',alignItems:'center',gap:6,opacity:ab.value===0?0.4:1}}>
                    <span style={{width:7,height:7,borderRadius:2,background:ab.value>0?'var(--gold)':'var(--border)',flexShrink:0}}/>
                    <span style={{minWidth:120}}>{ab.label_en} ({ab.label_zh})</span>
                    <b style={{color:ab.value>0?'var(--text)':'var(--text-muted)',fontFamily:'var(--font-mono)',fontSize:12}}>
                      {ab.format==='decimal'?ab.value.toFixed(2):ab.value>0?`${ab.value}/${ab.max}`:'—'}
                    </b>
                  </div>
                ))}
              </div>
            </div>

            {/* Career row */}
            {(data.career.rating || data.career.matches) && (
              <div style={{display:'flex',alignItems:'center',justifyContent:'center',gap:16,marginBottom:14,fontSize:13,color:'var(--text-secondary)'}}>
                {data.career.rating && <><span><span style={{fontFamily:'var(--font-display)',fontSize:20,fontWeight:700,color:'var(--text)'}}>{data.career.rating}</span> 生涯 Rating</span><span style={{color:'var(--border)'}}>|</span></>}
                {data.career.matches && <><span style={{fontFamily:'var(--font-display)',fontSize:20,fontWeight:700,color:'var(--text)'}}>{data.career.matches}</span> 比赛</>}
                {data.career.win_rate && <><span style={{color:'var(--border)'}}>|</span><span style={{fontFamily:'var(--font-display)',fontSize:20,fontWeight:700,color:'var(--text)'}}>{data.career.win_rate}</span> 胜率</>}
                {data.career.kd && <><span style={{color:'var(--border)'}}>|</span><span style={{fontFamily:'var(--font-display)',fontSize:20,fontWeight:700,color:'var(--text)'}}>{data.career.kd}</span> K/D</>}
              </div>
            )}

            {/* Honors row */}
            {data.honors && data.honors.length > 0 && (
              <div style={{display:'flex',gap:6,flexWrap:'wrap',justifyContent:'center',marginBottom:14}}>
                {data.honors.map(h => (
                  <span key={h.label} style={{fontSize:11,padding:'2px 10px',borderRadius:10,background:'rgba(196,138,10,0.06)',color:'var(--gold)',fontWeight:500}}>
                    {h.label} {h.value}×
                  </span>
                ))}
                {data.career.headshot_pct && <span style={{fontSize:11,padding:'2px 10px',borderRadius:10,background:'rgba(196,138,10,0.06)',color:'var(--gold)',fontWeight:500}}>爆头率 {data.career.headshot_pct}</span>}
                {data.career.win_streak > 0 && <span style={{fontSize:11,padding:'2px 10px',borderRadius:10,background:'rgba(196,138,10,0.06)',color:'var(--gold)',fontWeight:500}}>{data.career.win_streak} 连胜</span>}
              </div>
            )}

            {/* Recent matches */}
            {data.recent_matches && data.recent_matches.length > 0 && (
              <>
                <h2 style={{fontFamily:'var(--font-display)',fontSize:16,fontWeight:600,color:'var(--gold)',letterSpacing:'0.06em',textTransform:'uppercase',marginBottom:10,paddingBottom:8,borderBottom:'1px solid var(--border)'}}>近期 7 场比赛</h2>
                {data.recent_matches.map((m,i) => (
                  <div key={i} style={{display:'flex',alignItems:'center',gap:8,padding:'9px 0',borderBottom:i<data.recent_matches!.length-1?'1px solid rgba(0,0,0,0.04)':'none',fontSize:12}}>
                    <span style={{fontSize:10,color:'var(--text-muted)',minWidth:44}}>{m.date}</span>
                    <span style={{fontFamily:'var(--font-mono)',fontSize:12,fontWeight:700,color:'var(--gold)',minWidth:32,textAlign:'center'}}>{m.rating||'—'}</span>
                    <span style={{flex:1,minWidth:0}}>
                      <span style={{fontWeight:600}}>{m.team}</span> vs {m.opponent}
                      <div style={{fontSize:10,color:'var(--text-muted)',whiteSpace:'nowrap',overflow:'hidden',textOverflow:'ellipsis'}}>{m.event}</div>
                    </span>
                    <span style={{fontFamily:'var(--font-display)',fontSize:15,fontWeight:700,minWidth:30,textAlign:'center',color:m.result==='win'?'var(--green)':m.result==='loss'?'var(--red)':'var(--text)'}}>{m.score||'-:-'}</span>
                    <span style={{fontFamily:'var(--font-mono)',fontSize:11,color:'var(--text-secondary)',minWidth:50,textAlign:'center'}}>{m.kills>0?`${m.kills}-${m.deaths}`:'—'}</span>
                  </div>
                ))}
              </>
            )}
          </>
        )}
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Build and verify**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild/frontend && npx vite build --outDir ../dist 2>&1 | tail -1
```
Expected: `✓ built`

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/PlayerDetail.tsx
git commit -m "feat: add PlayerDetail modal card — radar, career stats, honors, recent matches"
```

---

### Task 3: 集成 — SearchableList 点击展开 + 最终验证

**Files:**
- Modify: `frontend/src/components/SearchableList.tsx` — 加 onClick + PlayerDetail

- [ ] **Step 1: Update SearchableList to handle click-to-expand**

Add import at top: `import PlayerDetail from './PlayerDetail'`

Add state: `const [selectedId, setSelectedId] = useState<number | null>(null)`

On each list item, add `onClick={() => setSelectedId(item.id)}` and `cursor:'pointer'` style.

Add at bottom of the component (before closing `</>`):

```tsx
{selectedId !== null && <PlayerDetail id={selectedId} onClose={() => setSelectedId(null)} />}
```

- [ ] **Step 2: Full build**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild/frontend && npx vite build --outDir ../dist && cd .. && go build -o hltv-mcp github.com/arcdent/hltv-mcp
```
Expected: both succeed

- [ ] **Step 3: Smoke test**

```bash
fuser -k 8082/tcp 2>/dev/null; sleep 0.5; ./hltv-mcp &
sleep 2
# Test player API
curl -s http://localhost:8082/api/players/11893 | head -c 200
kill %1
```

- [ ] **Step 4: Commit and push**

```bash
git add -A
git commit -m "feat: integrate PlayerDetail into SearchableList — click to expand"
git push
```
