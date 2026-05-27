const BASE = '/api'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  return res.json()
}

export const api = {
  health: () => request<{ status: string }>('/health'),
  status: () => request<any>('/status'),
  cacheStats: () => request<any>('/cache'),
  clearCache: () => request<any>('/cache', { method: 'DELETE' }),
  search: (q: string, type: 'team' | 'player') =>
    request<any>(`/search?q=${encodeURIComponent(q)}&type=${type}`),
  getTeam: (id: number) => request<any>(`/teams/${id}`),
  getPlayer: (id: number) => request<any>(`/players/${id}`),
  todayMatches: () => request<any>('/matches/today'),
  getEvents: (type: string, limit = 150) =>
    request<any>(`/events?type=${encodeURIComponent(type)}&limit=${limit}`),
  upcomingMatches: (params: Record<string, string>) =>
    request<any>(`/matches?${new URLSearchParams(params)}`),
  results: (params: Record<string, string>) =>
    request<any>(`/results?${new URLSearchParams(params)}`),
  realtimeNews: (limit = 25, offset = 0) =>
    request<any>(`/news/realtime?limit=${limit}&offset=${offset}`),
  newsDigest: (params: Record<string, string>) =>
    request<any>(`/news?${new URLSearchParams(params)}`),
  getNewsArticle: (url: string) =>
    request<any>(`/news/article?url=${encodeURIComponent(url)}`),
}
