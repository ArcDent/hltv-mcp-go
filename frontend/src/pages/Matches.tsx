import { useEffect, useState } from 'react'
import { api } from '../api/client'

type Tab = 'today' | 'upcoming' | 'results'

const tabs: { key: Tab; label: string }[] = [
  { key: 'today',    label: '今日赛程' },
  { key: 'upcoming', label: '即将开始' },
  { key: 'results',  label: '近期赛果' },
]

const nicknames: Record<string, string> = {
  'Vitality':'小蜜蜂','Spirit':'绿龙','Team Spirit':'绿龙','Natus Vincere':'天生赢家',
  'NAVI':'天生赢家','FaZe':'FaZe Clan','G2':'武士','MOUZ':'老鼠','Falcons':'猎鹰',
  'Astralis':'A队','Virtus.pro':'VP','Team Liquid':'液体','FURIA':'黑豹',
  'The MongolZ':'蒙古队','TYLOO':'天禄','3DMAX':'3DMAX','paiN':'paiN',
  'HEROIC':'HEROIC','Complexity':'coL','Ninjas in Pyjamas':'NIP',
  'Eternal Fire':'永火','fnatic':'橙黑','Rare Atom':'RA','Lynn Vision':'LVG',
  'Aurora':'欧若拉','RED Canids':'红犬','GamerLegion':'GL','PARIVISION':'PV',
}

function shortEvent(name?: string): string {
  if (!name) return ''
  return name
    .replace(/\bSeason\s*\d+\b/gi,'').replace(/\bSeries\s*\d+\b/gi,'')
    .replace(/\bRound\s*of\s*\d+\b/gi,'').replace(/\bCup\s*\d+\b/gi,'')
    .replace(/\bGroup\s*[A-D]\b/gi,'').replace(/\bStage\s*\d+\b/gi,'')
    .replace(/\bLower\s*Bracket\b/gi,'').replace(/\bUpper\s*Bracket\b/gi,'')
    .replace(/\bSemi-final\b/gi,'').replace(/\bQuarterfinal\b/gi,'')
    .replace(/\bFinals?\b/gi,'').replace(/\bClosed\s*Qualifier\b/gi,'预选')
    .replace(/\bQualifier\b/gi,'预选').replace(/\bOpen\b/gi,'')
    .replace(/\bChampionship\b/gi,'').replace(/\bMasters\b/gi,'大师赛')
    .replace(/\s{2,}/g,' ').trim()
}

export default function Matches() {
  const [tab, setTab] = useState<Tab>('today')
  const [data, setData] = useState<any>(null)
  const [filter, setFilter] = useState('')

  useEffect(() => {
    setData(null)
    if (tab === 'today')    api.todayMatches().then(setData)
    else if (tab === 'upcoming') api.upcomingMatches({ team: filter, limit: '40' }).then(setData)
    else api.results({ team: filter, limit: '40' }).then(setData)
  }, [tab, filter])

  const items: any[] = data?.items ?? []

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

  const inputStyle: React.CSSProperties = {
    background: 'var(--input-bg)', border: '1px solid var(--border)',
    borderRadius: 'var(--radius-sm)', color: 'var(--text)',
    fontSize: 14, padding: '8px 14px', outline: 'none', width: 200,
  }

  return (
    <div className="anim-in" style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      {/* Tab bar + filter */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 24 }}>
        {tabs.map(t => (
          <button key={t.key} onClick={() => setTab(t.key)} style={tabBtn(tab === t.key)}>
            {t.label}
          </button>
        ))}
        <div style={{ flex: 1 }} />
        {items.length > 0 && (
          <span style={{ fontSize: 14, color: 'var(--text-muted)' }}>{items.length} 场比赛</span>
        )}
        <input placeholder="筛选队伍..." value={filter} onChange={e => setFilter(e.target.value)}
          style={inputStyle}
          onFocus={e => { e.target.style.borderColor = 'var(--gold)'; e.target.style.boxShadow = '0 0 0 3px var(--gold-dim)'; }}
          onBlur={e => { e.target.style.borderColor = 'var(--border)'; e.target.style.boxShadow = 'none'; }}
        />
      </div>

      {/* Match grid */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 12 }}>
        {items.length === 0 && (
          <div style={{ ...cardStyle, gridColumn: '1 / -1', textAlign: 'center', padding: '80px 0',
            color: 'var(--text-muted)', fontSize: 15 }}>
            {data ? '暂无比赛数据' : '加载中...'}
          </div>
        )}

        {items.map((m, i) => {
          const c1 = nicknames[m.team1 ?? ''] ?? ''
          const c2 = nicknames[m.team2 ?? ''] ?? ''
          const evt = shortEvent(m.event)

          return (
            <div key={i} className="anim-in" style={{ ...cardStyle, display: 'flex', flexDirection: 'column', gap: 12, animationDelay: `${i * 25}ms` }}>
              {/* Teams + Score row */}
              <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                <div style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', minHeight: 50 }}>
                  <span style={{ fontSize: 17, fontWeight: 600, fontFamily: 'var(--font-display)',
                    color: 'var(--text)', letterSpacing: '0.03em', textAlign: 'center' }}>
                    {m.team1 || '待定'}
                  </span>
                  <span style={{ fontSize: 13, color: 'var(--text-muted)', height: 18 }}>{c1}</span>
                </div>

                <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', minWidth: 100 }}>
                  {m.score ? (
                    <span style={{ fontSize: 36, fontFamily: 'var(--font-display)', fontWeight: 700, color: 'var(--text)', lineHeight: 1 }}>
                      {m.score}
                    </span>
                  ) : (
                    <span style={{ fontSize: 36, fontFamily: 'var(--font-display)', fontWeight: 700, color: 'var(--gold)', lineHeight: 1 }}>
                      {m.scheduled_at ? m.scheduled_at.slice(11, 16) : '—:—'}
                    </span>
                  )}
                  <span style={{ fontSize: 11, color: 'var(--text-muted)', height: 16 }}>
                    {m.best_of ?? ''}
                  </span>
                </div>

                <div style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', minHeight: 50 }}>
                  <span style={{ fontSize: 17, fontWeight: 600, fontFamily: 'var(--font-display)',
                    color: 'var(--text)', letterSpacing: '0.03em', textAlign: 'center' }}>
                    {m.team2 || '待定'}
                  </span>
                  <span style={{ fontSize: 13, color: 'var(--text-muted)', height: 18 }}>{c2}</span>
                </div>
              </div>

              {/* Event */}
              {evt && (
                <div style={{ textAlign: 'center', fontSize: 13, color: 'var(--gold)',
                  fontWeight: 500, borderTop: '1px solid var(--border)', paddingTop: 10 }}>
                  {evt}
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
