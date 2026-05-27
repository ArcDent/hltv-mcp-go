import { Routes, Route, NavLink } from 'react-router-dom'
import Dashboard from './pages/Dashboard'
import Matches from './pages/Matches'
import Teams from './pages/Teams'
import Players from './pages/Players'
import News from './pages/News'
import Cache from './pages/Cache'

const navItems = [
  { to: '/',         label: '总览' },
  { to: '/matches',  label: '赛程' },
  { to: '/teams',    label: '队伍' },
  { to: '/players',  label: '选手' },
  { to: '/news',     label: '新闻' },
  { to: '/cache',    label: '缓存' },
]

export default function App() {
  return (
    <div className="h-full flex flex-col bg-bg">
      {/* ---- Top Bar ---- */}
      <header className="shrink-0 border-b border-border bg-surface">
        <div className="max-w-7xl mx-auto flex items-center px-8 h-16">
          {/* Brand */}
          <span className="font-display text-gold text-2xl tracking-[0.15em] font-bold shrink-0 mr-12">
            HLTV<span className="text-text-secondary font-normal text-base ml-1">MCP</span>
          </span>

          {/* Nav */}
          <nav className="flex gap-8">
            {navItems.map(({ to, label }) => (
              <NavLink
                key={to}
                to={to}
                end={to === '/'}
                className={({ isActive }) =>
                  `text-[16px] font-medium transition-colors border-b-[3px] pb-[15px] -mb-[15px] ${
                    isActive
                      ? 'text-gold border-gold'
                      : 'text-text-secondary border-transparent hover:text-text'
                  }`
                }
              >
                {label}
              </NavLink>
            ))}
          </nav>

          {/* Spacer pushes ONLINE to the right */}
          <div className="flex-1" />

          {/* ONLINE — far right */}
          <div className="flex items-center gap-2 text-text-muted text-[14px] shrink-0">
            <span className="w-2 h-2 rounded-full bg-green pulse-dot" />
            ONLINE
          </div>
        </div>
      </header>

      {/* ---- Main ---- */}
      <main className="flex-1 overflow-y-auto">
        <div className="max-w-7xl mx-auto px-8 py-10">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/matches" element={<Matches />} />
            <Route path="/teams" element={<Teams />} />
            <Route path="/players" element={<Players />} />
            <Route path="/news" element={<News />} />
            <Route path="/cache" element={<Cache />} />
          </Routes>
        </div>
      </main>
    </div>
  )
}
