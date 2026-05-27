import { useState, useEffect } from 'react'

const PRESETS = [
  { label: 'OpenAI',         url: 'https://api.openai.com/v1',        model: 'gpt-4o-mini' },
  { label: 'DeepSeek',       url: 'https://api.deepseek.com/v1',      model: 'deepseek-chat' },
  { label: 'Groq',           url: 'https://api.groq.com/openai/v1',   model: 'llama-3.3-70b-versatile' },
  { label: 'Ollama 本地',    url: 'http://localhost:11434/v1',        model: 'qwen2.5:7b' },
]

type Config = { provider_url: string; api_key: string; model: string; configured: boolean }

export function useTranslateConfig() {
  const [cfg, setCfg] = useState<Config | null>(null)
  const [realKey, setRealKey] = useState(() => sessionStorage.getItem('hltv_real_key') ?? '')
  const [open, setOpen] = useState(false)

  const fetchConfig = async () => {
    try {
      const r = await fetch('/api/translate/config')
      const c = await r.json()
      setCfg(c)
    } catch { setCfg({ provider_url: '', api_key: '', model: '', configured: false } as Config) }
  }

  useEffect(() => { fetchConfig() }, [])

  const save = async (url: string, key: string, model: string) => {
    sessionStorage.setItem('hltv_real_key', key)
    setRealKey(key)
    await fetch('/api/translate/config', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ provider_url: url, api_key: key, model }),
    })
    await fetchConfig()
    setOpen(false)
  }

  return { cfg, realKey, save, open, setOpen }
}

const overlay: React.CSSProperties = {
  position: 'fixed', inset: 0, zIndex: 100, background: 'rgba(0,0,0,0.5)',
  backdropFilter: 'blur(4px)', display: 'flex', alignItems: 'center', justifyContent: 'center',
  animation: 'fadeIn 0.2s ease',
}
const modal: React.CSSProperties = {
  background: 'var(--card)', border: '1px solid var(--border)',
  borderRadius: 'var(--radius)', width: 460, maxWidth: '90vw', padding: 32,
  boxShadow: '0 20px 60px rgba(0,0,0,0.3)', animation: 'slideUp 0.25s ease',
}
const inputS: React.CSSProperties = {
  width: '100%', background: 'var(--input-bg)', border: '1px solid var(--border)',
  borderRadius: 'var(--radius-sm)', color: 'var(--text)', fontSize: 14,
  padding: '10px 14px', outline: 'none', marginTop: 6, marginBottom: 16,
}
const labelS: React.CSSProperties = { fontSize: 13, fontWeight: 600, color: 'var(--text-secondary)' }

export function TranslateModal({ cfg, onSave, onClose }: {
  cfg: Config | null; onSave: (url: string, key: string, model: string) => void; onClose: () => void
}) {
  const [url, setUrl] = useState(cfg?.provider_url ?? '')
  const [key, setKey] = useState(cfg?.api_key ?? '')
  const [model, setModel] = useState(cfg?.model ?? '')

  const applyPreset = (p: typeof PRESETS[0]) => { setUrl(p.url); setModel(p.model) }

  return (
    <div style={overlay} onClick={onClose}>
      <div style={modal} onClick={e => e.stopPropagation()}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
          <h2 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 700,
            color: 'var(--gold)', letterSpacing: '0.06em', textTransform: 'uppercase' }}>翻译设置</h2>
          <button onClick={onClose} style={{ width: 30, height: 30, borderRadius: '50%',
            border: '1px solid var(--border)', background: 'var(--card)',
            color: 'var(--text-secondary)', fontSize: 16, cursor: 'pointer' }}>✕</button>
        </div>

        <label style={labelS}>API 地址</label>
        <input style={inputS} value={url} onChange={e => setUrl(e.target.value)}
          placeholder="https://api.openai.com/v1"
          onFocus={e => { e.target.style.borderColor = 'var(--gold)'; e.target.style.boxShadow = '0 0 0 3px var(--gold-dim)' }}
          onBlur={e => { e.target.style.borderColor = 'var(--border)'; e.target.style.boxShadow = 'none' }} />

        <label style={labelS}>API Key</label>
        <input style={inputS} type="password" value={key} onChange={e => setKey(e.target.value)}
          placeholder="sk-..."
          onFocus={e => { e.target.style.borderColor = 'var(--gold)'; e.target.style.boxShadow = '0 0 0 3px var(--gold-dim)' }}
          onBlur={e => { e.target.style.borderColor = 'var(--border)'; e.target.style.boxShadow = 'none' }} />

        <label style={labelS}>模型</label>
        <input style={{ ...inputS, marginBottom: 12 }} value={model} onChange={e => setModel(e.target.value)}
          onFocus={e => { e.target.style.borderColor = 'var(--gold)'; e.target.style.boxShadow = '0 0 0 3px var(--gold-dim)' }}
          onBlur={e => { e.target.style.borderColor = 'var(--border)'; e.target.style.boxShadow = 'none' }} />

        <div style={{ display: 'flex', gap: 6, marginBottom: 20, flexWrap: 'wrap' }}>
          {PRESETS.map(p => (
            <button key={p.label} onClick={() => applyPreset(p)} style={{
              padding: '4px 10px', fontSize: 12, borderRadius: 'var(--radius-sm)',
              border: '1px solid var(--border)', background: 'var(--input-bg)',
              color: (url === p.url && model === p.model) ? 'var(--gold)' : 'var(--text-muted)',
              cursor: 'pointer',
            }}>{p.label}</button>
          ))}
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <button onClick={() => onSave(url, key, model)} style={{
            padding: '10px 24px', background: 'var(--gold)', color: 'var(--bg)',
            border: 'none', borderRadius: 'var(--radius-sm)', fontSize: 14, fontWeight: 600,
            fontFamily: 'var(--font-display)', letterSpacing: '0.04em', textTransform: 'uppercase', cursor: 'pointer',
          }}>保存</button>
          <span style={{ fontSize: 12, color: cfg?.configured ? 'var(--green)' : 'var(--text-muted)' }}>
            {cfg?.configured ? '● 已配置' : '○ 未配置'}
          </span>
        </div>
      </div>
    </div>
  )
}
