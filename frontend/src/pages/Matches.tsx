import { useEffect, useState } from 'react'
import { api } from '../api/client'

export default function Matches() {
  const [tab, setTab] = useState<'today' | 'upcoming' | 'results'>('today')
  const [data, setData] = useState<any>(null)
  const [team, setTeam] = useState('')

  useEffect(() => {
    if (tab === 'today') api.todayMatches().then(setData)
    else if (tab === 'upcoming') api.upcomingMatches({ team, limit: '20' }).then(setData)
    else api.results({ team, limit: '20' }).then(setData)
  }, [tab, team])

  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">Matches</h1>
      <div className="flex gap-2 mb-4">
        {(['today', 'upcoming', 'results'] as const).map(t => (
          <button key={t} onClick={() => setTab(t)} className={`px-4 py-1 rounded ${tab === t ? 'bg-blue-600' : 'bg-gray-700'}`}>
            {t === 'today' ? 'Today' : t === 'upcoming' ? 'Upcoming' : 'Results'}
          </button>
        ))}
      </div>
      <input placeholder="Team filter" value={team} onChange={e => setTeam(e.target.value)}
        className="bg-gray-800 border border-gray-700 rounded px-3 py-1 mb-4 text-white w-64" />
      <div className="space-y-2">
        {data?.items?.map((m: any, i: number) => (
          <div key={i} className="bg-gray-800 p-3 rounded flex justify-between">
            <span>{m.team1} vs {m.team2}</span>
            {m.score && <span className="text-blue-400">{m.score}</span>}
            {m.event && <span className="text-gray-400">{m.event}</span>}
          </div>
        ))}
      </div>
    </div>
  )
}
