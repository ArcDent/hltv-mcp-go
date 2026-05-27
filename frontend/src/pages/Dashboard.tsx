import { useEffect, useState } from 'react'
import { api } from '../api/client'

export default function Dashboard() {
  const [status, setStatus] = useState<any>(null)
  useEffect(() => { api.status().then(setStatus).catch(() => {}) }, [])

  const stats = [
    { label: '运行时间', value: status ? `${status.uptime_sec}s` : '--' },
    { label: 'Go 版本', value: status?.go_version ?? '--' },
    { label: '内存占用', value: status ? `${status.memory_mb} MB` : '--' },
    { label: '缓存条目', value: status?.cache_entries ?? '--' },
  ]

  const sysRows = [
    { label: 'HTTP 服务', detail: '0.0.0.0:8082' },
    { label: 'MCP 连接', detail: 'stdio 已连接' },
    { label: 'Chrome', detail: 'chromedp 就绪' },
    { label: '数据源', detail: 'HTTP 直连 + chromedp 备用' },
  ]

  return (
    <div className="animate-in">
      <h2 className="text-[16px] font-semibold text-gold mb-4 pb-2 border-b border-border tracking-wide">
        ◈ 总览
      </h2>

      <div className="grid grid-cols-4 gap-3 mb-8">
        {stats.map((s, i) => (
          <div key={s.label} className="bg-panel border border-border rounded-md p-5 animate-in"
            style={{ animationDelay: `${i * 80}ms` }}>
            <div className="text-[13px] text-text-dim mb-2">{s.label}</div>
            <div className="text-[26px] font-bold text-text font-mono">{s.value}</div>
          </div>
        ))}
      </div>

      <div className="bg-panel border border-border rounded-md p-5">
        <div className="text-[14px] font-semibold text-text mb-4">系统状态</div>
        <div className="space-y-3">
          {sysRows.map((row) => (
            <div key={row.label} className="flex items-center gap-3 text-[14px]">
              <span className="w-2 h-2 rounded-full bg-[#3fb950]" />
              <span className="text-text w-28">{row.label}</span>
              <span className="text-text-dim">{row.detail}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
