import { useState } from 'react'
import { api } from '../api/client'

export default function Teams() {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<any[]>([])

  const search = async () => {
    const resp = await api.search(query, 'team')
    setResults(resp?.items ?? [])
  }

  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">Teams</h1>
      <div className="flex gap-2 mb-4">
        <input placeholder="Search teams..." value={query} onChange={e => setQuery(e.target.value)}
          onKeyDown={e => e.key === 'Enter' && search()}
          className="bg-gray-800 border border-gray-700 rounded px-3 py-2 flex-1 text-white" />
        <button onClick={search} className="px-4 py-2 bg-blue-600 rounded">Search</button>
      </div>
      <div className="space-y-2">
        {results.map((t: any, i: number) => (
          <div key={i} className="bg-gray-800 p-3 rounded">{t.name} <span className="text-gray-400">(id={t.id})</span></div>
        ))}
      </div>
    </div>
  )
}
