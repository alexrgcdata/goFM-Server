package server

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type rateWindow struct {
	start time.Time
	count int
}

type rateLimiter struct {
	mu      sync.Mutex
	limit   int
	windows map[string]rateWindow
}

func newRateLimiter(limit int) *rateLimiter {
	if limit < 1 {
		limit = 120
	}
	return &rateLimiter{limit: limit, windows: make(map[string]rateWindow)}
}

func (l *rateLimiter) allow(client string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.windows) > 10000 {
		for key, window := range l.windows {
			if now.Sub(window.start) >= time.Minute {
				delete(l.windows, key)
			}
		}
	}
	window := l.windows[client]
	if window.start.IsZero() || now.Sub(window.start) >= time.Minute {
		l.windows[client] = rateWindow{start: now, count: 1}
		return true
	}
	if window.count >= l.limit {
		return false
	}
	window.count++
	l.windows[client] = window
	return true
}

func clientAddress(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}
