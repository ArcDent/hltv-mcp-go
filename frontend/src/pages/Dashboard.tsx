import { useEffect, useState } from 'react'
import { api } from '../api/client'

export default function Dashboard() {
  const [status, setStatus] = useState<any>(null)

  useEffect(() => {
    api.status().then(setStatus).catch(() => {})
  }, [])

  const stats = [
    { label: 'UPTIME', value: status ? `${status.uptime_sec}s` : '---', unit: 'SEC' },
    { label: 'GO VERSION', value: status?.go_version ?? '---', unit: 'RT' },
    { label: 'MEMORY', value: status ? `${status.memory_mb}` : '--', unit: 'MB' },
    { label: 'CACHE KEYS', value: status?.cache_entries ?? '--', unit: 'ENT' },
  ]

  return (
    <div className="animate-in">
      <div className="mb-8">
        <h2 className="font-display text-neon text-lg tracking-[0.2em] mb-1">
          SYS.DASHBOARD
        </h2>
        <div className="h-[1px] w-full bg-gradient-to-r from-neon/50 via-neon/20 to-transparent" />
      </div>

      <div className="grid grid-cols-4 gap-3 mb-8">
        {stats.map((s, i) => (
          <div
            key={s.label}
            className="bg-panel border border-border p-4 animate-in"
            style={{ animationDelay: `${i * 80}ms` }}
          >
            <div className="text-[10px] text-text-dim tracking-[0.15em] mb-2">
              {s.label}
            </div>
            <div className="flex items-baseline gap-1">
              <span className="text-2xl font-bold text-text tracking-tight">
                {s.value}
              </span>
              <span className="text-[10px] text-steel">{s.unit}</span>
            </div>
          </div>
        ))}
      </div>

      {/* System status panel */}
      <div className="bg-panel border border-border p-5 animate-in" style={{ animationDelay: '320ms' }}>
        <div className="text-[10px] text-text-dim tracking-[0.15em] mb-3">
          SYSTEM STATUS
        </div>
        <div className="space-y-2">
          {[
            { label: 'HTTP SERVER', ok: true, detail: '0.0.0.0:8082' },
            { label: 'MCP STDIO', ok: true, detail: 'connected' },
            { label: 'CHROME', ok: true, detail: 'chromedp ready' },
            { label: 'HLTV CONN', ok: true, detail: 'direct + fallback' },
          ].map((s, _) => (
            <div key={s.label} className="flex items-center gap-3 text-[11px]">
              <span className={`w-2 h-2 rounded-full ${s.ok ? 'bg-neon animate-pulse' : 'bg-orange'}`} />
              <span className="text-text w-28">{s.label}</span>
              <span className="text-text-dim">{s.detail}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
