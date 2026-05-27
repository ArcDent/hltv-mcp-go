import { useEffect, useState } from 'react'
import { api } from '../api/client'

type Tab = 'realtime' | 'archive'

export default function News() {
  const [tab, setTab] = useState<Tab>('realtime')
  const [data, setData] = useState<any>(null)

  useEffect(() => {
    setData(null)
    if (tab === 'realtime') api.realtimeNews().then(setData)
    else api.newsDigest({ limit: '30' }).then(setData)
  }, [tab])

  const items: any[] = data?.items ?? []

  return (
    <div className="anim-in space-y-8">
      <div className="flex items-center gap-4">
        {[{ key: 'realtime', label: '实时新闻' }, { key: 'archive', label: '归档新闻' }].map(t => (
          <button key={t.key} onClick={() => setTab(t.key as Tab)}
            className={`text-[17px] font-display font-semibold tracking-wider uppercase pb-2 border-b-[3px] transition-colors ${
              tab === t.key
                ? 'text-gold border-gold'
                : 'text-text-muted border-transparent hover:text-text-secondary'
            }`}>
            {t.label}
          </button>
        ))}
      </div>

      <div className="space-y-[1px]">
        {items.length === 0 && (
          <div className="text-text-muted text-[16px] py-24 text-center bg-card border border-border rounded-lg">
            {data ? '暂无新闻' : '加载中...'}
          </div>
        )}
        {items.map((n, i) => (
          <div key={i}
            className="anim-in bg-card border border-border rounded-lg px-6 py-4 flex items-center gap-5 hover:border-gold/30 transition-colors group"
            style={{ animationDelay: `${i * 25}ms` }}>
            <span className="text-gold font-mono font-bold text-[16px] w-8 shrink-0">
              {String(i + 1).padStart(2, '0')}
            </span>
            <span className="flex-1 text-[17px] font-medium group-hover:text-gold transition-colors truncate">
              {n.title}
            </span>
            <span className="text-[14px] text-text-muted shrink-0">{n.published_at ?? n.relative_time ?? ''}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
