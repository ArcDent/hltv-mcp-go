# Match Display Date/Time Fix & UI Polish

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix date extraction from results/matches HTML pages, add pagination for team history, polish UI alignment

**Architecture:** Backend: modify `normalizeResultsCon` to traverse `.results-sublist` sections for dates (regex from headline), modify `NormalizeUpcomingMatches` to track `.matches-list-headline` dates via sibling traversal, add pagination loop in `GetTeamDetailCached`. Frontend: center match rows in TeamDetail, widen event name column, format MM/DD dates in event cards and modal detail

**Tech Stack:** Go 1.26, goquery, React 18, TypeScript

---

## File Map

| File | Change | Purpose |
|------|--------|---------|
| `internal/normalizer/match.go` | Modify 2 funcs | Date extraction from results (`.results-sublist`) + matches (`.matches-list-headline` sibling traversal) |
| `internal/facade/facade.go` | Modify `GetTeamDetailCached` | Paginate results offset 0/100/200; filter per page until 10 matches |
| `frontend/src/components/TeamDetail.tsx` | Modify lines 147-158 | Center match rows; widen event name column |
| `frontend/src/pages/Matches.tsx` | Modify lines 99-101, 137-178 | Event card MM/DD dates; modal date+time for scheduled matches |

---

### Task 1: Results page date extraction from section headlines

**Files:**
- Modify: `internal/normalizer/match.go:19-67` (`normalizeResultsCon`)

**Context:** HLTV `/results` page is organized as `.results-all > .results-sublist`. Each sublist has a `.standard-headline` (e.g. "Results for May 28th 2026") + many `.result-con` rows. Currently `.time, .date` selectors find nothing — dates must come from the headline.

- [ ] **Step 1: Add month name → number map and date parse function**

At top of `internal/normalizer/match.go`, after imports:

```go
import "regexp"

var monthMap = map[string]string{
	"January": "01", "February": "02", "March": "03", "April": "04",
	"May": "05", "June": "06", "July": "07", "August": "08",
	"September": "09", "October": "10", "November": "11", "December": "12",
}

var resultsDateRe = regexp.MustCompile(`Results for (\w+) (\d+)(?:st|nd|rd|th)? (\d{4})`)

func parseResultsDate(headline string) string {
	m := resultsDateRe.FindStringSubmatch(headline)
	if len(m) != 4 {
		return ""
	}
	month, ok := monthMap[m[1]]
	if !ok {
		return ""
	}
	day := m[2]
	if len(day) == 1 {
		day = "0" + day
	}
	return m[3] + "-" + month + "-" + day
}
```

- [ ] **Step 2: Rewrite `normalizeResultsCon` to walk `.results-sublist` sections**

Replace the function body:

```go
func normalizeResultsCon(doc *goquery.Document, perspective string) []types.NormalizedMatch {
	var matches []types.NormalizedMatch
	doc.Find(".results-sublist").Each(func(_ int, sublist *goquery.Selection) {
		date := parseResultsDate(cleanText(sublist.Find(".standard-headline").First().Text()))
		sublist.Find(".result-con").Each(func(_ int, s *goquery.Selection) {
			m := types.NormalizedMatch{Result: types.OutcomeUnknown}

			m.Team1 = cleanText(s.Find(".line-align.team1 .team").First().Text())
			m.Team2 = cleanText(s.Find(".line-align.team2 .team").First().Text())

			if score := cleanText(s.Find(".result-score").First().Text()); score != "" {
				m.Score = score
			}

			m.Event = cleanText(s.Find(".event-name, .map-text, .stars").First().Text())

			if href, ok := s.Find("a.a-reset").First().Attr("href"); ok && href != "" {
				if id := parseMatchID(href); id > 0 {
					m.MatchID = id
				}
			}

			m.PlayedAt = date

			if perspective != "" {
				if m.Team1 == perspective {
					m.Opponent = m.Team2
				} else if m.Team2 == perspective {
					m.Opponent = m.Team1
				}
			}

			if m.Team1 != "" || m.Team2 != "" {
				matches = append(matches, m)
			}
		})
	})
	return matches
}
```

- [ ] **Step 3: Build and verify**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 4: Verify API returns dates**

Run: `curl -s 'http://172.21.32.31:8082/api/events?type=results&limit=5' | python3 -c "import json,sys; d=json.load(sys.stdin); [print(m.get('played_at','?')) for e in d['data']['events'] for m in e['matches'][:2]]"`
Expected: dates like `2026-05-28`, NOT empty strings

- [ ] **Step 5: Commit**

```bash
git add internal/normalizer/match.go
git commit -m "fix: extract results dates from .results-sublist section headlines"
```

---

### Task 2: Matches page date extraction from .matches-list-headline

**Files:**
- Modify: `internal/normalizer/match.go:69-139` (`NormalizeUpcomingMatches`)

**Context:** HLTV `/matches` page has date headers `.matches-list-headline` (e.g. "Thursday - 2026-05-28") as siblings to match-wrapper containers. Live matches (`.liveMatches`) have no headline — use today's date. `scheduled_at` changes from `HH:MM` to `YYYY-MM-DD HH:MM`.

- [ ] **Step 1: Rewrite `NormalizeUpcomingMatches` with sibling traversal**

```go
func NormalizeUpcomingMatches(doc *goquery.Document, perspective string) []types.NormalizedMatch {
	var matches []types.NormalizedMatch
	currentDate := strings.Split(time.Now().UTC().Format(time.RFC3339), "T")[0] // "2006-01-02"

	// Find parent container holding both headlines and match-wrappers
	headline := doc.Find(".matches-list-headline").First()
	var parent *goquery.Selection
	if headline.Length() > 0 {
		parent = headline.Parent()
	}
	if parent.Length() == 0 {
		parent = doc.Find("body").First()
	}

	parent.Children().Each(func(_ int, child *goquery.Selection) {
		if child.HasClass("matches-list-headline") {
			text := cleanText(child.Text())
			if idx := strings.LastIndex(text, "- "); idx >= 0 {
				currentDate = strings.TrimSpace(text[idx+2:])
			}
			return
		}
		child.Find(".match").Each(func(_ int, s *goquery.Selection) {
			m := types.NormalizedMatch{Result: types.OutcomeScheduled}

			m.Event = cleanText(s.Find(".match-top").First().Text())

			infoText := cleanText(s.Find(".match-info").First().Text())
			if idx := strings.Index(infoText, " "); idx > 0 {
				m.ScheduledAt = currentDate + " " + infoText[:idx]
				m.BestOf = cleanText(infoText[idx:])
			} else {
				m.ScheduledAt = currentDate + " " + infoText
			}

			teamsText := cleanText(s.Find(".match-teams").First().Text())
			teamsText = strings.ReplaceAll(teamsText, "\n", " ")
			teamsText = strings.ReplaceAll(teamsText, "  ", " ")
			if idx := strings.Index(teamsText, " vs "); idx > 0 {
				m.Team1 = cleanText(teamsText[:idx])
				m.Team2 = cleanText(teamsText[idx+4:])
			} else {
				parts := strings.Fields(teamsText)
				if len(parts) >= 2 {
					m.Team1 = parts[0]
					if strings.ToLower(parts[len(parts)-2]) == "vs" {
						m.Team2 = parts[len(parts)-1]
					} else {
						m.Team2 = parts[len(parts)-1]
					}
				}
			}

			s.Find("a").Each(func(_ int, a *goquery.Selection) {
				if href, ok := a.Attr("href"); ok {
					if id := parseMatchID(href); id > 0 {
						m.MatchID = id
					}
				}
			})

			if perspective != "" {
				if m.Team1 == perspective {
					m.Opponent = m.Team2
				} else if m.Team2 == perspective {
					m.Opponent = m.Team1
				}
			}
			m.Team1 = TranslatePlaceholder(m.Team1)
			m.Team2 = TranslatePlaceholder(m.Team2)
			m.Opponent = TranslatePlaceholder(m.Opponent)

			if m.Team1 != "" && m.Team2 != "" {
				matches = append(matches, m)
			}
		})
	})
	return matches
}
```

Add `"time"` to the imports if not already present.

- [ ] **Step 2: Build and verify**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 3: Verify API returns full date-time**

Run: `curl -s 'http://172.21.32.31:8082/api/events?type=upcoming&limit=3' | python3 -c "import json,sys; d=json.load(sys.stdin); [print(m.get('scheduled_at','?')) for e in d['data']['events'] for m in e['matches'][:3]]"`
Expected: `2026-05-28 18:30` format, NOT just `18:30`

- [ ] **Step 4: Commit**

```bash
git add internal/normalizer/match.go
git commit -m "fix: extract upcoming match dates from .matches-list-headline sections"
```

---

### Task 3: Paginate results to find team historical matches

**Files:**
- Modify: `internal/facade/facade.go:119-171` (match fetching block in `GetTeamDetailCached`)

**Context:** The `/results` page only shows ~100 matches (~4 days). Teams with matches older than that get 0 historical data. Loop offset 0/100/200 until 10 team matches found. Filter per page to avoid collecting unnecessary data.

- [ ] **Step 1: Replace the single results fetch with a pagination loop**

Replace lines 119-128 (the `if td.Profile.Name != ""` block up to `if upcomingDoc`):

```go
		if td.Profile.Name != "" {
			name := td.Profile.Name
			var matches []types.NormalizedMatch

			// Paginate through results to find team matches (up to 3 pages)
			for offset := 0; offset < 300 && len(matches) < 10; offset += 100 {
				resultDoc, err := f.rs.GetResultsOffset(ctx, offset)
				if err != nil {
					break
				}
				pageResults := normalizer.NormalizeMatches(resultDoc, name)
				for _, m := range pageResults {
					if m.Team1 == name || m.Team2 == name || m.Opponent == name {
						matches = append(matches, m)
					}
				}
			}

			if upcomingDoc, err := f.ms.GetUpcoming(ctx); err == nil {
				allUpcoming := normalizer.NormalizeUpcomingMatches(upcomingDoc, name)
				for _, m := range allUpcoming {
					if m.Team1 == name || m.Team2 == name || m.Opponent == name {
						matches = append(matches, m)
					}
				}
			}

			normalizer.SortByPlayedAtDesc(matches)
			if len(matches) > 10 {
				matches = matches[:10]
			}
			td.RecentMatches = matches
```

And also update the `allMatches` variable declaration at line 120 — remove `var allMatches []types.NormalizedMatch` since we now declare `matches` inline.

- [ ] **Step 2: Build and verify**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 3: Verify team API returns more matches**

Run: `curl -s 'http://172.21.32.31:8082/api/teams/6665' | python3 -c "import json,sys; d=json.load(sys.stdin)['data']; print(f'Name: {d[\"profile\"][\"name\"]}'); print(f'Matches: {len(d[\"recent_matches\"])}'); [print(f'  {m[\"result\"]:9s} | {m.get(\"score\",\"?\"):5s} | {m.get(\"played_at\",\"?\"):10s} | {m.get(\"event\",\"?\")[:50]}') for m in d['recent_matches']]"`

Expected: More than 1 match (vs the 1 before), with `played_at` filled in

- [ ] **Step 4: Commit**

```bash
git add internal/facade/facade.go
git commit -m "fix: paginate results to find more team historical matches"
```

---

### Task 4: TeamDetail — center match rows + widen event name

**Files:**
- Modify: `frontend/src/components/TeamDetail.tsx:146-159`

**Context:** Two frontend fixes in one file:
1. Match rows are left-aligned (`flex:1` on team name) — change to centered layout
2. Event name column `maxWidth:80` too narrow — widen to `maxWidth:160`

- [ ] **Step 1: Replace the match row layout with centered version**

Replace the match map block (lines 146-159):

```tsx
{matches.map((m, i) => (
  <div key={i} style={{display:'flex',alignItems:'center',justifyContent:'center',gap:10,padding:'7px 0',borderBottom:i<matches.length-1?'1px solid rgba(128,128,128,0.06)':'none',fontSize:12}}>
    <span style={{minWidth:26,textAlign:'center',fontSize:10,fontWeight:700,fontFamily:'var(--font-mono)',padding:'2px 0',borderRadius:3,
      color:m.result==='win'?'var(--green)':m.result==='loss'?'var(--red)':'var(--text-muted)',
      background:m.result==='win'?'rgba(0,200,83,0.1)':m.result==='loss'?'rgba(255,82,82,0.1)':'var(--input-bg)'}}>
      {m.result==='win'?'W':m.result==='loss'?'L':'—'}
    </span>
    <span><b style={{fontWeight:600}}>{p.name}</b></span>
    {m.score ? (
      <span style={{fontFamily:'var(--font-mono)',fontSize:11,color:'var(--text-secondary)'}}>{m.score}</span>
    ) : (
      <span style={{fontFamily:'var(--font-mono)',fontSize:11,color:'var(--text-muted)'}}>vs</span>
    )}
    <span style={{fontWeight:600}}>{m.opponent || m.team2 || '待定'}</span>
    <span style={{fontSize:10,color:'var(--text-muted)',maxWidth:160,overflow:'hidden',textOverflow:'ellipsis',whiteSpace:'nowrap'}}>{m.event || ''}</span>
    <span style={{fontSize:10,color:'var(--text-muted)',minWidth:48,textAlign:'right'}}>{(m.played_at || m.scheduled_at || '').slice(5,10)}</span>
  </div>
))}
```

Key changes:
- Team name + "vs" / score + opponent now all centered as individual `<span>` elements (not one `flex:1` span)
- Event maxWidth: 80→160
- Date column uses `played_at` (historical) or `scheduled_at` (upcoming), both now `YYYY-MM-DD` format, sliced as `MM-DD`

- [ ] **Step 2: Build frontend**

Run: `cd frontend && npm run build`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/TeamDetail.tsx
git commit -m "fix: center match rows and widen event name in team detail"
```

---

### Task 5: Matches page — event card dates + modal date/time display

**Files:**
- Modify: `frontend/src/pages/Matches.tsx:99-101` (event card date line)
- Modify: `frontend/src/pages/Matches.tsx:137-178` (match row in modal)

**Context:** 
1. Event cards show time strings like "01:00 ~ Live" — after backend fix, `date_start`/`date_end` are `YYYY-MM-DD` — format as `MM/DD ~ MM/DD`
2. Modal: scheduled matches show only time (HH:MM) — add date below; right column shows only BO3 (no date)

- [ ] **Step 1: Fix event card date formatting**

Replace lines 99-101:

```tsx
{ev.date_start && ev.date_start.length === 10 ? ev.date_start.slice(5).replace('-','/') : ''}
{ev.date_end && ev.date_start !== ev.date_end ? ' ~ ' + ev.date_end.slice(5).replace('-','/') : ''}
```

This replaces:
```tsx
{ev.date_start || '?'} ~ {ev.date_end || '?'}
```

The old code unconditionally showed "~" even when start==end. New code: only show range when dates differ.

- [ ] **Step 2: Fix modal match row — add date below time, remove date from right column**

Replace the match map block in the modal (lines 137-178). The key changed section is the scheduled time display (replace lines 149-168) and the right column (line 174-176):

```tsx
{(selectedEvent.matches || []).map((m: any, i: number) => {
  const c1 = nicknames[m.team1 ?? ''] ?? ''
  const c2 = nicknames[m.team2 ?? ''] ?? ''
  return (
    <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 14, padding: '12px 0', borderTop: i > 0 ? '1px solid rgba(128,128,128,0.06)' : 'none', fontSize: 13 }}>
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 2 }}>
        <span style={{ fontSize: 15, fontWeight: 600, fontFamily: 'var(--font-display)', letterSpacing: '0.03em', color: 'var(--text)' }}>{m.team1 || '待定'}</span>
        <span style={{ fontSize: 11, color: 'var(--text-muted)', height: 16 }}>{c1}</span>
      </div>
      {m.score ? (
        <span style={{ fontFamily: 'var(--font-mono)', fontSize: 20, fontWeight: 700, color: 'var(--text)', minWidth: 50, textAlign: 'center' }}>{m.score}</span>
      ) : (
        (() => {
          const t = m.scheduled_at
          if (!t) return <span style={{ fontFamily: 'var(--font-display)', fontSize: 20, fontWeight: 700, color: 'var(--gold)', minWidth: 50, textAlign: 'center' }}>—:—</span>
          const parts = t.split(' ')
          const datePart = parts.length > 1 ? parts[0] : ''
          const timePart = parts.length > 1 ? parts[1] : t
          return (
            <div style={{ minWidth: 50, textAlign: 'center' }}>
              <div style={{ fontFamily: 'var(--font-display)', fontSize: 20, fontWeight: 700, color: 'var(--gold)', lineHeight: 1 }}>
                {timePart.length >= 5 ? timePart.slice(0, 5) : timePart}
              </div>
              {datePart && (
                <div style={{ fontFamily: 'var(--font-mono)', fontSize: 10, color: 'var(--text-muted)', marginTop: 2 }}>
                  {datePart.slice(5).replace('-', '/')}
                </div>
              )}
            </div>
          )
        })()
      )}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 2 }}>
        <span style={{ fontSize: 15, fontWeight: 600, fontFamily: 'var(--font-display)', letterSpacing: '0.03em', color: 'var(--text)' }}>{m.team2 || '待定'}</span>
        <span style={{ fontSize: 11, color: 'var(--text-muted)', height: 16 }}>{c2}</span>
      </div>
      <span style={{ fontSize: 11, color: m.score ? 'var(--text-muted)' : 'var(--gold)', minWidth: 60, textAlign: 'right' }}>
        {m.best_of ? `${m.best_of.toUpperCase()}` : ''}
      </span>
    </div>
  )
})}
```

Key changes:
- Scheduled matches: date always shown below time (not only >24h)
- `scheduled_at` now `YYYY-MM-DD HH:MM` format — split on space to get date and time parts
- Right column: only BO3, no date (removed `m.played_at` display)

- [ ] **Step 3: Build frontend**

Run: `cd frontend && npm run build`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add frontend/src/pages/Matches.tsx
git commit -m "fix: event card dates and modal date/time display with MM/DD formatting"
```

---

### Task 6: Rebuild + final verification

- [ ] **Step 1: Full rebuild**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./... && cd frontend && npm run build
```

- [ ] **Step 2: Verify all API endpoints return correct dates**

```bash
# Results: played_at should be YYYY-MM-DD
curl -s 'http://172.21.32.31:8082/api/events?type=results&limit=3' | python3 -c "import json,sys; d=json.load(sys.stdin); [print(m.get('played_at')) for e in d['data']['events'] for m in e['matches'][:2]]"

# Upcoming: scheduled_at should be YYYY-MM-DD HH:MM
curl -s 'http://172.21.32.31:8082/api/events?type=upcoming&limit=3' | python3 -c "import json,sys; d=json.load(sys.stdin); [print(m.get('scheduled_at')) for e in d['data']['events'] for m in e['matches'][:2]]"

# Team: should have >1 matches with dates
curl -s 'http://172.21.32.31:8082/api/teams/6665' | python3 -c "import json,sys; d=json.load(sys.stdin)['data']; print(f'Matches: {len(d[\"recent_matches\"])}'); [print(m.get('played_at')) for m in d['recent_matches']]"
```

Expected:
- Results: `2026-05-28` style dates
- Upcoming: `2026-05-29 18:30` style
- Team: >1 matches, dates filled

- [ ] **Step 3: Commit final verification notes**

```bash
git add -A && git commit -m "chore: final verification of date fix deployment"
```
