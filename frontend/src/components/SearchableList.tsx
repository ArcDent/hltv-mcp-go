import { useState } from 'react'
import PlayerDetail from './PlayerDetail'
import TeamDetail from './TeamDetail'
import { teamNicknames, playerNicknames } from '../data/nicknames'

type Props = {
  type: 'team' | 'player'
  placeholder: string
  emptyHint: string
  apiSearch: (q: string) => Promise<any>
}

const cardStyle: React.CSSProperties = {
  background: 'var(--card)', border: '1px solid var(--border)',
  borderRadius: 'var(--radius)', padding: '14px 20px', boxShadow: 'var(--card-shadow)',
  display: 'flex', alignItems: 'center', gap: 16,
}
const inputStyle: React.CSSProperties = {
  flex: 1, background: 'var(--input-bg)', border: '1px solid var(--border)',
  borderRadius: 'var(--radius-sm)', color: 'var(--text)', fontSize: 16,
  padding: '12px 18px', outline: 'none',
}
const btnStyle: React.CSSProperties = {
  padding: '12px 28px', background: 'var(--gold)', color: 'var(--bg)', border: 'none',
  borderRadius: 'var(--radius-sm)', fontSize: 16, fontWeight: 600,
  fontFamily: 'var(--font-display)', letterSpacing: '0.04em', textTransform: 'uppercase',
}

const focusIn = (e: React.FocusEvent<HTMLInputElement>) => {
  e.target.style.borderColor = 'var(--gold)'
  e.target.style.boxShadow = '0 0 0 3px var(--gold-dim)'
}
const focusOut = (e: React.FocusEvent<HTMLInputElement>) => {
  e.target.style.borderColor = 'var(--border)'
  e.target.style.boxShadow = 'none'
}

export default function SearchableList({ type, placeholder, emptyHint, apiSearch }: Props) {
  const [q, setQ] = useState('')
  const [list, setList] = useState<any[] | null>(null)
  const [loading, setLoading] = useState(false)
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [selectedTeamId, setSelectedTeamId] = useState<number | null>(null)

  const search = async () => {
    if (!q.trim()) return
    setLoading(true)
    try { const r = await apiSearch(q); setList(r?.items ?? []) } catch { setList([]) }
    setLoading(false)
  }

  return (
    <>
      <div className="anim-in" style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <div style={{ display: 'flex', gap: 12 }}>
          <input placeholder={placeholder}
            value={q} onChange={e => setQ(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && search()}
            style={inputStyle} onFocus={focusIn} onBlur={focusOut}
          />
          <button onClick={search} disabled={loading}
            style={{ ...btnStyle, opacity: loading ? 0.4 : 1, cursor: loading ? 'not-allowed' : 'pointer' }}>
            {loading ? '搜索中' : '搜索'}
          </button>
        </div>

        {list === null && (
          <div style={{ ...cardStyle, justifyContent: 'center', padding: '80px 0', color: 'var(--text-muted)', fontSize: 15 }}>
            {emptyHint}
          </div>
        )}
        {list?.length === 0 && (
          <div style={{ ...cardStyle, justifyContent: 'center', padding: '80px 0', color: 'var(--text-muted)', fontSize: 15 }}>
            无匹配结果
          </div>
        )}
        {list?.map((item, i) => (
          <div key={i} className="anim-in" onClick={() => {
              if (type === 'player' && item.id) setSelectedId(item.id)
              if (type === 'team' && item.id) setSelectedTeamId(item.id)
            }}
            style={{ ...cardStyle, animationDelay: `${i * 35}ms`, cursor: (type === 'player' || type === 'team') ? 'pointer' : 'default' }}>
            <span style={{ fontFamily: 'var(--font-mono)', fontSize: 15, fontWeight: 700, color: 'var(--gold)', minWidth: 28 }}>
              {String(i + 1).padStart(2, '0')}
            </span>
            <span style={{ flex: 1, fontSize: 17, fontWeight: 600, fontFamily: 'var(--font-display)', letterSpacing: '0.03em' }}>
              {item.name}
              {(playerNicknames[item.name] || teamNicknames[item.name]) && (
                <span style={{ fontSize: 12, color: 'var(--text-muted)', marginLeft: 8, fontWeight: 400 }}>
                  {playerNicknames[item.name] || teamNicknames[item.name]}
                </span>
              )}
            </span>
            <span style={{ fontSize: 13, color: 'var(--text-muted)', background: 'var(--input-bg)',
              border: '1px solid var(--border)', borderRadius: 'var(--radius-sm)', padding: '3px 10px',
              fontFamily: 'var(--font-mono)' }}>
              ID {item.id ?? '—'}
            </span>
            {item.slug && <span style={{ fontSize: 13, color: 'var(--text-muted)' }}>{item.slug}</span>}
          </div>
        ))}
      </div>
      {type === 'player' && selectedId !== null && <PlayerDetail id={selectedId} onClose={() => setSelectedId(null)} />}
      {type === 'team' && selectedTeamId !== null && <TeamDetail id={selectedTeamId} onClose={() => setSelectedTeamId(null)} />}
    </>
  )
}
