import { useEffect, useState } from 'react'
import { api } from '../api/client'

type Tab = 'today' | 'upcoming' | 'results'
const tabs: { key: Tab; label: string }[] = [
  { key: 'today', label: '今日' },
  { key: 'upcoming', label: '即将开始' },
  { key: 'results', label: '近期赛果' },
]

function isUpcomingSoon(scheduledAt: string): boolean {
  if (!scheduledAt) return false
  const t = new Date(scheduledAt).getTime()
  if (isNaN(t)) return false
  const diff = Math.abs(Date.now() - t)
  return diff < 2 * 60 * 60 * 1000
}

function getColloquial(name: string | undefined): string {
  if (!name) return ''
  const map: Record<string, string> = {
    'Vitality': '小蜜蜂', 'Team Spirit': '绿龙', 'Spirit': '绿龙',
    'Natus Vincere': '天生赢家', 'NAVI': '天生赢家', 'FaZe': 'FaZe Clan',
    'G2': '武士', 'MOUZ': '老鼠', 'Falcons': '猎鹰',
    'Astralis': 'A队', 'Virtus.pro': 'VP', 'Team Liquid': '液体',
    'FURIA': '黑豹', 'The MongolZ': '蒙古队', 'TYLOO': '天禄',
    '3DMAX': '3DMAX', 'paiN': 'paiN', 'HEROIC': 'HEROIC',
    'Complexity': 'coL', 'Ninjas in Pyjamas': 'NIP',
    'Eternal Fire': '永火', 'fnatic': '橙黑', 'Rare Atom': 'RA',
    'Lynn Vision': 'LVG', 'Aurora': '欧若拉', 'RED Canids': '红犬',
    'GamerLegion': 'GL', 'PARIVISION': 'PV',
  }
  return map[name] ?? ''
}

export default function Matches() {
  const [tab, setTab] = useState<Tab>('today')
  const [data, setData] = useState<any>(null)
  const [team, setTeam] = useState('')

  useEffect(() => {
    setData(null)
    if (tab === 'today') api.todayMatches().then(setData)
    else if (tab === 'upcoming') api.upcomingMatches({ team, limit: '20' }).then(setData)
    else api.results({ team, limit: '20' }).then(setData)
  }, [tab, team])

  const items: any[] = data?.items ?? []

  return (
    <div className="animate-in">
      <h2 className="text-[16px] font-semibold text-gold mb-4 pb-2 border-b border-border tracking-wide">
        ◆ 赛程
      </h2>

      <div className="flex gap-1 mb-6">
        {tabs.map((t) => (
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

      <div className="space-y-2">
        {items.length === 0 && (
          <div className="text-text-dim text-[14px] py-16 text-center border border-border rounded-md bg-panel">
            {data ? '暂无数据' : '加载中...'}
          </div>
        )}
        {items.map((m: any, i: number) => {
          const soon = isUpcomingSoon(m.scheduled_at)
          return (
            <div key={i}
              className="bg-panel border border-border rounded-md p-4 flex items-center gap-5 animate-in hover:border-gold-border transition-colors"
              style={{ animationDelay: `${i * 30}ms` }}>
              {m.score && !m.scheduled_at && (
                <span className="text-[12px] text-orange bg-orange-dim border border-orange/30 rounded px-2 py-0.5 font-medium shrink-0">
                  已结束
                </span>
              )}
              {soon && !m.score && (
                <span className="text-[12px] text-orange bg-orange-dim border border-orange/30 rounded px-2 py-0.5 font-medium shrink-0">
                  即将开始
                </span>
              )}

              <div className="flex-1 text-right">
                <div className="text-[16px] font-semibold text-text">{m.team1 ?? 'TBD'}</div>
                <div className="text-[12px] text-text-dim mt-0.5">{getColloquial(m.team1)}</div>
              </div>

              <div className="flex flex-col items-center min-w-[90px]">
                {m.score ? (
                  <span className="text-[28px] font-bold text-gold font-mono tracking-wider">{m.score}</span>
                ) : (
                  <span className="text-[28px] font-bold text-gold font-mono tracking-wider">
                    {m.scheduled_at ? m.scheduled_at.substring(11, 16) : '--:--'}
                  </span>
                )}
                {m.best_of && <span className="text-[11px] text-text-dim mt-0.5">{m.best_of}</span>}
              </div>

              <div className="flex-1">
                <div className="text-[16px] font-semibold text-text">{m.team2 ?? 'TBD'}</div>
                <div className="text-[12px] text-text-dim mt-0.5">{getColloquial(m.team2)}</div>
              </div>

              {m.event && (
                <span className="text-[12px] text-gold bg-gold-dim border border-gold-border rounded px-2 py-0.5 shrink-0">
                  {m.event}
                </span>
              )}
            </div>
          )
        })}
      </div>

      <div className="mt-6">
        <input
          placeholder="输入队伍名筛选..."
          value={team}
          onChange={(e) => setTeam(e.target.value)}
          className="w-64 bg-panel border-2 border-border text-text text-[15px] px-4 py-2.5 rounded-md focus:outline-none focus:border-gold placeholder:text-placeholder"
        />
      </div>
    </div>
  )
}
