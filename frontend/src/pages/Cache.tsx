import { useEffect, useState } from 'react'
import { api } from '../api/client'

export default function Cache() {
  const [stats, setStats] = useState<any>(null)
  const [cleared, setCleared] = useState(false)

  const refresh = () => { api.cacheStats().then(setStats).catch(() => {}) }
  useEffect(refresh, [])

  const handleClear = async () => {
    await api.clearCache(); setCleared(true); refresh()
    setTimeout(() => setCleared(false), 2500)
  }

  return (
    <div className="animate-in">
      <h2 className="text-[16px] font-semibold text-gold mb-4 pb-2 border-b border-border tracking-wide">
        ◫ 缓存管理
      </h2>

      <div className="grid grid-cols-3 gap-3 mb-6">
        {[
          { label: '缓存条目', value: stats?.entries ?? '--' },
          { label: '命中', value: stats?.hits ?? '--' },
          { label: '未命中', value: stats?.misses ?? '--' },
        ].map((s, i) => (
          <div key={s.label} className="bg-panel border border-border rounded-md p-5 animate-in"
            style={{ animationDelay: `${i * 80}ms` }}>
            <div className="text-[13px] text-text-dim mb-2">{s.label}</div>
            <div className="text-[26px] font-bold text-text font-mono">{s.value}</div>
          </div>
        ))}
      </div>

      <div className="flex items-center gap-4">
        <button onClick={handleClear}
          className="px-5 py-2.5 border border-orange/40 bg-orange-dim text-orange text-[14px] font-medium rounded-md hover:bg-orange/20 transition-colors">
          清除全部缓存
        </button>
        <button onClick={refresh}
          className="px-5 py-2.5 border border-border bg-panel text-text-dim text-[14px] font-medium rounded-md hover:text-text hover:border-gold-border transition-colors">
          刷新
        </button>
        {cleared && (
          <span className="text-[14px] text-gold animate-in">缓存已清除</span>
        )}
      </div>
    </div>
  )
}
