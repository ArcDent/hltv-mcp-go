import { useEffect, useState } from 'react'
import { api } from '../api/client'

export default function Dashboard() {
  const [status, setStatus] = useState<any>(null)
  useEffect(() => { api.status().then(setStatus).catch(() => {}) }, [])

  const bigs = [
    { label: '运行时间',    value: status ? `${status.uptime_sec} s` : '—',  icon: '⏱' },
    { label: 'Go 版本',     value: status?.go_version ?? '—',               icon: '●' },
    { label: '内存占用',    value: status ? `${status.memory_mb} MB`  : '—', icon: '◐' },
    { label: '缓存条目',    value: status?.cache_entries ?? '—',             icon: '◫' },
  ]

  const sys = [
    { label: 'HTTP 服务',  detail: '0.0.0.0:8082',              ok: true },
    { label: 'MCP 连接',   detail: 'stdio 已连接',               ok: true },
    { label: 'Chrome',     detail: 'chromedp 就绪',              ok: true },
    { label: '数据源',     detail: 'HTTP 直连 + chromedp 备用',  ok: true },
  ]

  return (
    <div className="anim-in space-y-8">

      {/* ---- Hero stat cards 2x2 ---- */}
      <div className="grid grid-cols-2 gap-4">
        {bigs.map((c, i) => (
          <div key={c.label}
            className="anim-in bg-card border border-border rounded-xl px-8 py-7 flex flex-col items-center text-center hover:border-gold/30 transition-colors"
            style={{ animationDelay: `${i * 100}ms` }}>
            <span className="text-[28px] mb-2 opacity-30">{c.icon}</span>
            <span className="text-[52px] font-display font-bold text-text leading-none tracking-tight">
              {c.value}
            </span>
            <span className="text-[16px] text-text-muted mt-3 font-medium tracking-wide">
              {c.label}
            </span>
          </div>
        ))}
      </div>

      {/* ---- System status + info ---- */}
      <div className="grid grid-cols-2 gap-4">
        {/* System status card */}
        <div className="bg-card border border-border rounded-xl px-8 py-7">
          <h3 className="font-display text-gold text-[20px] uppercase tracking-[0.15em] mb-6">
            ● 系统状态
          </h3>
          <div className="space-y-1">
            {sys.map((r) => (
              <div key={r.label} className="flex items-center gap-4 py-2.5">
                <span className={`w-2.5 h-2.5 rounded-full shrink-0 ${r.ok ? 'bg-green pulse-dot' : 'bg-red'}`} />
                <span className="text-[15px] font-medium w-24 shrink-0">{r.label}</span>
                <span className="text-[15px] text-text-secondary">{r.detail}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Quick info card */}
        <div className="bg-card border border-border rounded-xl px-8 py-7 flex flex-col justify-center items-center text-center gap-4">
          <span className="text-[40px] opacity-20">◈</span>
          <div>
            <div className="text-[20px] font-display font-bold text-gold tracking-widest">
              HLTV MCP
            </div>
            <div className="text-[15px] text-text-muted mt-2 leading-relaxed">
              CS2 电竞数据中心<br />
              队伍 · 选手 · 赛程 · 新闻
            </div>
            <div className="text-[13px] text-text-muted mt-4 pt-4 border-t border-border">
              配色方案 A — 深空金
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
