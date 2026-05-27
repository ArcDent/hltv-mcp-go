import { useEffect, useState } from 'react'
import { api } from '../api/client'

export default function Dashboard() {
  const [status, setStatus] = useState<any>(null)
  useEffect(() => { api.status().then(setStatus).catch(() => {}) }, [])

  const cards = [
    { label: '运行时间', value: status ? `${status.uptime_sec} s` : '—' },
    { label: 'Go 版本',  value: status?.go_version ?? '—' },
    { label: '内存占用', value: status ? `${status.memory_mb} MB` : '—' },
    { label: '缓存条目', value: status?.cache_entries ?? '—' },
  ]

  const sys = [
    { label: 'HTTP 服务', detail: '0.0.0.0:8082',            desc: 'REST API + 前端面板' },
    { label: 'MCP 连接',  detail: 'stdio',                    desc: 'MCP 协议已连接' },
    { label: 'Chrome',    detail: 'chromedp',                 desc: 'headless 反爬引擎就绪' },
    { label: '数据源',    detail: 'HTTP 直连 + chromedp 备用', desc: '5/6 端点可用' },
  ]

  const cardStyle: React.CSSProperties = {
    background: 'var(--card)', border: '1px solid var(--border)',
    borderRadius: 'var(--radius)', padding: '28px 24px',
    boxShadow: 'var(--card-shadow)',
    display: 'flex', flexDirection: 'column', alignItems: 'center', textAlign: 'center',
  }

  return (
    <div className="anim-in" style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
      {/* Stat cards 2x2 */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 16 }}>
        {cards.map((c, i) => (
          <div key={c.label} className="anim-in" style={{ ...cardStyle, animationDelay: `${i * 80}ms` }}>
            <span style={{ fontFamily: 'var(--font-display)', fontSize: 56, fontWeight: 700,
              color: 'var(--text)', lineHeight: 1, marginBottom: 12 }}>{c.value}</span>
            <span style={{ fontSize: 15, color: 'var(--text-secondary)', fontWeight: 500 }}>{c.label}</span>
          </div>
        ))}
      </div>

      {/* System cards 2x2 */}
      <h2 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 600,
        color: 'var(--gold-text)', letterSpacing: '0.06em', textTransform: 'uppercase' }}>
        系统状态
      </h2>
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 16 }}>
        {sys.map((s, i) => (
          <div key={s.label} className="anim-in" style={{ ...cardStyle, animationDelay: `${400 + i * 80}ms` }}>
            <span style={{ width: 10, height: 10, borderRadius: '50%', background: 'var(--green)', marginBottom: 16 }} />
            <span style={{ fontFamily: 'var(--font-display)', fontSize: 24, fontWeight: 600,
              color: 'var(--text)', marginBottom: 6, letterSpacing: '0.03em' }}>{s.detail}</span>
            <span style={{ fontSize: 14, color: 'var(--text-secondary)', fontWeight: 500, marginBottom: 4 }}>{s.label}</span>
            <span style={{ fontSize: 13, color: 'var(--text-muted)' }}>{s.desc}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
