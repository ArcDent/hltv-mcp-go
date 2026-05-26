import { useEffect, useState } from 'react'
import { api } from '../api/client'

export default function Cache() {
  const [stats, setStats] = useState<any>(null)
  const refresh = () => { api.cacheStats().then(setStats) }
  useEffect(refresh, [])

  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">Cache Management</h1>
      <div className="grid grid-cols-3 gap-4 mb-4">
        <div className="bg-gray-800 p-4 rounded"><div className="text-gray-400">Entries</div><div className="text-2xl">{stats?.entries ?? '...'}</div></div>
      </div>
      <button onClick={() => api.clearCache().then(refresh)} className="px-4 py-2 bg-red-600 rounded">Clear All Cache</button>
    </div>
  )
}
