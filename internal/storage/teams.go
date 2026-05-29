package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/arcdent/hltv-mcp/internal/types"
)

func (s *Store) UpsertTeam(td types.TeamDetail) error {
	statsJSON, _ := json.Marshal(td.Stats)
	achJSON, _ := json.Marshal(td.Achievements)
	rosterJSON, _ := json.Marshal(td.Roster)
	hlJSON, _ := json.Marshal(td.Highlights)
	matchesJSON, _ := json.Marshal(td.RecentMatches)
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := s.db.Exec(`
		INSERT INTO teams (id, name, slug, country, rank,
			stats_json, achievements_json, roster_json, highlights_json, recent_matches_json,
			fetched_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name=excluded.name, slug=excluded.slug, country=excluded.country, rank=excluded.rank,
			stats_json=excluded.stats_json, achievements_json=excluded.achievements_json,
			roster_json=excluded.roster_json, highlights_json=excluded.highlights_json,
			recent_matches_json=excluded.recent_matches_json,
			updated_at=excluded.updated_at`,
		td.Profile.ID, td.Profile.Name, td.Profile.Slug, td.Profile.Country, td.Ranking.WorldRank,
		string(statsJSON), string(achJSON), string(rosterJSON), string(hlJSON), string(matchesJSON),
		now, now)
	if err != nil {
		return fmt.Errorf("upsert team %d: %w", td.Profile.ID, err)
	}
	return nil
}

func (s *Store) GetTeam(id int) (types.TeamDetail, bool, error) {
	row := s.db.QueryRow(`
		SELECT id, name, slug, country, rank,
			stats_json, achievements_json, roster_json, highlights_json, recent_matches_json, fetched_at
		FROM teams WHERE id=?`, id)

	var td types.TeamDetail
	var statsJSON, achJSON, rosterJSON, hlJSON, matchesJSON sql.NullString
	var rank int
	var fetchedAt string

	err := row.Scan(&td.Profile.ID, &td.Profile.Name, &td.Profile.Slug, &td.Profile.Country,
		&rank, &statsJSON, &achJSON, &rosterJSON, &hlJSON, &matchesJSON, &fetchedAt)
	if err == sql.ErrNoRows {
		return td, false, nil
	}
	if err != nil {
		return td, false, fmt.Errorf("get team %d: %w", id, err)
	}

	td.Ranking.WorldRank = rank
	if statsJSON.Valid { json.Unmarshal([]byte(statsJSON.String), &td.Stats) }
	if achJSON.Valid { json.Unmarshal([]byte(achJSON.String), &td.Achievements) }
	if rosterJSON.Valid { json.Unmarshal([]byte(rosterJSON.String), &td.Roster) }
	if hlJSON.Valid { json.Unmarshal([]byte(hlJSON.String), &td.Highlights) }
	if matchesJSON.Valid { json.Unmarshal([]byte(matchesJSON.String), &td.RecentMatches) }

	return td, true, nil
}
