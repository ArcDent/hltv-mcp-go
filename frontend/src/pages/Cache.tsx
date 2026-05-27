import { useEffect, useState } from 'react'
import { api } from '../api/client'

export default function Cache() {
  const [stats, setStats] = useState<any>(null)
  const [cleared, setCleared] = useState(false)

  const refresh = () => { api.cacheStats().then(setStats).catch(() => {}) }
  useEffect(refresh, [])

  const clear = async () => {
    await api.clearCache(); setCleared(true); refresh()
    setTimeout(() => setCleared(false), 2500)
  }

  const cards = [
    { label: '缓存条目', value: stats?.entries ?? '—' },
    { label: '命中次数', value: stats?.hits    ?? '—' },
    { label: '未命中',   value: stats?.misses  ?? '—' },
  ]

  const cardStyle: React.CSSProperties = {
    background: 'var(--card)', border: '1px solid var(--border)',
    borderRadius: 'var(--radius)', padding: '28px 24px', boxShadow: 'var(--card-shadow)',
    display: 'flex', flexDirection: 'column', alignItems: 'center', textAlign: 'center',
  }

  return (
    <div className="anim-in" style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 16 }}>
        {cards.map((c, i) => (
          <div key={c.label} className="anim-in" style={{ ...cardStyle, animationDelay: `${i * 80}ms` }}>
            <span style={{ fontFamily: 'var(--font-display)', fontSize: 48, fontWeight: 700,
              color: 'var(--text)', lineHeight: 1, marginBottom: 10 }}>{c.value}</span>
            <span style={{ fontSize: 14, color: 'var(--text-secondary)', fontWeight: 500 }}>{c.label}</span>
          </div>
        ))}
      </div>

      <div style={{ display: 'flex', gap: 16, alignItems: 'center' }}>
        <button onClick={clear} style={{
          padding: '10px 24px', background: 'var(--red-dim)', color: 'var(--red)',
          border: '1px solid var(--red)', borderRadius: 'var(--radius-sm)',
          fontSize: 15, fontWeight: 600, fontFamily: 'var(--font-display)',
          letterSpacing: '0.04em', textTransform: 'uppercase', cursor: 'pointer',
        }}>清除全部缓存</button>
        <button onClick={refresh} style={{
          padding: '10px 24px', background: 'transparent', color: 'var(--text-secondary)',
          border: '1px solid var(--border)', borderRadius: 'var(--radius-sm)',
          fontSize: 15, fontWeight: 600, fontFamily: 'var(--font-display)',
          letterSpacing: '0.04em', textTransform: 'uppercase', cursor: 'pointer',
        }}>刷新</button>
        {cleared && <span style={{ fontSize: 15, color: 'var(--gold)', fontWeight: 500 }}>✓ 缓存已清除</span>}
      </div>
    </div>
  )
}
