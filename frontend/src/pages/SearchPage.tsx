import SearchableList from '../components/SearchableList'
import { api } from '../api/client'

type SearchPageProps = {
  type: 'team' | 'player'
  placeholder: string
  emptyHint: string
}

export default function SearchPage({ type, placeholder, emptyHint }: SearchPageProps) {
  return (
    <SearchableList
      key={type}
      type={type}
      placeholder={placeholder}
      emptyHint={emptyHint}
      apiSearch={q => api.search(q, type)}
    />
  )
}
