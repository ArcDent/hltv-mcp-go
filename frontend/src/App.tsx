import { useState } from 'react'
import { Routes, Route, NavLink } from 'react-router-dom'
import Dashboard from './pages/Dashboard'
import Matches from './pages/Matches'
import SearchPage from './pages/SearchPage'
import News from './pages/News'
import Cache from './pages/Cache'

const nav = [
  { to: '/',        label: '总览' },
  { to: '/matches', label: '赛程' },
  { to: '/teams',   label: '队伍' },
  { to: '/players', label: '选手' },
  { to: '/news',    label: '新闻' },
  { to: '/cache',   label: '缓存' },
]

export default function App() {
  const [dark, setDark] = useState(() => document.documentElement.classList.contains('dark'))
  const toggle = () => {
    document.documentElement.classList.toggle('dark')
    setDark(document.documentElement.classList.contains('dark'))
  }

  return (
    <div className="h-full flex" style={{ background: 'var(--bg)', color: 'var(--text)' }}>
      {/* Sidebar */}
      <nav style={{
        width: 180, flexShrink: 0, height: '100vh', position: 'sticky', top: 0,
        background: 'var(--card)', borderRight: '1px solid var(--border)',
        display: 'flex', flexDirection: 'column',
      }}>
        {/* Brand */}
        <div style={{ padding: '20px 16px', borderBottom: '1px solid var(--border)', textAlign: 'center' }}>
          <div style={{ fontFamily: 'var(--font-display)', fontSize: 20, fontWeight: 700,
            color: 'var(--gold)', letterSpacing: '0.08em' }}>
            HLTV<span style={{ color: 'var(--text-secondary)', fontWeight: 400, fontSize: 15 }}>MCP</span>
          </div>
          <div style={{ fontSize: 11, color: 'var(--text-muted)', marginTop: 4 }}>数据中心</div>
        </div>

        {/* Nav items */}
        <div style={{ flex: 1, padding: '12px 0' }}>
          {nav.map(({ to, label }) => (
            <NavLink key={to} to={to} end={to === '/'}
              style={({ isActive }) => ({
                display: 'flex', alignItems: 'center', gap: 10,
                padding: '10px 18px', fontSize: 15, fontWeight: 500,
                textDecoration: 'none',
                color: isActive ? 'var(--gold)' : 'var(--text-secondary)',
                background: isActive ? 'var(--gold-dim)' : 'transparent',
                borderLeft: isActive ? '2px solid var(--gold)' : '2px solid transparent',
                transition: 'all 0.15s ease',
              })}>
              {label}
            </NavLink>
          ))}
        </div>

        {/* Bottom status */}
        <div style={{ padding: '14px 18px', borderTop: '1px solid var(--border)',
          display: 'flex', flexDirection: 'column', gap: 8 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 6,
            fontSize: 12, color: 'var(--text-muted)' }}>
            <span style={{ width: 7, height: 7, borderRadius: '50%', background: 'var(--green)' }} />
            ONLINE
          </div>
          <button onClick={toggle} style={{
            width: 30, height: 30, borderRadius: '50%',
            border: '1px solid var(--border)', background: 'var(--card)',
            color: 'var(--text-secondary)', fontSize: 15,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            cursor: 'pointer',
          }}>
            {dark ? '☀' : '🌙'}
          </button>
        </div>
      </nav>

      {/* Main */}
      <main style={{ flex: 1, overflowY: 'auto', padding: '32px' }}>
        <div style={{ maxWidth: 1100, margin: '0 auto' }}>
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/matches" element={<Matches />} />
            <Route path="/teams" element={<SearchPage type="team" placeholder="搜索队伍 — 支持英文 / 中文 / 别名（如 Spirit、绿龙、小蜜蜂）" emptyHint="输入队名开始搜索" />} />
            <Route path="/players" element={<SearchPage type="player" placeholder="搜索选手 — 如 ZywOo、载物、s1mple" emptyHint="输入选手名开始搜索" />} />
            <Route path="/news" element={<News />} />
            <Route path="/cache" element={<Cache />} />
          </Routes>
        </div>
      </main>
    </div>
  )
}
