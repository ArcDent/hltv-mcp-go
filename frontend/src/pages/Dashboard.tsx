import { useEffect, useState } from 'react'
import { api } from '../api/client'

type Theme = 'dark' | 'light'

export default function Dashboard() {
  const [status, setStatus] = useState<any>(null)
  const [theme, setTheme] = useState<Theme>('dark')
  useEffect(() => { api.status().then(setStatus).catch(() => {}) }, [])

  const toggleTheme = () => {
    const next = theme === 'dark' ? 'light' : 'dark'
    document.documentElement.classList.toggle('light', next === 'light')
    setTheme(next)
  }

  const cards = [
    { label: '运行时间', value: status ? `${status.uptime_sec} s` : '—', icon: '⏱' },
    { label: 'Go 版本',  value: status?.go_version ?? '—',              icon: '●' },
    { label: '内存占用', value: status ? `${status.memory_mb} MB` : '—', icon: '◐' },
    { label: '缓存条目', value: status?.cache_entries ?? '—',            icon: '◫' },
  ]

  const sys = [
    { label: 'HTTP 服务', detail: '0.0.0.0:8082',             desc: 'REST API + 前端面板' },
    { label: 'MCP 连接',  detail: 'stdio',                     desc: 'MCP 协议已连接' },
    { label: 'Chrome',    detail: 'chromedp',                  desc: 'headless 反爬引擎就绪' },
    { label: '数据源',    detail: 'HTTP 直连',                  desc: 'chromedp 作为降级备用' },
  ]

  return (
    <div className="anim-in space-y-8">

      {/* ---- Stat cards 2x2 ---- */}
      <div className="grid grid-cols-2 gap-4">
        {cards.map((c, i) => (
          <div key={c.label}
            className="anim-in bg-card border border-border rounded-2xl px-8 py-8 flex flex-col items-center text-center hover:border-gold/30 transition-colors"
            style={{ animationDelay: `${i * 100}ms` }}>
            <span className="text-[28px] mb-3 opacity-25">{c.icon}</span>
            <span className="text-[56px] font-display font-bold text-text leading-none tracking-tight">
              {c.value}
            </span>
            <span className="text-[16px] text-text-muted mt-4 font-medium tracking-wide">
              {c.label}
            </span>
          </div>
        ))}
      </div>

      {/* ---- System cards 2x2 ---- */}
      <div className="grid grid-cols-2 gap-4">
        {sys.map((s, i) => (
          <div key={s.label}
            className="anim-in bg-card border border-border rounded-2xl px-8 py-7 flex flex-col items-center text-center hover:border-gold/30 transition-colors"
            style={{ animationDelay: `${400 + i * 100}ms` }}>
            <span className="w-3 h-3 rounded-full bg-green pulse-dot mb-4" />
            <span className="text-[22px] font-display font-semibold text-text tracking-wide">
              {s.detail}
            </span>
            <span className="text-[15px] text-text-muted mt-2 font-medium">
              {s.label}
            </span>
            <span className="text-[13px] text-text-muted mt-2 opacity-60">
              {s.desc}
            </span>
          </div>
        ))}
      </div>

      {/* ---- Theme toggle ---- */}
      <div className="flex justify-center">
        <button
          onClick={toggleTheme}
          className="group flex items-center gap-3 px-6 py-3 bg-card border border-border rounded-full hover:border-gold/40 transition-all duration-500"
        >
          <span className="text-[20px] transition-transform duration-500 group-hover:rotate-12">
            {theme === 'dark' ? '🌙' : '☀️'}
          </span>
          <span className="text-[15px] font-medium text-text-secondary">
            配色方案 — {theme === 'dark' ? '深空金' : '日光白'}
          </span>
          <span className="text-[11px] text-text-muted">点击切换</span>
        </button>
      </div>
    </div>
  )
}
