package storage

import (
	"fmt"
	"time"

	"github.com/arcdent/hltv-mcp/internal/types"
)

// BatchUpsertMatches inserts or updates matches.
// COALESCE semantic: empty fields do NOT overwrite existing non-empty values.
// This ensures a results scrape doesn't wipe scheduled_at from an upcoming scrape.
func (s *Store) BatchUpsertMatches(matches []types.NormalizedMatch) error {
	if len(matches) == 0 {
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339)

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("batch upsert matches: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO matches (match_id, team1, team2, team1_id, team2_id,
			opponent, opponent_id, event, score, result, winner, best_of,
			scheduled_at, played_at, map_text, fetched_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(match_id) DO UPDATE SET
			team1=COALESCE(NULLIF(excluded.team1,''), matches.team1),
			team2=COALESCE(NULLIF(excluded.team2,''), matches.team2),
			team1_id=COALESCE(NULLIF(excluded.team1_id,0), matches.team1_id),
			team2_id=COALESCE(NULLIF(excluded.team2_id,0), matches.team2_id),
			event=COALESCE(NULLIF(excluded.event,''), matches.event),
			score=COALESCE(NULLIF(excluded.score,''), matches.score),
			result=COALESCE(NULLIF(excluded.result,''), matches.result),
			winner=COALESCE(NULLIF(excluded.winner,''), matches.winner),
			best_of=COALESCE(NULLIF(excluded.best_of,''), matches.best_of),
			scheduled_at=COALESCE(NULLIF(excluded.scheduled_at,''), matches.scheduled_at),
			played_at=COALESCE(NULLIF(excluded.played_at,''), matches.played_at),
			map_text=COALESCE(NULLIF(excluded.map_text,''), matches.map_text),
			updated_at=excluded.updated_at`)
	if err != nil {
		return fmt.Errorf("batch upsert matches: prepare: %w", err)
	}
	defer stmt.Close()

	for _, m := range matches {
		if m.MatchID <= 0 {
			continue
		}
		_, err := stmt.Exec(m.MatchID, m.Team1, m.Team2, m.Team1ID, m.Team2ID,
			m.Opponent, m.OpponentID, m.Event, m.Score, string(m.Result), m.Winner,
			m.BestOf, m.ScheduledAt, m.PlayedAt, m.MapText, now, now)
		if err != nil {
			return fmt.Errorf("batch upsert matches: exec id=%d: %w", m.MatchID, err)
		}
	}
	return tx.Commit()
}

// QueryMatchesByTime returns matches filtered by time category.
// category: "upcoming" (scheduled_at >= today), "today" (scheduled_at starts today),
// "results" (played_at < today).
func (s *Store) QueryMatchesByTime(category string, limit int) ([]types.NormalizedMatch, error) {
	today := time.Now().UTC().Format("2006-01-02")

	var query string
	var args []any

	switch category {
	case "upcoming":
		query = `SELECT match_id, team1, team2, team1_id, team2_id,
			event, score, result, winner, best_of, scheduled_at, played_at, map_text
			FROM matches WHERE scheduled_at >= ?
			ORDER BY scheduled_at ASC`
		args = []any{today}
	case "today":
		query = `SELECT match_id, team1, team2, team1_id, team2_id,
			event, score, result, winner, best_of, scheduled_at, played_at, map_text
			FROM matches WHERE scheduled_at LIKE ?
			ORDER BY scheduled_at ASC`
		args = []any{today + "%"}
	case "results":
		query = `SELECT match_id, team1, team2, team1_id, team2_id,
			event, score, result, winner, best_of, scheduled_at, played_at, map_text
			FROM matches WHERE played_at < ?
			ORDER BY played_at DESC`
		args = []any{today}
	default:
		query = `SELECT match_id, team1, team2, team1_id, team2_id,
			event, score, result, winner, best_of, scheduled_at, played_at, map_text
			FROM matches ORDER BY COALESCE(scheduled_at, played_at) DESC`
	}

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query matches by time: %w", err)
	}
	defer rows.Close()

	var matches []types.NormalizedMatch
	for rows.Next() {
		var m types.NormalizedMatch
		var resultStr string
		if err := rows.Scan(&m.MatchID, &m.Team1, &m.Team2, &m.Team1ID, &m.Team2ID,
			&m.Event, &m.Score, &resultStr, &m.Winner,
			&m.BestOf, &m.ScheduledAt, &m.PlayedAt, &m.MapText); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		m.Result = types.MatchOutcome(resultStr)
		matches = append(matches, m)
	}
	if matches == nil {
		matches = []types.NormalizedMatch{}
	}
	return matches, rows.Err()
}
