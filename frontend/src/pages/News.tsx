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

  const card: React.CSSProperties = {
    background: 'var(--card)', border: '1px solid var(--border)',
    borderRadius: 'var(--radius)', padding: '14px 20px', boxShadow: 'var(--card-shadow)',
  }
  const tabBtn = (active: boolean): React.CSSProperties => ({
    fontSize: 16, fontWeight: 600, fontFamily: 'var(--font-display)',
    letterSpacing: '0.04em', textTransform: 'uppercase' as const,
    color: active ? 'var(--gold)' : 'var(--text-muted)',
    borderBottom: active ? '2px solid var(--gold)' : '2px solid transparent',
    paddingBottom: 6, background: 'none', cursor: 'pointer',
  })

  return (
    <div className="anim-in" style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div style={{ display: 'flex', gap: 24, borderBottom: '1px solid var(--border)', paddingBottom: 0 }}>
        {[{ key: 'realtime', label: '实时新闻' }, { key: 'archive', label: '归档新闻' }].map(t => (
          <button key={t.key} onClick={() => setTab(t.key as Tab)} style={tabBtn(tab === t.key)}>
            {t.label}
          </button>
        ))}
      </div>

      <div style={{ position: 'relative' }}>
        {/* Tab content with slide animation */}
        <div key={tab} style={{ animation: 'slideUp 0.3s ease both' }}>
          {items.length === 0 && (
            <div style={{ ...card, textAlign: 'center', padding: '80px 0', color: 'var(--text-muted)', fontSize: 15 }}>
              {data ? '暂无新闻' : '加载中...'}
            </div>
          )}
          {items.map((n, i) => (
            <div key={i} className="anim-in" style={{
              ...card, marginBottom: i < items.length - 1 ? 6 : 0,
              flexDirection: 'column', alignItems: 'stretch', gap: 4,
              animationDelay: `${i * 30}ms`,
            }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: 14, fontWeight: 700,
                  color: 'var(--gold)', minWidth: 24 }}>
                  {String(i + 1).padStart(2, '0')}
                </span>
                <span style={{ flex: 1, fontSize: 16, fontWeight: 500 }}>{n.title}</span>
                <span style={{ fontSize: 13, color: 'var(--text-muted)', whiteSpace: 'nowrap' }}>
                  {n.published_at ?? n.relative_time ?? ''}
                </span>
              </div>
              {/* Chinese translation / summary line */}
              {n.summary_hint && (
                <div style={{ fontSize: 13, color: 'var(--text-muted)', paddingLeft: 38, lineHeight: 1.5 }}>
                  {n.summary_hint}
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
