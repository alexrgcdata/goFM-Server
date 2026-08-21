package filemaker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ODataAdapter struct{ client *http.Client }

func NewODataAdapter(client *http.Client) *ODataAdapter {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &ODataAdapter{client: client}
}

var odataField = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.]*$`)

func (a *ODataAdapter) Execute(ctx context.Context, target Target, request Request, credential Credential) (Response, error) {
	base := target.BaseURL + "/fmi/odata/v4/" + url.PathEscape(request.Database) + "/" + url.PathEscape(request.Table)
	method, endpoint := http.MethodGet, base
	var body []byte
	switch request.Operation {
	case "find":
		values := url.Values{}
		values.Set("$count", "true")
		filters := make([]string, 0, len(request.Query))
		for _, clause := range request.Query {
			filter, err := odataFilter(clause)
			if err != nil {
				return Response{}, err
			}
			filters = append(filters, filter)
		}
		if len(filters) > 0 {
			values.Set("$filter", strings.Join(filters, " and "))
		}
		if request.Limit > 0 {
			values.Set("$top", strconv.Itoa(request.Limit))
		}
		if request.Offset > 1 {
			values.Set("$skip", strconv.Itoa(request.Offset-1))
		}
		if encoded := values.Encode(); encoded != "" {
			endpoint += "?" + encoded
		}
	case "get", "update", "delete":
		endpoint += "('" + url.PathEscape(strings.ReplaceAll(request.RecordID, "'", "''")) + "')"
		if request.Operation == "update" {
			method = http.MethodPatch
			body, _ = json.Marshal(request.Fields)
		}
		if request.Operation == "delete" {
			method = http.MethodDelete
		}
	case "create":
		method = http.MethodPost
		body, _ = json.Marshal(request.Fields)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		return Response{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(credential.Username, credential.Password)
	if request.Operation == "update" {
		req.Header.Set("If-Match", request.ExpectedETag)
	}
	response, err := a.client.Do(req)
	if err != nil {
		return Response{}, &APIError{Status: http.StatusBadGateway, Code: "filemaker_unavailable", Message: "FileMaker is unavailable"}
	}
	if method == http.MethodDelete && response.StatusCode >= 200 && response.StatusCode < 300 {
		response.Body.Close()
		return Response{Records: []map[string]any{}, Offset: 1, Meta: map[string]any{"adapter": "odata"}}, nil
	}
	var envelope struct {
		Value []map[string]any `json:"value"`
		Count int              `json:"@odata.count"`
	}
	var single map[string]any
	if request.Operation == "get" || request.Operation == "create" || request.Operation == "update" {
		if err := decodeJSON(response, &single); err != nil {
			return Response{}, err
		}
		envelope.Value = []map[string]any{single}
	} else if err := decodeJSON(response, &envelope); err != nil {
		return Response{}, err
	}
	found := envelope.Count
	if found == 0 {
		found = len(envelope.Value)
	}
	limit := request.Limit
	if limit == 0 {
		limit = len(envelope.Value)
	}
	offset := request.Offset
	if offset == 0 {
		offset = 1
	}
	return Response{Records: envelope.Value, FoundCount: found, Offset: offset, Limit: limit, Meta: map[string]any{"adapter": "odata"}}, nil
}

func odataFilter(clause QueryClause) (string, error) {
	if !odataField.MatchString(clause.Field) {
		return "", fmt.Errorf("invalid OData field name")
	}
	op := map[string]string{"eq": "eq", "ne": "ne", "gt": "gt", "gte": "ge", "lt": "lt", "lte": "le"}[clause.Op]
	if clause.Op == "contains" || clause.Op == "begins" {
		value := strings.ReplaceAll(fmt.Sprint(clause.Value), "'", "''")
		name := "contains"
		if clause.Op == "begins" {
			name = "startswith"
		}
		return fmt.Sprintf("%s(%s,'%s')", name, clause.Field, value), nil
	}
	if op == "" {
		return "", fmt.Errorf("unsupported OData operator")
	}
	return fmt.Sprintf("%s %s %s", clause.Field, op, odataLiteral(clause.Value)), nil
}
func odataLiteral(value any) string {
	switch typed := value.(type) {
	case bool:
		return strconv.FormatBool(typed)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case int:
		return strconv.Itoa(typed)
	default:
		return "'" + strings.ReplaceAll(fmt.Sprint(value), "'", "''") + "'"
	}
}
