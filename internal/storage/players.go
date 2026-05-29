package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/arcdent/hltv-mcp/internal/types"
)

func (s *Store) UpsertPlayer(pd types.PlayerDetail) error {
	ratingJSON, _ := json.Marshal(pd.Rating)
	careerJSON, _ := json.Marshal(pd.Career)
	abilitiesJSON, _ := json.Marshal(pd.Abilities)
	overviewJSON, _ := json.Marshal(pd.Summary)
	honorsJSON, _ := json.Marshal(pd.Honors)
	matchesJSON, _ := json.Marshal(pd.RecentMatches)
	top20JSON, _ := json.Marshal(pd.Top20Ranks)
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := s.db.Exec(`
		INSERT INTO players (id, name, slug, real_name, country, age, team,
			rating_json, career_json, abilities_json, overview_json,
			honors_json, recent_matches_json, top20_json,
			fetched_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name=excluded.name, slug=excluded.slug, real_name=excluded.real_name,
			country=excluded.country, age=excluded.age, team=excluded.team,
			rating_json=excluded.rating_json, career_json=excluded.career_json,
			abilities_json=excluded.abilities_json, overview_json=excluded.overview_json,
			honors_json=excluded.honors_json, recent_matches_json=excluded.recent_matches_json,
			top20_json=excluded.top20_json, updated_at=excluded.updated_at`,
		pd.Profile.ID, pd.Profile.Name, pd.Profile.Slug, pd.Profile.RealName,
		pd.Profile.Country, pd.Profile.Age, pd.Profile.Team,
		string(ratingJSON), string(careerJSON), string(abilitiesJSON),
		string(overviewJSON), string(honorsJSON), string(matchesJSON),
		string(top20JSON), now, now)
	if err != nil {
		return fmt.Errorf("upsert player %d: %w", pd.Profile.ID, err)
	}
	return nil
}

func (s *Store) GetPlayer(id int) (types.PlayerDetail, bool, error) {
	row := s.db.QueryRow(`
		SELECT id, name, slug, real_name, country, age, team,
			rating_json, career_json, abilities_json, overview_json,
			honors_json, recent_matches_json, top20_json, fetched_at
		FROM players WHERE id=?`, id)

	var pd types.PlayerDetail
	var ratingJSON, careerJSON, abilitiesJSON, overviewJSON sql.NullString
	var honorsJSON, matchesJSON, top20JSON sql.NullString
	var realName, country, team sql.NullString
	var age sql.NullInt64
	var fetchedAt string

	err := row.Scan(&pd.Profile.ID, &pd.Profile.Name, &pd.Profile.Slug,
		&realName, &country, &age, &team,
		&ratingJSON, &careerJSON, &abilitiesJSON, &overviewJSON,
		&honorsJSON, &matchesJSON, &top20JSON, &fetchedAt)
	if err == sql.ErrNoRows {
		return pd, false, nil
	}
	if err != nil {
		return pd, false, fmt.Errorf("get player %d: %w", id, err)
	}

	pd.Profile.RealName = realName.String
	pd.Profile.Country = country.String
	pd.Profile.Age = int(age.Int64)
	pd.Profile.Team = team.String
	if ratingJSON.Valid { json.Unmarshal([]byte(ratingJSON.String), &pd.Rating) }
	if careerJSON.Valid { json.Unmarshal([]byte(careerJSON.String), &pd.Career) }
	if abilitiesJSON.Valid { json.Unmarshal([]byte(abilitiesJSON.String), &pd.Abilities) }
	if overviewJSON.Valid { json.Unmarshal([]byte(overviewJSON.String), &pd.Summary) }
	if honorsJSON.Valid { json.Unmarshal([]byte(honorsJSON.String), &pd.Honors) }
	if matchesJSON.Valid { json.Unmarshal([]byte(matchesJSON.String), &pd.RecentMatches) }
	if top20JSON.Valid { json.Unmarshal([]byte(top20JSON.String), &pd.Top20Ranks) }

	return pd, true, nil
}
