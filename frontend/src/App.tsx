import { Routes, Route, NavLink } from 'react-router-dom'
import Dashboard from './pages/Dashboard'
import Matches from './pages/Matches'
import Teams from './pages/Teams'
import Players from './pages/Players'
import News from './pages/News'
import Cache from './pages/Cache'

const nav = [
  { to: '/', label: 'Dashboard' },
  { to: '/matches', label: 'Matches' },
  { to: '/teams', label: 'Teams' },
  { to: '/players', label: 'Players' },
  { to: '/news', label: 'News' },
  { to: '/cache', label: 'Cache' },
]

export default function App() {
  return (
    <div className="min-h-screen bg-gray-900 text-gray-100">
      <nav className="flex gap-4 p-4 bg-gray-800 border-b border-gray-700">
        {nav.map(({ to, label }) => (
          <NavLink key={to} to={to} className={({ isActive }) =>
            `px-3 py-1 rounded ${isActive ? 'bg-blue-600 text-white' : 'text-gray-400 hover:text-white'}`
          }>{label}</NavLink>
        ))}
      </nav>
      <main className="p-6 max-w-6xl mx-auto">
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
