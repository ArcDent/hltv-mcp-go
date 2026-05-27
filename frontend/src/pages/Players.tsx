import SearchableList from '../components/SearchableList'
import { api } from '../api/client'

export default function Players() {
  return (
    <SearchableList
      type="player"
      placeholder="搜索选手 — 如 ZywOo、载物、s1mple"
      emptyHint="输入选手名开始搜索"
      apiSearch={q => api.search(q, 'player')}
    />
  )
}
