import { useEffect, useState } from 'react'
import { api } from '../api/client'

export default function Cache() {
  const [stats, setStats] = useState<any>(null)
  const [cleared, setCleared] = useState(false)

  const refresh = () => { api.cacheStats().then(setStats).catch(() => {}) }
  useEffect(refresh, [])

  const clear = async () => {
    await api.clearCache(); setCleared(true); refresh()
    setTimeout(() => setCleared(false), 2500)
  }

  const cards = [
    { label: '缓存条目', value: stats?.entries ?? '—' },
    { label: '命中次数', value: stats?.hits    ?? '—' },
    { label: '未命中',   value: stats?.misses  ?? '—' },
  ]

  return (
    <div className="anim-in space-y-10">
      <div className="grid grid-cols-3 gap-4">
        {cards.map((c, i) => (
          <div key={c.label} className="bg-card border border-border rounded-lg px-6 py-5"
            style={{ animationDelay: `${i * 80}ms` }}>
            <div className="text-text-muted text-[14px] mb-2">{c.label}</div>
            <div className="text-[40px] font-display font-bold text-text leading-none">{c.value}</div>
          </div>
        ))}
      </div>

      <div className="flex items-center gap-5">
        <button onClick={clear}
          className="px-7 py-3 bg-red-muted border border-red/30 text-red text-[16px] font-display font-semibold tracking-wider rounded-lg uppercase hover:bg-red/20 transition-colors">
          清除全部缓存
        </button>
        <button onClick={refresh}
          className="px-7 py-3 bg-card border border-border text-text-secondary text-[16px] font-display font-semibold tracking-wider rounded-lg uppercase hover:border-gold/30 hover:text-text transition-colors">
          刷新
        </button>
        {cleared && (
          <span className="text-[16px] font-medium text-gold anim-in">✓ 缓存已清除</span>
        )}
      </div>
    </div>
  )
}
