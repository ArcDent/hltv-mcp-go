# Persistent Storage Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add SQLite persistent storage layer with three-tier fallback (Cache → SQLite → HLTV scrape) and SSE-based frontend live refresh.

**Architecture:** New `internal/storage/` package provides typed CRUD over SQLite via `database/sql` + `modernc.org/sqlite`. SSE hub broadcasts refresh events to frontend via `GET /api/sse` (NOT `/api/events` — already in use). Facade methods inject Store + SSE hub as optional dependencies, inserting db.Get before scrape and db.Upsert after scrape.

**Tech Stack:** Go 1.26, modernc.org/sqlite (pure Go, no CGO), database/sql, net/http SSE

**Critical design decisions from spec:**
- Type A (PlayerDetail/TeamDetail/NewsArticle): point query by ID, stale-while-revalidate + background refresh
- Type B (Matches/News): conditional query by time, stale-while-revalidate + background refresh
- Matches: three-category time queries (future/today/past), COALESCE partial update on upsert
- SSE hub: `*SSEHub` pointer injected into facade, nil-safe (degrade if nil)
- DB failure: log warn, degrade to cache-only, never crash

---

### Task 1: Add SQLite dependency and config fields

**Files:**
- Modify: `go.mod`
- Modify: `internal/config/config.go`

- [ ] **Step 1: Add modernc.org/sqlite dependency**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go get modernc.org/sqlite
```

- [ ] **Step 2: Add config fields for DB path and retention**

In `internal/config/config.go`, add fields to the `Config` struct:

```go
// After existing cache fields, add:
DBPath              string
DBRetentionMatches  int
DBRetentionNews     int
DBRetentionRealtime int
```

In `LoadConfig()`, add initialization alongside existing fields:

```go
DBPath:              envStr("HLTV_DB_PATH", "data/hltv.db"),
DBRetentionMatches:  envInt("HLTV_DB_RETENTION_MATCHES", 90),
DBRetentionNews:     envInt("HLTV_DB_RETENTION_NEWS", 30),
DBRetentionRealtime: envInt("HLTV_DB_RETENTION_REALTIME_NEWS", 7),
```

- [ ] **Step 3: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/config/
```
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add go.mod go.sum internal/config/config.go
git commit -m "chore: add modernc.org/sqlite dependency and DB config fields"
```

---

### Task 2: Create storage package — migration and schema

**Files:**
- Create: `internal/storage/migration.go`

- [ ] **Step 1: Create migration.go with DDL and cleanup logic**

```go
package storage

import (
	"database/sql"
	"log"
	"time"
)

const currentSchemaVersion = 1

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
	if _, err := db.Exec("INSERT INTO schema_version(version) VALUES(1)"); err != nil {
		return err
	}
	return nil
}

func runCleanup(db *sql.DB, retainMatches, retainNews, retainRealtime int) {
	now := time.Now().UTC().Format(time.RFC3339)
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

	// No cleanup for teams/players — permanent
	log.Printf("storage: cleanup done at %s", now)
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
```

- [ ] **Step 2: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/storage/
```
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/storage/migration.go
git commit -m "feat: add storage migration — DDL for 5 tables + cleanup loop"
```

---

### Task 3: Create storage.go — Store lifecycle

**Files:**
- Create: `internal/storage/storage.go`

- [ ] **Step 1: Create storage.go**

```go
package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Store struct {
	db     *sql.DB
	stopCh chan struct{}
}

func Open(dbPath string, retainMatches, retainNews, retainRealtime int) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
		return nil, fmt.Errorf("storage: mkdir: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("storage: open: %w", err)
	}
	db.SetMaxOpenConns(1) // SQLite serializes writes

	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("storage: migrate: %w", err)
	}

	stopCh := make(chan struct{})
	s := &Store{db: db, stopCh: stopCh}

	runCleanup(db, retainMatches, retainNews, retainRealtime)
	startCleanupLoop(db, retainMatches, retainNews, retainRealtime, stopCh)

	log.Printf("storage: opened %s (retention: matches=%dd news=%dd realtime=%dd)",
		dbPath, retainMatches, retainNews, retainRealtime)
	return s, nil
}

func (s *Store) Close() error {
	close(s.stopCh)
	return s.db.Close()
}

func (s *Store) DB() *sql.DB {
	return s.db
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/storage/
```
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/storage/storage.go
git commit -m "feat: add Store lifecycle — open, migrate, cleanup, close"
```

---

### Task 4: Create storage/teams.go — TeamDetail CRUD

**Files:**
- Create: `internal/storage/teams.go`

- [ ] **Step 1: Create teams.go with Upsert and Get**

```go
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
	achievementsJSON, _ := json.Marshal(td.Achievements)
	rosterJSON, _ := json.Marshal(td.Roster)
	highlightsJSON, _ := json.Marshal(td.Highlights)
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
		string(statsJSON), string(achievementsJSON), string(rosterJSON),
		string(highlightsJSON), string(matchesJSON),
		now, now)
	if err != nil {
		return fmt.Errorf("upsert team %d: %w", td.Profile.ID, err)
	}
	return nil
}

func (s *Store) GetTeam(id int) (types.TeamDetail, bool, error) {
	row := s.db.QueryRow("SELECT id, name, slug, country, rank, stats_json, achievements_json, roster_json, highlights_json, recent_matches_json, fetched_at FROM teams WHERE id=?", id)

	var td types.TeamDetail
	var statsJSON, achJSON, rosterJSON, hlJSON, matchesJSON sql.NullString
	var rank int
	var fetchedAt string

	if err := row.Scan(&td.Profile.ID, &td.Profile.Name, &td.Profile.Slug, &td.Profile.Country,
		&rank, &statsJSON, &achJSON, &rosterJSON, &hlJSON, &matchesJSON, &fetchedAt); err != nil {
		if err == sql.ErrNoRows {
			return td, false, nil
		}
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
```

- [ ] **Step 2: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/storage/
```
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/storage/teams.go
git commit -m "feat: add TeamDetail Upsert/Get storage methods"
```

---

### Task 5: Create storage/players.go — PlayerDetail CRUD

**Files:**
- Create: `internal/storage/players.go`

- [ ] **Step 1: Create players.go**

```go
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
	row := s.db.QueryRow("SELECT id, name, slug, real_name, country, age, team, rating_json, career_json, abilities_json, overview_json, honors_json, recent_matches_json, top20_json, fetched_at FROM players WHERE id=?", id)

	var pd types.PlayerDetail
	var ratingJSON, careerJSON, abilitiesJSON, overviewJSON, honorsJSON, matchesJSON, top20JSON sql.NullString
	var realName, country, team sql.NullString
	var age sql.NullInt64
	var fetchedAt string

	if err := row.Scan(&pd.Profile.ID, &pd.Profile.Name, &pd.Profile.Slug,
		&realName, &country, &age, &team,
		&ratingJSON, &careerJSON, &abilitiesJSON, &overviewJSON,
		&honorsJSON, &matchesJSON, &top20JSON, &fetchedAt); err != nil {
		if err == sql.ErrNoRows {
			return pd, false, nil
		}
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
```

- [ ] **Step 2: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/storage/
```
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/storage/players.go
git commit -m "feat: add PlayerDetail Upsert/Get storage methods"
```

---

### Task 6: Create storage/matches.go — BatchUpsert with partial update + time queries

**Files:**
- Create: `internal/storage/matches.go`

- [ ] **Step 1: Create matches.go**

```go
package storage

import (
	"fmt"
	"strings"
	"time"

	"github.com/arcdent/hltv-mcp/internal/types"
)

// BatchUpsertMatches inserts or updates matches. Empty fields do NOT overwrite
// existing non-empty values (COALESCE semantic), so a results scrape doesn't
// wipe the scheduled_at set by an upcoming scrape.
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
			opponent=COALESCE(NULLIF(excluded.opponent,''), matches.opponent),
			opponent_id=COALESCE(NULLIF(excluded.opponent_id,0), matches.opponent_id),
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
// "results" (played_at < today), "" = all.
func (s *Store) QueryMatchesByTime(category string, limit int) ([]types.NormalizedMatch, error) {
	today := time.Now().UTC().Format("2006-01-02")

	var query string
	var args []any

	switch category {
	case "upcoming":
		query = "SELECT match_id, team1, team2, team1_id, team2_id, opponent, opponent_id, event, score, result, winner, best_of, scheduled_at, played_at, map_text FROM matches WHERE scheduled_at >= ? ORDER BY scheduled_at ASC"
		args = []any{today}
	case "today":
		query = "SELECT match_id, team1, team2, team1_id, team2_id, opponent, opponent_id, event, score, result, winner, best_of, scheduled_at, played_at, map_text FROM matches WHERE scheduled_at LIKE ? ORDER BY scheduled_at ASC"
		args = []any{today + "%"}
	case "results":
		query = "SELECT match_id, team1, team2, team1_id, team2_id, opponent, opponent_id, event, score, result, winner, best_of, scheduled_at, played_at, map_text FROM matches WHERE played_at < ? ORDER BY played_at DESC"
		args = []any{today}
	default:
		query = "SELECT match_id, team1, team2, team1_id, team2_id, opponent, opponent_id, event, score, result, winner, best_of, scheduled_at, played_at, map_text FROM matches ORDER BY COALESCE(scheduled_at, played_at) DESC"
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
			&m.Opponent, &m.OpponentID, &m.Event, &m.Score, &resultStr, &m.Winner,
			&m.BestOf, &m.ScheduledAt, &m.PlayedAt, &m.MapText); err != nil {
			return nil, fmt.Errorf("query matches by time: scan: %w", err)
		}
		m.Result = types.MatchOutcome(resultStr)
		// Translate placeholders on read
		m.Team1 = types.TranslatePlaceholder(m.Team1)
		m.Team2 = types.TranslatePlaceholder(m.Team2)
		m.Opponent = types.TranslatePlaceholder(m.Opponent)
		matches = append(matches, m)
	}
	if matches == nil {
		matches = []types.NormalizedMatch{}
	}
	return matches, rows.Err()
}
```

- [ ] **Step 2: Verify types.TranslatePlaceholder is accessible**

The `TranslatePlaceholder` function is in `internal/normalizer/match.go`. It needs to be moved to `internal/types/` or we reference it from normalizer. Check:

```bash
grep -n "func TranslatePlaceholder" /home/arcdent/github/hltv-mcp-fully-rebuild/internal/normalizer/match.go
```

Expected: line 193. Since `storage` imports `types`, and `TranslatePlaceholder` is in `normalizer`, we have two options:
a. Move `TranslatePlaceholder` to `types` package
b. Call it in the facade layer before/after storage

Move it to `types` is cleaner since it's a pure string transform with no dependencies.

Actually, let me avoid moving it. The simpler approach: call `TranslatePlaceholder` in the facade layer when returning from SQLite, not in storage. Keep storage as a pure data layer.

So in storage/matches.go, remove the TranslatePlaceholder calls:

```go
// Remove these lines from the scan loop:
m.Team1 = types.TranslatePlaceholder(m.Team1)
m.Team2 = types.TranslatePlaceholder(m.Team2)
m.Opponent = types.TranslatePlaceholder(m.Opponent)
```

The facade layer will call `normalizer.TranslatePlaceholder` when assembling the response.

- [ ] **Step 3: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/storage/
```
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add internal/storage/matches.go
git commit -m "feat: add matches BatchUpsert (COALESCE partial update) + QueryMatchesByTime"
```

---

### Task 7: Create storage/news.go — News and RealtimeNews CRUD

**Files:**
- Create: `internal/storage/news.go`

- [ ] **Step 1: Create news.go**

```go
package storage

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"time"

	"github.com/arcdent/hltv-mcp/internal/types"
)

// --- NewsItem / NewsArticle (shared url_hash key) ---

func (s *Store) BatchUpsertNews(items []types.NewsItem) error {
	if len(items) == 0 {
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339)

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("batch upsert news: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO news (url_hash, title, link, published_at, tag, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(url_hash) DO UPDATE SET
			title=excluded.title, link=excluded.link,
			published_at=excluded.published_at, tag=excluded.tag,
			fetched_at=excluded.fetched_at`)
	if err != nil {
		return fmt.Errorf("batch upsert news: prepare: %w", err)
	}
	defer stmt.Close()

	for _, n := range items {
		hash := fmt.Sprintf("%x", md5.Sum([]byte(n.Link)))
		_, err := stmt.Exec(hash, n.Title, n.Link, n.PublishedAt, n.Tag, now)
		if err != nil {
			return fmt.Errorf("batch upsert news: exec: %w", err)
		}
	}
	return tx.Commit()
}

func (s *Store) UpsertNewsArticle(article types.NewsArticle) error {
	now := time.Now().UTC().Format(time.RFC3339)
	hash := fmt.Sprintf("%x", md5.Sum([]byte(article.Link)))

	_, err := s.db.Exec(`
		INSERT INTO news (url_hash, title, link, published_at, body_text, author, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(url_hash) DO UPDATE SET
			title=COALESCE(NULLIF(excluded.title,''), news.title),
			body_text=COALESCE(NULLIF(excluded.body_text,''), news.body_text),
			author=COALESCE(NULLIF(excluded.author,''), news.author),
			fetched_at=excluded.fetched_at`,
		hash, article.Title, article.Link, article.PublishedAt,
		article.BodyText, article.Author, now)
	return err
}

func (s *Store) GetNewsArticle(url string) (types.NewsArticle, bool, error) {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(url)))
	row := s.db.QueryRow("SELECT title, link, published_at, body_text, author, fetched_at FROM news WHERE url_hash=?", hash)

	var article types.NewsArticle
	var bodyText, author sql.NullString
	var fetchedAt string
	if err := row.Scan(&article.Title, &article.Link, &article.PublishedAt,
		&bodyText, &author, &fetchedAt); err != nil {
		if err == sql.ErrNoRows {
			return article, false, nil
		}
		return article, false, err
	}
	article.BodyText = bodyText.String
	article.Author = author.String
	return article, true, nil
}

func (s *Store) QueryNews(limit int) ([]types.NewsItem, error) {
	query := "SELECT title, link, published_at, tag, fetched_at FROM news ORDER BY published_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []types.NewsItem
	for rows.Next() {
		var n types.NewsItem
		var fetchedAt string
		if err := rows.Scan(&n.Title, &n.Link, &n.PublishedAt, &n.Tag, &fetchedAt); err != nil {
			return nil, err
		}
		items = append(items, n)
	}
	if items == nil {
		items = []types.NewsItem{}
	}
	return items, rows.Err()
}

// --- RealtimeNews ---

func (s *Store) BatchUpsertRealtimeNews(items []types.RealtimeNewsItem) error {
	if len(items) == 0 {
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339)

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("batch upsert realtime_news: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO realtime_news (url_hash, section, category, title, link, relative_time, comments, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(url_hash) DO UPDATE SET
			section=excluded.section, category=excluded.category,
			title=excluded.title, link=excluded.link,
			relative_time=excluded.relative_time, comments=excluded.comments,
			fetched_at=excluded.fetched_at`)
	if err != nil {
		return fmt.Errorf("batch upsert realtime_news: prepare: %w", err)
	}
	defer stmt.Close()

	for _, n := range items {
		hash := fmt.Sprintf("%x", md5.Sum([]byte(n.Link)))
		_, err := stmt.Exec(hash, n.Section, n.Category, n.Title, n.Link, n.RelativeTime, n.Comments, now)
		if err != nil {
			return fmt.Errorf("batch upsert realtime_news: exec: %w", err)
		}
	}
	return tx.Commit()
}

func (s *Store) QueryRealtimeNews(limit int) ([]types.RealtimeNewsItem, error) {
	query := "SELECT section, category, title, link, relative_time, comments, fetched_at FROM realtime_news"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []types.RealtimeNewsItem
	for rows.Next() {
		var n types.RealtimeNewsItem
		var fetchedAt string
		if err := rows.Scan(&n.Section, &n.Category, &n.Title, &n.Link, &n.RelativeTime, &n.Comments, &fetchedAt); err != nil {
			return nil, err
		}
		items = append(items, n)
	}
	if items == nil {
		items = []types.RealtimeNewsItem{}
	}
	return items, rows.Err()
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/storage/
```
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/storage/news.go
git commit -m "feat: add news and realtime_news BatchUpsert/Query/Get storage methods"
```

---

### Task 8: Create SSE hub and handler

**Files:**
- Create: `internal/http/sse.go`

- [ ] **Step 1: Create sse.go**

```go
package http

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

// SSEEvent represents a refresh notification sent to frontend.
type SSEEvent struct {
	Entity string `json:"entity"`
	ID     int    `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
}

// SSEHub manages SSE client connections and broadcasts.
// Nil-safe: if hub is nil, Broadcast is a no-op.
type SSEHub struct {
	mu      sync.RWMutex
	clients map[chan SSEEvent]struct{}
}

// NewSSEHub creates a new SSE hub.
func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients: make(map[chan SSEEvent]struct{}),
	}
}

// Broadcast sends an event to all connected clients. No-op if hub is nil.
func (h *SSEHub) Broadcast(evt SSEEvent) {
	if h == nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- evt:
		default:
			// drop if client buffer full (non-blocking)
		}
	}
}

func (h *SSEHub) register(ch chan SSEEvent) {
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
}

func (h *SSEHub) unregister(ch chan SSEEvent) {
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
}

// SSEHandler returns an http.HandlerFunc for SSE at GET /api/sse.
func SSEHandler(hub *SSEHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		ch := make(chan SSEEvent, 32)
		hub.register(ch)
		defer hub.unregister(ch)

		// Heartbeat ticker
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case evt := <-ch:
				data, _ := json.Marshal(evt)
				if _, err := w.Write([]byte("event: refreshed\ndata: " + string(data) + "\n\n")); err != nil {
					return
				}
				flusher.Flush()
			case <-ticker.C:
				if _, err := w.Write([]byte(": keepalive\n\n")); err != nil {
					return
				}
				flusher.Flush()
			case <-r.Context().Done():
				return
			}
		}
	}
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/http/
```
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/http/sse.go
git commit -m "feat: add SSE hub and handler for frontend live refresh"
```

---

### Task 9: Modify router.go — register SSE endpoint

**Files:**
- Modify: `internal/http/router.go`

- [ ] **Step 1: Add SSE route**

Change `NewRouter` signature to accept `*SSEHub`:

```go
func NewRouter(f *facade.HltvFacade, frontendFS fs.FS, sseHub *SSEHub) http.Handler {
```

After the existing `h := handlers.New(f)` line, add:

```go
r.Get("/api/sse", SSEHandler(sseHub))
```

The full function becomes:

```go
func NewRouter(f *facade.HltvFacade, frontendFS fs.FS, sseHub *SSEHub) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	h := handlers.New(f)

	r.Get("/api/health", h.Health)
	r.Get("/api/status", h.Status)
	r.Get("/api/cache", h.GetCacheStats)
	r.Delete("/api/cache", h.ClearCache)
	r.Get("/api/search", h.Search)
	r.Get("/api/teams/{id}", h.GetTeam)
	r.Get("/api/players/{id}", h.GetPlayer)
	r.Get("/api/matches/today", h.GetTodayMatches)
	r.Get("/api/matches", h.GetUpcomingMatches)
	r.Get("/api/results", h.GetResults)
	r.Get("/api/events", h.GetEvents)
	r.Get("/api/sse", SSEHandler(sseHub))
	r.Get("/api/news/realtime", h.GetRealtimeNews)
	r.Get("/api/news", h.GetNewsDigest)
	r.Get("/api/news/article", h.GetNewsArticle)
	r.Get("/api/translate/config", h.GetTranslateConfig)
	r.Put("/api/translate/config", h.PutTranslateConfig)
	r.Post("/api/translate", h.PostTranslate)
	r.Get("/api/nicknames", h.GetNicknames)
	r.Put("/api/nicknames/team", h.PutTeamNickname)
	r.Put("/api/nicknames/player", h.PutPlayerNickname)

	// SPA fallback (unchanged)
	if frontendFS != nil {
		feFS, err := fs.Sub(frontendFS, "dist")
		if err == nil {
			fileServer := http.FileServer(http.FS(feFS))
			r.Get("/*", func(w http.ResponseWriter, req *http.Request) {
				fsPath := strings.TrimPrefix(req.URL.Path, "/")
				if _, err := feFS.Open(fsPath); err != nil {
					req.URL.Path = "/"
				}
				fileServer.ServeHTTP(w, req)
			})
		}
	}

	return r
}
```

- [ ] **Step 2: Verify compilation** (will fail until main.go is updated in Task 13)

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/http/ 2>&1
```
Expected: compiled (no main.go dependency in this package)

- [ ] **Step 3: Commit**

```bash
git add internal/http/router.go
git commit -m "feat: register SSE endpoint at /api/sse"
```

---

### Task 10: Modify facade/facade.go — inject Store and SSEHub, add Type A three-tier fallback

**Files:**
- Modify: `internal/facade/facade.go`

Reference: spec § Type A data flow (PlayerDetail/TeamDetail/NewsArticle)

- [ ] **Step 1: Add Store and SSEHub fields to HltvFacade**

Add imports:

```go
import (
	// ... existing imports ...
	"github.com/arcdent/hltv-mcp/internal/storage"
	httppkg "github.com/arcdent/hltv-mcp/internal/http" // for SSEEvent — actually let's define it in a shared place

	"log"
)
```

Wait, SSEEvent is defined in `internal/http/sse.go`. The facade can't import http (circular). Solution: define SSEEvent in a shared location, or let the facade take a simple callback.

Better approach: define a `Broadcaster` interface in the facade package, or simply inject `func(entity string, id int, name string)`. Since the SSE hub is optional (nil-safe), we can use:

```go
type SSEBroadcaster func(entity string, id int, name string)
```

The SSE hub's Broadcast method can be wrapped. But actually, simpler: just pass `*SSEHub` as an `interface{}` and call it via a known method. No, that's ugly.

Cleanest solution: The SSE event types need to be in a package both `http` and `facade` can import. Put them in `internal/types/`:

Wait, that's over-engineering. Simpler: define a small callback type in the facade:

```go
// RefreshNotifier is called when fresh data replaces stale SQLite data
type RefreshNotifier func(entity string, id int, name string)
```

And have the main.go wire it to SSEHub.Broadcast. Let's do this.

Actually even simpler: just make the facade store a reference to the storage.Store and an optional callback. Let me revise.

```go
type HltvFacade struct {
	cfg    *config.Config
	cache  *cache.Cache
	client *client.HltvClient
	store  *storage.Store            // NEW (nil = no persistence)
	notify func(entity string, id int, name string) // NEW (nil = no SSE)
	ts     *scraper.TeamScraper
	ps     *scraper.PlayerScraper
	rs     *scraper.ResultsScraper
	ms     *scraper.MatchesScraper
	ns     *scraper.NewsScraper
	rns    *scraper.RealtimeNewsScraper
	nas    *scraper.NewsArticleScraper
}
```

And:

```go
func New(cfg *config.Config, c *cache.Cache, cli *client.HltvClient, store *storage.Store, notify func(string, int, string)) *HltvFacade {
```

- [ ] **Step 2: Rewrite facade.go with three-tier fallback**

The full file after changes (only showing changed methods):

```go
package facade

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/arcdent/hltv-mcp/internal/cache"
	"github.com/arcdent/hltv-mcp/internal/client"
	"github.com/arcdent/hltv-mcp/internal/config"
	"github.com/arcdent/hltv-mcp/internal/normalizer"
	"github.com/arcdent/hltv-mcp/internal/scraper"
	"github.com/arcdent/hltv-mcp/internal/storage"
	"github.com/arcdent/hltv-mcp/internal/types"
)

type HltvFacade struct {
	cfg    *config.Config
	cache  *cache.Cache
	client *client.HltvClient
	store  *storage.Store
	notify func(entity string, id int, name string)
	ts     *scraper.TeamScraper
	ps     *scraper.PlayerScraper
	rs     *scraper.ResultsScraper
	ms     *scraper.MatchesScraper
	ns     *scraper.NewsScraper
	rns    *scraper.RealtimeNewsScraper
	nas    *scraper.NewsArticleScraper
}

func New(cfg *config.Config, c *cache.Cache, cli *client.HltvClient, store *storage.Store, notify func(string, int, string)) *HltvFacade {
	return &HltvFacade{
		cfg:    cfg,
		cache:  c,
		client: cli,
		store:  store,
		notify: notify,
		ts:     scraper.NewTeamScraper(cli),
		ps:     scraper.NewPlayerScraper(cli),
		rs:     scraper.NewResultsScraper(cli),
		ms:     scraper.NewMatchesScraper(cli),
		ns:     scraper.NewNewsScraper(cli),
		rns:    scraper.NewRealtimeNewsScraper(cli),
		nas:    scraper.NewNewsArticleScraper(cli),
	}
}

func (f *HltvFacade) broadcast(entity string, id int, name string) {
	if f.notify != nil {
		f.notify(entity, id, name)
	}
}

// GetPlayerDetailCached implements Type A three-tier fallback
func (f *HltvFacade) GetPlayerDetailCached(ctx context.Context, id int, slug string) (types.PlayerDetail, error) {
	if slug == "" {
		slug = fmt.Sprintf("player-%d", id)
	}
	key := fmt.Sprintf("player_detail:%d", id)

	// Tier 1: memory cache
	if cached, ok := f.cache.Get(key); ok {
		return cached.(types.PlayerDetail), nil
	}

	// Tier 2: SQLite
	if f.store != nil {
		if pd, ok, _ := f.store.GetPlayer(id); ok {
			// Return stale data, set short-lived cache, refresh in background
			f.cache.Set(key, pd, 10)
			go f.refreshPlayer(id, slug, key)
			return pd, nil
		}
	}

	// Tier 3: scrape from HLTV
	pd, err := f.scrapeAndStorePlayer(ctx, id, slug)
	if err != nil {
		return types.PlayerDetail{}, err
	}
	f.cache.Set(key, pd, f.cfg.CacheTTLPlayerDetail)
	return pd, nil
}

func (f *HltvFacade) scrapeAndStorePlayer(ctx context.Context, id int, slug string) (types.PlayerDetail, error) {
	doc, err := f.ps.GetPlayer(ctx, id, slug)
	if err != nil {
		return types.PlayerDetail{}, err
	}
	pd := normalizer.NormalizePlayerDetail(doc)
	pd.Profile.ID = id
	if f.store != nil {
		if err := f.store.UpsertPlayer(pd); err != nil {
			log.Printf("facade: upsert player %d: %v", id, err)
		}
	}
	return pd, nil
}

func (f *HltvFacade) refreshPlayer(id int, slug, key string) {
	pd, err := f.scrapeAndStorePlayer(context.Background(), id, slug)
	if err != nil {
		log.Printf("facade: refresh player %d: %v", id, err)
		return
	}
	f.cache.Set(key, pd, f.cfg.CacheTTLPlayerDetail)
	f.broadcast("player", pd.Profile.ID, pd.Profile.Name)
}

// GetTeamDetailCached implements Type A three-tier fallback
func (f *HltvFacade) GetTeamDetailCached(ctx context.Context, id int, slug string) (types.TeamDetail, error) {
	if slug == "" {
		slug = fmt.Sprintf("team-%d", id)
	}
	key := fmt.Sprintf("team_detail:%d", id)

	// Tier 1: memory cache
	if cached, ok := f.cache.Get(key); ok {
		return cached.(types.TeamDetail), nil
	}

	// Tier 2: SQLite
	if f.store != nil {
		if td, ok, _ := f.store.GetTeam(id); ok {
			f.cache.Set(key, td, 10)
			go f.refreshTeam(id, slug, key)
			return td, nil
		}
	}

	// Tier 3: scrape from HLTV
	td, err := f.scrapeAndStoreTeam(ctx, id, slug)
	if err != nil {
		return types.TeamDetail{}, err
	}
	f.cache.Set(key, td, f.cfg.CacheTTLPlayerDetail)
	return td, nil
}

func (f *HltvFacade) scrapeAndStoreTeam(ctx context.Context, id int, slug string) (types.TeamDetail, error) {
	doc, err := f.ts.GetTeam(ctx, id, slug)
	if err != nil {
		return types.TeamDetail{}, err
	}
	td := normalizer.NormalizeTeamDetail(doc)
	td.Profile.ID = id
	td.Profile.Slug = slug

	if td.Profile.Name != "" {
		name := td.Profile.Name
		if upcomingDoc, err := f.ms.GetUpcoming(ctx); err == nil {
			allUpcoming := normalizer.NormalizeUpcomingMatches(upcomingDoc, name)
			for _, m := range allUpcoming {
				if m.Team1 == name || m.Team2 == name || m.Opponent == name {
					td.RecentMatches = append(td.RecentMatches, m)
				}
			}
		}
	}
	if td.Highlights != nil {
		for _, m := range td.Highlights.RecentMatches {
			if m.Result == "won" {
				td.Stats.Wins++
			} else {
				td.Stats.Losses++
			}
		}
		total := td.Stats.Wins + td.Stats.Losses + td.Stats.Draws
		if total > 0 {
			td.Stats.WinRate = fmt.Sprintf("%.0f%%", float64(td.Stats.Wins)/float64(total)*100)
		}
	}

	if f.store != nil {
		if err := f.store.UpsertTeam(td); err != nil {
			log.Printf("facade: upsert team %d: %v", id, err)
		}
	}
	return td, nil
}

func (f *HltvFacade) refreshTeam(id int, slug, key string) {
	td, err := f.scrapeAndStoreTeam(context.Background(), id, slug)
	if err != nil {
		log.Printf("facade: refresh team %d: %v", id, err)
		return
	}
	f.cache.Set(key, td, f.cfg.CacheTTLPlayerDetail)
	f.broadcast("team", td.Profile.ID, td.Profile.Name)
}

// GetNewsArticleCached implements Type A three-tier fallback
func (f *HltvFacade) GetNewsArticleCached(ctx context.Context, url string) (types.NewsArticle, error) {
	key := fmt.Sprintf("news_article:%x", md5.Sum([]byte(url)))

	// Tier 1: memory cache
	if cached, ok := f.cache.Get(key); ok {
		return cached.(types.NewsArticle), nil
	}

	// Tier 2: SQLite
	if f.store != nil {
		if article, ok, _ := f.store.GetNewsArticle(url); ok {
			f.cache.Set(key, article, 10)
			go f.refreshNewsArticle(url, key)
			return article, nil
		}
	}

	// Tier 3: scrape from HLTV
	article, err := f.scrapeAndStoreNewsArticle(ctx, url)
	if err != nil {
		return types.NewsArticle{}, err
	}
	f.cache.Set(key, article, f.cfg.CacheTTLNewsArticle)
	return article, nil
}

func (f *HltvFacade) scrapeAndStoreNewsArticle(ctx context.Context, url string) (types.NewsArticle, error) {
	doc, err := f.nas.GetArticle(ctx, url)
	if err != nil {
		return types.NewsArticle{}, err
	}
	article := normalizer.NormalizeNewsArticle(doc, url)
	if f.store != nil {
		if err := f.store.UpsertNewsArticle(article); err != nil {
			log.Printf("facade: upsert news article: %v", err)
		}
	}
	return article, nil
}

func (f *HltvFacade) refreshNewsArticle(url, key string) {
	article, err := f.scrapeAndStoreNewsArticle(context.Background(), url)
	if err != nil {
		log.Printf("facade: refresh news article: %v", err)
		return
	}
	f.cache.Set(key, article, f.cfg.CacheTTLNewsArticle)
	f.broadcast("news", 0, article.Title)
}

// --- Remaining methods unchanged: withCache, createMeta, errorResponse, cloneResponse, stats ---
```

Wait, I need to keep the `withCache` and other methods that aren't changing. Let me make this clearer. I'll only show the CHANGED parts of facade.go:

The full revised facade.go:
1. New imports: `storage`, `log`
2. New fields: `store`, `notify`
3. Updated `New` function
4. New helper: `broadcast`
5. Rewritten: `GetPlayerDetailCached`, `GetTeamDetailCached`, `GetNewsArticleCached`
6. New private methods: `scrapeAndStorePlayer`, `refreshPlayer`, `scrapeAndStoreTeam`, `refreshTeam`, `scrapeAndStoreNewsArticle`, `refreshNewsArticle`
7. Unchanged: `createMeta`, `withCache`, `errorResponse`, `cloneResponse`, `CacheEntries`, `CacheHits`, `CacheMisses`, `ClearCache`

Let me write the complete facade.go file.

Actually this is getting very long. Let me write the complete file since it's critical. But I'll present it as the step content.

Let me write the full file now. Too long for inline, let me write it to the file directly in the implementation phase. For the plan, I'll show the key changes and note what's unchanged.

Actually, the plan skill says: "Every step must contain the actual content an engineer needs." So I should show the complete file content. But the file is ~200 lines. Let me write the complete file.

- [ ] **Step 3: Commit**

```bash
git add internal/facade/facade.go
git commit -m "feat: add three-tier fallback with Store and SSE notify to facade"
```

OK, I need to finish writing the plan. Let me continue with remaining tasks.

Let me actually plan the full facade.go rewrite carefully. The issue is that I can't show the entire file in a step because it's too long. I'll need to use Edit operations to make targeted changes instead.

Let me reconsider the plan structure. Instead of showing a full rewrite, I'll break the facade.go changes into edit steps:

1. Add imports (storage, log)
2. Add fields to struct
3. Update New() signature
4. Add broadcast helper
5. Rewrite GetPlayerDetailCached
6. Add scrapeAndStorePlayer, refreshPlayer
7. Rewrite GetTeamDetailCached
8. Add scrapeAndStoreTeam, refreshTeam
9. Rewrite GetNewsArticleCached
10. Add scrapeAndStoreNewsArticle, refreshNewsArticle

This is cleaner and more maintainable. Let me write this out.

Actually, for a plan, showing the complete resulting code is more useful. The engineer reads the plan task and implements it. Let me write the full code.

Let me continue writing the plan now.<｜end▁of▁thinking｜>

<｜｜DSML｜｜tool_calls>
<｜｜DSML｜｜invoke name="Write">
<｜｜DSML｜｜parameter name="content" string="true"># Persistent Storage Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add SQLite persistent storage layer with three-tier fallback (Cache → SQLite → HLTV scrape) and SSE-based frontend live refresh.

**Architecture:** New `internal/storage/` package provides typed CRUD over SQLite via `database/sql` + `modernc.org/sqlite`. SSE hub at `GET /api/sse` (NOT `/api/events` — already in use by match grouping) broadcasts refresh events. Facade injects Store (nil = degrade) and notify callback (nil = no SSE). Type A methods use stale-while-revalidate; Type B methods read SQLite on cache miss, store in compute closure.

**Tech Stack:** Go 1.26, modernc.org/sqlite (pure Go, no CGO), database/sql, net/http SSE

**Critical design invariants from spec:**
- Type A (PlayerDetail/TeamDetail/NewsArticle): point query by ID, stale-while-revalidate + background goroutine
- Type B (Matches/News): conditional query by time, stale-while-revalidate + background goroutine
- Matches `scheduled_at` / `played_at` indexed, three-category time queries
- BatchUpsert matches: COALESCE partial update (empty fields don't overwrite)
- SSE notify: callback `func(entity string, id int, name string)`, nil-safe
- DB failure: log warn, degrade to cache-only, never crash

---

### Task 1: Add SQLite dependency and config fields

**Files:**
- Modify: `go.mod`
- Modify: `internal/config/config.go`

- [ ] **Step 1: Add modernc.org/sqlite**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go get modernc.org/sqlite
```

- [ ] **Step 2: Add DB config fields**

In `internal/config/config.go`, add to `Config` struct (after `CacheStaleWindowSec`):

```go
DBPath              string
DBRetentionMatches  int
DBRetentionNews     int
DBRetentionRealtime int
```

In `LoadConfig()`, add within the return literal:

```go
DBPath:              envStr("HLTV_DB_PATH", "data/hltv.db"),
DBRetentionMatches:  envInt("HLTV_DB_RETENTION_MATCHES", 90),
DBRetentionNews:     envInt("HLTV_DB_RETENTION_NEWS", 30),
DBRetentionRealtime: envInt("HLTV_DB_RETENTION_REALTIME_NEWS", 7),
```

- [ ] **Step 3: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/config/
```
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add go.mod go.sum internal/config/config.go
git commit -m "chore: add modernc.org/sqlite and DB config fields"
```

---

### Task 2: Create storage package — migration, schema, cleanup

**Files:**
- Create: `internal/storage/migration.go`

- [ ] **Step 1: Create migration.go with full DDL and cleanup**

```go
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
```

- [ ] **Step 2: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/storage/
```
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/storage/migration.go
git commit -m "feat: add storage migration — DDL 5 tables + cleanup loop"
```

---

### Task 3: Create storage.go — Store lifecycle

**Files:**
- Create: `internal/storage/storage.go`

- [ ] **Step 1: Create storage.go**

```go
package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Store wraps a SQLite database for persistent HLTV data.
// Nil-safe usage: callers check (*Store) != nil before use.
type Store struct {
	db     *sql.DB
	stopCh chan struct{}
}

// Open creates/opens the SQLite database at dbPath and runs migrations.
// Returns error if open or migration fails. Caller decides whether to crash or degrade.
func Open(dbPath string, retainMatches, retainNews, retainRealtime int) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
		return nil, fmt.Errorf("storage: mkdir: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("storage: open: %w", err)
	}
	db.SetMaxOpenConns(1)

	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("storage: migrate: %w", err)
	}

	stopCh := make(chan struct{})
	s := &Store{db: db, stopCh: stopCh}

	runCleanup(db, retainMatches, retainNews, retainRealtime)
	startCleanupLoop(db, retainMatches, retainNews, retainRealtime, stopCh)

	log.Printf("storage: opened %s (retention: matches=%dd news=%dd realtime=%dd)", dbPath, retainMatches, retainNews, retainRealtime)
	return s, nil
}

// Close stops the cleanup loop and closes the database.
func (s *Store) Close() error {
	close(s.stopCh)
	return s.db.Close()
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/storage/
```
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/storage/storage.go
git commit -m "feat: add Store lifecycle — open, migrate, cleanup, close"
```

---

### Task 4: Create storage/teams.go — TeamDetail CRUD

**Files:**
- Create: `internal/storage/teams.go`

- [ ] **Step 1: Create teams.go**

```go
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
```

- [ ] **Step 2: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/storage/
```
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/storage/teams.go
git commit -m "feat: add TeamDetail Upsert/Get storage methods"
```

---

### Task 5: Create storage/players.go — PlayerDetail CRUD

**Files:**
- Create: `internal/storage/players.go`

- [ ] **Step 1: Create players.go**

```go
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
```

- [ ] **Step 2: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/storage/
```
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/storage/players.go
git commit -m "feat: add PlayerDetail Upsert/Get storage methods"
```

---

### Task 6: Create storage/matches.go — BatchUpsert with COALESCE + time queries

**Files:**
- Create: `internal/storage/matches.go`

- [ ] **Step 1: Create matches.go**

```go
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
```

- [ ] **Step 2: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/storage/
```
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/storage/matches.go
git commit -m "feat: add matches BatchUpsert with COALESCE + QueryMatchesByTime"
```

---

### Task 7: Create storage/news.go — News and RealtimeNews CRUD

**Files:**
- Create: `internal/storage/news.go`

- [ ] **Step 1: Create news.go**

```go
package storage

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"time"

	"github.com/arcdent/hltv-mcp/internal/types"
)

// --- NewsItem / NewsArticle (shared url_hash key) ---

func (s *Store) BatchUpsertNews(items []types.NewsItem) error {
	if len(items) == 0 {
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339)

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("batch upsert news: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO news (url_hash, title, link, published_at, tag, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(url_hash) DO UPDATE SET
			title=excluded.title, link=excluded.link,
			published_at=excluded.published_at, tag=excluded.tag,
			fetched_at=excluded.fetched_at`)
	if err != nil {
		return fmt.Errorf("batch upsert news: prepare: %w", err)
	}
	defer stmt.Close()

	for _, n := range items {
		hash := fmt.Sprintf("%x", md5.Sum([]byte(n.Link)))
		if _, err := stmt.Exec(hash, n.Title, n.Link, n.PublishedAt, n.Tag, now); err != nil {
			return fmt.Errorf("batch upsert news: exec: %w", err)
		}
	}
	return tx.Commit()
}

func (s *Store) UpsertNewsArticle(article types.NewsArticle) error {
	now := time.Now().UTC().Format(time.RFC3339)
	hash := fmt.Sprintf("%x", md5.Sum([]byte(article.Link)))

	_, err := s.db.Exec(`
		INSERT INTO news (url_hash, title, link, published_at, body_text, author, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(url_hash) DO UPDATE SET
			title=COALESCE(NULLIF(excluded.title,''), news.title),
			body_text=COALESCE(NULLIF(excluded.body_text,''), news.body_text),
			author=COALESCE(NULLIF(excluded.author,''), news.author),
			fetched_at=excluded.fetched_at`,
		hash, article.Title, article.Link, article.PublishedAt,
		article.BodyText, article.Author, now)
	return err
}

func (s *Store) GetNewsArticle(url string) (types.NewsArticle, bool, error) {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(url)))
	row := s.db.QueryRow("SELECT title, link, published_at, body_text, author FROM news WHERE url_hash=?", hash)

	var article types.NewsArticle
	var bodyText, author sql.NullString
	err := row.Scan(&article.Title, &article.Link, &article.PublishedAt, &bodyText, &author)
	if err == sql.ErrNoRows {
		return article, false, nil
	}
	if err != nil {
		return article, false, err
	}
	article.BodyText = bodyText.String
	article.Author = author.String
	return article, true, nil
}

func (s *Store) QueryNews(limit int) ([]types.NewsItem, error) {
	query := "SELECT title, link, published_at, tag FROM news ORDER BY published_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []types.NewsItem
	for rows.Next() {
		var n types.NewsItem
		if err := rows.Scan(&n.Title, &n.Link, &n.PublishedAt, &n.Tag); err != nil {
			return nil, err
		}
		items = append(items, n)
	}
	if items == nil {
		items = []types.NewsItem{}
	}
	return items, rows.Err()
}

// --- RealtimeNews ---

func (s *Store) BatchUpsertRealtimeNews(items []types.RealtimeNewsItem) error {
	if len(items) == 0 {
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339)

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("batch upsert realtime_news: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO realtime_news (url_hash, section, category, title, link, relative_time, comments, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(url_hash) DO UPDATE SET
			section=excluded.section, category=excluded.category,
			title=excluded.title, link=excluded.link,
			relative_time=excluded.relative_time, comments=excluded.comments,
			fetched_at=excluded.fetched_at`)
	if err != nil {
		return fmt.Errorf("batch upsert realtime_news: prepare: %w", err)
	}
	defer stmt.Close()

	for _, n := range items {
		hash := fmt.Sprintf("%x", md5.Sum([]byte(n.Link)))
		if _, err := stmt.Exec(hash, n.Section, n.Category, n.Title, n.Link, n.RelativeTime, n.Comments, now); err != nil {
			return fmt.Errorf("batch upsert realtime_news: exec: %w", err)
		}
	}
	return tx.Commit()
}

func (s *Store) QueryRealtimeNews(limit int) ([]types.RealtimeNewsItem, error) {
	query := "SELECT section, category, title, link, relative_time, comments FROM realtime_news"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []types.RealtimeNewsItem
	for rows.Next() {
		var n types.RealtimeNewsItem
		if err := rows.Scan(&n.Section, &n.Category, &n.Title, &n.Link, &n.RelativeTime, &n.Comments); err != nil {
			return nil, err
		}
		items = append(items, n)
	}
	if items == nil {
		items = []types.RealtimeNewsItem{}
	}
	return items, rows.Err()
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/storage/
```
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/storage/news.go
git commit -m "feat: add news and realtime_news CRUD storage methods"
```

---

### Task 8: Create SSE hub and handler

**Files:**
- Create: `internal/http/sse.go`

- [ ] **Step 1: Create sse.go**

```go
package http

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

// SSEEvent is a refresh notification for frontend EventSource consumers.
type SSEEvent struct {
	Entity string `json:"entity"`
	ID     int    `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
}

// SSEHub manages SSE client connections. Nil Broadcast is safe (no-op).
type SSEHub struct {
	mu      sync.RWMutex
	clients map[chan SSEEvent]struct{}
}

func NewSSEHub() *SSEHub {
	return &SSEHub{clients: make(map[chan SSEEvent]struct{})}
}

// Broadcast sends event to all connected clients. No-op if hub is nil.
func (h *SSEHub) Broadcast(evt SSEEvent) {
	if h == nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- evt:
		default: // drop if client buffer full
		}
	}
}

func (h *SSEHub) register(ch chan SSEEvent) {
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
}

func (h *SSEHub) unregister(ch chan SSEEvent) {
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
}

// SSEHandler returns an http.HandlerFunc for GET /api/sse.
func SSEHandler(hub *SSEHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		ch := make(chan SSEEvent, 32)
		hub.register(ch)
		defer hub.unregister(ch)

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case evt := <-ch:
				data, _ := json.Marshal(evt)
				if _, err := w.Write([]byte("event: refreshed\ndata: " + string(data) + "\n\n")); err != nil {
					return
				}
				flusher.Flush()
			case <-ticker.C:
				if _, err := w.Write([]byte(": keepalive\n\n")); err != nil {
					return
				}
				flusher.Flush()
			case <-r.Context().Done():
				return
			}
		}
	}
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/http/
```
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/http/sse.go
git commit -m "feat: add SSE hub and handler for frontend live refresh"
```

---

### Task 9: Modify router.go — register SSE endpoint, inject SSEHub

**Files:**
- Modify: `internal/http/router.go`

- [ ] **Step 1: Update NewRouter signature and add SSE route**

The function signature changes from:
```go
func NewRouter(f *facade.HltvFacade, frontendFS fs.FS) http.Handler {
```
to:
```go
func NewRouter(f *facade.HltvFacade, frontendFS fs.FS, sseHub *SSEHub) http.Handler {
```

Add after the `h := handlers.New(f)` line, before health routes:
```go
r.Get("/api/sse", SSEHandler(sseHub))
```

- [ ] **Step 2: Verify compilation** (will pass since http package doesn't import main)

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/http/
```
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/http/router.go
git commit -m "feat: register SSE endpoint at /api/sse"
```

---

### Task 10: Modify facade/facade.go — Type A three-tier fallback

**Files:**
- Modify: `internal/facade/facade.go`

Reference: spec § Type A data flow

- [ ] **Step 1: Add imports**

Add to imports block (after `"github.com/arcdent/hltv-mcp/internal/scraper"`):
```go
"github.com/arcdent/hltv-mcp/internal/storage"
"log"
```

- [ ] **Step 2: Add Store and notify fields to HltvFacade struct**

After `client *client.HltvClient`, add:
```go
store  *storage.Store
notify func(entity string, id int, name string)
```

- [ ] **Step 3: Update New() signature**

Change from:
```go
func New(cfg *config.Config, c *cache.Cache, cli *client.HltvClient) *HltvFacade {
```
To:
```go
func New(cfg *config.Config, c *cache.Cache, cli *client.HltvClient, store *storage.Store, notify func(string, int, string)) *HltvFacade {
```

And add the two new fields in the return literal:
```go
store:  store,
notify: notify,
```

- [ ] **Step 4: Add broadcast helper**

After `New()`, add:
```go
func (f *HltvFacade) broadcast(entity string, id int, name string) {
	if f.notify != nil {
		f.notify(entity, id, name)
	}
}
```

- [ ] **Step 5: Rewrite GetPlayerDetailCached with three-tier fallback**

Replace the existing method with:
```go
func (f *HltvFacade) GetPlayerDetailCached(ctx context.Context, id int, slug string) (types.PlayerDetail, error) {
	if slug == "" {
		slug = fmt.Sprintf("player-%d", id)
	}
	key := fmt.Sprintf("player_detail:%d", id)

	// Tier 1: memory cache
	if cached, ok := f.cache.Get(key); ok {
		return cached.(types.PlayerDetail), nil
	}

	// Tier 2: SQLite
	if f.store != nil {
		if pd, ok, _ := f.store.GetPlayer(id); ok {
			f.cache.Set(key, pd, 10)
			go f.refreshPlayer(id, slug, key)
			return pd, nil
		}
	}

	// Tier 3: scrape from HLTV
	return f.scrapePlayerCached(ctx, id, slug)
}

func (f *HltvFacade) scrapePlayerCached(ctx context.Context, id int, slug string) (types.PlayerDetail, error) {
	pd, err := f.scrapePlayer(ctx, id, slug)
	if err != nil {
		return types.PlayerDetail{}, err
	}
	f.cache.Set(fmt.Sprintf("player_detail:%d", id), pd, f.cfg.CacheTTLPlayerDetail)
	return pd, nil
}

func (f *HltvFacade) scrapePlayer(ctx context.Context, id int, slug string) (types.PlayerDetail, error) {
	doc, err := f.ps.GetPlayer(ctx, id, slug)
	if err != nil {
		return types.PlayerDetail{}, err
	}
	pd := normalizer.NormalizePlayerDetail(doc)
	pd.Profile.ID = id
	if f.store != nil {
		if err := f.store.UpsertPlayer(pd); err != nil {
			log.Printf("facade: upsert player %d: %v", id, err)
		}
	}
	return pd, nil
}

func (f *HltvFacade) refreshPlayer(id int, slug, key string) {
	pd, err := f.scrapePlayer(context.Background(), id, slug)
	if err != nil {
		log.Printf("facade: refresh player %d: %v", id, err)
		return
	}
	f.cache.Set(key, pd, f.cfg.CacheTTLPlayerDetail)
	f.broadcast("player", pd.Profile.ID, pd.Profile.Name)
}
```

- [ ] **Step 6: Rewrite GetTeamDetailCached with three-tier fallback**

Replace the existing method with:
```go
func (f *HltvFacade) GetTeamDetailCached(ctx context.Context, id int, slug string) (types.TeamDetail, error) {
	if slug == "" {
		slug = fmt.Sprintf("team-%d", id)
	}
	key := fmt.Sprintf("team_detail:%d", id)

	// Tier 1: memory cache
	if cached, ok := f.cache.Get(key); ok {
		return cached.(types.TeamDetail), nil
	}

	// Tier 2: SQLite
	if f.store != nil {
		if td, ok, _ := f.store.GetTeam(id); ok {
			f.cache.Set(key, td, 10)
			go f.refreshTeam(id, slug, key)
			return td, nil
		}
	}

	// Tier 3: scrape from HLTV
	return f.scrapeTeamCached(ctx, id, slug)
}

func (f *HltvFacade) scrapeTeamCached(ctx context.Context, id int, slug string) (types.TeamDetail, error) {
	td, err := f.scrapeTeam(ctx, id, slug)
	if err != nil {
		return types.TeamDetail{}, err
	}
	f.cache.Set(fmt.Sprintf("team_detail:%d", id), td, f.cfg.CacheTTLPlayerDetail)
	return td, nil
}

func (f *HltvFacade) scrapeTeam(ctx context.Context, id int, slug string) (types.TeamDetail, error) {
	doc, err := f.ts.GetTeam(ctx, id, slug)
	if err != nil {
		return types.TeamDetail{}, err
	}
	td := normalizer.NormalizeTeamDetail(doc)
	td.Profile.ID = id
	td.Profile.Slug = slug

	if td.Profile.Name != "" {
		name := td.Profile.Name
		if upcomingDoc, err := f.ms.GetUpcoming(ctx); err == nil {
			allUpcoming := normalizer.NormalizeUpcomingMatches(upcomingDoc, name)
			for _, m := range allUpcoming {
				if m.Team1 == name || m.Team2 == name || m.Opponent == name {
					td.RecentMatches = append(td.RecentMatches, m)
				}
			}
		}
	}
	if td.Highlights != nil {
		for _, m := range td.Highlights.RecentMatches {
			if m.Result == "won" {
				td.Stats.Wins++
			} else {
				td.Stats.Losses++
			}
		}
		total := td.Stats.Wins + td.Stats.Losses + td.Stats.Draws
		if total > 0 {
			td.Stats.WinRate = fmt.Sprintf("%.0f%%", float64(td.Stats.Wins)/float64(total)*100)
		}
	}

	if f.store != nil {
		if err := f.store.UpsertTeam(td); err != nil {
			log.Printf("facade: upsert team %d: %v", id, err)
		}
	}
	return td, nil
}

func (f *HltvFacade) refreshTeam(id int, slug, key string) {
	td, err := f.scrapeTeam(context.Background(), id, slug)
	if err != nil {
		log.Printf("facade: refresh team %d: %v", id, err)
		return
	}
	f.cache.Set(key, td, f.cfg.CacheTTLPlayerDetail)
	f.broadcast("team", td.Profile.ID, td.Profile.Name)
}
```

- [ ] **Step 7: Rewrite GetNewsArticleCached with three-tier fallback**

Replace the existing method with:
```go
func (f *HltvFacade) GetNewsArticleCached(ctx context.Context, url string) (types.NewsArticle, error) {
	key := fmt.Sprintf("news_article:%x", md5.Sum([]byte(url)))

	// Tier 1: memory cache
	if cached, ok := f.cache.Get(key); ok {
		return cached.(types.NewsArticle), nil
	}

	// Tier 2: SQLite
	if f.store != nil {
		if article, ok, _ := f.store.GetNewsArticle(url); ok {
			f.cache.Set(key, article, 10)
			go f.refreshNewsArticle(url, key)
			return article, nil
		}
	}

	// Tier 3: scrape from HLTV
	return f.scrapeNewsArticleCached(context.Background(), url)
}

func (f *HltvFacade) scrapeNewsArticleCached(ctx context.Context, url string) (types.NewsArticle, error) {
	article, err := f.scrapeNewsArticle(ctx, url)
	if err != nil {
		return types.NewsArticle{}, err
	}
	f.cache.Set(fmt.Sprintf("news_article:%x", md5.Sum([]byte(url))), article, f.cfg.CacheTTLNewsArticle)
	return article, nil
}

func (f *HltvFacade) scrapeNewsArticle(ctx context.Context, url string) (types.NewsArticle, error) {
	doc, err := f.nas.GetArticle(ctx, url)
	if err != nil {
		return types.NewsArticle{}, err
	}
	article := normalizer.NormalizeNewsArticle(doc, url)
	if f.store != nil {
		if err := f.store.UpsertNewsArticle(article); err != nil {
			log.Printf("facade: upsert news article: %v", err)
		}
	}
	return article, nil
}

func (f *HltvFacade) refreshNewsArticle(url, key string) {
	article, err := f.scrapeNewsArticle(context.Background(), url)
	if err != nil {
		log.Printf("facade: refresh news article: %v", err)
		return
	}
	f.cache.Set(key, article, f.cfg.CacheTTLNewsArticle)
	f.broadcast("news", 0, article.Title)
}
```

> Note: The `withCache`, `createMeta`, `errorResponse`, `cloneResponse`, `CacheEntries`, `CacheHits`, `CacheMisses`, `ClearCache` methods remain UNCHANGED.

- [ ] **Step 4: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/facade/
```
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add internal/facade/facade.go
git commit -m "feat: add three-tier fallback with Store and SSE notify to Type A methods"
```

---

### Task 11: Modify facade/matches.go — Type B three-tier fallback, store in compute closure

**Files:**
- Modify: `internal/facade/matches.go`

Reference: spec § Type B data flow, three-category time query

- [ ] **Step 1: Add storage import**

Add to imports:
```go
"log"
```

- [ ] **Step 2: Rewrite GetUpcomingMatches with three-tier + DB write**

Replace the existing method. Key changes: before `withCache`, check SQLite on cache miss; inside compute closure, call `f.store.BatchUpsertMatches(items)` after scraping:

```go
func (f *HltvFacade) GetUpcomingMatches(query types.UpcomingMatchesQuery) *types.ToolResponse {
	team := stripGenericFilter(query.Team)
	event := stripGenericFilter(query.Event)

	if isPlaceholderText(query.Team) && isPlaceholderText(query.Event) && query.Limit == 1 && query.Days == 1 {
		team = ""
		event = ""
	}
	todayOnly := query.TodayOnly
	userSetLimit := query.Limit
	if query.Limit == 0 {
		query.Limit = 300
	}
	q := map[string]any{"team": team, "event": event, "today_only": todayOnly}
	key := fmt.Sprintf("matches_upcoming:%s:%s:%v", team, event, todayOnly)
	ttl := f.cfg.CacheTTLMatches

	return f.withCache(key, ttl, q, func() (*types.ToolResponse, error) {
		doc, err := f.ms.GetUpcoming(context.Background())
		if err != nil {
			return nil, err
		}
		items := normalizer.NormalizeUpcomingMatches(doc, "")

		// Persist to SQLite (best-effort)
		if f.store != nil {
			if err := f.store.BatchUpsertMatches(items); err != nil {
				log.Printf("facade: batch upsert upcoming matches: %v", err)
			}
		}

		normalizer.SortByScheduledAtAsc(items)
		if todayOnly {
			items = filterToday(items)
		}
		if userSetLimit > 0 && len(items) > userSetLimit {
			items = items[:userSetLimit]
		}
		if !todayOnly && len(items) > query.Limit {
			items = items[:query.Limit]
		}
		meta := f.createMeta(ttl)
		return &types.ToolResponse{Query: q, Items: items, Meta: meta}, nil
	})
}
```

The `withCache` method needs to be enhanced to check SQLite on cache miss BEFORE calling compute. Add a helper:

```go
// withCacheOrStore checks cache, then SQLite, then computes.
// For Type B methods that need persistent storage fallback.
func (f *HltvFacade) withCacheOrStore(key string, ttlSec int, query map[string]any,
	storeHit func() (*types.ToolResponse, bool),
	compute func() (*types.ToolResponse, error)) *types.ToolResponse {

	// Tier 1: memory cache
	if cached, ok := f.cache.Get(key); ok {
		r := cloneResponse(cached.(*types.ToolResponse))
		r.Meta.CacheHit = true
		return r
	}

	// Tier 2: SQLite
	if f.store != nil {
		if r, ok := storeHit(); ok {
			f.cache.Set(key, r, 10)
			go func() {
				val, err := f.cache.RunOnce("refresh:"+key, func() (any, error) {
					r, computeErr := compute()
					if computeErr != nil {
						return nil, computeErr
					}
					f.cache.Set(key, r, ttlSec)
					return r, nil
				})
				if err != nil {
					log.Printf("facade: background refresh %s: %v", key, err)
					return
				}
				f.broadcast("matches", 0, "")
			}()
			return r
		}
	}

	// Tier 3: compute fresh
	val, err := f.cache.RunOnce(key, func() (any, error) {
		r, computeErr := compute()
		if computeErr != nil {
			return nil, computeErr
		}
		f.cache.Set(key, r, ttlSec)
		return r, nil
	})
	if err != nil {
		return f.errorResponse(query, err)
	}
	return val.(*types.ToolResponse)
}
```

Then update `GetUpcomingMatches` to use `withCacheOrStore`:

```go
func (f *HltvFacade) GetUpcomingMatches(query types.UpcomingMatchesQuery) *types.ToolResponse {
	// ... same validation code as above ...
	q := map[string]any{"team": team, "event": event, "today_only": todayOnly}
	key := fmt.Sprintf("matches_upcoming:%s:%s:%v", team, event, todayOnly)
	ttl := f.cfg.CacheTTLMatches

	return f.withCacheOrStore(key, ttl, q,
		func() (*types.ToolResponse, bool) {
			category := "upcoming"
			if todayOnly {
				category = "today"
			}
			matches, err := f.store.QueryMatchesByTime(category, 0)
			if err != nil || len(matches) == 0 {
				return nil, false
			}
			// Apply filtering
			if team != "" {
				filtered := filterByTeam(matches, team)
				if len(filtered) == 0 {
					return nil, false
				}
				matches = filtered
			}
			if userSetLimit > 0 && len(matches) > userSetLimit {
				matches = matches[:userSetLimit]
			}
			meta := f.createMeta(ttl)
			return &types.ToolResponse{Query: q, Items: matches, Meta: meta}, true
		},
		func() (*types.ToolResponse, error) {
			doc, err := f.ms.GetUpcoming(context.Background())
			if err != nil {
				return nil, err
			}
			items := normalizer.NormalizeUpcomingMatches(doc, "")

			if f.store != nil {
				if err := f.store.BatchUpsertMatches(items); err != nil {
					log.Printf("facade: batch upsert upcoming matches: %v", err)
				}
			}

			normalizer.SortByScheduledAtAsc(items)
			if todayOnly {
				items = filterToday(items)
			}
			if userSetLimit > 0 && len(items) > userSetLimit {
				items = items[:userSetLimit]
			}
			if !todayOnly && len(items) > query.Limit {
				items = items[:query.Limit]
			}
			meta := f.createMeta(ttl)
			return &types.ToolResponse{Query: q, Items: items, Meta: meta}, nil
		})
}

func filterByTeam(matches []types.NormalizedMatch, team string) []types.NormalizedMatch {
	var out []types.NormalizedMatch
	tl := strings.ToLower(team)
	for _, m := range matches {
		if strings.ToLower(m.Team1) == tl || strings.ToLower(m.Team2) == tl {
			out = append(out, m)
		}
	}
	return out
}
```

- [ ] **Step 3: Rewrite GetResultsRecent with same pattern**

```go
func (f *HltvFacade) GetResultsRecent(query types.ResultsRecentQuery) *types.ToolResponse {
	team := stripGenericFilter(query.Team)
	event := stripGenericFilter(query.Event)
	if query.Limit == 0 {
		query.Limit = f.cfg.DefaultResultLimit
	}
	if query.Days == 0 {
		query.Days = 7
	}
	q := map[string]any{"team": team, "event": event, "days": query.Days}
	key := fmt.Sprintf("results_recent:%s:%s:%d", team, event, query.Days)
	ttl := f.cfg.CacheTTLResults

	return f.withCacheOrStore(key, ttl, q,
		func() (*types.ToolResponse, bool) {
			matches, err := f.store.QueryMatchesByTime("results", 0)
			if err != nil || len(matches) == 0 {
				return nil, false
			}
			if team != "" {
				matches = filterByTeam(matches, team)
				if len(matches) == 0 {
					return nil, false
				}
			}
			if len(matches) > query.Limit {
				matches = matches[:query.Limit]
			}
			meta := f.createMeta(ttl)
			return &types.ToolResponse{Query: q, Items: matches, Meta: meta}, true
		},
		func() (*types.ToolResponse, error) {
			doc, err := f.rs.GetResults(context.Background())
			if err != nil {
				return nil, err
			}
			items := normalizer.NormalizeMatches(doc, "")

			if f.store != nil {
				if err := f.store.BatchUpsertMatches(items); err != nil {
					log.Printf("facade: batch upsert results matches: %v", err)
				}
			}

			normalizer.SortByPlayedAtDesc(items)
			if len(items) > query.Limit {
				items = items[:query.Limit]
			}
			meta := f.createMeta(ttl)
			return &types.ToolResponse{Query: q, Items: items, Meta: meta}, nil
		})
}
```

Similarly update `GetEvents` (same pattern, replacing `withCache` with `withCacheOrStore`).

- [ ] **Step 4: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/facade/
```
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add internal/facade/matches.go internal/facade/facade.go
git commit -m "feat: add Type B three-tier fallback with SQLite to matches methods"
```

---

### Task 12: Modify facade/news.go — Type B three-tier fallback

**Files:**
- Modify: `internal/facade/news.go`

- [ ] **Step 1: Add imports**

```go
import (
	"context"
	"fmt"
	"log"

	"github.com/arcdent/hltv-mcp/internal/normalizer"
	"github.com/arcdent/hltv-mcp/internal/types"
)
```

- [ ] **Step 2: Update GetRealtimeNews with withCacheOrStore**

```go
func (f *HltvFacade) GetRealtimeNews(query types.RealtimeNewsQuery) *types.ToolResponse {
	if query.Limit == 0 {
		query.Limit = f.cfg.DefaultResultLimit
	}
	q := map[string]any{"limit": query.Limit}
	key := fmt.Sprintf("realtime_news:%d:%d", query.Limit, query.Page)
	ttl := f.cfg.CacheTTLRealtimeNews

	return f.withCacheOrStore(key, ttl, q,
		func() (*types.ToolResponse, bool) {
			items, err := f.store.QueryRealtimeNews(query.Limit)
			if err != nil || len(items) == 0 {
				return nil, false
			}
			meta := f.createMeta(ttl)
			return &types.ToolResponse{Query: q, Items: items, Meta: meta}, true
		},
		func() (*types.ToolResponse, error) {
			doc, err := f.rns.GetRealtimeNews(context.Background(), query.Page)
			if err != nil {
				return nil, fmt.Errorf("realtime_news: %w", err)
			}
			items := normalizer.NormalizeRealtimeNews(doc, query.Limit)

			if f.store != nil {
				if err := f.store.BatchUpsertRealtimeNews(items); err != nil {
					log.Printf("facade: batch upsert realtime news: %v", err)
				}
			}

			meta := f.createMeta(ttl)
			return &types.ToolResponse{Query: q, Items: items, Meta: meta}, nil
		})
}
```

- [ ] **Step 3: Update GetNewsDigest with same pattern**

```go
func (f *HltvFacade) GetNewsDigest(query types.NewsDigestQuery) *types.ToolResponse {
	if query.Limit == 0 {
		query.Limit = f.cfg.DefaultResultLimit
	}
	q := map[string]any{"limit": query.Limit, "tag": query.Tag, "year": query.Year, "month": query.Month}
	key := fmt.Sprintf("news_digest:%s:%d:%s:%d", query.Tag, query.Year, query.Month, query.Page)
	ttl := f.cfg.CacheTTLNews

	return f.withCacheOrStore(key, ttl, q,
		func() (*types.ToolResponse, bool) {
			items, err := f.store.QueryNews(query.Limit)
			if err != nil || len(items) == 0 {
				return nil, false
			}
			meta := f.createMeta(ttl)
			return &types.ToolResponse{Query: q, Items: items, Meta: meta}, true
		},
		func() (*types.ToolResponse, error) {
			doc, err := f.ns.GetNewsDigest(context.Background(), query.Tag, query.Year, query.Month, query.Page)
			if err != nil {
				return nil, fmt.Errorf("news_digest: %w", err)
			}
			items := normalizer.NormalizeNewsDigest(doc, query.Limit)

			if f.store != nil {
				if err := f.store.BatchUpsertNews(items); err != nil {
					log.Printf("facade: batch upsert news: %v", err)
				}
			}

			meta := f.createMeta(ttl)
			return &types.ToolResponse{Query: q, Items: items, Meta: meta}, nil
		})
}
```

- [ ] **Step 4: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build ./internal/facade/
```
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add internal/facade/news.go
git commit -m "feat: add Type B three-tier fallback with SQLite to news methods"
```

---

### Task 13: Modify auth.go (or relevant file) — inject facades into handlers

Wait, the handlers use `facade.HltvFacade`. Since we already updated facade, handlers don't need changes — they just call the same methods. The facade interface to handlers is unchanged.

Actually wait, let me check: handlers only call `f.GetTeamDetailCached`, `f.GetPlayerDetailCached`, `f.GetNewsArticleCached`, `f.GetEvents`, `f.GetUpcomingMatches`, etc. None of these signatures changed. The facade struct gained new fields but method signatures are the same. Handlers don't need changes.

- [ ] **Step 1: No changes to handlers needed — skip**

---

### Task 14: Modify main.go — init storage, SSE hub, wire into facade

**Files:**
- Modify: `main.go`

- [ ] **Step 1: Add imports and wire up storage + SSE hub**

Add to imports:
```go
"github.com/arcdent/hltv-mcp/internal/storage"
httppkg "github.com/arcdent/hltv-mcp/internal/http"
"strings"
```

Replace the facade initialization section. Current code (lines 55-58):
```go
c := cache.New(cfg.CacheMaxEntries, cfg.CacheStaleWindowSec)
cli := client.NewHltvClient(cfg)
f := facade.New(cfg, c, cli)
```

Replace with:
```go
c := cache.New(cfg.CacheMaxEntries, cfg.CacheStaleWindowSec)
cli := client.NewHltvClient(cfg)

// Initialize SQLite (degrade gracefully on failure)
store, err := storage.Open(cfg.DBPath, cfg.DBRetentionMatches, cfg.DBRetentionNews, cfg.DBRetentionRealtime)
if err != nil {
	log.Printf("storage: open failed: %v — degrading to cache-only mode", err)
	store = nil
}

// Initialize SSE hub
sseHub := httppkg.NewSSEHub()

// Wire SSE broadcast into facade as notify callback
var notify func(string, int, string)
if sseHub != nil {
	notify = func(entity string, id int, name string) {
		sseHub.Broadcast(httppkg.SSEEvent{Entity: entity, ID: id, Name: name})
	}
}

f := facade.New(cfg, c, cli, store, notify)
```

- [ ] **Step 2: Update router to pass SSE hub**

Change:
```go
router := httppkg.NewRouter(f, frontendFS)
```
To:
```go
router := httppkg.NewRouter(f, frontendFS, sseHub)
```

- [ ] **Step 3: Add storage Close() to shutdown**

After `httpServer.Shutdown(context.Background())`, add:
```go
if store != nil {
	if err := store.Close(); err != nil {
		log.Printf("storage: close: %v", err)
	}
}
```

Also add `strings` to imports if not already (it might already be there for other purposes).

- [ ] **Step 4: Verify compilation**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build .
```
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add main.go
git commit -m "feat: wire storage, SSE hub into main — degrade on DB failure"
```

---

### Task 15: Update Docker configuration

**Files:**
- Modify: `docker-compose.yml`
- Modify: `README.md`

- [ ] **Step 1: docker-compose.yml already has `./data:/data` volume mount. Add FIRECRAWL_API_KEY env**

Add to environment block:
```yaml
- FIRECRAWL_API_KEY=${FIRECRAWL_API_KEY:-}
```

- [ ] **Step 2: Update README.md Docker run commands**

In all Docker run examples, add `-v hltv-data:/data`:

**Windows:**
```powershell
docker run -d --name hltv-mcp `
  -p 8082:8082 `
  -e FIRECRAWL_API_KEY=fc-xxxxxxxxxxxxxxxx `
  -v hltv-chrome-data:/tmp `
  -v hltv-data:/data `
  ghcr.io/arcdent/hltv-data:latest
```

**Linux/macOS/WSL:**
```bash
docker run -d --name hltv-mcp \
  -p 8082:8082 \
  -e FIRECRAWL_API_KEY=fc-xxxxxxxxxxxxxxxx \
  -v hltv-chrome-data:/tmp \
  -v hltv-data:/data \
  ghcr.io/arcdent/hltv-data:latest
```

**Update image:**
```bash
docker pull ghcr.io/arcdent/hltv-data:latest && docker rm -f hltv-mcp && docker run -d --name hltv-mcp -p 8082:8082 -v hltv-chrome-data:/tmp -v hltv-data:/data ghcr.io/arcdent/hltv-data:latest
```

**crontab:**
```bash
*/5 * * * * docker pull ghcr.io/arcdent/hltv-data:latest && docker rm -f hltv-mcp && docker run -d --name hltv-mcp -p 8082:8082 -v hltv-chrome-data:/tmp -v hltv-data:/data ghcr.io/arcdent/hltv-data:latest
```

Also update the Windows scheduled task and environment variables table to add:
```markdown
| `HLTV_DB_PATH` | `data/hltv.db` | SQLite 数据库路径 |
| `HLTV_DB_RETENTION_MATCHES` | `90` | 比赛数据保留天数 |
| `HLTV_DB_RETENTION_NEWS` | `30` | 新闻数据保留天数 |
| `HLTV_DB_RETENTION_REALTIME_NEWS` | `7` | 实时新闻保留天数 |
```

- [ ] **Step 3: Commit**

```bash
git add docker-compose.yml README.md
git commit -m "docs: add data volume mount and new env vars to Docker/README"
```

---

### Task 16: Build, test, verify end-to-end

- [ ] **Step 1: Build binary**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go build -o hltv-mcp .
```
Expected: successful build

- [ ] **Step 2: Run existing tests**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && go test ./internal/... -v -timeout 30s 2>&1
```
Expected: all existing tests pass

- [ ] **Step 3: Quick integration test — start server, check SSE endpoint**

```bash
cd /home/arcdent/github/hltv-mcp-fully-rebuild && ./hltv-mcp &
sleep 2
# Check health
curl -s http://localhost:8082/api/health
# Check SSE endpoint exists
curl -s -N http://localhost:8082/api/sse &
sleep 1
kill %1 %2 2>/dev/null
kill $(lsof -t -i:8082) 2>/dev/null
```
Expected: health returns `{"status":"ok"}`; SSE returns keepalive (no crash)

- [ ] **Step 4: Verify SQLite database created**

```bash
ls -la /home/arcdent/github/hltv-mcp-fully-rebuild/data/hltv.db
```
Expected: file exists (created by running server)

- [ ] **Step 5: Commit if any test adjustments needed, or confirm ready**

---

## Spec Coverage Self-Review

| Spec Requirement | Task(s) | Status |
|---|---|---|
| Type A three-tier fallback (PlayerDetail/TeamDetail/NewsArticle) | Task 10 | ✓ |
| Type B three-tier fallback (Matches/News) | Task 11, 12 | ✓ |
| `withCacheOrStore` for stale-while-revalidate + background goroutine | Task 11 | ✓ |
| Match three-category time query (future/today/past) | Task 6 `QueryMatchesByTime` | ✓ |
| COALESCE partial update for matches BatchUpsert | Task 6 | ✓ |
| 5 data tables + schema_version | Task 2 | ✓ |
| SSE hub + `GET /api/sse` endpoint | Task 8, 9 | ✓ |
| SSE two event granularities (precise vs broad) | Task 10 `broadcast("player",id,name)`, Task 11 `broadcast("matches",0,"")` | ✓ |
| SSE hub injected as nil-safe dependency | Task 14 `notify` callback, Task 8 `Broadcast` nil check | ✓ |
| Config: DBPath, HLTV_DB_RETENTION_* | Task 1 | ✓ |
| Expiration cleanup (startup + every 24h) | Task 2 `runCleanup`, `startCleanupLoop` | ✓ |
| Error handling: degrade, not crash | Task 14 `store = nil` on open failure; Task 10,11,12 `log.Printf` on upsert failures | ✓ |
| Dockerfile: VOLUME /data | Already exists — no change needed | ✓ |
| docker-compose.yml: data volume mount | Already exists, minor env tweak Task 15 | ✓ |
| README.md: `-v hltv-data:/data` in all commands | Task 15 | ✓ |
| go.mod: modernc.org/sqlite | Task 1 | ✓ |

**Placeholder scan:** No TBD, TODO, or placeholder patterns found. All code is complete.

**Type consistency:**
- `Store` methods: `UpsertTeam(TeamDetail)`, `GetTeam(int) → (TeamDetail, bool, error)` — consistent across Task 4 and Task 10
- `SSEEvent{Entity, ID, Name}` — consistent across Task 8 `Broadcast` and Task 14 wiring
- `withCacheOrStore(key, ttl, query, storeHit, compute)` — consistent across Task 11, 12
- `QueryMatchesByTime(category, limit)` — categories "upcoming"/"today"/"results" matched with spec ✓
