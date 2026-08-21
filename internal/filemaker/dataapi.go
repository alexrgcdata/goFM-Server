package filemaker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type cachedSession struct {
	token   string
	expires time.Time
}

type DataAPIAdapter struct {
	client   *http.Client
	mu       sync.Mutex
	sessions map[string]cachedSession
}

func NewDataAPIAdapter(client *http.Client) *DataAPIAdapter {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &DataAPIAdapter{client: client, sessions: make(map[string]cachedSession)}
}

func (a *DataAPIAdapter) Execute(ctx context.Context, target Target, request Request, credential Credential) (Response, error) {
	token, err := a.session(ctx, target, request.Database, credential)
	if err != nil {
		return Response{}, err
	}
	result, unauthorized, err := a.executeWithToken(ctx, target, request, token)
	if !unauthorized {
		return result, err
	}
	a.forget(target, request.Database, credential.Username)
	token, err = a.session(ctx, target, request.Database, credential)
	if err != nil {
		return Response{}, err
	}
	result, _, err = a.executeWithToken(ctx, target, request, token)
	return result, err
}

func (a *DataAPIAdapter) session(ctx context.Context, target Target, database string, credential Credential) (string, error) {
	key := target.Name + "\x00" + database + "\x00" + credential.Username
	a.mu.Lock()
	if session, ok := a.sessions[key]; ok && time.Now().Before(session.expires) {
		a.mu.Unlock()
		return session.token, nil
	}
	a.mu.Unlock()
	endpoint := target.BaseURL + "/fmi/data/vLatest/databases/" + url.PathEscape(database) + "/sessions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader("{}"))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(credential.Username, credential.Password)
	response, err := a.client.Do(req)
	if err != nil {
		return "", &APIError{Status: http.StatusBadGateway, Code: "filemaker_unavailable", Message: "FileMaker is unavailable"}
	}
	var envelope struct {
		Response struct {
			Token string `json:"token"`
		} `json:"response"`
		Messages []struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"messages"`
	}
	if err := decodeJSON(response, &envelope); err != nil {
		return "", err
	}
	if envelope.Response.Token == "" {
		return "", &APIError{Status: http.StatusBadGateway, Code: "filemaker_login_failed", Message: "FileMaker login did not return a session token"}
	}
	a.mu.Lock()
	a.sessions[key] = cachedSession{token: envelope.Response.Token, expires: time.Now().Add(14 * time.Minute)}
	a.mu.Unlock()
	return envelope.Response.Token, nil
}

func (a *DataAPIAdapter) forget(target Target, database, username string) {
	a.mu.Lock()
	delete(a.sessions, target.Name+"\x00"+database+"\x00"+username)
	a.mu.Unlock()
}

func (a *DataAPIAdapter) executeWithToken(ctx context.Context, target Target, request Request, token string) (Response, bool, error) {
	base := target.BaseURL + "/fmi/data/vLatest/databases/" + url.PathEscape(request.Database) + "/layouts/" + url.PathEscape(request.Layout)
	method, endpoint := http.MethodGet, base
	var body []byte
	switch request.Operation {
	case "find":
		method, endpoint = http.MethodPost, base+"/_find"
		query := make(map[string]any)
		for _, clause := range request.Query {
			query[clause.Field] = dataAPICriterion(clause.Op, clause.Value)
		}
		payload := map[string]any{"query": []any{query}}
		if request.Offset > 0 {
			payload["offset"] = request.Offset
		}
		if request.Limit > 0 {
			payload["limit"] = request.Limit
		}
		body, _ = json.Marshal(payload)
	case "get":
		endpoint = base + "/records/" + url.PathEscape(request.RecordID)
	case "create":
		method, endpoint = http.MethodPost, base+"/records"
		body, _ = json.Marshal(map[string]any{"fieldData": request.Fields})
	case "update":
		method, endpoint = http.MethodPatch, base+"/records/"+url.PathEscape(request.RecordID)
		body, _ = json.Marshal(map[string]any{"fieldData": request.Fields, "modId": request.ExpectedModID})
	case "delete":
		method, endpoint = http.MethodDelete, base+"/records/"+url.PathEscape(request.RecordID)
	}
	if request.ScriptAfter != nil {
		values := url.Values{"script": []string{request.ScriptAfter.Name}}
		if request.ScriptAfter.Parameter != "" {
			values.Set("script.param", request.ScriptAfter.Parameter)
		}
		endpoint += "?" + values.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		return Response{}, false, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	response, err := a.client.Do(req)
	if err != nil {
		return Response{}, false, &APIError{Status: http.StatusBadGateway, Code: "filemaker_unavailable", Message: "FileMaker is unavailable"}
	}
	if response.StatusCode == http.StatusUnauthorized {
		response.Body.Close()
		return Response{}, true, nil
	}
	var envelope struct {
		Response struct {
			Data []struct {
				FieldData map[string]any `json:"fieldData"`
				RecordID  string         `json:"recordId"`
				ModID     string         `json:"modId"`
			} `json:"data"`
			DataInfo struct {
				FoundCount    int `json:"foundCount"`
				ReturnedCount int `json:"returnedCount"`
			} `json:"dataInfo"`
		} `json:"response"`
		Messages []struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"messages"`
	}
	if err := decodeJSON(response, &envelope); err != nil {
		return Response{}, false, err
	}
	if len(envelope.Messages) > 0 && envelope.Messages[0].Code != "0" {
		return Response{}, false, &APIError{Status: http.StatusBadGateway, Code: "filemaker_error", Message: "FileMaker rejected the operation"}
	}
	records := make([]map[string]any, 0, len(envelope.Response.Data))
	for _, item := range envelope.Response.Data {
		record := item.FieldData
		record["_record_id"] = item.RecordID
		record["_mod_id"] = item.ModID
		records = append(records, record)
	}
	limit := request.Limit
	if limit == 0 {
		limit = envelope.Response.DataInfo.ReturnedCount
	}
	offset := request.Offset
	if offset == 0 {
		offset = 1
	}
	return Response{Records: records, FoundCount: envelope.Response.DataInfo.FoundCount, Offset: offset, Limit: limit, Meta: map[string]any{"adapter": "dataapi"}}, false, nil
}

func dataAPICriterion(op string, value any) string {
	text := fmt.Sprint(value)
	switch op {
	case "eq":
		return "==" + text
	case "ne":
		return "!=" + text
	case "gt":
		return ">" + text
	case "gte":
		return ">=" + text
	case "lt":
		return "<" + text
	case "lte":
		return "<=" + text
	case "contains":
		return "*" + text + "*"
	case "begins":
		return text + "*"
	default:
		return text
	}
}
