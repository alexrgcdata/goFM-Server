package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type RequestRecord struct {
	ID                 string
	StartedAt          time.Time
	DurationMS         int64
	Method             string
	Path               string
	Status             int
	Route              string
	Upstream           string
	Outcome            string
	ResponsePreview    string
	HookAfterRequested bool
}

type Store struct{ db *sql.DB }

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, fmt.Errorf("storage path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, fmt.Errorf("create storage directory: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	store := &Store{db: db}
	if _, err := db.Exec(`PRAGMA journal_mode=WAL; CREATE TABLE IF NOT EXISTS request_logs (
id TEXT PRIMARY KEY, started_at TEXT NOT NULL, duration_ms INTEGER NOT NULL, method TEXT NOT NULL,
path TEXT NOT NULL, status INTEGER NOT NULL, route TEXT, upstream TEXT, outcome TEXT NOT NULL,
response_preview TEXT, hook_after_requested INTEGER NOT NULL DEFAULT 0
); CREATE INDEX IF NOT EXISTS idx_request_logs_started_at ON request_logs(started_at DESC);`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("initialize sqlite: %w", err)
	}
	return store, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Append(ctx context.Context, record RequestRecord, maxEntries int) error {
	if s == nil || s.db == nil {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `INSERT OR REPLACE INTO request_logs
(id,started_at,duration_ms,method,path,status,route,upstream,outcome,response_preview,hook_after_requested)
VALUES(?,?,?,?,?,?,?,?,?,?,?)`, record.ID, record.StartedAt.UTC().Format(time.RFC3339Nano), record.DurationMS, record.Method, record.Path, record.Status, record.Route, record.Upstream, record.Outcome, record.ResponsePreview, record.HookAfterRequested)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `DELETE FROM request_logs WHERE id NOT IN (SELECT id FROM request_logs ORDER BY started_at DESC LIMIT ?)`, maxEntries)
	return err
}

func (s *Store) Recent(ctx context.Context, limit int) ([]RequestRecord, error) {
	if s == nil || s.db == nil {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,started_at,duration_ms,method,path,status,route,upstream,outcome,response_preview,hook_after_requested FROM request_logs ORDER BY started_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []RequestRecord
	for rows.Next() {
		var record RequestRecord
		var started string
		var hook int
		if err := rows.Scan(&record.ID, &started, &record.DurationMS, &record.Method, &record.Path, &record.Status, &record.Route, &record.Upstream, &record.Outcome, &record.ResponsePreview, &hook); err != nil {
			return nil, err
		}
		record.StartedAt, err = time.Parse(time.RFC3339Nano, started)
		if err != nil {
			return nil, err
		}
		record.HookAfterRequested = hook != 0
		records = append(records, record)
	}
	return records, rows.Err()
}
