package storage

import (
	"database/sql"
	"log"
	"time"
)

func migrate(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_version (
		version     INTEGER PRIMARY KEY,
		applied_at  TEXT DEFAULT (datetime('now'))
	)`)
	if err != nil {
		return err
	}

	var v int
	if err := db.QueryRow("SELECT COALESCE(MAX(version),0) FROM schema_version").Scan(&v); err != nil {
		return err
	}
	if v < 1 {
		if err := applyV1(db); err != nil {
			return err
		}
	}
	return nil
}

func applyV1(db *sql.DB) error {
	ddl := `
	CREATE TABLE IF NOT EXISTS teams (
		id                  INTEGER PRIMARY KEY,
		name                TEXT NOT NULL,
		slug                TEXT,
		country             TEXT,
		rank                INTEGER,
		stats_json          TEXT,
		achievements_json   TEXT,
		roster_json         TEXT,
		highlights_json     TEXT,
		recent_matches_json TEXT,
		fetched_at          TEXT,
		updated_at          TEXT
	);

	CREATE TABLE IF NOT EXISTS players (
		id                    INTEGER PRIMARY KEY,
		name                  TEXT NOT NULL,
		slug                  TEXT,
		real_name             TEXT,
		country               TEXT,
		age                   INTEGER,
		team                  TEXT,
		rating_json           TEXT,
		career_json           TEXT,
		abilities_json        TEXT,
		overview_json         TEXT,
		honors_json           TEXT,
		recent_matches_json   TEXT,
		top20_json            TEXT,
		fetched_at            TEXT,
		updated_at            TEXT
	);

	CREATE TABLE IF NOT EXISTS matches (
		match_id     INTEGER PRIMARY KEY,
		team1        TEXT,
		team2        TEXT,
		team1_id     INTEGER,
		team2_id     INTEGER,
		opponent     TEXT,
		opponent_id  INTEGER,
		event        TEXT,
		score        TEXT,
		result       TEXT,
		winner       TEXT,
		best_of      TEXT,
		scheduled_at TEXT,
		played_at    TEXT,
		map_text     TEXT,
		fetched_at   TEXT,
		updated_at   TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_matches_scheduled ON matches(scheduled_at);
	CREATE INDEX IF NOT EXISTS idx_matches_played ON matches(played_at);

	CREATE TABLE IF NOT EXISTS news (
		url_hash     TEXT PRIMARY KEY,
		title        TEXT NOT NULL,
		link         TEXT,
		published_at TEXT,
		tag          TEXT,
		body_text    TEXT,
		author       TEXT,
		fetched_at   TEXT
	);

	CREATE TABLE IF NOT EXISTS realtime_news (
		url_hash      TEXT PRIMARY KEY,
		section       TEXT,
		category      TEXT,
		title         TEXT NOT NULL,
		link          TEXT,
		relative_time TEXT,
		comments      TEXT,
		fetched_at    TEXT
	);
	`
	if _, err := db.Exec(ddl); err != nil {
		return err
	}
	_, err := db.Exec("INSERT INTO schema_version(version) VALUES(1)")
	return err
}

func runCleanup(db *sql.DB, retainMatches, retainNews, retainRealtime int) {
	cutoffMatches := time.Now().UTC().AddDate(0, 0, -retainMatches).Format("2006-01-02")
	cutoffNews := time.Now().UTC().AddDate(0, 0, -retainNews).Format("2006-01-02")
	cutoffRealtime := time.Now().UTC().AddDate(0, 0, -retainRealtime).Format("2006-01-02")

	if _, err := db.Exec("DELETE FROM matches WHERE played_at < ? AND scheduled_at < ?", cutoffMatches, cutoffMatches); err != nil {
		log.Printf("storage: cleanup matches: %v", err)
	}
	if _, err := db.Exec("DELETE FROM news WHERE fetched_at < ?", cutoffNews); err != nil {
		log.Printf("storage: cleanup news: %v", err)
	}
	if _, err := db.Exec("DELETE FROM realtime_news WHERE fetched_at < ?", cutoffRealtime); err != nil {
		log.Printf("storage: cleanup realtime_news: %v", err)
	}
	log.Printf("storage: cleanup complete (matches>%dd news>%dd realtime>%dd)", retainMatches, retainNews, retainRealtime)
}

func startCleanupLoop(db *sql.DB, retainMatches, retainNews, retainRealtime int, stop <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				runCleanup(db, retainMatches, retainNews, retainRealtime)
			case <-stop:
				return
			}
		}
	}()
}
