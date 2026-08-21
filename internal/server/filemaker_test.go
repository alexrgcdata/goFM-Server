package server

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"gofm-server/internal/credentials"
	"gofm-server/internal/filemaker"
)

func TestFileMakerEndpointRequiresApplicationTokenAndKeepsCredentialsServerSide(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)
	t.Setenv("GOFM_TEST_FM_VAULT", base64.StdEncoding.EncodeToString(key))
	host := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/sessions") {
			user, password, _ := r.BasicAuth()
			if user != "fm-user" || password != "fm-pass" {
				t.Error("server did not use vault credential")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"response": map[string]any{"token": "private-session"}, "messages": []any{map[string]string{"code": "0"}}})
			return
		}
		if strings.Contains(r.Header.Get("Authorization"), "app-token") {
			t.Error("application token leaked to FileMaker")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"response": map[string]any{"data": []any{}, "dataInfo": map[string]any{"foundCount": 0, "returnedCount": 0}}, "messages": []any{map[string]string{"code": "0"}}})
	}))
	defer host.Close()
	vault, err := credentials.Open(filepath.Join(t.TempDir(), "credentials.enc"), "GOFM_TEST_FM_VAULT")
	if err != nil {
		t.Fatal(err)
	}
	if err := vault.Put("primary", credentials.Entry{Username: "fm-user", Password: "fm-pass"}); err != nil {
		t.Fatal(err)
	}
	server := New(Config{Tokens: []string{"app-token"}, AdminTokens: []string{"admin-token"}, Security: SecurityConfig{MaxRequestBodyBytes: 1 << 20, RateLimitPerMinute: 100}, FileMakerConnections: []filemaker.Target{{Name: "demo", Adapter: "dataapi", BaseURL: host.URL, Credential: "primary", DefaultDatabase: "CRM", AllowedDatabases: []string{"CRM"}, AllowedLayouts: []string{"Customers"}, AllowedOperations: []string{"find"}}}})
	server.vault = vault
	server.fm = filemaker.NewGateway(server.config.FileMakerConnections, vault, host.Client())
	body := `{"connection":"demo","operation":"find","layout":"Customers","query":[],"limit":50}`
	unauthorized := newRecorder()
	unauthorizedRequest := newRequest(t, http.MethodPost, "/api/filemaker/execute", strings.NewReader(body))
	unauthorizedRequest.Header.Set("Content-Type", "application/json")
	server.ServeHTTP(unauthorized, unauthorizedRequest)
	if unauthorized.status != http.StatusUnauthorized {
		t.Fatalf("unauthorized status=%d", unauthorized.status)
	}
	response := newRecorder()
	request := newRequest(t, http.MethodPost, "/api/filemaker/execute", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer app-token")
	server.ServeHTTP(response, request)
	if response.status != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.status, response.body.String())
	}
	if strings.Contains(response.body.String(), "private-session") || strings.Contains(response.body.String(), "fm-pass") {
		t.Fatal("FileMaker secret leaked to caller")
	}
}
