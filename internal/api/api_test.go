package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const testToken = "test-only-token-012345678901234567890123456789"

func testConfig() Config {
	return Config{Address: "127.0.0.1:8080", Token: testToken, MaxRequestBytes: 1024}
}

func TestDispatch(t *testing.T) {
	s, err := New(testConfig())
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, method, path, token, contentType, body string
		status                                       int
		result, code                                 string
	}{
		{"hello", "POST", "/api", testToken, "application/json", `{"action":"hello","params":[]}`, 200, `"hello world"`, ""},
		{"success", "POST", "/api", testToken, "application/json", `{"action":"success","params":[]}`, 200, `"successful"`, ""},
		{"boolean", "POST", "/api", testToken, "application/json", `{"action":"boolean","params":[]}`, 200, `true`, ""},
		{"dynamic params", "POST", "/api", testToken, "application/json; charset=utf-8", `{"action":"echo","params":["hi",1234567890123456789,true,{"name":"Alex"},null],"time":"2026-10-30T12:00:00Z"}`, 200, `["hi",1234567890123456789,true,{"name":"Alex"},null]`, ""},
		{"public health", "GET", "/health", "", "", "", 200, `"ok"`, ""},
		{"missing auth", "POST", "/api", "", "application/json", `not json`, 401, `null`, "unauthorized"},
		{"bad auth", "POST", "/api", "incorrect", "application/json", `{"action":"hello","params":[]}`, 401, `null`, "unauthorized"},
		{"unknown action", "POST", "/api", testToken, "application/json", `{"action":"os.exec","params":[]}`, 404, `null`, "unknown_action"},
		{"invalid params", "POST", "/api", testToken, "application/json", `{"action":"hello","params":[1]}`, 400, `null`, "invalid_params"},
		{"missing params", "POST", "/api", testToken, "application/json", `{"action":"hello"}`, 400, `null`, "invalid_request"},
		{"null params", "POST", "/api", testToken, "application/json", `{"action":"hello","params":null}`, 400, `null`, "invalid_request"},
		{"object params", "POST", "/api", testToken, "application/json", `{"action":"hello","params":{}}`, 400, `null`, "invalid_json"},
		{"null body", "POST", "/api", testToken, "application/json", `null`, 400, `null`, "invalid_request"},
		{"extra field", "POST", "/api", testToken, "application/json", `{"action":"hello","params":[],"host":"untrusted"}`, 400, `null`, "invalid_json"},
		{"trailing value", "POST", "/api", testToken, "application/json", `{"action":"hello","params":[]} {}`, 400, `null`, "invalid_json"},
		{"broken JSON", "POST", "/api", testToken, "application/json", `{`, 400, `null`, "invalid_json"},
		{"empty body", "POST", "/api", testToken, "application/json", ``, 400, `null`, "invalid_json"},
		{"invalid time", "POST", "/api", testToken, "application/json", `{"action":"hello","params":[],"time":"yesterday"}`, 400, `null`, "invalid_time"},
		{"wrong content type", "POST", "/api", testToken, "text/plain", `{}`, 415, `null`, "unsupported_media_type"},
		{"wrong method", "GET", "/api", testToken, "", "", 405, `null`, "method_not_allowed"},
		{"health method", "POST", "/health", "", "", "", 405, `null`, "method_not_allowed"},
		{"missing route", "GET", "/other", "", "", "", 404, `null`, "not_found"},
		{"large body", "POST", "/api", testToken, "application/json", `{"action":"echo","params":["` + strings.Repeat("x", 1100) + `"]}`, 413, `null`, "request_too_large"},
		{"large trailing whitespace", "POST", "/api", testToken, "application/json", `{"action":"hello","params":[]}` + strings.Repeat(" ", 1100), 413, `null`, "request_too_large"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.token != "" {
				r.Header.Set("Authorization", "Bearer "+tc.token)
			}
			if tc.contentType != "" {
				r.Header.Set("Content-Type", tc.contentType)
			}
			w := httptest.NewRecorder()
			s.ServeHTTP(w, r)
			if w.Code != tc.status {
				t.Fatalf("status %d, want %d: %s", w.Code, tc.status, w.Body)
			}
			var result struct {
				Success   bool
				Result    json.RawMessage
				Error     *APIError
				RequestID string `json:"request_id"`
				Time      string
			}
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if string(result.Result) != tc.result || result.Success != (tc.status < 400) {
				t.Fatalf("unexpected response: %s", w.Body)
			}
			if tc.code != "" && (result.Error == nil || result.Error.Code != tc.code) {
				t.Fatalf("wrong error: %s", w.Body)
			}
			if tc.code == "" && result.Error != nil {
				t.Fatalf("unexpected error: %s", w.Body)
			}
			if len(result.RequestID) != 32 || result.RequestID != w.Header().Get("X-Request-ID") {
				t.Fatal("missing request ID")
			}
			if _, err := time.Parse(time.RFC3339Nano, result.Time); err != nil {
				t.Fatal(err)
			}
			if w.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("response can be cached")
			}
		})
	}
}

func TestAuthenticationPreventsDispatch(t *testing.T) {
	s, _ := New(testConfig())
	called := false
	s.actions["probe"] = func(context.Context, []json.RawMessage) (any, error) { called = true; return true, nil }
	for _, headers := range [][]string{{}, {"Basic " + testToken}, {"Bearer wrong"}, {"Bearer " + testToken, "Bearer " + testToken}} {
		r := httptest.NewRequest("POST", "/api", strings.NewReader(`{"action":"probe","params":[]}`))
		r.Header.Set("Content-Type", "application/json")
		for _, header := range headers {
			r.Header.Add("Authorization", header)
		}
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		if w.Code != 401 || called {
			t.Fatal("unauthenticated action dispatched")
		}
	}
}

func TestInternalErrorsStayPrivate(t *testing.T) {
	s, _ := New(testConfig())
	s.actions["fail"] = func(context.Context, []json.RawMessage) (any, error) { return nil, errors.New("private credentials") }
	r := httptest.NewRequest("POST", "/api", strings.NewReader(`{"action":"fail","params":[]}`))
	r.Header.Set("Authorization", "Bearer "+testToken)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != 500 || strings.Contains(w.Body.String(), "private credentials") {
		t.Fatal("internal error exposed")
	}
}

func TestConfig(t *testing.T) {
	for _, change := range []func(*Config){
		func(c *Config) { c.Token = "" }, func(c *Config) { c.Token = "short" },
		func(c *Config) { c.Address = ":invalid" }, func(c *Config) { c.Address = ":0" },
		func(c *Config) { c.MaxRequestBytes = 0 }, func(c *Config) { c.MaxRequestBytes = 2 << 20 },
	} {
		c := testConfig()
		change(&c)
		if _, err := New(c); err == nil {
			t.Fatal("invalid config accepted")
		}
	}
	path := filepath.Join(t.TempDir(), "config.json")
	data, _ := json.Marshal(testConfig())
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfig(path); err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{`{}`, `null`, string(data) + ` {}`, `{"tokens":["secret"]}`, `{"token":"DO-NOT-ECHO",`} {
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		_, err := LoadConfig(path)
		if err == nil || strings.Contains(err.Error(), "DO-NOT-ECHO") {
			t.Fatal("bad config accepted or echoed")
		}
	}
}

func TestConcurrentRequests(t *testing.T) {
	s, _ := New(testConfig())
	host := httptest.NewServer(s)
	defer host.Close()
	finished := make(chan error, 20)
	for i := 0; i < 20; i++ {
		go func() {
			r, _ := http.NewRequest("POST", host.URL+"/api", strings.NewReader(`{"action":"hello","params":[]}`))
			r.Header.Set("Authorization", "Bearer "+testToken)
			r.Header.Set("Content-Type", "application/json")
			response, err := host.Client().Do(r)
			if err == nil {
				response.Body.Close()
				if response.StatusCode != 200 {
					err = errors.New("concurrent request failed")
				}
			}
			finished <- err
		}()
	}
	for i := 0; i < 20; i++ {
		if err := <-finished; err != nil {
			t.Fatal(err)
		}
	}
}
