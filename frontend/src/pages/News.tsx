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

  const items: any[] = data?.items ?? []

  return (
    <div className="animate-in">
      <h2 className="text-[16px] font-semibold text-gold mb-4 pb-2 border-b border-border tracking-wide">
        ◉ 新闻
      </h2>

      <div className="flex gap-1 mb-6">
        {[{ key: 'realtime' as Tab, label: '实时新闻' }, { key: 'archive' as Tab, label: '归档新闻' }].map((t) => (
          <button key={t.key} onClick={() => setTab(t.key)}
            className={`px-5 py-2 text-[14px] font-medium rounded-md border transition-all ${
              tab === t.key
                ? 'border-gold bg-gold-dim text-gold'
                : 'border-border bg-panel text-text-dim hover:text-text hover:border-gold-border'
            }`}>
            {t.label}
          </button>
        ))}
      </div>

      <div className="space-y-1">
        {items.length === 0 && (
          <div className="text-text-dim text-[14px] py-16 text-center border border-border rounded-md bg-panel">
            {data ? '暂无新闻' : '加载中...'}
          </div>
        )}
        {items.map((n: any, i: number) => (
          <div key={i} className="bg-panel border border-border rounded-md p-3.5 flex items-center gap-4 animate-in hover:border-gold-border transition-colors group"
            style={{ animationDelay: `${i * 25}ms` }}>
            <span className="text-gold text-[14px] font-mono font-semibold w-7 shrink-0">{String(i + 1).padStart(2, '0')}</span>
            <span className="flex-1 text-[15px] text-text group-hover:text-gold transition-colors truncate">{n.title}</span>
            <span className="text-[13px] text-text-dim shrink-0">{n.published_at ?? n.relative_time ?? ''}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
