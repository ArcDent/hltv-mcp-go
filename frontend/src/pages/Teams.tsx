import { useState } from 'react'
import { api } from '../api/client'

export default function Teams() {
  const [q, setQ] = useState('')
  const [list, setList] = useState<any[] | null>(null)
  const [loading, setLoading] = useState(false)

  const search = async () => {
    if (!q.trim()) return
    setLoading(true)
    try { const r = await api.search(q, 'team'); setList(r?.items ?? []) }
    catch { setList([]) }
    setLoading(false)
  }

  return (
    <div className="anim-in space-y-8">
      <div className="flex gap-4">
        <input
          placeholder="搜索队伍 — 支持英文 / 中文 / 别名（如 Spirit、绿龙、小蜜蜂）"
          value={q}
          onChange={e => setQ(e.target.value)}
          onKeyDown={e => e.key === 'Enter' && search()}
          className="flex-1 bg-card border border-border text-text text-[17px] px-5 py-3.5 rounded-lg focus:outline-none focus:border-gold placeholder:text-text-muted"
        />
        <button onClick={search} disabled={loading}
          className="px-8 py-3.5 bg-gold text-bg text-[17px] font-display font-semibold tracking-wider rounded-lg hover:brightness-105 transition-all disabled:opacity-30 uppercase">
          {loading ? '搜索中' : '搜索'}
        </button>
      </div>

      <div className="space-y-[1px]">
        {list === null && (
          <div className="text-text-muted text-[16px] py-24 text-center bg-card border border-border rounded-lg">
            输入队名开始搜索
          </div>
        )}
        {list?.length === 0 && (
          <div className="text-text-muted text-[16px] py-24 text-center bg-card border border-border rounded-lg">
            无匹配结果
          </div>
        )}
        {list?.map((t, i) => (
          <div key={i}
            className="anim-in bg-card border border-border rounded-lg px-6 py-4 flex items-center gap-5 hover:border-gold/30 transition-colors"
            style={{ animationDelay: `${i * 35}ms` }}>
            <span className="text-gold font-mono font-bold text-[16px] w-8 shrink-0">
              {String(i + 1).padStart(2, '0')}
            </span>
            <span className="flex-1 text-[18px] font-display font-semibold tracking-wide">{t.name}</span>
            <span className="text-[14px] text-text-muted bg-surface border border-border rounded-md px-3 py-1 font-mono">
              ID {t.id ?? '—'}
            </span>
            {t.slug && <span className="text-[14px] text-text-muted">{t.slug}</span>}
          </div>
        ))}
      </div>
    </div>
  )
}
