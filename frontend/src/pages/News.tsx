import { useEffect, useState } from 'react'
import { api } from '../api/client'

export default function News() {
  const [tab, setTab] = useState<'realtime' | 'archive'>('realtime')
  const [data, setData] = useState<any>(null)

  useEffect(() => {
    if (tab === 'realtime') api.realtimeNews().then(setData)
    else api.newsDigest({ limit: '25' }).then(setData)
  }, [tab])

  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">News</h1>
      <div className="flex gap-2 mb-4">
        {(['realtime', 'archive'] as const).map(t => (
          <button key={t} onClick={() => setTab(t)} className={`px-4 py-1 rounded ${tab === t ? 'bg-blue-600' : 'bg-gray-700'}`}>
            {t === 'realtime' ? 'Realtime' : 'Archive'}
          </button>
        ))}
      </div>
      <div className="space-y-2">
        {data?.items?.map((n: any, i: number) => (
          <div key={i} className="bg-gray-800 p-3 rounded">
            <div className="font-medium">{n.title}</div>
            <div className="text-sm text-gray-400">{n.published_at || n.relative_time}</div>
          </div>
        ))}
      </div>
    </div>
  )
}
