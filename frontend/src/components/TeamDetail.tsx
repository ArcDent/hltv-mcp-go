import { useEffect, useState } from 'react'
import Modal from './Modal'
import PlayerDetail from './PlayerDetail'
import useNicknames from '../hooks/useNicknames'

type TeamData = {
  profile: { id: number; name: string; slug: string; country?: string; region?: string }
  ranking: { world_rank: number; points: number }
  stats: { wins: number; losses: number; draws: number; win_rate: string; recent_form: string }
  achievements?: { label: string; count: number; tier: string }[]
  roster?: { id: number; name: string; slug: string; rating: number; country?: string }[]
  recent_matches?: { team1?: string; team2?: string; opponent?: string; score?: string; result: string; event?: string; played_at?: string; scheduled_at?: string; map_text?: string; best_of?: string }[]
  highlights?: { win_rate: string; win_streak: number; recent_matches?: { opponent: string; result: string }[] }
}

export default function TeamDetail({ id, onClose }: { id: number; onClose: () => void }) {
  const [data, setData] = useState<TeamData | null>(null)
  const [loading, setLoading] = useState(true)
  const [selectedPlayerId, setSelectedPlayerId] = useState<number | null>(null)
  const { teamNicknames, playerNicknames, saveTeamNickname, savePlayerNickname } = useNicknames()
  const [editingTeamNick, setEditingTeamNick] = useState(false)
  const [editingPlayerId, setEditingPlayerId] = useState<number | null>(null)

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
  const hl = data?.highlights

  const cnName = teamNicknames[p?.name ?? '']

  return (
    <Modal onClose={onClose} width={840} maxHeight="90vh">

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
                  {editingTeamNick ? (
                    <input
                      autoFocus
                      defaultValue={cnName}
                      style={{padding:'2px 8px',borderRadius:4,fontSize:11,background:'var(--input-bg)',border:'1px solid var(--gold)',color:'var(--text)',width:80,outline:'none'}}
                      onKeyDown={e => {
                        if (e.key === 'Enter') { saveTeamNickname(p?.name ?? '', (e.target as HTMLInputElement).value); setEditingTeamNick(false) }
                        if (e.key === 'Escape') setEditingTeamNick(false)
                      }}
                      onBlur={e => { saveTeamNickname(p?.name ?? '', e.target.value); setEditingTeamNick(false) }}
                    />
                  ) : (
                    <span style={{padding:'2px 10px',borderRadius:4,fontSize:11,background:'var(--gold-dim)',color:'var(--gold)',fontWeight:600,display:'inline-flex',alignItems:'center',gap:4}}>
                      {cnName || '无简称'}
                      <span onClick={() => setEditingTeamNick(true)} style={{cursor:'pointer',opacity:0.6,fontSize:10}} title="编辑简称">✏️</span>
                    </span>
                  )}
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
            <div style={{display:'flex',marginBottom:18,border:'1px solid var(--border)',borderRadius:'var(--radius-sm)',overflow:'hidden'}}>
              <div style={{flex:1,textAlign:'center',padding:'10px 6px',background:'var(--input-bg)'}}>
                <div style={{fontFamily:'var(--font-display)',fontSize:22,fontWeight:700,color:'var(--green)',lineHeight:1}}>{stats?.wins ?? 0}</div>
                <div style={{fontSize:10,color:'var(--text-muted)',marginTop:3,textTransform:'uppercase',letterSpacing:'0.05em'}}>胜</div>
              </div>
              <div style={{flex:1,textAlign:'center',padding:'10px 6px',background:'var(--input-bg)',borderLeft:'1px solid var(--border)'}}>
                <div style={{fontFamily:'var(--font-display)',fontSize:22,fontWeight:700,color:'var(--red)',lineHeight:1}}>{stats?.losses ?? 0}</div>
                <div style={{fontSize:10,color:'var(--text-muted)',marginTop:3,textTransform:'uppercase',letterSpacing:'0.05em'}}>负</div>
              </div>
              <div style={{flex:1,textAlign:'center',padding:'10px 6px',background:'var(--input-bg)',borderLeft:'1px solid var(--border)'}}>
                <div style={{fontFamily:'var(--font-display)',fontSize:22,fontWeight:700,color:'var(--text)',lineHeight:1}}>{hl?.win_rate || stats?.win_rate || '—'}</div>
                <div style={{fontSize:10,color:'var(--text-muted)',marginTop:3,textTransform:'uppercase',letterSpacing:'0.05em'}}>胜率</div>
              </div>
              <div style={{flex:1,textAlign:'center',padding:'10px 6px',background:'var(--input-bg)',borderLeft:'1px solid var(--border)'}}>
                <div style={{fontFamily:'var(--font-display)',fontSize:22,fontWeight:700,color:'var(--gold)',lineHeight:1}}>{hl?.win_streak ?? '—'}</div>
                <div style={{fontSize:10,color:'var(--text-muted)',marginTop:3,textTransform:'uppercase',letterSpacing:'0.05em'}}>连胜</div>
              </div>
            </div>

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

              {/* Left: Recent matches from highlights */}
              <div>
                <div style={{fontFamily:'var(--font-display)',fontSize:14,fontWeight:600,color:'var(--gold)',letterSpacing:'0.05em',textTransform:'uppercase',marginBottom:10,paddingBottom:6,borderBottom:'1px solid var(--border)',display:'flex',justifyContent:'space-between'}}>
                  近期战绩
                  <span style={{fontSize:11,fontWeight:400,color:'var(--text-muted)',fontFamily:'var(--font-body)',textTransform:'none',letterSpacing:0}}>{(hl?.recent_matches || matches).length} 场</span>
                </div>
                {(hl?.recent_matches?.length || 0) === 0 && matches.length === 0 && <div style={{fontSize:12,color:'var(--text-muted)',textAlign:'center',padding:'20px 0'}}>暂无数据</div>}
                {(hl?.recent_matches || matches).map((m: any, i: number) => (
                  <div key={i} style={{display:'flex',alignItems:'center',gap:10,padding:'7px 0',borderBottom:i<(hl?.recent_matches || matches).length-1?'1px solid rgba(128,128,128,0.06)':'none',fontSize:12}}>
                    <span style={{minWidth:26,textAlign:'center',fontSize:10,fontWeight:700,fontFamily:'var(--font-mono)',padding:'2px 0',borderRadius:3,
                      color:m.result==='won'||m.result==='win'?'var(--green)':m.result==='lost'||m.result==='loss'?'var(--red)':'var(--text-muted)',
                      background:m.result==='won'||m.result==='win'?'rgba(0,200,83,0.1)':m.result==='lost'||m.result==='loss'?'rgba(255,82,82,0.1)':'var(--input-bg)'}}>
                      {m.result==='won'||m.result==='win'?'W':m.result==='lost'||m.result==='loss'?'L':'—'}
                    </span>
                    <span style={{flex:1}}><b style={{fontWeight:600}}>{p.name}</b><span style={{margin:'0 6px',color:'var(--text-muted)'}}>vs</span><span style={{fontWeight:600}}>{m.opponent || m.team2 || '待定'}</span></span>
                    <span style={{fontSize:10,color:'var(--text-muted)',maxWidth:160,overflow:'hidden',textOverflow:'ellipsis',whiteSpace:'nowrap'}}>{m.event || ''}</span>
                    <span style={{fontSize:10,color:'var(--text-muted)',minWidth:48,textAlign:'right'}}>{(m.played_at || m.scheduled_at || '').slice(5,10)}</span>
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
                      {(playerNicknames[pl.name] || editingPlayerId === pl.id) ? (
                        editingPlayerId === pl.id ? (
                          <input
                            autoFocus
                            defaultValue={playerNicknames[pl.name] ?? ''}
                            style={{fontSize:11,background:'var(--input-bg)',border:'1px solid var(--gold)',borderRadius:3,padding:'1px 4px',color:'var(--text)',width:60,outline:'none',marginLeft:4}}
                            onKeyDown={e => {
                              if (e.key === 'Enter') { savePlayerNickname(pl.name, (e.target as HTMLInputElement).value); setEditingPlayerId(null) }
                              if (e.key === 'Escape') setEditingPlayerId(null)
                            }}
                            onBlur={e => { savePlayerNickname(pl.name, e.target.value); setEditingPlayerId(null) }}
                            onClick={e => e.stopPropagation()}
                          />
                        ) : (
                          <span style={{fontSize:11,color:'var(--text-muted)',marginLeft:4,fontWeight:400}}>
                            {playerNicknames[pl.name]}
                            <span onClick={e => { e.stopPropagation(); setEditingPlayerId(pl.id) }} style={{cursor:'pointer',opacity:0.4,fontSize:9,marginLeft:2}} title="编辑简称">✏️</span>
                          </span>
                        )
                      ) : !playerNicknames[pl.name] && editingPlayerId !== pl.id ? (
                        <span onClick={e => { e.stopPropagation(); setEditingPlayerId(pl.id) }} style={{cursor:'pointer',opacity:0.4,fontSize:9,marginLeft:4}} title="添加简称">+</span>
                      ) : null}
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
      {selectedPlayerId !== null && <PlayerDetail id={selectedPlayerId} onClose={() => setSelectedPlayerId(null)} />}
    </Modal>
  )
}
