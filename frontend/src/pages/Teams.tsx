import { useState } from 'react'
import { api } from '../api/client'

export default function Teams() {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<any[] | null>(null)
  const [loading, setLoading] = useState(false)

  const search = async () => {
    if (!query.trim()) return
    setLoading(true)
    try { const r = await api.search(query, 'team'); setResults(r?.items ?? []) }
    catch { setResults([]) }
    setLoading(false)
  }

  return (
    <div className="animate-in">
      <h2 className="text-[16px] font-semibold text-gold mb-4 pb-2 border-b border-border tracking-wide">
        ▣ 队伍搜索
      </h2>

      <div className="flex gap-3 mb-6">
        <input
          placeholder="输入队伍名（支持中英文/别名，如 Spirit、绿龙、小蜜蜂）"
          value={query} onChange={(e) => setQuery(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && search()}
          className="flex-1 bg-panel border-2 border-border text-text text-[15px] px-4 py-2.5 rounded-md focus:outline-none focus:border-gold placeholder:text-placeholder"
        />
        <button onClick={search} disabled={loading}
          className="px-6 py-2.5 bg-gold text-[#0d1117] text-[15px] font-semibold rounded-md hover:brightness-110 transition-all disabled:opacity-40">
          {loading ? '搜索中...' : '搜索'}
        </button>
      </div>

      <div className="space-y-1">
        {results === null && (
          <div className="text-text-dim text-[14px] py-16 text-center border border-border rounded-md bg-panel">
            输入队名开始搜索
          </div>
        )}
        {results?.length === 0 && (
          <div className="text-text-dim text-[14px] py-16 text-center border border-border rounded-md bg-panel">
            无匹配结果
          </div>
        )}
        {results?.map((t: any, i: number) => (
          <div key={i} className="bg-panel border border-border rounded-md p-3.5 flex items-center gap-4 animate-in hover:border-gold-border transition-colors"
            style={{ animationDelay: `${i * 35}ms` }}>
            <span className="text-gold text-[14px] font-mono font-semibold w-7">{String(i + 1).padStart(2, '0')}</span>
            <span className="flex-1 text-[16px] font-semibold text-text">{t.name}</span>
            <span className="text-[12px] text-text-dim bg-[#0d1117] border border-border rounded px-2.5 py-1 font-mono">
              ID:{t.id ?? '--'}
            </span>
            {t.slug && <span className="text-[12px] text-text-dim">{t.slug}</span>}
          </div>
        ))}
      </div>
    </div>
  )
}
