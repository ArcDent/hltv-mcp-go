import { useEffect, useState } from 'react'
import { api } from '../api/client'

export default function Dashboard() {
  const [status, setStatus] = useState<any>(null)
  useEffect(() => { api.status().then(setStatus).catch(() => {}) }, [])

  const cards = [
    { label: '运行时间',       value: status ? `${status.uptime_sec} s`  : '—', unit: '' },
    { label: 'Go 版本',        value: status?.go_version ?? '—',         unit: '' },
    { label: '内存占用',       value: status ? `${status.memory_mb}`     : '—', unit: 'MB' },
    { label: '缓存条目',       value: status?.cache_entries ?? '—',      unit: '' },
  ]

  const rows = [
    { label: 'HTTP 服务',  detail: '0.0.0.0:8082' },
    { label: 'MCP 连接',   detail: 'stdio 已连接' },
    { label: 'Chrome',     detail: 'chromedp 就绪' },
    { label: '数据源',     detail: 'HTTP 直连 + chromedp 备用' },
  ]

  return (
    <div className="anim-in space-y-10">
      {/* ---- Stat Cards ---- */}
      <div className="grid grid-cols-4 gap-4">
        {cards.map((c, i) => (
          <div key={c.label} className="bg-card border border-border rounded-lg px-6 py-5"
            style={{ animationDelay: `${i * 80}ms` }}>
            <div className="text-text-muted text-[14px] mb-2">{c.label}</div>
            <div className="flex items-baseline gap-1">
              <span className="text-[40px] font-display font-bold text-text leading-none">
                {c.value}
              </span>
              <span className="text-[15px] text-text-muted">{c.unit}</span>
            </div>
          </div>
        ))}
      </div>

      {/* ---- System Status ---- */}
      <div>
        <h2 className="font-display text-gold text-[20px] uppercase tracking-widest mb-5">
          系统状态
        </h2>
        <div className="bg-card border border-border rounded-lg divide-y divide-border">
          {rows.map((r) => (
            <div key={r.label} className="flex items-center gap-4 px-6 py-4">
              <span className="w-[10px] h-[10px] rounded-full bg-green pulse-dot shrink-0" />
              <span className="text-[15px] font-medium w-28 shrink-0">{r.label}</span>
              <span className="text-[15px] text-text-secondary">{r.detail}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
