import { useEffect, useState } from 'react'
import { api } from '../api/client'
import useNicknames from '../hooks/useNicknames'
import Modal from '../components/Modal'

type Tab = 'today' | 'upcoming' | 'results'

const tabs: { key: Tab; label: string }[] = [
  { key: 'today',    label: '今日赛程' },
  { key: 'upcoming', label: '即将开始' },
  { key: 'results',  label: '近期赛果' },
]

export default function Matches() {
  const [tab, setTab] = useState<Tab>('today')
  const [events, setEvents] = useState<any[]>([])
  const [other, setOther] = useState<any[]>([])
  const [loading, setLoading] = useState(false)
  const [selectedEvent, setSelectedEvent] = useState<any>(null)
  const { teamNicknames: nicknames } = useNicknames()

  useEffect(() => {
    setLoading(true)
    setEvents([])
    setOther([])
    api.getEvents(tab, 150).then(d => {
      setEvents(d?.data?.events ?? [])
      setOther(d?.data?.other ?? [])
      setLoading(false)
    }).catch(() => setLoading(false))
  }, [tab])

  const cardStyle: React.CSSProperties = {
    background: 'var(--card)', border: '1px solid var(--border)',
    borderRadius: 'var(--radius)', padding: '16px 20px',
    boxShadow: 'var(--card-shadow)',
  }

  const tabBtn = (active: boolean): React.CSSProperties => ({
    fontSize: 16, fontWeight: 600, fontFamily: 'var(--font-display)',
    letterSpacing: '0.04em', textTransform: 'uppercase' as const,
    color: active ? 'var(--gold)' : 'var(--text-muted)',
    borderBottom: active ? '2px solid var(--gold)' : '2px solid transparent',
    paddingBottom: 6, background: 'none', cursor: 'pointer',
  })

  const totalEvents = events.length + (other.length > 0 ? 1 : 0)

  return (
    <div className="anim-in" style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>

      {/* Tab bar */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 24 }}>
        {tabs.map(t => (
          <button key={t.key} onClick={() => setTab(t.key)} style={tabBtn(tab === t.key)}>
            {t.label}
          </button>
        ))}
        <div style={{ flex: 1 }} />
        {!loading && totalEvents > 0 && (
          <span style={{ fontSize: 14, color: 'var(--text-muted)' }}>{totalEvents} 个赛事</span>
        )}
      </div>

      {/* Event cards grid */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 12 }}>
        {!loading && totalEvents === 0 && (
          <div style={{ ...cardStyle, gridColumn: '1 / -1', textAlign: 'center', padding: '80px 0', color: 'var(--text-muted)', fontSize: 15 }}>
            暂无赛事数据
          </div>
        )}

        {loading && (
          <div style={{ ...cardStyle, gridColumn: '1 / -1', textAlign: 'center', padding: '80px 0', color: 'var(--text-muted)', fontSize: 15 }}>
            加载中...
          </div>
        )}

        {events.map((ev, i) => (
          <div key={i} className="anim-in" style={{ ...cardStyle, cursor: 'pointer', animationDelay: `${i * 30}ms` }}
            onClick={() => setSelectedEvent(ev)}
            onMouseEnter={e => { (e.currentTarget as HTMLDivElement).style.borderColor = 'var(--gold)' }}
            onMouseLeave={e => { (e.currentTarget as HTMLDivElement).style.borderColor = 'var(--border)' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
              <span style={{ flex: 1, fontSize: 16, fontWeight: 700, fontFamily: 'var(--font-display)', letterSpacing: '0.03em', color: 'var(--text)' }}>
                {ev.name}
              </span>
              <span style={{ background: 'var(--gold-dim)', color: 'var(--gold)', fontFamily: 'var(--font-mono)', fontSize: 12, fontWeight: 700, padding: '3px 10px', borderRadius: 20 }}>
                {ev.match_count}
              </span>
            </div>
            <div style={{ marginTop: 8, fontSize: 12, color: 'var(--text-muted)' }}>
              {ev.date_start && ev.date_start.length === 10 ? ev.date_start.slice(5).replace('-','/') : '?'}
              {ev.date_end && ev.date_start !== ev.date_end && ev.date_end.length === 10 ? ' ~ ' + ev.date_end.slice(5).replace('-','/') : ''}
            </div>
          </div>
        ))}

        {/* Other bucket */}
        {other.length > 0 && (
          <div className="anim-in" style={{ ...cardStyle, cursor: 'pointer' }}
            onClick={() => setSelectedEvent({ name: 'Other', date_start: '—', date_end: '—', match_count: other.length, matches: other })}
            onMouseEnter={e => { (e.currentTarget as HTMLDivElement).style.borderColor = 'var(--gold)' }}
            onMouseLeave={e => { (e.currentTarget as HTMLDivElement).style.borderColor = 'var(--border)' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
              <span style={{ flex: 1, fontSize: 16, fontWeight: 700, fontFamily: 'var(--font-display)', letterSpacing: '0.03em', color: 'var(--text-secondary)' }}>
                Other
              </span>
              <span style={{ background: 'var(--input-bg)', color: 'var(--text-muted)', fontFamily: 'var(--font-mono)', fontSize: 12, fontWeight: 700, padding: '3px 10px', borderRadius: 20 }}>
                {other.length}
              </span>
            </div>
            <div style={{ marginTop: 8, fontSize: 12, color: 'var(--text-muted)' }}>未分配赛事</div>
          </div>
        )}
      </div>

      {/* Event Detail Modal */}
      {selectedEvent && (
        <Modal onClose={() => setSelectedEvent(null)}>

            <div style={{ fontFamily: 'var(--font-display)', fontSize: 20, fontWeight: 700, color: 'var(--gold)', letterSpacing: '0.04em', marginBottom: 6 }}>
              {selectedEvent.name}
            </div>
            <div style={{ fontSize: 13, color: 'var(--text-muted)', marginBottom: 18, paddingBottom: 14, borderBottom: '1px solid var(--border)' }}>
              {selectedEvent.date_start || '?'} ~ {selectedEvent.date_end || '?'} · {selectedEvent.match_count} 场比赛
            </div>

            {(selectedEvent.matches || []).map((m: any, i: number) => {
              const c1 = nicknames[m.team1 ?? ''] ?? ''
              const c2 = nicknames[m.team2 ?? ''] ?? ''
              return (
                <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 14, padding: '12px 0', borderTop: i > 0 ? '1px solid rgba(128,128,128,0.06)' : 'none', fontSize: 13 }}>
                  <div style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 2 }}>
                    <span style={{ fontSize: 15, fontWeight: 600, fontFamily: 'var(--font-display)', letterSpacing: '0.03em', color: 'var(--text)' }}>{m.team1 || '待定'}</span>
                    <span style={{ fontSize: 11, color: 'var(--text-muted)', height: 16 }}>{c1}</span>
                  </div>
                  {m.score ? (
                    <span style={{ fontFamily: 'var(--font-mono)', fontSize: 20, fontWeight: 700, color: 'var(--text)', minWidth: 50, textAlign: 'center' }}>{m.score}</span>
                  ) : (
                    (() => {
                      const t = m.scheduled_at
                      if (!t) return <span style={{ fontFamily: 'var(--font-display)', fontSize: 20, fontWeight: 700, color: 'var(--gold)', minWidth: 50, textAlign: 'center' }}>—:—</span>
                      const parts = t.split(' ')
                      const datePart = parts.length > 1 ? parts[0] : ''
                      const timePart = parts.length > 1 ? parts[1] : t
                      return (
                        <div style={{ minWidth: 50, textAlign: 'center' }}>
                          <div style={{ fontFamily: 'var(--font-display)', fontSize: 20, fontWeight: 700, color: 'var(--gold)', lineHeight: 1 }}>
                            {timePart.length >= 5 ? timePart.slice(0, 5) : timePart}
                          </div>
                          {datePart && (
                            <div style={{ fontFamily: 'var(--font-mono)', fontSize: 10, color: 'var(--text-muted)', marginTop: 2 }}>
                              {datePart.slice(5).replace('-', '/')}
                            </div>
                          )}
                        </div>
                      )
                    })()
                  )}
                  <div style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 2 }}>
                    <span style={{ fontSize: 15, fontWeight: 600, fontFamily: 'var(--font-display)', letterSpacing: '0.03em', color: 'var(--text)' }}>{m.team2 || '待定'}</span>
                    <span style={{ fontSize: 11, color: 'var(--text-muted)', height: 16 }}>{c2}</span>
                  </div>
                  <span style={{ fontSize: 11, color: m.score ? 'var(--text-muted)' : 'var(--gold)', minWidth: 60, textAlign: 'right' }}>
                    {m.best_of ? `${m.best_of.toUpperCase()}` : ''}
                  </span>
                </div>
              )
            })}
        </Modal>
      )}
    </div>
  )
}
