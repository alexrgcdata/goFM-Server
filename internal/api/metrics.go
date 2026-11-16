package api

import (
	"net/http"
	"sync"
	"time"
)

// Counters live in memory and reset when the Go process restarts.
type requestMetrics struct {
	mu      sync.Mutex
	started time.Time
	count   uint64
	errors  uint64
	total   time.Duration
}

type metricsSnapshot struct {
	UptimeSeconds     int64   `json:"uptime_seconds"`
	Requests          uint64  `json:"requests"`
	Errors            uint64  `json:"errors"`
	AverageResponseMS float64 `json:"average_response_ms"`
}

func (m *requestMetrics) record(status int, elapsed time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.count++
	m.total += elapsed
	if status >= http.StatusBadRequest {
		m.errors++
	}
}

func (m *requestMetrics) snapshot() metricsSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()

	snapshot := metricsSnapshot{
		UptimeSeconds: int64(time.Since(m.started) / time.Second),
		Requests:      m.count,
		Errors:        m.errors,
	}
	if m.count > 0 {
		snapshot.AverageResponseMS = float64(m.total) / float64(time.Millisecond) / float64(m.count)
	}

	return snapshot
}
