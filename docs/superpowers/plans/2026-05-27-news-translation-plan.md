# 新闻翻译 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为新闻页面添加基于 OpenAI 兼容 API 的中文翻译，配置持久化到 Go 后端，双语展示英文原标题 + 中文翻译。

**Architecture:** Go 后端新增 `/api/translate/config`（GET/PUT）存储配置到 `translate_config.json`。前端新增 `TranslateProvider` 配置弹窗组件，`News.tsx` 加载翻译并逐条展示。

**Tech Stack:** Go 1.26, chi, React 18, TypeScript, fetch API

**Spec:** `docs/superpowers/specs/2026-05-27-news-translation.md`

---

### Task 1: Go 后端翻译配置端点

**Files:**
- Create: `internal/http/handlers/translate.go`
- Modify: `internal/http/router.go`（注册新路由）

- [ ] **Step 1: Write translate.go — GET/PUT handler + file persistence**

```go
package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const translateConfigFile = "translate_config.json"

type TranslateConfig struct {
	ProviderURL string `json:"provider_url"`
	APIKey      string `json:"api_key"`
	Model       string `json:"model"`
}

func configPath() string {
	exec, _ := os.Executable()
	dir := filepath.Dir(exec)
	// Fallback: if binary path unavailable, use cwd
	if dir == "" || dir == "." {
		dir, _ = os.Getwd()
	}
	return filepath.Join(dir, translateConfigFile)
}

func loadTranslateConfig() (TranslateConfig, error) {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return TranslateConfig{}, err
	}
	var cfg TranslateConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return TranslateConfig{}, err
	}
	return cfg, nil
}

func saveTranslateConfig(cfg TranslateConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0600)
}

func maskKey(key string) string {
	if len(key) <= 6 {
		return strings.Repeat("*", len(key))
	}
	return key[:3] + strings.Repeat("*", len(key)-6) + key[len(key)-3:]
}

func (h *Handlers) GetTranslateConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := loadTranslateConfig()
	if err != nil {
		writeJSON(w, map[string]any{
			"provider_url": "",
			"api_key":      "",
			"model":        "",
			"configured":   false,
		})
		return
	}
	writeJSON(w, map[string]any{
		"provider_url": cfg.ProviderURL,
		"api_key":      maskKey(cfg.APIKey),
		"model":        cfg.Model,
		"configured":   cfg.ProviderURL != "" && cfg.APIKey != "",
	})
}

func (h *Handlers) PutTranslateConfig(w http.ResponseWriter, r *http.Request) {
	var cfg TranslateConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	// If key is masked (contains ***), keep the existing key
	if strings.Contains(cfg.APIKey, "***") {
		existing, err := loadTranslateConfig()
		if err == nil {
			cfg.APIKey = existing.APIKey
		}
	}
	if err := saveTranslateConfig(cfg); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save config")
		return
	}
	writeJSON(w, map[string]string{"status": "saved"})
}
```

- [ ] **Step 2: Register routes in router.go**

In `internal/http/router.go`, add after existing routes:

```go
r.Get("/api/translate/config", h.GetTranslateConfig)
r.Put("/api/translate/config", h.PutTranslateConfig)
```

- [ ] **Step 3: Build and verify**

```bash
go build github.com/arcdent/hltv-mcp/internal/...
```
Expected: success

- [ ] **Step 4: Commit**

```bash
git add internal/http/handlers/translate.go internal/http/router.go
git commit -m "feat: add translate config API endpoints (GET/PUT)"
```

---

### Task 2: 前端翻译配置弹窗组件

**Files:**
- Create: `frontend/src/components/TranslateProvider.tsx`

- [ ] **Step 1: Write TranslateProvider.tsx**

```tsx
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
  const [open, setOpen] = useState(false)

  const fetch = async () => {
    try {
      const r = await fetch('/api/translate/config')
      setCfg(await r.json())
    } catch { setCfg({ provider_url: '', api_key: '', model: '', configured: false }) }
  }

  useEffect(() => { fetch() }, [])

  const save = async (url: string, key: string, model: string) => {
    await fetch('/api/translate/config', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ provider_url: url, api_key: key, model }),
    })
    await fetch()
    setOpen(false)
  }

  return { cfg, save, open, setOpen }
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
```

- [ ] **Step 2: Build and verify**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild/frontend && npx vite build --outDir ../dist 2>&1 | tail -1
```
Expected: `✓ built`

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/TranslateProvider.tsx
git commit -m "feat: add TranslateProvider config panel component"
```

---

### Task 3: 新闻翻译展示 — News.tsx 集成

**Files:**
- Modify: `frontend/src/pages/News.tsx`

- [ ] **Step 1: Write updated News.tsx with translation logic**

```tsx
import { useEffect, useState, useCallback } from 'react'
import { api } from '../api/client'
import { useTranslateConfig, TranslateModal } from '../components/TranslateProvider'

type Tab = 'realtime' | 'archive'

// Translation cache helpers
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
  const { cfg, save, open, setOpen } = useTranslateConfig()
  const [translations, setTranslations] = useState<Record<string, string>>({})
  const [translating, setTranslating] = useState<Set<string>>(new Set())

  useEffect(() => { setData(null); if (tab === 'realtime') api.realtimeNews().then(setData)
    else api.newsDigest({ limit: '30' }).then(setData) }, [tab])

  // Translate items on load
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

    // Concurrent translate (max 3)
    let active = 0; let idx = 0
    const run = async () => {
      while (idx < toTranslate.length && active < 3) {
        const title = toTranslate[idx++]
        active++
        setTranslating(prev => new Set(prev).add(title))
        try {
          const res = await fetch(`${cfg!.provider_url}/chat/completions`, {
            method: 'POST', headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${cfg!.api_key}` },
            body: JSON.stringify({ model: cfg!.model, messages: [
              { role: 'system', content: '将以下CS电竞新闻标题翻译为简体中文，只输出翻译结果，不要任何解释' },
              { role: 'user', content: title },
            ], temperature: 0.1 }),
          })
          const j = await res.json()
          const zh = j?.choices?.[0]?.message?.content?.trim() ?? ''
          if (zh) {
            cache[hashTitle(title)] = { zh, ts: Date.now() }
            setTranslations(prev => ({ ...prev, [title]: zh }))
          }
        } catch {}
        active--
        setTranslating(prev => { const s = new Set(prev); s.delete(title); return s })
      }
    }
    // Start 3 workers
    run(); run(); run()
  }, [data, cfg?.configured])

  const items: any[] = data?.items ?? []

  const card: React.CSSProperties = { background: 'var(--card)', border: '1px solid var(--border)',
    borderRadius: 'var(--radius)', padding: '14px 20px', boxShadow: 'var(--card-shadow)' }
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
          <button key={t.key} onClick={() => setTab(t.key as Tab)} style={tabBtn(tab === t.key)}>{t.label}</button>
        ))}
        <div style={{ flex: 1 }} />
        <button onClick={() => setOpen(true)} title="翻译设置" style={{
          width: 30, height: 30, borderRadius: '50%', border: '1px solid var(--border)',
          background: 'var(--card)', color: cfg?.configured ? 'var(--gold)' : 'var(--text-muted)',
          fontSize: 16, cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center',
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
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: 14, fontWeight: 700, color: 'var(--gold)', minWidth: 24 }}>
                  {String(i + 1).padStart(2, '0')}
                </span>
                <span style={{ flex: 1, fontSize: 16, fontWeight: 500 }}>{n.title}</span>
                <span style={{ fontSize: 13, color: 'var(--text-muted)', whiteSpace: 'nowrap' }}>
                  {n.published_at ?? n.relative_time ?? ''}
                </span>
              </div>
              {/* Translation line */}
              {cfg?.configured && (
                <div style={{ fontSize: 13, color: 'var(--text-muted)', paddingLeft: 38, lineHeight: 1.5 }}>
                  {loading ? '翻译中...' : (zh ?? '')}
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Build and verify**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild/frontend && npx vite build --outDir ../dist 2>&1 | tail -1
```
Expected: `✓ built`

- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/News.tsx
git commit -m "feat: integrate translation into News — OpenAI API, bilingual display"
```

---

### Task 4: Final integration — build + test + AGENTS.md

- [ ] **Step 1: Full build**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild/frontend && npx vite build --outDir ../dist && cd .. && go build -o hltv-mcp github.com/arcdent/hltv-mcp
```

- [ ] **Step 2: Verify translate config API**

```bash
./hltv-mcp &
sleep 2
curl -s http://localhost:8082/api/translate/config
curl -s -X PUT http://localhost:8082/api/translate/config \
  -H 'Content-Type: application/json' \
  -d '{"provider_url":"https://test.example.com/v1","api_key":"sk-test123456","model":"gpt-test"}'
curl -s http://localhost:8082/api/translate/config
kill %1
```
Expected: first GET returns `{"configured":false}`, PUT returns `{"status":"saved"}`, second GET returns config with masked key `sk-***456`.

- [ ] **Step 3: Delete test config file + commit**

```bash
rm -f translate_config.json
git add -A && git commit -m "feat: complete news translation — backend config + frontend panel + bilingual display" && git push
```
