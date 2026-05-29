import { useEffect, useState } from 'react'
import useNicknames from '../hooks/useNicknames'
import Modal from './Modal'

type PlayerData = {
  profile: { id: number; name: string; real_name?: string; slug: string; country?: string; age?: number; team?: string; prize_money?: string }
  rating: { value: number; maps: number }
  abilities: { key: string; label_en: string; label_zh: string; value: number; max: number; format?: string }[]
  career: { rating?: number; matches?: number; win_rate?: string; kd?: number; headshot_pct?: string; win_streak?: number }
  top20_ranks?: Record<string, number>
  honors?: { label: string; value: number }[]
  recent_matches?: { date: string; team: string; opponent: string; score: string; result: string; rating: number; kills: number; deaths: number; event: string }[]
}

export default function PlayerDetail({ id, onClose }: { id: number; onClose: () => void }) {
  const [data, setData] = useState<PlayerData | null>(null)
  const [loading, setLoading] = useState(true)
  const { playerNicknames, savePlayerNickname } = useNicknames()
  const [editingNick, setEditingNick] = useState(false)

  useEffect(() => {
    setLoading(true)
    fetch(`/api/players/${id}`).then(r => r.json()).then(d => {
      setData(d.data ?? null); setLoading(false)
    }).catch(() => setLoading(false))
  }, [id])

  const p = data?.profile
  const abilities = data?.abilities ?? []
  const top20 = data?.top20_ranks ? Object.entries(data.top20_ranks).sort((a,b) => Number(b[0])-Number(a[0])) : []

  return (
    <Modal onClose={onClose} width={580} maxHeight="90vh">

        {loading && <div style={{textAlign:'center',padding:60,color:'var(--text-muted)'}}>加载中...</div>}
        {!loading && !p && <div style={{textAlign:'center',padding:60,color:'var(--text-muted)'}}>详情暂时不可用</div>}

        {!loading && p && (
          <>
            <div style={{display:'flex',gap:14,marginBottom:14}}>
              <div style={{width:56,height:56,borderRadius:'50%',background:'linear-gradient(135deg,var(--gold),#c48a0a)',display:'flex',alignItems:'center',justifyContent:'center',fontSize:24,color:'#fff',fontWeight:700,fontFamily:'var(--font-display)',flexShrink:0}}>
                {p.name.charAt(0)}
              </div>
              <div style={{flex:1}}>
                <div style={{fontSize:22,fontWeight:700,color:'var(--text)',lineHeight:1.2}}>{p.name}</div>
                {p.real_name ? <div style={{fontSize:13,color:'var(--text-muted)'}}>{p.real_name}</div> : <div style={{fontSize:13,color:'var(--text-muted)'}}>暂无</div>}
                <div style={{display:'flex',flexWrap:'wrap',gap:6,marginTop:4,alignItems:'center'}}>
                  {editingNick ? (
                    <input
                      autoFocus
                      defaultValue={playerNicknames[p.name] ?? ''}
                      style={{fontSize:12,background:'var(--input-bg)',border:'1px solid var(--gold)',borderRadius:3,padding:'2px 6px',color:'var(--text)',width:100,outline:'none'}}
                      onKeyDown={e => {
                        if (e.key === 'Enter') { savePlayerNickname(p.name, (e.target as HTMLInputElement).value); setEditingNick(false) }
                        if (e.key === 'Escape') setEditingNick(false)
                      }}
                      onBlur={e => { savePlayerNickname(p.name, e.target.value); setEditingNick(false) }}
                    />
                  ) : (
                    <>
                      {playerNicknames[p.name] ? (
                        <span onClick={() => setEditingNick(true)} style={{padding:'2px 8px',borderRadius:4,fontSize:11,background:'var(--gold-dim)',color:'var(--gold)',fontWeight:600,cursor:'pointer'}} title="点击编辑简称">
                          {playerNicknames[p.name]}
                        </span>
                      ) : (
                        <span onClick={() => setEditingNick(true)} style={{cursor:'pointer',fontSize:11,color:'var(--text-muted)',opacity:0.5}} title="添加简称">+ 添加简称</span>
                      )}
                    </>
                  )}
                  {p.country ? <span style={{padding:'2px 8px',background:'var(--input-bg)',borderRadius:4,fontSize:11,color:'var(--text-secondary)'}}>{p.country}</span> : <span style={{padding:'2px 8px',background:'var(--input-bg)',borderRadius:4,fontSize:11,color:'var(--text-muted)'}}>未知国籍</span>}
                  {p.age ? <span style={{padding:'2px 8px',background:'var(--input-bg)',borderRadius:4,fontSize:11,color:'var(--text-secondary)'}}>Age {p.age}</span> : null}
                  <span style={{padding:'2px 8px',background: p.team ? 'var(--gold-dim)' : 'var(--input-bg)',borderRadius:4,fontSize:11,color: p.team ? 'var(--gold)' : 'var(--text-muted)',fontWeight: p.team ? 600 : 400}}>{p.team || '暂无队伍'}</span>
                  {p.prize_money && <span style={{padding:'2px 8px',background:'var(--input-bg)',borderRadius:4,fontSize:11,color:'var(--text-secondary)'}}>{p.prize_money}</span>}
                </div>
              </div>
            </div>

            {top20.length > 0 && (
              <div style={{display:'flex',gap:4,flexWrap:'wrap',justifyContent:'center',marginBottom:16}}>
                {top20.map(([year, rank]) => (
                  <span key={year} style={{fontSize:11,fontWeight:700,padding:'2px 7px',borderRadius:10,
                    background: rank===1?'linear-gradient(135deg,#f0c040,#c48a0a)':rank===2?'#e0e0d8':'rgba(196,138,10,0.12)',
                    color: rank===1?'#fff':rank===2?'#6b7280':'var(--gold)'}}>{year} #{rank}</span>
                ))}
              </div>
            )}

            <div style={{fontFamily:'var(--font-display)',fontSize:16,fontWeight:600,color:'var(--gold)',letterSpacing:'0.06em',textTransform:'uppercase',marginBottom:12,paddingBottom:8,borderBottom:'1px solid var(--border)'}}>
              能力评分 <span style={{fontSize:13,fontWeight:400,color:'var(--text-muted)'}}>近 3 月 · {data.rating.maps} maps</span>
            </div>

            <div style={{display:'flex',justifyContent:'center',alignItems:'center',gap:24,marginBottom:16}}>
              <svg width="140" height="140" viewBox="0 0 140 140">
                {[66,48,30,12].map(r => <circle key={r} cx="70" cy="70" r={r} fill="none" stroke="var(--border)" strokeWidth="1"/>)}
                {[0,45,90,135].map(a => (
                  <line key={a} x1={70+66*Math.cos(a*Math.PI/180)} y1={70+66*Math.sin(a*Math.PI/180)} x2={70-66*Math.cos(a*Math.PI/180)} y2={70-66*Math.sin(a*Math.PI/180)} stroke="var(--border)" strokeWidth="0.5"/>
                ))}
                {abilities.length >= 1 && (
                  <polygon
                    points={abilities.slice(0,8).map((ab,i) => {
                      const angle = (i*45-90)*Math.PI/180
                      const v = ab.format === 'decimal' ? ab.value/2*66 : (ab.value||0)/100*66
                      return `${(70+v*Math.cos(angle)).toFixed(0)},${(70+v*Math.sin(angle)).toFixed(0)}`
                    }).join(' ')}
                    fill="rgba(196,138,10,0.12)" stroke="var(--gold)" strokeWidth="1.5"
                  />
                )}
              </svg>
              <div style={{display:'flex',flexDirection:'column',gap:3,fontSize:11,color:'var(--text-secondary)'}}>
                {abilities.map(ab => (
                  <div key={ab.key} style={{display:'flex',alignItems:'center',gap:6,opacity:ab.value===0&&ab.format!=='decimal'?0.4:1}}>
                    <span style={{width:7,height:7,borderRadius:2,background:ab.value>0||ab.format==='decimal'?'var(--gold)':'var(--border)',flexShrink:0}}/>
                    <span style={{minWidth:120}}>{ab.label_en} ({ab.label_zh})</span>
                    <b style={{color:ab.value>0||ab.format==='decimal'?'var(--text)':'var(--text-muted)',fontFamily:'var(--font-mono)',fontSize:12}}>
                      {ab.format==='decimal'?ab.value.toFixed(2):ab.value>0?`${ab.value}/${ab.max}`:'—'}
                    </b>
                  </div>
                ))}
              </div>
            </div>

            {(data.career.rating || data.career.matches) && (
              <div style={{display:'flex',alignItems:'center',justifyContent:'center',gap:16,marginBottom:14,fontSize:13,color:'var(--text-secondary)'}}>
                {data.career.rating && <><span><span style={{fontFamily:'var(--font-display)',fontSize:20,fontWeight:700,color:'var(--text)'}}>{data.career.rating}</span> 生涯 Rating</span><span style={{color:'var(--border)'}}>|</span></>}
                {(data.career.matches ?? 0) > 0 && <><span style={{fontFamily:'var(--font-display)',fontSize:20,fontWeight:700,color:'var(--text)'}}>{data.career.matches}</span> 比赛</>}
                {data.career.win_rate && <><span style={{color:'var(--border)'}}>|</span><span style={{fontFamily:'var(--font-display)',fontSize:20,fontWeight:700,color:'var(--text)'}}>{data.career.win_rate}</span> 胜率</>}
                {(data.career.kd ?? 0) > 0 && <><span style={{color:'var(--border)'}}>|</span><span style={{fontFamily:'var(--font-display)',fontSize:20,fontWeight:700,color:'var(--text)'}}>{data.career.kd}</span> K/D</>}
              </div>
            )}

            {data.honors && data.honors.length > 0 && (
              <div style={{display:'flex',gap:6,flexWrap:'wrap',justifyContent:'center',marginBottom:14}}>
                {data.honors.map(h => (
                  <span key={h.label} style={{fontSize:11,padding:'2px 10px',borderRadius:10,background:'rgba(196,138,10,0.06)',color:'var(--gold)',fontWeight:500}}>{h.label} {h.value}×</span>
                ))}
                {data.career.headshot_pct && <span style={{fontSize:11,padding:'2px 10px',borderRadius:10,background:'rgba(196,138,10,0.06)',color:'var(--gold)',fontWeight:500}}>爆头率 {data.career.headshot_pct}</span>}
                {(data.career.win_streak ?? 0) > 0 && <span style={{fontSize:11,padding:'2px 10px',borderRadius:10,background:'rgba(196,138,10,0.06)',color:'var(--gold)',fontWeight:500}}>{data.career.win_streak} 连胜</span>}
              </div>
            )}

            {data.recent_matches && data.recent_matches.length > 0 && (
              <>
                <div style={{fontFamily:'var(--font-display)',fontSize:16,fontWeight:600,color:'var(--gold)',letterSpacing:'0.06em',textTransform:'uppercase',marginBottom:10,paddingBottom:8,borderBottom:'1px solid var(--border)'}}>近期比赛</div>
                {data.recent_matches!.map((m,i) => (
                  <div key={i} style={{display:'flex',alignItems:'center',gap:10,padding:'8px 0',borderBottom:i<data.recent_matches!.length-1?'1px solid rgba(0,0,0,0.04)':'none',fontSize:12}}>
                    <span style={{minWidth:24,textAlign:'center',fontSize:10,fontWeight:700,fontFamily:'var(--font-mono)',
                      color:m.result==='win'?'var(--green)':m.result==='loss'?'var(--red)':'var(--text-muted)',
                      background:m.result==='win'?'rgba(0,200,83,0.1)':m.result==='loss'?'rgba(255,82,82,0.1)':'var(--input-bg)',
                      padding:'2px 0',borderRadius:3}}>
                      {m.result==='win'?'W':m.result==='loss'?'L':'—'}
                    </span>
                    <span style={{flex:1,minWidth:0}}>
                      <span style={{fontWeight:600}}>{m.team || '待定'}</span> <span style={{color:'var(--text-muted)'}}>vs</span> {m.opponent || '待定'}
                      <div style={{fontSize:10,color:'var(--text-muted)',whiteSpace:'nowrap',overflow:'hidden',textOverflow:'ellipsis'}}>{m.event}</div>
                    </span>
                    {m.result !== 'scheduled' && m.score && <span style={{fontFamily:'var(--font-mono)',fontSize:11,color:'var(--text-secondary)',whiteSpace:'nowrap'}}>{m.score}</span>}
                    <span style={{fontSize:10,color:'var(--text-muted)',minWidth:52,textAlign:'right'}}>{m.date}</span>
                  </div>
                ))}
              </>
            )}
          </>
        )}
      </Modal>
    )
}
