import { useState, useEffect, useCallback } from 'react'

type NicknameDict = Record<string, string>

let cachedTeams: NicknameDict | null = null
let cachedPlayers: NicknameDict | null = null
let fetchPromise: Promise<void> | null = null

async function ensureLoaded(): Promise<void> {
  if (cachedTeams && cachedPlayers) return
  if (fetchPromise) {
    await fetchPromise
    return
  }
  fetchPromise = (async () => {
    const resp = await fetch('/api/nicknames')
    const data = await resp.json()
    cachedTeams = data.teams ?? {}
    cachedPlayers = data.players ?? {}
  })()
  await fetchPromise
  fetchPromise = null
}

async function saveNickname(type: 'team' | 'player', name: string, nickname: string): Promise<void> {
  const resp = await fetch(`/api/nicknames/${type}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, nickname }),
  })
  if (!resp.ok) {
    const err = await resp.json().catch(() => ({}))
    throw new Error((err as any).error ?? 'save failed')
  }
  // Update local cache
  if (type === 'team') {
    cachedTeams = { ...cachedTeams, [name]: nickname }
  } else {
    cachedPlayers = { ...cachedPlayers, [name]: nickname }
  }
}

export default function useNicknames() {
  const [teamNicknames, setTeamNicknames] = useState<NicknameDict>(cachedTeams ?? {})
  const [playerNicknames, setPlayerNicknames] = useState<NicknameDict>(cachedPlayers ?? {})
  const [loading, setLoading] = useState(!cachedTeams)

  useEffect(() => {
    ensureLoaded().then(() => {
      setTeamNicknames(cachedTeams!)
      setPlayerNicknames(cachedPlayers!)
      setLoading(false)
    })
  }, [])

  const saveTeamNickname = useCallback(async (name: string, nickname: string) => {
    await saveNickname('team', name, nickname)
    setTeamNicknames({ ...cachedTeams! })
  }, [])

  const savePlayerNickname = useCallback(async (name: string, nickname: string) => {
    await saveNickname('player', name, nickname)
    setPlayerNicknames({ ...cachedPlayers! })
  }, [])

  return { teamNicknames, playerNicknames, saveTeamNickname, savePlayerNickname, loading }
}
