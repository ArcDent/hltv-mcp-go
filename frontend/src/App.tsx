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
