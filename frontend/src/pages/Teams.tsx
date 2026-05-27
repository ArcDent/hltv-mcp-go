import { useState } from 'react'
import { api } from '../api/client'

export default function Teams() {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<any[] | null>(null)
  const [loading, setLoading] = useState(false)

  const search = async () => {
    if (!query.trim()) return
    setLoading(true)
    try {
      const resp = await api.search(query, 'team')
      setResults(resp?.items ?? [])
    } catch { setResults([]) }
    setLoading(false)
  }

  return (
    <div className="animate-in">
      <div className="mb-8">
        <h2 className="font-display text-neon text-lg tracking-[0.2em] mb-1">
          INTEL.TEAMS
        </h2>
        <div className="h-[1px] w-full bg-gradient-to-r from-neon/50 via-neon/20 to-transparent" />
      </div>

      <div className="flex gap-2 mb-6">
        <input
          placeholder="SEARCH TEAMS..."
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && search()}
          className="flex-1 bg-panel border border-border text-text text-[12px] px-4 py-2 tracking-wider focus:outline-none focus:border-neon/50 placeholder:text-text-dim/50"
        />
        <button
          onClick={search}
          className="px-6 py-2 border border-neon/30 bg-neon-dim text-neon text-[11px] tracking-[0.15em] hover:bg-neon/20 transition-colors disabled:opacity-30"
          disabled={loading}
        >
          {loading ? '...' : 'EXECUTE'}
        </button>
      </div>

      <div className="space-y-0.5">
        {results === null && (
          <div className="text-text-dim text-[11px] tracking-wider py-12 text-center border border-border bg-panel">
            ENTER SEARCH QUERY
          </div>
        )}
        {results?.length === 0 && (
          <div className="text-text-dim text-[11px] tracking-wider py-8 text-center border border-border bg-panel">
            NO RESULTS
          </div>
        )}
        {results?.map((t: any, i: number) => (
          <div
            key={i}
            className="bg-panel border border-border p-3 flex items-center gap-4 animate-in hover:border-neon/20 transition-colors"
            style={{ animationDelay: `${i * 40}ms` }}
          >
            <span className="text-neon text-[10px] w-6">{String(i + 1).padStart(2, '0')}</span>
            <span className="flex-1 text-[12px] text-text font-semibold tracking-wide">{t.name}</span>
            <span className="text-[10px] text-steel bg-surface border border-border px-2 py-0.5">
              ID:{t.id ?? '--'}
            </span>
            {t.slug && (
              <span className="text-[10px] text-text-dim">{t.slug}</span>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
