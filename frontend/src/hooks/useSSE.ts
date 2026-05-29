import { useEffect } from 'react'

type SSEEvent = { entity: string; id?: number; name?: string }
type Callback = (evt: SSEEvent) => void

let eventSource: EventSource | null = null
const listeners = new Map<string, Set<Callback>>()

function connect(): EventSource {
  if (eventSource) return eventSource
  const es = new EventSource('/api/sse')
  es.addEventListener('refreshed', (e: MessageEvent) => {
    try {
      const evt: SSEEvent = JSON.parse(e.data)
      listeners.get(evt.entity)?.forEach(cb => cb(evt))
      // also notify wildcard listeners
      listeners.get('*')?.forEach(cb => cb(evt))
    } catch {
      // ignore malformed events
    }
  })
  es.onerror = () => {
    // EventSource auto-reconnects; no action needed
  }
  eventSource = es
  return es
}

/**
 * Subscribe to SSE refresh events. Callback fires when backend finishes
 * a background scrape and pushes a `refreshed` event.
 *
 * @param entity - entity type to listen for ("player", "team", "matches", "news"), or "*" for all
 * @param callback - called with the SSE event payload
 */
export function useSSE(entity: string, callback: Callback) {
  useEffect(() => {
    const es = connect()
    // Ensure es is connected (EventSource constructor triggers connection)
    void es

    if (!listeners.has(entity)) {
      listeners.set(entity, new Set())
    }
    listeners.get(entity)!.add(callback)

    return () => {
      listeners.get(entity)?.delete(callback)
      if (listeners.get(entity)?.size === 0) {
        listeners.delete(entity)
      }
      // Don't close EventSource — it's a singleton shared across components
    }
  }, [entity, callback])
}
