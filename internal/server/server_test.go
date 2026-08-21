package server

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

func TestSQLiteStoresMetadataWithoutResponsePreview(t *testing.T) {
	server := New(Config{Tokens: []string{"app"}, AdminTokens: []string{"admin"}, Logs: LogConfig{MaxEntries: 10, CaptureBodyPreview: true}, Storage: StorageConfig{DBFile: filepath.Join(t.TempDir(), "gofm.sqlite")}})
	defer server.store.Close()
	server.ServeHTTP(newRecorder(), newRequest(t, http.MethodGet, "/health", nil))
	records, err := server.store.Recent(context.Background(), 10)
	if err != nil || len(records) != 1 {
		t.Fatalf("read SQLite metadata: records=%d err=%v", len(records), err)
	}
	if records[0].ResponsePreview != "" {
		t.Fatal("SQLite stored a response preview in plaintext")
	}
	if len(server.logs.list()) != 1 || server.logs.list()[0].ResponsePreview == "" {
		t.Fatal("encrypted-capable log store did not retain the configured preview")
	}
}

type responseRecorder struct {
	header http.Header
	body   strings.Builder
	status int
}

func newRecorder() *responseRecorder            { return &responseRecorder{header: make(http.Header)} }
func (r *responseRecorder) Header() http.Header { return r.header }
func (r *responseRecorder) Write(data []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.body.Write(data)
}
func (r *responseRecorder) WriteHeader(status int) { r.status = status }
func newRequest(t *testing.T, method, target string, body io.Reader) *http.Request {
	t.Helper()
	request, err := http.NewRequest(method, target, body)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func testServer(upstream string) *Server {
	return New(Config{Tokens: []string{"test-token"}, Routes: []Route{{Path: "/api/echo", Methods: []string{"GET", "POST"}, Target: Target{Type: "http", URL: upstream}}}})
}

func TestHealthIsPublic(t *testing.T) {
	response := newRecorder()
	testServer("http://example.invalid").ServeHTTP(response, newRequest(t, http.MethodGet, "/health", nil))
	if response.status != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.status)
	}
	var body map[string]string
	if err := json.NewDecoder(strings.NewReader(response.body.String())).Decode(&body); err != nil || body["status"] != "ok" {
		t.Fatalf("unexpected health response: %#v, %v", body, err)
	}
}

func TestConfiguredBasePathKeepsGatewaySecondary(t *testing.T) {
	server := New(Config{
		BasePath: "/openbridge/api",
		Tokens:   []string{"test-token"},
		Routes: []Route{{
			Path:    "/echo",
			Methods: []string{"GET"},
			Target:  Target{Type: "http", URL: "http://example.invalid"},
		}},
	})
	health := newRecorder()
	server.ServeHTTP(health, newRequest(t, http.MethodGet, "/openbridge/api/health", nil))
	if health.status != http.StatusOK {
		t.Fatalf("prefixed health status = %d, want 200", health.status)
	}
	outside := newRecorder()
	server.ServeHTTP(outside, newRequest(t, http.MethodGet, "/openbridge/health", nil))
	if outside.status != http.StatusNotFound {
		t.Fatalf("outside base path status = %d, want 404", outside.status)
	}
	route := newRecorder()
	request := newRequest(t, http.MethodGet, "/openbridge/api/echo", nil)
	request.Header.Set("Authorization", "Bearer test-token")
	server.ServeHTTP(route, request)
	if route.status == http.StatusNotFound || route.status == http.StatusUnauthorized {
		t.Fatalf("prefixed route was not matched, status = %d", route.status)
	}
}

func TestProtectedRouteRequiresBearerToken(t *testing.T) {
	response := newRecorder()
	testServer("http://example.invalid").ServeHTTP(response, newRequest(t, http.MethodGet, "/api/echo", nil))
	if response.status != http.StatusUnauthorized || response.Header().Get("WWW-Authenticate") != "Bearer" {
		t.Fatalf("got status %d and WWW-Authenticate %q", response.status, response.Header().Get("WWW-Authenticate"))
	}
}

func TestRequestBodyLimitIsEnforced(t *testing.T) {
	server := testServer("http://example.invalid")
	server.config.Security.MaxRequestBodyBytes = 1024
	request := newRequest(t, http.MethodPost, "/api/echo", strings.NewReader(strings.Repeat("x", 1025)))
	request.Header.Set("Authorization", "Bearer test-token")
	response := newRecorder()
	server.ServeHTTP(response, request)
	if response.status != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", response.status)
	}
	logs := server.logs.list()
	if len(logs) != 1 || logs[0].Outcome != "failure" || logs[0].Status != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized request failure was not logged: %#v", logs)
	}
}

func TestRateLimitIsEnforced(t *testing.T) {
	server := testServer("http://example.invalid")
	server.limits = newRateLimiter(1)
	for i, want := range []int{http.StatusOK, http.StatusTooManyRequests} {
		response := newRecorder()
		request := newRequest(t, http.MethodGet, "/health", nil)
		request.RemoteAddr = "127.0.0.1:5000"
		server.ServeHTTP(response, request)
		if response.status != want {
			t.Fatalf("request %d status = %d, want %d", i+1, response.status, want)
		}
	}
}

func TestAdminAuthorizedRouteRejectsApplicationToken(t *testing.T) {
	server := New(Config{Tokens: []string{"app"}, AdminTokens: []string{"admin"}, Routes: []Route{{Path: "/private", Methods: []string{"GET"}, Auth: "admin", Target: Target{Type: "http", URL: "http://example.invalid"}}}})
	request := newRequest(t, http.MethodGet, "/private", nil)
	request.Header.Set("Authorization", "Bearer app")
	response := newRecorder()
	server.ServeHTTP(response, request)
	if response.status != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.status)
	}
}

func TestProxyForwardsMatchingRouteWithoutAuthHeader(t *testing.T) {
	var received *http.Request
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	upstream := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Clone(r.Context())
		body, _ := io.ReadAll(r.Body)
		received.Body = io.NopCloser(strings.NewReader(string(body)))
		w.Header().Set("X-Upstream", "yes")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	})}
	go upstream.Serve(listener)
	t.Cleanup(func() { _ = upstream.Close() })
	response := newRecorder()
	request := newRequest(t, http.MethodPost, "/api/echo?name=goFM", strings.NewReader("hello"))
	request.Header.Set("Authorization", "Bearer test-token")
	testServer("http://"+listener.Addr().String()+"/base").ServeHTTP(response, request)
	if received == nil || received.URL.Path != "/base/api/echo" || received.URL.RawQuery != "name=goFM" || received.Header.Get("Authorization") != "" || received.Method != http.MethodPost {
		t.Fatalf("unexpected upstream request: %#v", received)
	}
	body, _ := io.ReadAll(received.Body)
	if string(body) != "hello" {
		t.Fatalf("body = %q", body)
	}
	if response.status != http.StatusCreated || response.body.String() != "created" || response.Header().Get("X-Upstream") != "yes" {
		t.Fatalf("unexpected proxy response: %d, %q, %q", response.status, response.body.String(), response.Header().Get("X-Upstream"))
	}
}

func TestRouteAndMethodMismatchesReturnNotFound(t *testing.T) {
	server := testServer("http://example.invalid")
	for _, request := range []*http.Request{newRequest(t, http.MethodDelete, "/api/echo", nil), newRequest(t, http.MethodGet, "/unknown", nil)} {
		response := newRecorder()
		server.ServeHTTP(response, request)
		if response.status != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", response.status)
		}
	}
}

func TestPreflightAllowsConfiguredOriginAndRoute(t *testing.T) {
	server := testServer("http://example.invalid")
	server.config.CORSOrigins = []string{"http://localhost:5173"}
	request := newRequest(t, http.MethodOptions, "/api/echo", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	request.Header.Set("Access-Control-Request-Headers", "Authorization, Content-Type")
	response := newRecorder()
	server.ServeHTTP(response, request)
	if response.status != http.StatusNoContent || response.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" || response.Header().Get("Access-Control-Allow-Methods") != "GET, POST" {
		t.Fatalf("unexpected preflight response: %d, %#v", response.status, response.Header())
	}
}

func TestPreflightRejectsUnconfiguredOrigin(t *testing.T) {
	server := testServer("http://example.invalid")
	server.config.CORSOrigins = []string{"http://localhost:5173"}
	request := newRequest(t, http.MethodOptions, "/api/echo", nil)
	request.Header.Set("Origin", "https://untrusted.example")
	request.Header.Set("Access-Control-Request-Method", http.MethodGet)
	response := newRecorder()
	server.ServeHTTP(response, request)
	if response.status != http.StatusForbidden || response.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("unexpected preflight response: %d, %#v", response.status, response.Header())
	}
}

func TestConfigRejectsUnsafeCORSOrigin(t *testing.T) {
	config := Config{CORSOrigins: []string{"http://localhost:5173/path"}}
	if err := config.Validate(); err == nil {
		t.Fatal("expected CORS origin validation error")
	}
}
