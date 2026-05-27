import { useEffect, useState } from 'react'
import { api } from '../api/client'

export default function Dashboard() {
  const [status, setStatus] = useState<any>(null)
  const [modal, setModal] = useState(false)
  useEffect(() => { api.status().then(setStatus).catch(() => {}) }, [])

  const cards = [
    { label: '运行时间', value: status ? `${status.uptime_sec} s` : '—' },
    { label: 'Go 版本',  value: status?.go_version ?? '—' },
    { label: '内存占用', value: status ? `${status.memory_mb} MB` : '—' },
    { label: '缓存条目', value: status?.cache_entries ?? '—' },
  ]

  const epStatus = (status?.endpoints as any[]) ?? []
  const epOk = status?.endpoints_ok ?? 0
  const epTotal = status?.endpoints_total ?? 6
  const chromeOk = epStatus.some((e: any) => e.name === '/matches' && e.ok)

  const sys = [
    { label: 'HTTP 服务', detail: '0.0.0.0:8082',                     desc: 'REST API + 前端面板' },
    { label: 'MCP 连接',  detail: 'stdio',                             desc: 'MCP 协议已连接' },
    { label: 'Chrome',    detail: chromeOk ? 'chromedp 就绪' : '不可用', desc: chromeOk ? 'headless 反爬引擎就绪' : '部分端点不可用' },
    { label: '数据源',    detail: 'HTTP 直连 + chromedp 备用',          desc: `${epOk}/${epTotal} 端点可用`, clickable: true },
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
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 16 }}>
        {sys.map((s, i) => (
          <div key={s.label} className="anim-in"
            onClick={s.clickable ? () => setModal(true) : undefined}
            style={{
              ...cardStyle, animationDelay: `${400 + i * 80}ms`,
              cursor: s.clickable ? 'pointer' : 'default',
              transition: 'border-color 0.2s ease',
            }}
            onMouseEnter={e => { if (s.clickable) e.currentTarget.style.borderColor = 'var(--gold)' }}
            onMouseLeave={e => { if (s.clickable) e.currentTarget.style.borderColor = 'var(--border)' }}
          >
            <span style={{ width: 10, height: 10, borderRadius: '50%', background: 'var(--green)', marginBottom: 16 }} />
            <span style={{ fontFamily: 'var(--font-display)', fontSize: 24, fontWeight: 600,
              color: 'var(--text)', marginBottom: 6, letterSpacing: '0.03em' }}>{s.detail}</span>
            <span style={{ fontSize: 14, color: 'var(--text-secondary)', fontWeight: 500, marginBottom: 4 }}>{s.label}</span>
            <span style={{ fontSize: 13, color: 'var(--text-muted)' }}>{s.desc}</span>
            {s.clickable && <span style={{ fontSize: 11, color: 'var(--gold)', marginTop: 6 }}>点击查看详情 →</span>}
          </div>
        ))}
      </div>

      {/* Modal overlay */}
      {modal && (
        <div onClick={() => setModal(false)} style={{
          position: 'fixed', inset: 0, zIndex: 100,
          background: 'rgba(0,0,0,0.5)', backdropFilter: 'blur(4px)',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          animation: 'fadeIn 0.2s ease',
        }}>
          <div onClick={e => e.stopPropagation()} style={{
            background: 'var(--card)', border: '1px solid var(--border)',
            borderRadius: 'var(--radius)', width: 520, maxWidth: '90vw',
            padding: '32px', boxShadow: '0 20px 60px rgba(0,0,0,0.3)',
            animation: 'slideUp 0.25s ease',
          }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
              <h2 style={{ fontFamily: 'var(--font-display)', fontSize: 20, fontWeight: 700,
                color: 'var(--gold)', letterSpacing: '0.06em', textTransform: 'uppercase' }}>
                ● 数据源状态
              </h2>
              <button onClick={() => setModal(false)} style={{
                width: 30, height: 30, borderRadius: '50%', border: '1px solid var(--border)',
                background: 'var(--card)', color: 'var(--text-secondary)', fontSize: 16,
                display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer',
              }}>✕</button>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
              {epStatus.length === 0 && <div style={{ textAlign: 'center', color: 'var(--text-muted)', padding: 20 }}>加载中...</div>}
              {epStatus.map((ep: any) => (
                <div key={ep.name} style={{
                  display: 'flex', alignItems: 'center', gap: 14, padding: '12px 16px',
                  background: 'var(--input-bg)', borderRadius: 'var(--radius-sm)',
                  border: '1px solid var(--border)',
                }}>
                  <span style={{
                    width: 8, height: 8, borderRadius: '50%',
                    background: ep.ok ? 'var(--green)' : 'var(--red)',
                  }} />
                  <span style={{ flex: 1, fontSize: 15, fontFamily: 'var(--font-mono)' }}>{ep.name}</span>
                  <span style={{
                    fontSize: 12, fontWeight: 600, padding: '2px 10px', borderRadius: 10,
                    background: 'var(--gold-dim)', color: 'var(--gold)',
                  }}>{ep.method}</span>
                  <span style={{
                    fontSize: 12, fontWeight: 600, padding: '2px 10px', borderRadius: 10,
                    background: ep.ok ? 'rgba(94,201,124,0.12)' : 'var(--red-dim)',
                    color: ep.ok ? 'var(--green)' : 'var(--red)',
                  }}>{ep.ok ? '✓ 正常' : '✗ 异常'}</span>
                </div>
              ))}
            </div>
            <div style={{ marginTop: 16, fontSize: 13, color: 'var(--text-muted)', textAlign: 'center' }}>
              点击遮罩关闭
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
