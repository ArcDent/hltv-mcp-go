# 前端重设计实现计划 — 电竞数据看板

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将前端从终端绿暗黑风重新设计为深蓝灰+金色电竞数据看板，中文优先，全屏铺满，字号加大，高对比度输入框。

**Architecture:** 单页应用，左侧 sticky 侧栏(200px) + 右侧滚动主内容区。所有页面共享同一套 CSS 变量和组件模式。TypeScript 保持现有 API 调用层不变，仅替换 UI 层。

**Tech Stack:** React 18, TypeScript, Vite, Tailwind CSS v4, React Router

**Spec:** `docs/superpowers/specs/2026-05-27-frontend-redesign.md`

---

### Task 1: Design System — index.html + index.css

**Files:**
- Modify: `frontend/index.html`
- Modify: `frontend/src/index.css`

- [ ] **Step 1: Update index.html with Chinese font CDN**

```html
<!doctype html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <link rel="icon" type="image/svg+xml" href="/favicon.svg" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>HLTV MCP — 电竞数据中心</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;600;700&display=swap" rel="stylesheet">
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

- [ ] **Step 2: Write index.css with full design system**

```css
@import "tailwindcss";

@theme {
  --color-bg: #0d1117;
  --color-panel: #161b22;
  --color-border: #30363d;
  --color-gold: #f0c040;
  --color-gold-dim: rgba(240, 192, 64, 0.08);
  --color-gold-border: rgba(240, 192, 64, 0.25);
  --color-orange: #ff8c42;
  --color-orange-dim: rgba(255, 140, 66, 0.12);
  --color-text: #e6edf3;
  --color-text-dim: #8b949e;
  --color-placeholder: #484f58;
  --font-mono: 'JetBrains Mono', monospace;
  --font-sans: 'PingFang SC', 'Microsoft YaHei', 'Segoe UI', system-ui, sans-serif;
}

* { margin: 0; padding: 0; box-sizing: border-box; }

html, body, #root {
  height: 100%;
  overflow: hidden;
}

body {
  font-family: var(--font-sans);
  font-size: 15px;
  line-height: 1.6;
  color: var(--color-text);
  background: var(--color-bg);
  -webkit-font-smoothing: antialiased;
}

/* Scrollbar */
::-webkit-scrollbar { width: 6px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: var(--color-border); border-radius: 3px; }
::-webkit-scrollbar-thumb:hover { background: var(--color-gold-border); }

/* Staggered entry animation */
@keyframes fade-in-up {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

.animate-in {
  animation: fade-in-up 0.35s ease-out both;
}
```

- [ ] **Step 3: Build and verify**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild/frontend && npx vite build --outDir ../dist 2>&1 | tail -3
```
Expected: `✓ built in ...` (3 files output)

- [ ] **Step 4: Commit**

```bash
git add frontend/index.html frontend/src/index.css
git commit -m "feat: add design system — Chinese fonts, gold+dark palette, CSS variables"
```

---

### Task 2: Layout — App.tsx with sticky sidebar

**Files:**
- Modify: `frontend/src/App.tsx`
- Delete: `frontend/src/App.css` (no longer needed)

- [ ] **Step 1: Write App.tsx with sticky sidebar layout**

```tsx
import { Routes, Route, NavLink } from 'react-router-dom'
import Dashboard from './pages/Dashboard'
import Matches from './pages/Matches'
import Teams from './pages/Teams'
import Players from './pages/Players'
import News from './pages/News'
import Cache from './pages/Cache'

const nav = [
  { to: '/', label: '总览', icon: '◈' },
  { to: '/matches', label: '赛程', icon: '◆' },
  { to: '/teams', label: '队伍', icon: '▣' },
  { to: '/players', label: '选手', icon: '◎' },
  { to: '/news', label: '新闻', icon: '◉' },
  { to: '/cache', label: '缓存', icon: '◫' },
]

export default function App() {
  return (
    <div className="h-full flex">
      {/* Sticky sidebar */}
      <nav className="w-48 shrink-0 border-r border-border bg-panel flex flex-col sticky top-0 h-screen">
        <div className="px-4 py-5 border-b border-border text-center">
          <h1 className="text-[17px] font-bold text-gold tracking-wide">HLTV MCP</h1>
          <p className="text-[11px] text-text-dim mt-1">数据中心</p>
        </div>

        <div className="flex flex-col py-3 flex-1">
          {nav.map(({ to, label, icon }) => (
            <NavLink
              key={to}
              to={to}
              end={to === '/'}
              className={({ isActive }) =>
                `flex items-center gap-3 px-4 py-2.5 text-[15px] transition-all border-l-2 ${
                  isActive
                    ? 'text-gold border-gold bg-gold-dim'
                    : 'text-text-dim border-transparent hover:text-text hover:bg-panel'
                }`
              }
            >
              <span className="text-sm">{icon}</span>
              {label}
            </NavLink>
          ))}
        </div>

        <div className="px-4 py-3 border-t border-border flex items-center gap-2 text-[12px] text-text-dim">
          <span className="w-2 h-2 rounded-full bg-[#3fb950]" />
          在线
        </div>
      </nav>

      {/* Scrollable main content */}
      <main className="flex-1 overflow-y-auto p-8">
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

- [ ] **Step 2: Delete old App.css**

```bash
rm /home/arcdent/github/hltv-mcp-fully-rebuild/frontend/src/App.css
```

- [ ] **Step 3: Build and verify**

```bash
npx vite build --outDir ../dist 2>&1 | tail -3
```
Expected: `✓ built in ...`

- [ ] **Step 4: Commit**

```bash
git add frontend/src/App.tsx frontend/src/App.css
git commit -m "feat: add sticky sidebar layout with Chinese navigation"
```

---

### Task 3: Dashboard page

**Files:**
- Modify: `frontend/src/pages/Dashboard.tsx`

- [ ] **Step 1: Write Dashboard.tsx**

```tsx
import { useEffect, useState } from 'react'
import { api } from '../api/client'

export default function Dashboard() {
  const [status, setStatus] = useState<any>(null)
  useEffect(() => { api.status().then(setStatus).catch(() => {}) }, [])

  const stats = [
    { label: '运行时间', value: status ? `${status.uptime_sec}s` : '--' },
    { label: 'Go 版本', value: status?.go_version ?? '--' },
    { label: '内存占用', value: status ? `${status.memory_mb} MB` : '--' },
    { label: '缓存条目', value: status?.cache_entries ?? '--' },
  ]

  const sysRows = [
    { label: 'HTTP 服务', detail: '0.0.0.0:8082' },
    { label: 'MCP 连接', detail: 'stdio 已连接' },
    { label: 'Chrome', detail: 'chromedp 就绪' },
    { label: '数据源', detail: 'HTTP 直连 + chromedp 备用' },
  ]

  return (
    <div className="animate-in">
      <h2 className="text-[16px] font-semibold text-gold mb-4 pb-2 border-b border-border tracking-wide">
        ◈ 总览
      </h2>

      <div className="grid grid-cols-4 gap-3 mb-8">
        {stats.map((s, i) => (
          <div key={s.label} className="bg-panel border border-border rounded-md p-5 animate-in"
            style={{ animationDelay: `${i * 80}ms` }}>
            <div className="text-[13px] text-text-dim mb-2">{s.label}</div>
            <div className="text-[26px] font-bold text-text font-mono">{s.value}</div>
          </div>
        ))}
      </div>

      <div className="bg-panel border border-border rounded-md p-5">
        <div className="text-[14px] font-semibold text-text mb-4">系统状态</div>
        <div className="space-y-3">
          {sysRows.map((row) => (
            <div key={row.label} className="flex items-center gap-3 text-[14px]">
              <span className="w-2 h-2 rounded-full bg-[#3fb950]" />
              <span className="text-text w-28">{row.label}</span>
              <span className="text-text-dim">{row.detail}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Build and verify**

```bash
npx vite build --outDir ../dist 2>&1 | tail -3
```
Expected: `✓ built in ...`

- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/Dashboard.tsx
git commit -m "feat: add Dashboard page — stat cards + system status panel"
```

---

### Task 4: Matches page

**Files:**
- Modify: `frontend/src/pages/Matches.tsx`

- [ ] **Step 1: Write Matches.tsx**

```tsx
import { useEffect, useState } from 'react'
import { api } from '../api/client'

type Tab = 'today' | 'upcoming' | 'results'
const tabs: { key: Tab; label: string }[] = [
  { key: 'today', label: '今日' },
  { key: 'upcoming', label: '即将开始' },
  { key: 'results', label: '近期赛果' },
]

function isUpcomingSoon(scheduledAt: string): boolean {
  if (!scheduledAt) return false
  const t = new Date(scheduledAt).getTime()
  if (isNaN(t)) return false
  const diff = Math.abs(Date.now() - t)
  return diff < 2 * 60 * 60 * 1000
}

export default function Matches() {
  const [tab, setTab] = useState<Tab>('today')
  const [data, setData] = useState<any>(null)
  const [team, setTeam] = useState('')

  useEffect(() => {
    setData(null)
    if (tab === 'today') api.todayMatches().then(setData)
    else if (tab === 'upcoming') api.upcomingMatches({ team, limit: '20' }).then(setData)
    else api.results({ team, limit: '20' }).then(setData)
  }, [tab, team])

  const items: any[] = data?.items ?? []

  return (
    <div className="animate-in">
      <h2 className="text-[16px] font-semibold text-gold mb-4 pb-2 border-b border-border tracking-wide">
        ◆ 赛程
      </h2>

      <div className="flex gap-1 mb-6">
        {tabs.map((t) => (
          <button key={t.key} onClick={() => setTab(t.key)}
            className={`px-5 py-2 text-[14px] font-medium rounded-md border transition-all ${
              tab === t.key
                ? 'border-gold bg-gold-dim text-gold'
                : 'border-border bg-panel text-text-dim hover:text-text hover:border-gold-border'
            }`}>
            {t.label}
          </button>
        ))}
      </div>

      <div className="space-y-2">
        {items.length === 0 && (
          <div className="text-text-dim text-[14px] py-16 text-center border border-border rounded-md bg-panel">
            {data ? '暂无数据' : '加载中...'}
          </div>
        )}
        {items.map((m: any, i: number) => {
          const soon = isUpcomingSoon(m.scheduled_at)
          return (
            <div key={i}
              className="bg-panel border border-border rounded-md p-4 flex items-center gap-5 animate-in hover:border-gold-border transition-colors"
              style={{ animationDelay: `${i * 30}ms` }}>
              {m.score && !m.scheduled_at && (
                <span className="text-[12px] text-orange bg-orange-dim border border-orange/30 rounded px-2 py-0.5 font-medium shrink-0">
                  已结束
                </span>
              )}
              {soon && !m.score && (
                <span className="text-[12px] text-orange bg-orange-dim border border-orange/30 rounded px-2 py-0.5 font-medium shrink-0">
                  即将开始
                </span>
              )}

              <div className="flex-1 text-right">
                <div className="text-[16px] font-semibold text-text">{m.team1 ?? 'TBD'}</div>
                <div className="text-[12px] text-text-dim mt-0.5">{getColloquial(m.team1)}</div>
              </div>

              <div className="flex flex-col items-center min-w-[90px]">
                {m.score ? (
                  <span className="text-[28px] font-bold text-gold font-mono tracking-wider">{m.score}</span>
                ) : (
                  <span className="text-[28px] font-bold text-gold font-mono tracking-wider">
                    {m.scheduled_at ? m.scheduled_at.substring(11, 16) : '--:--'}
                  </span>
                )}
                {m.best_of && <span className="text-[11px] text-text-dim mt-0.5">{m.best_of}</span>}
              </div>

              <div className="flex-1">
                <div className="text-[16px] font-semibold text-text">{m.team2 ?? 'TBD'}</div>
                <div className="text-[12px] text-text-dim mt-0.5">{getColloquial(m.team2)}</div>
              </div>

              {m.event && (
                <span className="text-[12px] text-gold bg-gold-dim border border-gold-border rounded px-2 py-0.5 shrink-0">
                  {m.event}
                </span>
              )}
            </div>
          )
        })}
      </div>

      <div className="mt-6">
        <input
          placeholder="输入队伍名筛选..."
          value={team}
          onChange={(e) => setTeam(e.target.value)}
          className="w-64 bg-panel border-2 border-border text-text text-[15px] px-4 py-2.5 rounded-md focus:outline-none focus:border-gold placeholder:text-placeholder"
        />
      </div>
    </div>
  )
}

function getColloquial(name: string | undefined): string {
  if (!name) return ''
  const map: Record<string, string> = {
    'Vitality': '小蜜蜂', 'Team Spirit': '绿龙', 'Spirit': '绿龙',
    'Natus Vincere': '天生赢家', 'NAVI': '天生赢家', 'FaZe': 'FaZe Clan',
    'G2': '武士', 'MOUZ': '老鼠', 'Falcons': '猎鹰',
    'Astralis': 'A队', 'Virtus.pro': 'VP', 'Team Liquid': '液体',
    'FURIA': '黑豹', 'The MongolZ': '蒙古队', 'TYLOO': '天禄',
  }
  return map[name] ?? ''
}
```

- [ ] **Step 2: Build and verify**

```bash
npx vite build --outDir ../dist 2>&1 | tail -3
```
Expected: `✓ built in ...`

- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/Matches.tsx
git commit -m "feat: add Matches page — match cards with scores, upcoming labels, Chinese team nicknames"
```

---

### Task 5: Teams + Players pages

**Files:**
- Modify: `frontend/src/pages/Teams.tsx`
- Modify: `frontend/src/pages/Players.tsx`

- [ ] **Step 1: Write Teams.tsx**

```tsx
import { useState } from 'react'
import { api } from '../api/client'

export default function Teams() {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<any[] | null>(null)
  const [loading, setLoading] = useState(false)

  const search = async () => {
    if (!query.trim()) return
    setLoading(true)
    try { const r = await api.search(query, 'team'); setResults(r?.items ?? []) }
    catch { setResults([]) }
    setLoading(false)
  }

  return (
    <div className="animate-in">
      <h2 className="text-[16px] font-semibold text-gold mb-4 pb-2 border-b border-border tracking-wide">
        ▣ 队伍搜索
      </h2>

      <div className="flex gap-3 mb-6">
        <input
          placeholder="输入队伍名（支持中英文/别名，如 Spirit、绿龙、小蜜蜂）"
          value={query} onChange={(e) => setQuery(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && search()}
          className="flex-1 bg-panel border-2 border-border text-text text-[15px] px-4 py-2.5 rounded-md focus:outline-none focus:border-gold placeholder:text-placeholder"
        />
        <button onClick={search} disabled={loading}
          className="px-6 py-2.5 bg-gold text-[#0d1117] text-[15px] font-semibold rounded-md hover:brightness-110 transition-all disabled:opacity-40">
          {loading ? '搜索中...' : '搜索'}
        </button>
      </div>

      <div className="space-y-1">
        {results === null && (
          <div className="text-text-dim text-[14px] py-16 text-center border border-border rounded-md bg-panel">
            输入队名开始搜索
          </div>
        )}
        {results?.length === 0 && (
          <div className="text-text-dim text-[14px] py-16 text-center border border-border rounded-md bg-panel">
            无匹配结果
          </div>
        )}
        {results?.map((t: any, i: number) => (
          <div key={i} className="bg-panel border border-border rounded-md p-3.5 flex items-center gap-4 animate-in hover:border-gold-border transition-colors"
            style={{ animationDelay: `${i * 35}ms` }}>
            <span className="text-gold text-[14px] font-mono font-semibold w-7">{String(i + 1).padStart(2, '0')}</span>
            <span className="flex-1 text-[16px] font-semibold text-text">{t.name}</span>
            <span className="text-[12px] text-text-dim bg-[#0d1117] border border-border rounded px-2.5 py-1 font-mono">
              ID:{t.id ?? '--'}
            </span>
            {t.slug && <span className="text-[12px] text-text-dim">{t.slug}</span>}
          </div>
        ))}
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Write Players.tsx** (same pattern, different API call)

```tsx
import { useState } from 'react'
import { api } from '../api/client'

export default function Players() {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<any[] | null>(null)
  const [loading, setLoading] = useState(false)

  const search = async () => {
    if (!query.trim()) return
    setLoading(true)
    try { const r = await api.search(query, 'player'); setResults(r?.items ?? []) }
    catch { setResults([]) }
    setLoading(false)
  }

  return (
    <div className="animate-in">
      <h2 className="text-[16px] font-semibold text-gold mb-4 pb-2 border-b border-border tracking-wide">
        ◎ 选手搜索
      </h2>

      <div className="flex gap-3 mb-6">
        <input
          placeholder="输入选手名（如 ZywOo、载物、s1mple）"
          value={query} onChange={(e) => setQuery(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && search()}
          className="flex-1 bg-panel border-2 border-border text-text text-[15px] px-4 py-2.5 rounded-md focus:outline-none focus:border-gold placeholder:text-placeholder"
        />
        <button onClick={search} disabled={loading}
          className="px-6 py-2.5 bg-gold text-[#0d1117] text-[15px] font-semibold rounded-md hover:brightness-110 transition-all disabled:opacity-40">
          {loading ? '搜索中...' : '搜索'}
        </button>
      </div>

      <div className="space-y-1">
        {results === null && (
          <div className="text-text-dim text-[14px] py-16 text-center border border-border rounded-md bg-panel">
            输入选手名开始搜索
          </div>
        )}
        {results?.length === 0 && (
          <div className="text-text-dim text-[14px] py-16 text-center border border-border rounded-md bg-panel">
            无匹配结果
          </div>
        )}
        {results?.map((p: any, i: number) => (
          <div key={i} className="bg-panel border border-border rounded-md p-3.5 flex items-center gap-4 animate-in hover:border-gold-border transition-colors"
            style={{ animationDelay: `${i * 35}ms` }}>
            <span className="text-gold text-[14px] font-mono font-semibold w-7">{String(i + 1).padStart(2, '0')}</span>
            <span className="flex-1 text-[16px] font-semibold text-text">{p.name}</span>
            <span className="text-[12px] text-text-dim bg-[#0d1117] border border-border rounded px-2.5 py-1 font-mono">
              ID:{p.id ?? '--'}
            </span>
            {p.slug && <span className="text-[12px] text-text-dim">{p.slug}</span>}
          </div>
        ))}
      </div>
    </div>
  )
}
```

- [ ] **Step 3: Build and verify + commit**

```bash
npx vite build --outDir ../dist 2>&1 | tail -3
```
```bash
git add frontend/src/pages/Teams.tsx frontend/src/pages/Players.tsx
git commit -m "feat: add Teams and Players search pages with high-contrast inputs"
```

---

### Task 6: News + Cache pages

**Files:**
- Modify: `frontend/src/pages/News.tsx`
- Modify: `frontend/src/pages/Cache.tsx`

- [ ] **Step 1: Write News.tsx**

```tsx
import { useEffect, useState } from 'react'
import { api } from '../api/client'

type Tab = 'realtime' | 'archive'

export default function News() {
  const [tab, setTab] = useState<Tab>('realtime')
  const [data, setData] = useState<any>(null)

  useEffect(() => {
    setData(null)
    if (tab === 'realtime') api.realtimeNews().then(setData)
    else api.newsDigest({ limit: '25' }).then(setData)
  }, [tab])

  const items: any[] = data?.items ?? []

  return (
    <div className="animate-in">
      <h2 className="text-[16px] font-semibold text-gold mb-4 pb-2 border-b border-border tracking-wide">
        ◉ 新闻
      </h2>

      <div className="flex gap-1 mb-6">
        {[{ key: 'realtime' as Tab, label: '实时新闻' }, { key: 'archive' as Tab, label: '归档新闻' }].map((t) => (
          <button key={t.key} onClick={() => setTab(t.key)}
            className={`px-5 py-2 text-[14px] font-medium rounded-md border transition-all ${
              tab === t.key
                ? 'border-gold bg-gold-dim text-gold'
                : 'border-border bg-panel text-text-dim hover:text-text hover:border-gold-border'
            }`}>
            {t.label}
          </button>
        ))}
      </div>

      <div className="space-y-1">
        {items.length === 0 && (
          <div className="text-text-dim text-[14px] py-16 text-center border border-border rounded-md bg-panel">
            {data ? '暂无新闻' : '加载中...'}
          </div>
        )}
        {items.map((n: any, i: number) => (
          <div key={i} className="bg-panel border border-border rounded-md p-3.5 flex items-center gap-4 animate-in hover:border-gold-border transition-colors group"
            style={{ animationDelay: `${i * 25}ms` }}>
            <span className="text-gold text-[14px] font-mono font-semibold w-7 shrink-0">{String(i + 1).padStart(2, '0')}</span>
            <span className="flex-1 text-[15px] text-text group-hover:text-gold transition-colors truncate">{n.title}</span>
            <span className="text-[13px] text-text-dim shrink-0">{n.published_at ?? n.relative_time ?? ''}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Write Cache.tsx**

```tsx
import { useEffect, useState } from 'react'
import { api } from '../api/client'

export default function Cache() {
  const [stats, setStats] = useState<any>(null)
  const [cleared, setCleared] = useState(false)

  const refresh = () => { api.cacheStats().then(setStats).catch(() => {}) }
  useEffect(refresh, [])

  const handleClear = async () => {
    await api.clearCache(); setCleared(true); refresh()
    setTimeout(() => setCleared(false), 2500)
  }

  return (
    <div className="animate-in">
      <h2 className="text-[16px] font-semibold text-gold mb-4 pb-2 border-b border-border tracking-wide">
        ◫ 缓存管理
      </h2>

      <div className="grid grid-cols-3 gap-3 mb-6">
        {[
          { label: '缓存条目', value: stats?.entries ?? '--' },
          { label: '命中', value: stats?.hits ?? '--' },
          { label: '未命中', value: stats?.misses ?? '--' },
        ].map((s, i) => (
          <div key={s.label} className="bg-panel border border-border rounded-md p-5 animate-in"
            style={{ animationDelay: `${i * 80}ms` }}>
            <div className="text-[13px] text-text-dim mb-2">{s.label}</div>
            <div className="text-[26px] font-bold text-text font-mono">{s.value}</div>
          </div>
        ))}
      </div>

      <div className="flex items-center gap-4">
        <button onClick={handleClear}
          className="px-5 py-2.5 border border-orange/40 bg-orange-dim text-orange text-[14px] font-medium rounded-md hover:bg-orange/20 transition-colors">
          清除全部缓存
        </button>
        <button onClick={refresh}
          className="px-5 py-2.5 border border-border bg-panel text-text-dim text-[14px] font-medium rounded-md hover:text-text hover:border-gold-border transition-colors">
          刷新
        </button>
        {cleared && (
          <span className="text-[14px] text-gold animate-in">缓存已清除</span>
        )}
      </div>
    </div>
  )
}
```

- [ ] **Step 3: Build and commit**

```bash
npx vite build --outDir ../dist 2>&1 | tail -3
```
```bash
git add frontend/src/pages/News.tsx frontend/src/pages/Cache.tsx
git commit -m "feat: add News and Cache pages"
```

---

### Task 7: Final integration — rebuild Go binary and verify

- [ ] **Step 1: Build frontend + Go + smoke test**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild/frontend && npx vite build --outDir ../dist
cd .. && go build -o hltv-mcp github.com/arcdent/hltv-mcp

# Start and verify
./hltv-mcp &
sleep 2
curl -s http://localhost:8082/api/health
curl -s -o /dev/null -w "%{http_code} %{size_download}" http://localhost:8082/
curl -s -o /dev/null -w " %{http_code} %{size_download}" "http://localhost:8082$(curl -s http://localhost:8082/ | grep -oP 'src=\"[^\"]*\.js\"' | head -1 | sed 's/src=\"//;s/\"//')"
kill %1
```

Expected: API 200, HTML 200 with content, JS 200 with >200KB

- [ ] **Step 2: Commit and push**

```bash
git add -A
git commit -m "feat: complete frontend redesign — CS esports dashboard theme"
git push
```
