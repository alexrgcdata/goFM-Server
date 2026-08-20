package filemaker

import (
	"context"
	"fmt"
)

// QueryClause is a transport-neutral filter. Adapters translate it into
// FileMaker Data API queries or OData expressions after allow-list validation.
type QueryClause struct {
	Field string `json:"field"`
	Op    string `json:"op"`
	Value any    `json:"value"`
}

type Request struct {
	Operation string         `json:"operation"`
	Database  string         `json:"database"`
	Layout    string         `json:"layout"`
	Table     string         `json:"table,omitempty"`
	RecordID  string         `json:"record_id,omitempty"`
	Query     []QueryClause  `json:"query,omitempty"`
	Fields    map[string]any `json:"fields,omitempty"`
	Limit     int            `json:"limit,omitempty"`
	Offset    int            `json:"offset,omitempty"`
}

type Response struct {
	Records    []map[string]any `json:"records"`
	FoundCount int              `json:"found_count"`
	Offset     int              `json:"offset"`
	Limit      int              `json:"limit"`
	RawMeta    map[string]any   `json:"raw_meta,omitempty"`
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
	Name       string `json:"name"`
	Adapter    string `json:"adapter"`
	BaseURL    string `json:"base_url"`
	Database   string `json:"database"`
	Layout     string `json:"layout"`
	Credential string `json:"credential"`
}

func (r Request) Validate() error {
	switch r.Operation {
	case "find", "get", "create", "update", "delete":
	default:
		return fmt.Errorf("unsupported FileMaker operation %q", r.Operation)
	}
	if r.Database == "" || r.Layout == "" {
		return fmt.Errorf("database and layout are required")
	}
	if r.Limit < 0 || r.Limit > 1000 {
		return fmt.Errorf("limit must be between 0 and 1000")
	}
	if r.Offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}
	return nil
}
