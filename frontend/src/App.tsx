import { Routes, Route, NavLink } from 'react-router-dom'
import Dashboard from './pages/Dashboard'
import Matches from './pages/Matches'
import Teams from './pages/Teams'
import Players from './pages/Players'
import News from './pages/News'
import Cache from './pages/Cache'

const nav = [
  { to: '/', label: 'DASH', icon: '◈' },
  { to: '/matches', label: 'MATCHES', icon: '◆' },
  { to: '/teams', label: 'TEAMS', icon: '▣' },
  { to: '/players', label: 'PLAYERS', icon: '◎' },
  { to: '/news', label: 'NEWS', icon: '◉' },
  { to: '/cache', label: 'CACHE', icon: '◫' },
]

export default function App() {
  return (
    <div className="min-h-screen bg-bg text-text font-mono">
      {/* Top status bar */}
      <div className="h-[2px] bg-neon animate-pulse" />

      <div className="flex gap-0">
        {/* Sidebar navigation */}
        <nav className="w-48 shrink-0 border-r border-border min-h-screen bg-surface flex flex-col">
          <div className="px-4 py-5 border-b border-border">
            <h1 className="font-display text-neon text-sm tracking-[0.3em] text-center">
              HLTV·MCP
            </h1>
            <div className="text-[9px] text-text-dim text-center mt-1 tracking-widest">
              TACTICAL OPS
            </div>
          </div>

          <div className="flex flex-col py-3 flex-1">
            {nav.map(({ to, label, icon }) => (
              <NavLink
                key={to}
                to={to}
                className={({ isActive }) =>
                  `flex items-center gap-3 px-4 py-2.5 text-[11px] tracking-[0.15em] transition-all duration-150 border-l-2 ${
                    isActive
                      ? 'bg-neon-dim text-neon border-neon'
                      : 'text-steel border-transparent hover:text-text hover:bg-surface hover:border-steel-dim'
                  }`
                }
              >
                <span className="text-sm">{icon}</span>
                {label}
              </NavLink>
            ))}
          </div>

          {/* Footer status */}
          <div className="px-4 py-3 border-t border-border">
            <div className="flex items-center gap-2 text-[10px] text-text-dim">
              <span className="w-1.5 h-1.5 rounded-full bg-neon animate-pulse" />
              <span>SYS ONLINE</span>
            </div>
          </div>
        </nav>

        {/* Main content */}
        <main className="flex-1 p-6 min-w-0">
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
    </div>
  )
}
