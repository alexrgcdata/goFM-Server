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
	totalMS float64
	recent  []trafficEntry
}

type trafficEntry struct {
	Time       string  `json:"time"`
	RequestID  string  `json:"request_id"`
	Action     string  `json:"action,omitempty"`
	Status     int     `json:"status"`
	DurationMS float64 `json:"duration_ms"`
	ErrorCode  string  `json:"error_code,omitempty"`
}

type metricsSnapshot struct {
	UptimeSeconds     int64          `json:"uptime_seconds"`
	Requests          uint64         `json:"requests"`
	Errors            uint64         `json:"errors"`
	AverageResponseMS float64        `json:"average_response_ms"`
	Recent            []trafficEntry `json:"recent"`
}

func (m *requestMetrics) record(entry trafficEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.count++
	m.totalMS += entry.DurationMS
	if entry.Status >= http.StatusBadRequest {
		m.errors++
	}

	// Bound memory even when the router receives a lot of traffic.
	const limit = 50
	if len(m.recent) == limit {
		copy(m.recent, m.recent[1:])
		m.recent = m.recent[:limit-1]
	}
	m.recent = append(m.recent, entry)
}

func (m *requestMetrics) snapshot() metricsSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()

	snapshot := metricsSnapshot{
		UptimeSeconds: int64(time.Since(m.started) / time.Second),
		Requests:      m.count,
		Errors:        m.errors,
		Recent:        make([]trafficEntry, 0, len(m.recent)),
	}
	for index := len(m.recent) - 1; index >= 0; index-- {
		snapshot.Recent = append(snapshot.Recent, m.recent[index])
	}
	if m.count > 0 {
		snapshot.AverageResponseMS = m.totalMS / float64(m.count)
	}

	return snapshot
}
