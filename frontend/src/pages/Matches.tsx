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

// Shorten event names: strip verbose suffixes, keep core identity
function shortEvent(name?: string): string {
  if (!name) return ''
  let s = name
  // Strip verbose suffixes
  s = s.replace(/\bSeason\s*\d+\b/gi, '')
  s = s.replace(/\bSeries\s*\d+\b/gi, '')
  s = s.replace(/\bRound\s*of\s*\d+\b/gi, '')
  s = s.replace(/\bCup\s*\d+\b/gi, '')
  s = s.replace(/\bGroup\s*[A-D]\b/gi, '')
  s = s.replace(/\bStage\s*\d+\b/gi, '')
  s = s.replace(/\bDivision\s*\d+\b/gi, '')
  s = s.replace(/\bLower\s*Bracket\b/gi, '')
  s = s.replace(/\bUpper\s*Bracket\b/gi, '')
  s = s.replace(/\bSemi-final\b/gi, '')
  s = s.replace(/\bQuarterfinal\b/gi, '')
  s = s.replace(/\bFinals?\b/gi, '')
  s = s.replace(/\bClosed\s*Qualifier\b/gi, '预选')
  s = s.replace(/\bQualifier\b/gi, '预选')
  s = s.replace(/\bOpen\b/gi, '')
  s = s.replace(/\bClosed\b/gi, '')
  s = s.replace(/\bChampionship\b/gi, '')
  s = s.replace(/\bMasters\b/gi, '大师赛')
  s = s.replace(/\s{2,}/g, ' ')
  return s.trim()
}

function cn(name?: string) { return nicknames[name ?? ''] ?? '' }

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

  return (
    <div className="anim-in space-y-6">
      {/* ---- Tab bar ---- */}
      <div className="flex items-center gap-4">
        {tabs.map(t => (
          <button key={t.key} onClick={() => setTab(t.key)}
            className={`text-[18px] font-display font-semibold tracking-wider uppercase pb-2 border-b-[3px] transition-colors ${
              tab === t.key
                ? 'text-gold border-gold'
                : 'text-text-muted border-transparent hover:text-text-secondary'
            }`}>
            {t.label}
          </button>
        ))}
        <div className="flex-1" />
        {/* Star count */}
        {items.length > 0 && (
          <span className="text-text-muted text-[15px]">{items.length} 场比赛</span>
        )}
        <input
          placeholder="筛选队伍..."
          value={filter}
          onChange={e => setFilter(e.target.value)}
          className="w-52 bg-card border border-border text-text text-[15px] px-4 py-2 rounded-lg focus:outline-none focus:border-gold placeholder:text-text-muted"
        />
      </div>

      {/* ---- 2-col match grid ---- */}
      <div className="grid grid-cols-2 gap-3">
        {items.length === 0 && (
          <div className="col-span-2 text-text-muted text-[16px] py-24 text-center bg-card border border-border rounded-lg">
            {data ? '暂无比赛数据' : '加载中...'}
          </div>
        )}

        {items.map((m, i) => {
          const c1 = cn(m.team1)
          const c2 = cn(m.team2)
          const evt = shortEvent(m.event)

          return (
            <div key={i}
              className="anim-in bg-card border border-border rounded-lg px-5 py-4 flex flex-col gap-3 hover:border-gold/30 transition-colors"
              style={{ animationDelay: `${i * 25}ms` }}>

              {/* ---- Teams + Score row ---- */}
              <div className="flex items-center gap-3">
                {/* Team 1 — fixed width, center-aligned, reserves height for nickname */}
                <div className="flex-1 min-w-0 flex flex-col items-center justify-center min-h-[52px]">
                  <span className="text-[18px] font-display font-semibold text-text leading-tight tracking-wide truncate max-w-full">
                    {m.team1 ?? 'TBD'}
                  </span>
                  <span className="text-[13px] text-text-muted mt-0.5 h-[18px]">
                    {c1}
                  </span>
                </div>

                {/* Score or Time — fixed width */}
                <div className="shrink-0 w-[110px] flex flex-col items-center justify-center">
                  {m.score ? (
                    <span className="text-[38px] font-display font-bold text-text leading-none tracking-tight">
                      {m.score}
                    </span>
                  ) : (
                    <span className="text-[38px] font-display font-bold text-gold leading-none tracking-tight">
                      {m.scheduled_at ? m.scheduled_at.slice(11, 16) : '—:—'}
                    </span>
                  )}
                  <span className="text-[12px] text-text-muted mt-0.5 h-[16px] flex items-center">
                    {m.best_of && m.best_of}
                  </span>
                </div>

                {/* Team 2 — fixed width, center-aligned, reserves height for nickname */}
                <div className="flex-1 min-w-0 flex flex-col items-center justify-center min-h-[52px]">
                  <span className="text-[18px] font-display font-semibold text-text leading-tight tracking-wide truncate max-w-full">
                    {m.team2 ?? 'TBD'}
                  </span>
                  <span className="text-[13px] text-text-muted mt-0.5 h-[18px]">
                    {c2}
                  </span>
                </div>
              </div>

              {/* ---- Event row ---- */}
              {evt && (
                <div className="text-center text-[13px] text-gold font-medium tracking-wide border-t border-border pt-2.5 truncate">
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
