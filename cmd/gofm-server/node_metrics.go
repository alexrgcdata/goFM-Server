package main

import (
	"io"
	"net/http"
	"time"
)

var nodeMetricsClient = &http.Client{
	Timeout: 4 * time.Second,
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func readNodeMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "use GET", http.StatusMethodNotAllowed)
		return
	}

	// This URL is fixed. The browser cannot choose a host, path, or query.
	request, err := http.NewRequestWithContext(r.Context(), http.MethodGet, "http://127.0.0.1:8081/admin/metrics", nil)
	if err != nil {
		http.Error(w, "Node metrics unavailable", http.StatusBadGateway)
		return
	}
	for _, value := range r.Header.Values("Authorization") {
		request.Header.Add("Authorization", value)
	}

	response, err := nodeMetricsClient.Do(request)
	if err != nil {
		http.Error(w, "Node metrics unavailable", http.StatusBadGateway)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusUnauthorized {
		http.Error(w, "Node metrics unavailable", http.StatusBadGateway)
		return
	}

	// Buffer the tiny response so an upstream failure cannot look like success.
	const maxBytes = 16 * 1024
	body, err := io.ReadAll(io.LimitReader(response.Body, maxBytes+1))
	if err != nil || len(body) > maxBytes {
		http.Error(w, "Node metrics unavailable", http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(response.StatusCode)
	_, _ = w.Write(body)
}
