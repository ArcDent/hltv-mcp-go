import SearchableList from '../components/SearchableList'
import { api } from '../api/client'

export default function Teams() {
  return (
    <SearchableList
      type="team"
      placeholder="搜索队伍 — 支持英文 / 中文 / 别名（如 Spirit、绿龙、小蜜蜂）"
      emptyHint="输入队名开始搜索"
      apiSearch={q => api.search(q, 'team')}
    />
  )
}
