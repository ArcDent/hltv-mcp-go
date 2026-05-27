import { useEffect, useState } from 'react'
import { api } from '../api/client'

export default function Cache() {
  const [stats, setStats] = useState<any>(null)
  const [cleared, setCleared] = useState(false)

  const refresh = () => { api.cacheStats().then(setStats).catch(() => {}) }
  useEffect(refresh, [])

  const handleClear = async () => {
    await api.clearCache()
    setCleared(true)
    refresh()
    setTimeout(() => setCleared(false), 2500)
  }

  return (
    <div className="animate-in">
      <div className="mb-8">
        <h2 className="font-display text-neon text-lg tracking-[0.2em] mb-1">
          SYS.CACHE
        </h2>
        <div className="h-[1px] w-full bg-gradient-to-r from-neon/50 via-neon/20 to-transparent" />
      </div>

      <div className="grid grid-cols-3 gap-3 mb-6">
        {[
          { label: 'ENTRIES', value: stats?.entries ?? '--' },
          { label: 'HITS', value: stats?.hits ?? '--' },
          { label: 'MISSES', value: stats?.misses ?? '--' },
        ].map((s, i) => (
          <div
            key={s.label}
            className="bg-panel border border-border p-4 animate-in"
            style={{ animationDelay: `${i * 80}ms` }}
          >
            <div className="text-[10px] text-text-dim tracking-[0.15em] mb-2">{s.label}</div>
            <div className="text-2xl font-bold text-text">{s.value}</div>
          </div>
        ))}
      </div>

      <div className="flex items-center gap-4">
        <button
          onClick={handleClear}
          className="px-5 py-2 border border-orange/40 bg-orange-dim text-orange text-[11px] tracking-[0.15em] hover:bg-orange/20 transition-colors"
        >
          PURGE ALL CACHE
        </button>
        <button
          onClick={refresh}
          className="px-5 py-2 border border-border bg-panel text-steel text-[11px] tracking-[0.15em] hover:text-text hover:border-steel-dim transition-colors"
        >
          REFRESH
        </button>
        {cleared && (
          <span className="text-neon text-[11px] animate-in">CACHE PURGED</span>
        )}
      </div>
    </div>
  )
}
