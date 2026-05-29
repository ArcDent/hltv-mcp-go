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
