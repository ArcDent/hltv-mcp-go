import { useState } from 'react'
import PlayerDetail from './PlayerDetail'

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

const teamNicknames: Record<string, string> = {
  'Vitality':'小蜜蜂','Spirit':'绿龙','Team Spirit':'绿龙','Natus Vincere':'天生赢家',
  'NAVI':'天生赢家','FaZe':'FaZe','G2':'武士','MOUZ':'老鼠','Falcons':'猎鹰',
  'Astralis':'A队','Virtus.pro':'VP','Team Liquid':'液体','FURIA':'黑豹',
  'The MongolZ':'蒙古队','TYLOO':'天禄','3DMAX':'3DMAX','paiN':'paiN',
  'HEROIC':'HEROIC','Complexity':'coL','Ninjas in Pyjamas':'NIP',
  'Eternal Fire':'永火','fnatic':'橙黑','Rare Atom':'RA','Lynn Vision':'LVG',
  'Aurora':'欧若拉','RED Canids':'红犬','GamerLegion':'GL','PARIVISION':'PV',
}

const playerNicknames: Record<string, string> = {
  'ZywOo': '载物', 's1mple': '简单', 'm0NESY': '小孩', 'donk': '洞克',
  'NiKo': '尼扣', 'dev1ce': '设备', 'ropz': '肉铺子', 'karrigan': '卡里根',
  'apEX': 'A队长', 'flameZ': '火焰', 'Spinx': '斯宾克斯', 'mezii': '梅子',
  'jL': '杰L', 'Aleksib': '阿列克西', 'b1t': '比特', 'iM': '爱慕',
  'w0nderful': '神奇', 'broky': '布洛基', 'frozen': '冰封人', 'Twistzz': '总监',
  'huNter-': '猎人', 'jks': '杰克S', 'NAF': '纳夫', 'YEKINDAR': '叶金达',
  'cadiaN': '卡点', 'stavn': '斯塔文', 'jabbi': '贾比', 'TeSeS': '特塞斯',
  'EliGE': '一粒鸡', 'Magisk': '魔法男孩', 'dupreeh': '杜普瑞', 'Xyp9x': '九爷',
  'gla1ve': '格莱乌', 'electroNic': '电子哥', 'Perfecto': '完美', 'Boombl4': '胖球',
  'sh1ro': '细弱', 'Ax1Le': '阿列克斯', 'Hobbit': '霍比特', 'KSCERATO': '卡斯赛拉托',
  'yuurih': '优日', 'arT': '阿特', 'FalleN': '教父',
}

export default function SearchableList({ type, placeholder, emptyHint, apiSearch }: Props) {
  const [q, setQ] = useState('')
  const [list, setList] = useState<any[] | null>(null)
  const [loading, setLoading] = useState(false)
  const [selectedId, setSelectedId] = useState<number | null>(null)

  const search = async () => {
    if (!q.trim()) return
    setLoading(true)
    try { const r = await apiSearch(q); setList(r?.items ?? []) } catch { setList([]) }
    setLoading(false)
  }

  return (
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
        <div key={i} className="anim-in" onClick={() => type === 'player' && item.id && setSelectedId(item.id)}
          style={{ ...cardStyle, animationDelay: `${i * 35}ms`, cursor: type === 'player' ? 'pointer' : 'default' }}>
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
  )
}
