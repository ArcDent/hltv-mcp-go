import { useState } from 'react'
import { api } from '../api/client'

export default function Players() {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<any[]>([])

  const search = async () => {
    const resp = await api.search(query, 'player')
    setResults(resp?.items ?? [])
  }

  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">Players</h1>
      <div className="flex gap-2 mb-4">
        <input placeholder="Search players..." value={query} onChange={e => setQuery(e.target.value)}
          onKeyDown={e => e.key === 'Enter' && search()}
          className="bg-gray-800 border border-gray-700 rounded px-3 py-2 flex-1 text-white" />
        <button onClick={search} className="px-4 py-2 bg-blue-600 rounded">Search</button>
      </div>
      <div className="space-y-2">
        {results.map((p: any, i: number) => (
          <div key={i} className="bg-gray-800 p-3 rounded">{p.name} <span className="text-gray-400">(id={p.id})</span></div>
        ))}
      </div>
    </div>
  )
}
