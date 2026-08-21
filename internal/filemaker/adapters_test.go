package filemaker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDataAPIAdapterLogsInAndNormalizesFind(t *testing.T) {
	loginCount := 0
	host := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/sessions"):
			loginCount++
			user, password, ok := r.BasicAuth()
			if !ok || user != "fm-user" || password != "fm-pass" {
				t.Error("missing FileMaker Basic auth")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"response": map[string]any{"token": "fm-session"}, "messages": []any{map[string]string{"code": "0", "message": "OK"}}})
		case strings.HasSuffix(r.URL.Path, "/_find"):
			if r.Header.Get("Authorization") != "Bearer fm-session" {
				t.Error("missing FileMaker session token")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"response": map[string]any{"data": []any{map[string]any{"fieldData": map[string]any{"Name": "Alex"}, "recordId": "1", "modId": "3"}}, "dataInfo": map[string]any{"foundCount": 1, "returnedCount": 1}}, "messages": []any{map[string]string{"code": "0", "message": "OK"}}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer host.Close()
	adapter := NewDataAPIAdapter(host.Client())
	target := Target{Name: "demo", BaseURL: host.URL}
	response, err := adapter.Execute(context.Background(), target, Request{Operation: "find", Database: "CRM", Layout: "Customers", Limit: 50, Offset: 1}, Credential{Username: "fm-user", Password: "fm-pass"})
	if err != nil {
		t.Fatal(err)
	}
	if loginCount != 1 || response.FoundCount != 1 || response.Records[0]["Name"] != "Alex" || response.Records[0]["_record_id"] != "1" {
		t.Fatalf("unexpected normalized response: %#v", response)
	}
}

func TestODataAdapterBuildsSafeFilterAndNormalizesRecords(t *testing.T) {
	host := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok || user != "fm-user" || password != "fm-pass" {
			t.Error("missing OData Basic auth")
		}
		if !strings.Contains(r.URL.Query().Get("$filter"), "Status eq 'Active'") {
			t.Errorf("unexpected filter: %s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"value": []any{map[string]any{"ROWID": 1, "Name": "Alex"}}, "@odata.count": 1})
	}))
	defer host.Close()
	response, err := NewODataAdapter(host.Client()).Execute(context.Background(), Target{BaseURL: host.URL}, Request{Operation: "find", Database: "CRM", Table: "Customers", Query: []QueryClause{{Field: "Status", Op: "eq", Value: "Active"}}, Limit: 10, Offset: 1}, Credential{Username: "fm-user", Password: "fm-pass"})
	if err != nil {
		t.Fatal(err)
	}
	if response.FoundCount != 1 || response.Records[0]["Name"] != "Alex" {
		t.Fatalf("unexpected normalized response: %#v", response)
	}
	if _, err := odataFilter(QueryClause{Field: "Name) or true", Op: "eq", Value: "x"}); err == nil {
		t.Fatal("unsafe OData field was accepted")
	}
}
