package filemaker

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// QueryClause is a transport-neutral filter. Adapters translate it into
// FileMaker Data API queries or OData expressions after allow-list validation.
type QueryClause struct {
	Field string `json:"field"`
	Op    string `json:"op"`
	Value any    `json:"value"`
}

type Request struct {
	Connection    string         `json:"connection"`
	Operation     string         `json:"operation"`
	Database      string         `json:"database"`
	Layout        string         `json:"layout"`
	Table         string         `json:"table,omitempty"`
	RecordID      string         `json:"record_id,omitempty"`
	ExpectedModID string         `json:"expected_mod_id,omitempty"`
	ExpectedETag  string         `json:"expected_etag,omitempty"`
	Query         []QueryClause  `json:"query,omitempty"`
	Fields        map[string]any `json:"fields,omitempty"`
	Limit         int            `json:"limit,omitempty"`
	Offset        int            `json:"offset,omitempty"`
	ScriptAfter   *ScriptRequest `json:"script_after,omitempty"`
}

type ScriptRequest struct {
	Name      string `json:"name"`
	Parameter string `json:"parameter,omitempty"`
}

type Response struct {
	Records    []map[string]any `json:"records"`
	FoundCount int              `json:"found_count"`
	Offset     int              `json:"offset"`
	Limit      int              `json:"limit"`
	Meta       map[string]any   `json:"meta,omitempty"`
}

type Credential struct {
	Username string
	Password string
	APIKey   string
	Token    string
}

type CredentialProvider interface {
	Credentials(context.Context, string) (Credential, error)
}
type Adapter interface {
	Execute(context.Context, Request) (Response, error)
	Close() error
}
type Factory interface {
	New(context.Context, Target) (Adapter, error)
}

type Target struct {
	Name              string   `json:"name"`
	Adapter           string   `json:"adapter"`
	BaseURL           string   `json:"base_url"`
	Credential        string   `json:"credential"`
	DefaultDatabase   string   `json:"default_database,omitempty"`
	AllowedDatabases  []string `json:"allowed_databases"`
	AllowedLayouts    []string `json:"allowed_layouts,omitempty"`
	AllowedTables     []string `json:"allowed_tables,omitempty"`
	AllowedOperations []string `json:"allowed_operations,omitempty"`
	AllowedScripts    []string `json:"allowed_scripts,omitempty"`
}

func (r Request) Validate() error {
	switch r.Operation {
	case "find", "get", "create", "update", "delete":
	default:
		return fmt.Errorf("unsupported FileMaker operation %q", r.Operation)
	}
	if r.Database == "" {
		return fmt.Errorf("database is required")
	}
	if r.Layout == "" && r.Table == "" {
		return fmt.Errorf("layout or table is required")
	}
	if r.Limit < 0 || r.Limit > 1000 {
		return fmt.Errorf("limit must be between 0 and 1000")
	}
	if r.Offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}
	if (r.Operation == "get" || r.Operation == "update" || r.Operation == "delete") && r.RecordID == "" {
		return fmt.Errorf("record_id is required for this operation")
	}
	if (r.Operation == "create" || r.Operation == "update") && len(r.Fields) == 0 {
		return fmt.Errorf("fields are required for this operation")
	}
	for _, clause := range r.Query {
		if strings.TrimSpace(clause.Field) == "" {
			return fmt.Errorf("query field is required")
		}
		switch clause.Op {
		case "eq", "ne", "gt", "gte", "lt", "lte", "contains", "begins":
		default:
			return fmt.Errorf("unsupported query operator %q", clause.Op)
		}
	}
	return nil
}

func (t *Target) Validate() error {
	t.Name = strings.TrimSpace(t.Name)
	if t.Name == "" {
		return fmt.Errorf("connection name is required")
	}
	if t.Adapter != "dataapi" && t.Adapter != "odata" {
		return fmt.Errorf("connection %s adapter must be dataapi or odata", t.Name)
	}
	parsed, err := url.Parse(t.BaseURL)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("connection %s base_url must be an absolute HTTPS URL without credentials, query, or fragment", t.Name)
	}
	t.BaseURL = strings.TrimRight(t.BaseURL, "/")
	if t.Credential == "" {
		return fmt.Errorf("connection %s credential is required", t.Name)
	}
	if t.DefaultDatabase != "" && !contains(t.AllowedDatabases, t.DefaultDatabase) {
		return fmt.Errorf("connection %s default_database must be allow-listed", t.Name)
	}
	if len(t.AllowedDatabases) == 0 {
		return fmt.Errorf("connection %s requires allowed_databases", t.Name)
	}
	if len(t.AllowedOperations) == 0 {
		t.AllowedOperations = []string{"find", "get"}
	}
	for _, operation := range t.AllowedOperations {
		switch operation {
		case "find", "get", "create", "update", "delete":
		default:
			return fmt.Errorf("connection %s has unsupported operation %s", t.Name, operation)
		}
	}
	return nil
}

func (t Target) Authorize(r *Request) error {
	if r.Database == "" {
		r.Database = t.DefaultDatabase
	}
	if !contains(t.AllowedDatabases, r.Database) {
		return fmt.Errorf("database is not allowed for this connection")
	}
	if !contains(t.AllowedOperations, r.Operation) {
		return fmt.Errorf("operation is not allowed for this connection")
	}
	if t.Adapter == "dataapi" {
		if r.Layout == "" || !contains(t.AllowedLayouts, r.Layout) {
			return fmt.Errorf("layout is not allowed for this connection")
		}
	} else {
		if r.Table == "" || !contains(t.AllowedTables, r.Table) {
			return fmt.Errorf("table is not allowed for this connection")
		}
		if r.ScriptAfter != nil {
			return fmt.Errorf("script_after is supported only by the Data API adapter")
		}
	}
	if r.Operation == "update" && t.Adapter == "dataapi" && r.ExpectedModID == "" {
		return fmt.Errorf("expected_mod_id is required for Data API updates")
	}
	if r.Operation == "update" && t.Adapter == "odata" && r.ExpectedETag == "" {
		return fmt.Errorf("expected_etag is required for OData updates")
	}
	if r.ScriptAfter != nil && !contains(t.AllowedScripts, r.ScriptAfter.Name) {
		return fmt.Errorf("script is not allowed for this connection")
	}
	return r.Validate()
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

type Gateway struct {
	targets  map[string]Target
	provider CredentialProvider
	dataAPI  *DataAPIAdapter
	odata    *ODataAdapter
}

func NewGateway(targets []Target, provider CredentialProvider, client *http.Client) *Gateway {
	configured := make(map[string]Target, len(targets))
	for _, target := range targets {
		configured[target.Name] = target
	}
	return &Gateway{targets: configured, provider: provider, dataAPI: NewDataAPIAdapter(client), odata: NewODataAdapter(client)}
}

func (g *Gateway) Execute(ctx context.Context, request Request) (Response, error) {
	target, ok := g.targets[request.Connection]
	if !ok {
		return Response{}, fmt.Errorf("connection is not configured")
	}
	if err := target.Authorize(&request); err != nil {
		return Response{}, err
	}
	if g.provider == nil {
		return Response{}, fmt.Errorf("credential vault is unavailable")
	}
	credential, err := g.provider.Credentials(ctx, target.Credential)
	if err != nil {
		return Response{}, fmt.Errorf("configured credential is unavailable")
	}
	if target.Adapter == "dataapi" {
		return g.dataAPI.Execute(ctx, target, request, credential)
	}
	return g.odata.Execute(ctx, target, request, credential)
}
