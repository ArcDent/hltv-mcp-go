import { useEffect, useState } from 'react'
import PlayerDetail from './PlayerDetail'

type TeamData = {
  profile: { id: number; name: string; slug: string; country?: string; region?: string }
  ranking: { world_rank: number; points: number }
  stats: { wins: number; losses: number; draws: number; win_rate: string; recent_form: string }
  achievements?: { label: string; count: number; tier: string }[]
  roster?: { id: number; name: string; slug: string; rating: number; country?: string }[]
  recent_matches?: { team1?: string; team2?: string; opponent?: string; score?: string; result: string; event?: string; played_at?: string; map_text?: string; best_of?: string }[]
}

const teamNicknames: Record<string, string> = {
  'Vitality':'小蜜蜂','Spirit':'绿龙','Team Spirit':'绿龙','Natus Vincere':'天生赢家',
  'NAVI':'天生赢家','FaZe':'FaZe Clan','G2':'武士','MOUZ':'老鼠','Falcons':'猎鹰',
  'Astralis':'A队','Virtus.pro':'VP','Team Liquid':'液体','FURIA':'黑豹',
  'The MongolZ':'蒙古队','TYLOO':'天禄','3DMAX':'3DMAX','paiN':'paiN',
  'HEROIC':'HEROIC','Complexity':'coL','Ninjas in Pyjamas':'NIP',
  'Eternal Fire':'永火','fnatic':'橙黑','Rare Atom':'RA','Lynn Vision':'LVG',
  'Aurora':'欧若拉','RED Canids':'红犬','GamerLegion':'GL','PARIVISION':'PV',
}

const playerNicknames: Record<string, string> = {
  'ZywOo': '载物', 's1mple': '森破', 'm0NESY': '小孩', 'donk': '洞克',
  'NiKo': '尼扣', 'dev1ce': '设备', 'ropz': '肉铺子', 'karrigan': '大表哥',
  'apEX': 'A队长', 'flameZ': '火焰', 'Spinx': '斯宾克斯', 'mezii': '梅子',
  'jL': '杰L', 'Aleksib': '阿列克西', 'b1t': '比特', 'iM': '爱慕',
  'w0nderful': '神奇', 'broky': '布洛基', 'frozen': '寒王', 'Twistzz': '总监',
  'huNter-': '猎人', 'jks': '杰克S', 'NAF': '纳夫', 'YEKINDAR': '叶金达',
  'cadiaN': '卡点', 'stavn': '斯塔文', 'jabbi': '贾比', 'TeSeS': '特塞斯',
  'EliGE': '一粒鸡', 'Magisk': '魔法男孩', 'dupreeh': '杜普瑞', 'Xyp9x': '九爷',
  'gla1ve': '格莱乌', 'electroNic': '电子哥', 'Perfecto': '完美', 'Boombl4': '胖球',
  'sh1ro': '细弱', 'Ax1Le': '阿列克斯', 'Hobbit': '霍比特', 'KSCERATO': '卡斯赛拉托',
  'yuurih': '优日', 'arT': '阿特', 'FalleN': '教父',
}

export default function TeamDetail({ id, onClose }: { id: number; onClose: () => void }) {
  const [data, setData] = useState<TeamData | null>(null)
  const [loading, setLoading] = useState(true)
  const [selectedPlayerId, setSelectedPlayerId] = useState<number | null>(null)

  useEffect(() => {
    setLoading(true)
    fetch(`/api/teams/${id}`).then(r => r.json()).then(d => {
      setData(d.data ?? null); setLoading(false)
    }).catch(() => setLoading(false))
  }, [id])

  const p = data?.profile
  const rank = data?.ranking
  const stats = data?.stats
  const achievements = data?.achievements ?? []
  const roster = data?.roster ?? []
  const matches = data?.recent_matches ?? []

  const cnName = teamNicknames[p?.name ?? '']

  return (
    <div onClick={onClose} style={{position:'fixed',inset:0,zIndex:100,background:'rgba(0,0,0,0.5)',backdropFilter:'blur(4px)',display:'flex',alignItems:'center',justifyContent:'center',animation:'fadeIn 0.2s ease'}}>
      <div onClick={e => e.stopPropagation()} style={{position:'relative',background:'var(--card)',border:'1px solid var(--border)',borderRadius:'var(--radius)',width:840,maxWidth:'95vw',maxHeight:'90vh',overflowY:'auto',padding:28,boxShadow:'0 20px 60px rgba(0,0,0,0.3)',animation:'slideUp 0.25s ease'}}>

        <button onClick={onClose} style={{position:'absolute',top:14,right:14,width:30,height:30,borderRadius:'50%',border:'1px solid var(--border)',background:'var(--card)',color:'var(--text-secondary)',fontSize:16,cursor:'pointer',display:'flex',alignItems:'center',justifyContent:'center'}}>✕</button>

        {loading && <div style={{textAlign:'center',padding:60,color:'var(--text-muted)'}}>加载中...</div>}
        {!loading && !p && <div style={{textAlign:'center',padding:60,color:'var(--text-muted)'}}>详情暂时不可用</div>}

        {!loading && p && (
          <>
            {/* Header */}
            <div style={{display:'flex',gap:14,marginBottom:18}}>
              <div style={{width:60,height:60,borderRadius:12,background:'linear-gradient(135deg,var(--gold),#8b6914)',display:'flex',alignItems:'center',justifyContent:'center',fontSize:26,color:'#fff',fontWeight:700,fontFamily:'var(--font-display)',flexShrink:0}}>
                {p.name.charAt(0)}
              </div>
              <div style={{flex:1}}>
                <div style={{fontSize:24,fontWeight:700,color:'var(--text)',lineHeight:1.2}}>{p.name}</div>
                <div style={{fontSize:12,color:'var(--text-muted)',marginTop:2}}>{p.country || '—'}{roster.length > 0 ? ` · 队员 ${roster.length} 人` : ''}{p.region ? ` · ${p.region}` : ''}</div>
                <div style={{display:'flex',flexWrap:'wrap',gap:6,marginTop:6}}>
                  {p.country ? <span style={{padding:'2px 10px',borderRadius:4,fontSize:11,background:'var(--input-bg)',color:'var(--text-secondary)',fontWeight:500}}>{p.country}</span> : null}
                  {cnName ? <span style={{padding:'2px 10px',borderRadius:4,fontSize:11,background:'var(--gold-dim)',color:'var(--gold)',fontWeight:600}}>{cnName}</span> : null}
                </div>
              </div>
            </div>

            {/* Ranking + Points */}
            {rank && rank.world_rank > 0 && (
              <div style={{display:'flex',justifyContent:'center',gap:10,marginBottom:16}}>
                <span style={{display:'flex',alignItems:'center',gap:8,padding:'6px 16px',borderRadius:20,background:'linear-gradient(135deg,#f0c040,#c48a0a)',color:'#1a1d29',fontFamily:'var(--font-display)',fontSize:14,fontWeight:700,letterSpacing:'0.04em'}}>
                  {'🏆'} World #{rank.world_rank}
                </span>
                {rank.points > 0 && (
                  <span style={{padding:'6px 16px',borderRadius:20,background:'var(--input-bg)',color:'var(--text-secondary)',fontSize:12,fontWeight:500}}>
                    积分 <span style={{fontFamily:'var(--font-display)',fontSize:15,fontWeight:700,color:'var(--text)'}}>{rank.points}</span> pts
                  </span>
                )}
              </div>
            )}

            {/* Stats Bar */}
            {(stats?.wins !== undefined) && (
              <div style={{display:'flex',marginBottom:18,border:'1px solid var(--border)',borderRadius:'var(--radius-sm)',overflow:'hidden'}}>
                <div style={{flex:1,textAlign:'center',padding:'10px 6px',background:'var(--input-bg)'}}>
                  <div style={{fontFamily:'var(--font-display)',fontSize:22,fontWeight:700,color:'var(--green)',lineHeight:1}}>{stats!.wins}</div>
                  <div style={{fontSize:10,color:'var(--text-muted)',marginTop:3,textTransform:'uppercase',letterSpacing:'0.05em'}}>胜</div>
                </div>
                <div style={{flex:1,textAlign:'center',padding:'10px 6px',background:'var(--input-bg)',borderLeft:'1px solid var(--border)'}}>
                  <div style={{fontFamily:'var(--font-display)',fontSize:22,fontWeight:700,color:'var(--red)',lineHeight:1}}>{stats!.losses}</div>
                  <div style={{fontSize:10,color:'var(--text-muted)',marginTop:3,textTransform:'uppercase',letterSpacing:'0.05em'}}>负</div>
                </div>
                <div style={{flex:1,textAlign:'center',padding:'10px 6px',background:'var(--input-bg)',borderLeft:'1px solid var(--border)'}}>
                  <div style={{fontFamily:'var(--font-display)',fontSize:22,fontWeight:700,color:'var(--text)',lineHeight:1}}>{stats!.draws}</div>
                  <div style={{fontSize:10,color:'var(--text-muted)',marginTop:3,textTransform:'uppercase',letterSpacing:'0.05em'}}>平</div>
                </div>
                <div style={{flex:1,textAlign:'center',padding:'10px 6px',background:'var(--input-bg)',borderLeft:'1px solid var(--border)'}}>
                  <div style={{fontFamily:'var(--font-display)',fontSize:22,fontWeight:700,color:'var(--text)',lineHeight:1}}>{stats!.win_rate || '—'}</div>
                  <div style={{fontSize:10,color:'var(--text-muted)',marginTop:3,textTransform:'uppercase',letterSpacing:'0.05em'}}>胜率</div>
                  {stats!.recent_form && <div style={{fontFamily:'var(--font-mono)',fontSize:10,color:'var(--gold)',marginTop:2}}>近5场 {stats!.recent_form}</div>}
                </div>
              </div>
            )}

            {/* Achievements */}
            {achievements.length > 0 && (
              <div style={{display:'flex',gap:6,flexWrap:'wrap',justifyContent:'center',marginBottom:20}}>
                {achievements.map((a, i) => (
                  <span key={i} style={{
                    fontSize:11,padding:'3px 10px',borderRadius:10,fontWeight:a.tier==='major'?600:500,display:'flex',alignItems:'center',gap:3,
                    background: a.tier==='major'?'linear-gradient(135deg,rgba(240,192,64,0.15),rgba(196,138,10,0.1))':'rgba(196,138,10,0.06)',
                    color: a.tier==='major'?'#f0c040':'var(--gold)',
                  }}>
                    {a.tier==='major'?'🏆 ':''}{a.label} <span style={{fontFamily:'var(--font-mono)',fontWeight:700,opacity:0.8}}>{a.count}&times;</span>
                  </span>
                ))}
              </div>
            )}

            {/* Two columns: recent matches + roster */}
            <div style={{display:'grid',gridTemplateColumns:'1fr 1fr',gap:24}}>

              {/* Left: Recent 10 matches */}
              <div>
                <div style={{fontFamily:'var(--font-display)',fontSize:14,fontWeight:600,color:'var(--gold)',letterSpacing:'0.05em',textTransform:'uppercase',marginBottom:10,paddingBottom:6,borderBottom:'1px solid var(--border)',display:'flex',justifyContent:'space-between'}}>
                  近期战绩
                  <span style={{fontSize:11,fontWeight:400,color:'var(--text-muted)',fontFamily:'var(--font-body)',textTransform:'none',letterSpacing:0}}>{matches.length} 场</span>
                </div>
                {matches.length === 0 && <div style={{fontSize:12,color:'var(--text-muted)',textAlign:'center',padding:'20px 0'}}>暂无数据</div>}
                {matches.map((m, i) => (
                  <div key={i} style={{display:'flex',alignItems:'center',gap:10,padding:'7px 0',borderBottom:i<matches.length-1?'1px solid rgba(128,128,128,0.06)':'none',fontSize:12}}>
                    <span style={{minWidth:26,textAlign:'center',fontSize:10,fontWeight:700,fontFamily:'var(--font-mono)',padding:'2px 0',borderRadius:3,
                      color:m.result==='win'?'var(--green)':m.result==='loss'?'var(--red)':'var(--text-muted)',
                      background:m.result==='win'?'rgba(0,200,83,0.1)':m.result==='loss'?'rgba(255,82,82,0.1)':'var(--input-bg)'}}>
                      {m.result==='win'?'W':m.result==='loss'?'L':'—'}
                    </span>
                    <span style={{flex:1,minWidth:0,whiteSpace:'nowrap',overflow:'hidden',textOverflow:'ellipsis'}}>
                      <b style={{fontWeight:600}}>{p.name}</b> vs {m.opponent || m.team2 || '待定'}
                    </span>
                    {m.score && <span style={{fontFamily:'var(--font-mono)',fontSize:11,color:'var(--text-secondary)',minWidth:30,textAlign:'center'}}>{m.score}</span>}
                    <span style={{fontSize:10,color:'var(--text-muted)',maxWidth:80,whiteSpace:'nowrap',overflow:'hidden',textOverflow:'ellipsis'}}>{m.event || ''}</span>
                    <span style={{fontSize:10,color:'var(--text-muted)',minWidth:48,textAlign:'right'}}>{(m.played_at || '').slice(5,10)}</span>
                  </div>
                ))}
              </div>

              {/* Right: Roster */}
              <div>
                <div style={{fontFamily:'var(--font-display)',fontSize:14,fontWeight:600,color:'var(--gold)',letterSpacing:'0.05em',textTransform:'uppercase',marginBottom:10,paddingBottom:6,borderBottom:'1px solid var(--border)',display:'flex',justifyContent:'space-between'}}>
                  队员阵容
                  <span style={{fontSize:11,fontWeight:400,color:'var(--text-muted)',fontFamily:'var(--font-body)',textTransform:'none',letterSpacing:0}}>{roster.length} 人</span>
                </div>
                {roster.length === 0 && <div style={{fontSize:12,color:'var(--text-muted)',textAlign:'center',padding:'20px 0'}}>暂无数据</div>}
                {roster.map((pl, i) => (
                  <div key={i} onClick={() => pl.id > 0 && setSelectedPlayerId(pl.id)}
                    style={{
                      display:'flex',alignItems:'center',gap:10,padding:'7px 4px',fontSize:13,
                      borderBottom:i<roster.length-1?'1px solid rgba(128,128,128,0.06)':'none',
                      cursor: pl.id > 0 ? 'pointer' : 'default', borderRadius:4,
                    }}
                    onMouseEnter={e => { if(pl.id>0) e.currentTarget.style.background='var(--gold-dim)' }}
                    onMouseLeave={e => { e.currentTarget.style.background='transparent' }}>
                    <span style={{fontFamily:'var(--font-mono)',fontSize:11,fontWeight:700,color:'var(--text-muted)',minWidth:18}}>
                      {String(i+1).padStart(2,'0')}
                    </span>
                    <span style={{fontWeight:600,flex:1}}>
                      {pl.name}
                      {(playerNicknames[pl.name]) && <span style={{fontSize:11,color:'var(--text-muted)',marginLeft:4,fontWeight:400}}>{playerNicknames[pl.name]}</span>}
                    </span>
                    {pl.rating > 0 && <span style={{fontFamily:'var(--font-mono)',fontSize:11,color:'var(--text-secondary)',background:'var(--input-bg)',padding:'2px 7px',borderRadius:4}}>Rating {pl.rating.toFixed(2)}</span>}
                    {pl.id > 0 && <span style={{fontSize:10,color:'var(--gold)',opacity:0.5}}>→</span>}
                  </div>
                ))}
              </div>
            </div>

            <div style={{marginTop:14,textAlign:'center',fontSize:11,color:'var(--text-muted)'}}>点击队员可查看选手详情 · ESC 关闭</div>
          </>
        )}
      </div>
      {selectedPlayerId !== null && <PlayerDetail id={selectedPlayerId} onClose={() => setSelectedPlayerId(null)} />}
    </div>
  )
}
