import { useState } from 'react'
import { api } from '../api/client'

const card: React.CSSProperties = {
  background: 'var(--card)', border: '1px solid var(--border)',
  borderRadius: 'var(--radius)', padding: '14px 20px', boxShadow: 'var(--card-shadow)',
  display: 'flex', alignItems: 'center', gap: 16,
}
const inputS: React.CSSProperties = {
  flex: 1, background: 'var(--input-bg)', border: '1px solid var(--border)',
  borderRadius: 'var(--radius-sm)', color: 'var(--text)', fontSize: 16,
  padding: '12px 18px', outline: 'none',
}

export default function Teams() {
  const [q, setQ] = useState('')
  const [list, setList] = useState<any[] | null>(null)
  const [loading, setLoading] = useState(false)

  const search = async () => {
    if (!q.trim()) return
    setLoading(true)
    try { const r = await api.search(q, 'team'); setList(r?.items ?? []) } catch { setList([]) }
    setLoading(false)
  }

  return (
    <div className="anim-in" style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div style={{ display: 'flex', gap: 12 }}>
        <input placeholder="搜索队伍 — 支持英文 / 中文 / 别名（如 Spirit、绿龙、小蜜蜂）"
          value={q} onChange={e => setQ(e.target.value)} onKeyDown={e => e.key === 'Enter' && search()}
          style={inputS}
          onFocus={e => { e.target.style.borderColor = 'var(--gold)'; e.target.style.boxShadow = '0 0 0 3px var(--gold-dim)'; }}
          onBlur={e => { e.target.style.borderColor = 'var(--border)'; e.target.style.boxShadow = 'none'; }}
        />
        <button onClick={search} disabled={loading} style={{
          padding: '12px 28px', background: 'var(--gold)', color: 'var(--bg)', border: 'none',
          borderRadius: 'var(--radius-sm)', fontSize: 16, fontWeight: 600,
          fontFamily: 'var(--font-display)', letterSpacing: '0.04em', textTransform: 'uppercase',
          opacity: loading ? 0.4 : 1, cursor: loading ? 'not-allowed' : 'pointer',
        }}>{loading ? '搜索中' : '搜索'}</button>
      </div>

      {list === null && (
        <div style={{ ...card, justifyContent: 'center', padding: '80px 0', color: 'var(--text-muted)', fontSize: 15 }}>
          输入队名开始搜索
        </div>
      )}
      {list?.length === 0 && (
        <div style={{ ...card, justifyContent: 'center', padding: '80px 0', color: 'var(--text-muted)', fontSize: 15 }}>
          无匹配结果
        </div>
      )}
      {list?.map((t, i) => (
        <div key={i} className="anim-in" style={{ ...card, animationDelay: `${i * 35}ms` }}>
          <span style={{ fontFamily: 'var(--font-mono)', fontSize: 15, fontWeight: 700, color: 'var(--gold)', minWidth: 28 }}>
            {String(i + 1).padStart(2, '0')}
          </span>
          <span style={{ flex: 1, fontSize: 17, fontWeight: 600, fontFamily: 'var(--font-display)', letterSpacing: '0.03em' }}>
            {t.name}
          </span>
          <span style={{ fontSize: 13, color: 'var(--text-muted)', background: 'var(--input-bg)',
            border: '1px solid var(--border)', borderRadius: 'var(--radius-sm)', padding: '3px 10px',
            fontFamily: 'var(--font-mono)' }}>
            ID {t.id ?? '—'}
          </span>
          {t.slug && <span style={{ fontSize: 13, color: 'var(--text-muted)' }}>{t.slug}</span>}
        </div>
      ))}
    </div>
  )
}
