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

  const items = data?.items ?? []

  return (
    <div className="animate-in">
      <div className="mb-8">
        <h2 className="font-display text-neon text-lg tracking-[0.2em] mb-1">
          COM.NEWS
        </h2>
        <div className="h-[1px] w-full bg-gradient-to-r from-neon/50 via-neon/20 to-transparent" />
      </div>

      <div className="flex gap-1 mb-5">
        {([{ key: 'realtime', label: 'REALTIME' }, { key: 'archive', label: 'ARCHIVE' }] as const).map((t) => (
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

      <div className="space-y-0.5">
        {items.length === 0 && (
          <div className="text-text-dim text-[11px] tracking-wider py-8 text-center border border-border bg-panel">
            {data ? 'NO DATA' : 'LOADING...'}
          </div>
        )}
        {items.map((n: any, i: number) => (
          <div
            key={i}
            className="bg-panel border border-border p-3 flex items-center gap-4 animate-in hover:border-neon/20 transition-colors group"
            style={{ animationDelay: `${i * 25}ms` }}
          >
            <span className="text-neon text-[10px] w-6 shrink-0">{String(i + 1).padStart(2, '0')}</span>
            <span className="flex-1 text-[11px] text-text tracking-wide group-hover:text-neon/80 transition-colors line-clamp-1">
              {n.title}
            </span>
            <span className="text-[10px] text-text-dim shrink-0">
              {n.published_at ?? n.relative_time ?? ''}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}
