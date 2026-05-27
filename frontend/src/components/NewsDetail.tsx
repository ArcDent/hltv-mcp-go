import { useEffect, useState } from 'react'
import Modal from './Modal'
import { useTranslateConfig } from './TranslateProvider'

type ArticleData = {
  title: string; published_at: string; link: string; body_text: string; author?: string;
}

export default function NewsDetail({ url, onClose }: { url: string; onClose: () => void }) {
  const [data, setData] = useState<ArticleData | null>(null)
  const [loading, setLoading] = useState(true)
  const [translating, setTranslating] = useState(false)
  const [translated, setTranslated] = useState('')
  const { cfg, realKey } = useTranslateConfig()

  useEffect(() => {
    setLoading(true)
    fetch(`/api/news/article?url=${encodeURIComponent(url)}`).then(r => r.json()).then(d => {
      setData(d.data ?? null); setLoading(false)
    }).catch(() => setLoading(false))
  }, [url])

  // Check localStorage for cached translation
  useEffect(() => {
    if (!data?.body_text) return
    try {
      let hash = 0; for (let i = 0; i < url.length; i++) { hash = (hash * 31 + url.charCodeAt(i)) >>> 0 }
      const key = `news_trans:${hash.toString(16)}`
      const cached = localStorage.getItem(key)
      if (cached) {
        const { zh } = JSON.parse(cached)
        setTranslated(zh)
      }
    } catch { /* ignore corrupt cache */ }
  }, [data, url])

  const doTranslate = async () => {
    if (!data?.body_text || !cfg?.configured) return
    setTranslating(true)
    try {
      const res = await fetch(`${cfg!.provider_url}/chat/completions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${realKey}` },
        body: JSON.stringify({
          model: cfg!.model,
          messages: [
            { role: 'system', content: '将以下CS电竞新闻正文翻译为简体中文' },
            { role: 'user', content: data.body_text.slice(0, 8000) },
          ],
          temperature: 0.1,
        }),
      })
      const body = await res.text()
      if (!res.ok) throw new Error(body)
      const j = JSON.parse(body)
      const zh = (j?.choices?.[0]?.message?.content as string)?.trim() ?? ''
      if (zh) {
        setTranslated(zh)
        try {
          let hash = 0; for (let i = 0; i < url.length; i++) { hash = (hash * 31 + url.charCodeAt(i)) >>> 0 }
          localStorage.setItem(`news_trans:${hash.toString(16)}`, JSON.stringify({ zh, ts: Date.now() }))
        } catch { /* ignore storage errors */ }
      }
    } catch (e) { console.error('translate article failed:', e) }
    setTranslating(false)
  }

  return (
    <Modal onClose={onClose} width={800} maxHeight="90vh">

        {loading && <div style={{textAlign:'center',padding:60,color:'var(--text-muted)'}}>抓取中...</div>}
        {!loading && !data && <div style={{textAlign:'center',padding:60,color:'var(--text-muted)'}}>文章暂时不可用</div>}

        {!loading && data && (
          <>
            <div style={{fontSize:22,fontWeight:700,color:'var(--text)',lineHeight:1.3,marginBottom:8}}>{data.title}</div>
            <div style={{display:'flex',alignItems:'center',gap:12,fontSize:12,color:'var(--text-muted)',marginBottom:18,paddingBottom:14,borderBottom:'1px solid var(--border)'}}>
              {data.published_at && <span>{data.published_at}</span>}
              {data.author && <span>· {data.author}</span>}
            </div>

            <div style={{fontSize:14,lineHeight:1.8,color:'var(--text)',whiteSpace:'pre-wrap',marginBottom:20}}>
              {data.body_text}
            </div>

            {translated && (
              <>
                <div style={{fontFamily:'var(--font-display)',fontSize:14,fontWeight:600,color:'var(--gold)',letterSpacing:'0.05em',textTransform:'uppercase',marginBottom:10,paddingBottom:8,borderBottom:'1px solid var(--border)'}}>
                  中文翻译
                </div>
                <div style={{fontSize:14,lineHeight:1.8,color:'var(--text-secondary)',whiteSpace:'pre-wrap',marginBottom:20}}>
                  {translated}
                </div>
              </>
            )}

            <div style={{display:'flex',gap:12,justifyContent:'center',paddingTop:14,borderTop:'1px solid var(--border)'}}>
              {cfg?.configured && !translated && (
                <button onClick={doTranslate} disabled={translating} style={{
                  padding:'8px 20px',background:'var(--gold)',color:'#fff',border:'none',borderRadius:'var(--radius-sm)',
                  fontSize:14,fontWeight:600,fontFamily:'var(--font-display)',letterSpacing:'0.04em',textTransform:'uppercase',
                  cursor:translating?'not-allowed':'pointer',opacity:translating?0.5:1,
                }}>
                  {translating ? '翻译中...' : '翻译正文'}
                </button>
              )}
              {data.link && (
                <a href={data.link} target="_blank" rel="noopener noreferrer" style={{
                  padding:'8px 20px',background:'var(--input-bg)',color:'var(--text-secondary)',border:'1px solid var(--border)',
                  borderRadius:'var(--radius-sm)',fontSize:14,textDecoration:'none',
                }}>
                  在 HLTV 阅读原文 →
                </a>
              )}
            </div>
          </>
        )}
    </Modal>
  )
}
