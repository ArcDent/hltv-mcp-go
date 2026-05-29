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
