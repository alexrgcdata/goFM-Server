package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type demoJob struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

// NewWithDemoJobs enables the optional PostgreSQL example.
func NewWithDemoJobs(config Config, database *sql.DB) (*Server, error) {
	server, err := New(config)
	if err != nil {
		return nil, err
	}

	server.actions["demo.jobs"] = func(ctx context.Context, params []json.RawMessage) (any, error) {
		if err := noParams(params); err != nil {
			return nil, err
		}
		return readDemoJobs(ctx, database)
	}
	return server, nil
}

func readDemoJobs(ctx context.Context, database *sql.DB) ([]demoJob, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// The caller cannot change this query or choose a different table.
	rows, err := database.QueryContext(ctx, `
		SELECT id, title, status
		FROM public.go_demo_jobs
		ORDER BY id
		LIMIT 25
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := make([]demoJob, 0)
	for rows.Next() {
		var job demoJob
		if err := rows.Scan(&job.ID, &job.Title, &job.Status); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}
