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
  'NAVI':'天生赢家','FaZe':'FaZe','G2':'武士','MOUZ':'老鼠','Falcons':'猎鹰',
  'Astralis':'A队','Virtus.pro':'VP','Team Liquid':'液体','FURIA':'黑豹',
  'The MongolZ':'蒙古队','TYLOO':'天禄','3DMAX':'3DMAX','paiN':'paiN',
  'HEROIC':'HEROIC','Complexity':'coL','Ninjas in Pyjamas':'NIP',
  'Eternal Fire':'永火','fnatic':'橙黑','Rare Atom':'RA','Lynn Vision':'LVG',
  'Aurora':'欧若拉','RED Canids':'红犬','GamerLegion':'GL','PARIVISION':'PV',
}

function cn(name?: string) { return nicknames[name ?? ''] ?? '' }

export default function Matches() {
  const [tab, setTab] = useState<Tab>('today')
  const [data, setData] = useState<any>(null)
  const [filter, setFilter] = useState('')

  useEffect(() => {
    setData(null)
    if (tab === 'today')    api.todayMatches().then(setData)
    else if (tab === 'upcoming') api.upcomingMatches({ team: filter, limit: '30' }).then(setData)
    else api.results({ team: filter, limit: '30' }).then(setData)
  }, [tab, filter])

  const items: any[] = data?.items ?? []

  return (
    <div className="anim-in space-y-8">
      {/* ---- Tab bar ---- */}
      <div className="flex items-center gap-4">
        {tabs.map(t => (
          <button key={t.key} onClick={() => setTab(t.key)}
            className={`text-[17px] font-display font-semibold tracking-wider uppercase pb-2 border-b-[3px] transition-colors ${
              tab === t.key
                ? 'text-gold border-gold'
                : 'text-text-muted border-transparent hover:text-text-secondary'
            }`}>
            {t.label}
          </button>
        ))}
        <div className="flex-1" />
        <input
          placeholder="筛选队伍..."
          value={filter}
          onChange={e => setFilter(e.target.value)}
          className="w-56 bg-card border border-border text-text text-[15px] px-4 py-2 rounded-lg focus:outline-none focus:border-gold placeholder:text-text-muted"
        />
      </div>

      {/* ---- Match Cards ---- */}
      <div className="space-y-3">
        {items.length === 0 && (
          <div className="text-text-muted text-[16px] py-24 text-center bg-card border border-border rounded-lg">
            {data ? '暂无比赛数据' : '加载中...'}
          </div>
        )}

        {items.map((m, i) => {
          const live = !m.score && m.scheduled_at && Date.now() - new Date(m.scheduled_at).getTime() > -7200000 && Date.now() - new Date(m.scheduled_at).getTime() < 7200000

          return (
            <div key={i}
              className="bg-card border border-border rounded-lg px-6 py-5 flex items-center gap-6 anim-in hover:border-gold/30 transition-colors"
              style={{ animationDelay: `${i * 35}ms` }}>

              {/* Left team */}
              <div className="flex-1 text-right min-w-0">
                <div className="text-[20px] font-display font-semibold text-text leading-tight tracking-wide truncate">
                  {m.team1 ?? 'TBD'}
                </div>
                <div className="text-[14px] text-text-muted mt-0.5">{cn(m.team1)}</div>
              </div>

              {/* Score / Time */}
              <div className="flex flex-col items-center shrink-0 w-[140px]">
                {m.score ? (
                  <span className="text-[48px] font-display font-bold text-text leading-none tracking-tight">
                    {m.score}
                  </span>
                ) : (
                  <span className="text-[48px] font-display font-bold text-gold leading-none tracking-tight">
                    {m.scheduled_at ? m.scheduled_at.slice(11, 16) : '—:—'}
                  </span>
                )}
                {live && !m.score && (
                  <span className="text-[13px] font-bold text-red mt-1 px-3 py-0.5 border border-red/30 bg-red-muted rounded-full tracking-widest uppercase">
                    LIVE
                  </span>
                )}
                {m.best_of && <span className="text-[13px] text-text-muted mt-1">{m.best_of}</span>}
              </div>

              {/* Right team */}
              <div className="flex-1 min-w-0">
                <div className="text-[20px] font-display font-semibold text-text leading-tight tracking-wide truncate">
                  {m.team2 ?? 'TBD'}
                </div>
                <div className="text-[14px] text-text-muted mt-0.5">{cn(m.team2)}</div>
              </div>

              {/* Event tag */}
              {m.event && (
                <span className="shrink-0 text-[13px] text-gold bg-gold-muted border border-gold/20 rounded-full px-4 py-1.5 font-medium">
                  {m.event}
                </span>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
