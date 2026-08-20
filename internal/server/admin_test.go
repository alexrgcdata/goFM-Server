package server

import (
	"net/http"
	"strings"
	"testing"
)

func TestAdminEndpointsRequireAdminTokenAndExposeBoundedLogs(t *testing.T) {
	server := New(Config{Tokens: []string{"route-token"}, AdminTokens: []string{"admin-token"}, Logs: LogConfig{MaxEntries: 2}, Routes: []Route{{Path: "/api/echo", Methods: []string{"GET"}, Target: Target{Type: "http", URL: "http://example.invalid"}}}})
	unauthorized := newRecorder()
	server.ServeHTTP(unauthorized, newRequest(t, http.MethodGet, "/__gofm/logs", nil))
	if unauthorized.status != http.StatusUnauthorized {
		t.Fatalf("admin status = %d, want 401", unauthorized.status)
	}

	for i := 0; i < 3; i++ {
		response := newRecorder()
		request := newRequest(t, http.MethodGet, "/health", nil)
		server.ServeHTTP(response, request)
	}
	authorized := newRequest(t, http.MethodGet, "/__gofm/logs", nil)
	authorized.Header.Set("Authorization", "Bearer admin-token")
	response := newRecorder()
	server.ServeHTTP(response, authorized)
	if response.status != http.StatusOK {
		t.Fatalf("admin logs status = %d, want 200", response.status)
	}
	if len(server.logs.list()) != 2 {
		t.Fatalf("log count = %d, want bounded count 2", len(server.logs.list()))
	}
}

func TestAdminCanAddValidatedRuntimeRoute(t *testing.T) {
	server := New(Config{Tokens: []string{"token"}, AdminTokens: []string{"admin"}, Routes: []Route{}})
	request := newRequest(t, http.MethodPost, "/__gofm/routes", strings.NewReader(`{"id":"demo","path":"/api/demo","methods":["GET"],"target":{"type":"http","url":"http://127.0.0.1:3001"}}`))
	request.Header.Set("Authorization", "Bearer admin")
	request.Header.Set("Content-Type", "application/json")
	response := newRecorder()
	server.ServeHTTP(response, request)
	if response.status != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", response.status, response.body.String())
	}
	if server.match(http.MethodGet, "/api/demo") == nil {
		t.Fatal("runtime route was not added")
	}
}
