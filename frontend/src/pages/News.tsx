import { useEffect, useState } from 'react'
import { api } from '../api/client'
import { useTranslateConfig, TranslateModal } from '../components/TranslateProvider'

type Tab = 'realtime' | 'archive'

const CACHE_KEY = 'hltv_translations'
const CACHE_TTL = 7 * 24 * 3600 * 1000

function loadCache(): Record<string, { zh: string; ts: number }> {
  try { return JSON.parse(localStorage.getItem(CACHE_KEY) ?? '{}') } catch { return {} }
}
function saveCache(c: Record<string, { zh: string; ts: number }>) {
  localStorage.setItem(CACHE_KEY, JSON.stringify(c))
}
function hashTitle(t: string) {
  let h = 0; for (let i = 0; i < t.length; i++) { h = (h * 31 + t.charCodeAt(i)) >>> 0 }
  return h.toString(16)
}

export default function News() {
  const [tab, setTab] = useState<Tab>('realtime')
  const [data, setData] = useState<any>(null)
  const { cfg, realKey, save, open, setOpen } = useTranslateConfig()
  const [translations, setTranslations] = useState<Record<string, string>>({})
  const [translating, setTranslating] = useState<Set<string>>(new Set())

  useEffect(() => {
    setData(null)
    if (tab === 'realtime') api.realtimeNews().then(setData)
    else api.newsDigest({ limit: '30' }).then(setData)
  }, [tab])

  useEffect(() => {
    const items: any[] = data?.items ?? []
    if (!cfg?.configured || items.length === 0) return

    const cache = loadCache()
    const toTranslate: string[] = []
    const known: Record<string, string> = {}

    for (const item of items) {
      if (!item.title) continue
      const h = hashTitle(item.title)
      const cached = cache[h]
      if (cached && Date.now() - cached.ts < CACHE_TTL) {
        known[item.title] = cached.zh
      } else if (!translations[item.title]) {
        toTranslate.push(item.title)
      }
    }

    setTranslations(prev => ({ ...prev, ...known }))

    let active = 0; let idx = 0
    const run = async () => {
      while (idx < toTranslate.length && active < 3) {
        const title = toTranslate[idx++]
        active++
        setTranslating(prev => new Set(prev).add(title))
        try {
          const res = await fetch(`${cfg!.provider_url}/chat/completions`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${realKey}` },
            body: JSON.stringify({
              model: cfg!.model,
              messages: [
                { role: 'system', content: '将以下CS电竞新闻标题翻译为简体中文，只输出翻译结果，不要任何解释' },
                { role: 'user', content: title },
              ],
              temperature: 0.1,
            }),
          })
          const body = await res.text()
          if (!res.ok) { console.error('translate API error', res.status, body); throw new Error(body) }
          const j = JSON.parse(body)
          const zh = (j?.choices?.[0]?.message?.content as string)?.trim() ?? ''
          if (zh) {
            cache[hashTitle(title)] = { zh, ts: Date.now() }
            saveCache(cache)
            setTranslations(prev => ({ ...prev, [title]: zh }))
          }
        } catch (e) { console.error('translate failed:', title, e) }
        active--
        setTranslating(prev => { const s = new Set(prev); s.delete(title); return s })
      }
    }
    run(); run(); run()
  }, [data, cfg?.configured])

  const items: any[] = data?.items ?? []

  const card: React.CSSProperties = {
    background: 'var(--card)', border: '1px solid var(--border)',
    borderRadius: 'var(--radius)', padding: '14px 20px', boxShadow: 'var(--card-shadow)',
  }
  const tabBtn = (active: boolean): React.CSSProperties => ({
    fontSize: 16, fontWeight: 600, fontFamily: 'var(--font-display)',
    letterSpacing: '0.04em', textTransform: 'uppercase' as const,
    color: active ? 'var(--gold)' : 'var(--text-muted)',
    borderBottom: active ? '2px solid var(--gold)' : '2px solid transparent',
    paddingBottom: 6, background: 'none', cursor: 'pointer',
  })

  return (
    <div className="anim-in" style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div style={{ display: 'flex', gap: 24, borderBottom: '1px solid var(--border)', paddingBottom: 0, alignItems: 'center' }}>
        {[{ key: 'realtime', label: '实时新闻' }, { key: 'archive', label: '归档新闻' }].map(t => (
          <button key={t.key} onClick={() => setTab(t.key as Tab)} style={tabBtn(tab === t.key)}>
            {t.label}
          </button>
        ))}
        <div style={{ flex: 1 }} />
        <button onClick={() => setOpen(true)} title="翻译设置" style={{
          width: 32, height: 32, borderRadius: '50%',
          border: '1px solid var(--border)', background: 'var(--card)',
          color: cfg?.configured ? 'var(--gold)' : 'var(--text-muted)',
          fontSize: 16, cursor: 'pointer',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
        }}>⚙</button>
      </div>

      {open && cfg && <TranslateModal cfg={cfg} onSave={save} onClose={() => setOpen(false)} />}

      <div key={tab} style={{ animation: 'slideUp 0.3s ease both' }}>
        {items.length === 0 && (
          <div style={{ ...card, textAlign: 'center', padding: '80px 0', color: 'var(--text-muted)', fontSize: 15 }}>
            {data ? '暂无新闻' : '加载中...'}
          </div>
        )}
        {items.map((n, i) => {
          const zh = translations[n.title]
          const loading = translating.has(n.title)
          return (
            <div key={i} className="anim-in" style={{
              ...card, marginBottom: i < items.length - 1 ? 6 : 0,
              flexDirection: 'column', alignItems: 'stretch', gap: 4,
              animationDelay: `${i * 30}ms`,
            }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: 14, fontWeight: 700,
                  color: 'var(--gold)', minWidth: 24 }}>
                  {String(i + 1).padStart(2, '0')}
                </span>
                <span style={{ flex: 1, fontSize: 16, fontWeight: 500 }}>{n.title}</span>
                <span style={{ fontSize: 13, color: 'var(--text-muted)', whiteSpace: 'nowrap' }}>
                  {n.published_at ?? n.relative_time ?? ''}
                </span>
              </div>
              {cfg?.configured && (
                <div style={{ fontSize: 13, color: 'var(--text-muted)', paddingLeft: 38, lineHeight: 1.5 }}>
                  {loading ? '翻译中...' : (zh || '')}
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
