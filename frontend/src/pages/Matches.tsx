import { useEffect, useState } from 'react'
import { api } from '../api/client'

type Tab = 'today' | 'upcoming' | 'results'

const tabs: { key: Tab; label: string }[] = [
  { key: 'today', label: 'TODAY' },
  { key: 'upcoming', label: 'UPCOMING' },
  { key: 'results', label: 'RESULTS' },
]

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

  const items = data?.items ?? []

  return (
    <div className="animate-in">
      <div className="mb-8">
        <h2 className="font-display text-neon text-lg tracking-[0.2em] mb-1">
          TAC.MATCHES
        </h2>
        <div className="h-[1px] w-full bg-gradient-to-r from-neon/50 via-neon/20 to-transparent" />
      </div>

      {/* Tabs */}
      <div className="flex gap-1 mb-5">
        {tabs.map((t) => (
          <button
            key={t.key}
            onClick={() => setTab(t.key)}
            className={`px-4 py-1.5 text-[11px] tracking-[0.12em] border transition-all ${
              tab === t.key
                ? 'border-neon bg-neon-dim text-neon'
                : 'border-border bg-panel text-steel hover:text-text hover:border-steel-dim'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {/* Filter input */}
      <input
        placeholder="FILTER BY TEAM..."
        value={team}
        onChange={(e) => setTeam(e.target.value)}
        className="w-56 bg-panel border border-border text-text text-[11px] px-3 py-1.5 mb-5 tracking-wider focus:outline-none focus:border-neon/50 placeholder:text-text-dim/50"
      />

      {/* Match list */}
      <div className="space-y-0.5">
        {items.length === 0 && (
          <div className="text-text-dim text-[11px] tracking-wider py-8 text-center border border-border bg-panel">
            {data ? 'NO DATA' : 'LOADING...'}
          </div>
        )}
        {items.map((m: any, i: number) => (
          <div
            key={i}
            className="bg-panel border border-border p-3 flex items-center gap-4 animate-in hover:border-neon/20 transition-colors"
            style={{ animationDelay: `${i * 30}ms` }}
          >
            {/* Team 1 */}
            <div className="flex-1 text-right text-[12px] font-semibold text-text tracking-wide truncate">
              {m.team1 ?? 'TBD'}
            </div>

            {/* Score or Time */}
            <div className="flex flex-col items-center min-w-[80px]">
              {m.score ? (
                <span className="text-sm font-bold text-neon tracking-wider">{m.score}</span>
              ) : (
                <span className="text-sm font-bold text-orange tracking-wider">
                  {m.scheduled_at ?? '--:--'}
                </span>
              )}
              {m.best_of && (
                <span className="text-[9px] text-steel mt-0.5">{m.best_of}</span>
              )}
            </div>

            {/* Team 2 */}
            <div className="flex-1 text-[12px] font-semibold text-text tracking-wide truncate">
              {m.team2 ?? 'TBD'}
            </div>

            {/* Event */}
            <div className="w-48 text-right text-[10px] text-steel truncate">
              {m.event ?? ''}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
