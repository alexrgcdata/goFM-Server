package server

import (
	"encoding/base64"
	"net/http"
	"path/filepath"
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

func TestAdminCanDisableAndBoundRuntimeLogs(t *testing.T) {
	server := New(Config{Tokens: []string{"token"}, AdminTokens: []string{"admin"}, Logs: LogConfig{MaxEntries: 5}})
	request := newRequest(t, http.MethodPut, "/__gofm/settings/logs", strings.NewReader(`{"max_entries":0}`))
	request.Header.Set("Authorization", "Bearer admin")
	response := newRecorder()
	server.ServeHTTP(response, request)
	if response.status != http.StatusOK || server.logs.limit() != 0 {
		t.Fatalf("disable logs status=%d limit=%d body=%s", response.status, server.logs.limit(), response.body.String())
	}
	server.ServeHTTP(newRecorder(), newRequest(t, http.MethodGet, "/health", nil))
	if len(server.logs.list()) != 0 {
		t.Fatal("disabled logger captured a request")
	}
}

func TestAdminStoresCredentialAndAddsFileMakerConnection(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	t.Setenv("GOFM_TEST_ADMIN_VAULT", base64.StdEncoding.EncodeToString(key))
	server := New(Config{Tokens: []string{"token"}, AdminTokens: []string{"admin"}, Credentials: VaultConfig{File: filepath.Join(t.TempDir(), "credentials.enc"), EncryptionKeyEnv: "GOFM_TEST_ADMIN_VAULT"}})
	body := `{"connection":{"name":"crm","adapter":"dataapi","base_url":"https://fm.example.com","credential":"crm-login","default_database":"CRM","allowed_databases":["CRM"],"allowed_layouts":["Customers"],"allowed_operations":["find","get"]},"credential":{"username":"fm-user","password":"fm-pass"}}`
	request := newRequest(t, http.MethodPost, "/__gofm/filemaker-connections", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer admin")
	response := newRecorder()
	server.ServeHTTP(response, request)
	if response.status != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", response.status, response.body.String())
	}
	if !strings.Contains(response.body.String(), `"credential_stored":true`) || strings.Contains(response.body.String(), "fm-pass") {
		t.Fatalf("response did not confirm safe credential storage: %s", response.body.String())
	}
}

func TestAdminCanAddValidatedRuntimeRoute(t *testing.T) {
	server := New(Config{Tokens: []string{"token"}, AdminTokens: []string{"admin"}, Security: SecurityConfig{AllowedUpstreamHosts: []string{"127.0.0.1:3001"}}, Routes: []Route{}})
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
	if !strings.Contains(response.body.String(), `"persistent":false`) {
		t.Fatal("runtime route response did not disclose memory-only persistence")
	}
}
