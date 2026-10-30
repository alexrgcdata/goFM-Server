package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"
)

// Request is the public contract. Params is always an array; each action
// validates its own parameter count and types. Time is optional client metadata.
type Request struct {
	Action string            `json:"action"`
	Params []json.RawMessage `json:"params"`
	Time   string            `json:"time,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Response struct {
	Success   bool      `json:"success"`
	Result    any       `json:"result"`
	Error     *APIError `json:"error"`
	RequestID string    `json:"request_id"`
	Time      string    `json:"time"`
}

type actionFunc func(context.Context, []json.RawMessage) (any, error)

// actionRoutes is the complete allow-list. Add a named function in functions.go
// and register it here. Requests cannot invoke arbitrary Go functions or URLs.
func actionRoutes() map[string]actionFunc {
	return map[string]actionFunc{
		"hello":   hello,   // [] -> "hello world"
		"success": success, // [] -> "successful"
		"boolean": boolean, // [] -> true (a JSON boolean, not a string)
		"echo":    echo,    // [any JSON values...] -> those same values
	}
}

type Server struct {
	tokenHash [32]byte
	maxBytes  int64
	actions   map[string]actionFunc
}

// New fails closed: no server can be created with missing authentication.
func New(c Config) (*Server, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &Server{tokenHash: sha256.Sum256([]byte(c.Token)), maxBytes: c.MaxRequestBytes, actions: actionRoutes()}, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		http.Error(w, "request ID unavailable", http.StatusInternalServerError)
		return
	}
	id := hex.EncodeToString(idBytes)
	w.Header().Set("X-Request-ID", id)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	reply := func(status int, result any, code, message string) {
		response := Response{Success: status < 400, Result: result, RequestID: id, Time: time.Now().UTC().Format(time.RFC3339Nano)}
		if code != "" {
			response.Error = &APIError{Code: code, Message: message}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(response)
	}
	// HTTP routes live here; the action routes above choose the application job.
	switch r.URL.Path {
	case "/health":
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			reply(405, nil, "method_not_allowed", "use GET")
			return
		}
		reply(200, "ok", "", "")
		return
	case "/api":
		// Authenticate before parsing the caller's JSON or invoking a function.
		headers := r.Header.Values("Authorization")
		parts := strings.Fields(r.Header.Get("Authorization"))
		var candidate string
		if len(headers) == 1 && len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			candidate = parts[1]
		}
		hash := sha256.Sum256([]byte(candidate))
		if candidate == "" || subtle.ConstantTimeCompare(hash[:], s.tokenHash[:]) != 1 {
			w.Header().Set("WWW-Authenticate", `Bearer realm="gofm"`)
			reply(401, nil, "unauthorized", "a valid bearer token is required")
			return
		}
	default:
		reply(404, nil, "not_found", "route not found")
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		reply(405, nil, "method_not_allowed", "use POST")
		return
	}
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		reply(415, nil, "unsupported_media_type", "use Content-Type: application/json")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, s.maxBytes)
	defer r.Body.Close()
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	var request Request
	err = d.Decode(&request)
	if err == nil {
		if trailing := d.Decode(new(any)); trailing != io.EOF {
			if trailing == nil {
				trailing = errors.New("extra JSON value")
			}
			err = trailing
		}
	}
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			reply(413, nil, "request_too_large", "request body exceeds the configured limit")
		} else {
			reply(400, nil, "invalid_json", "provide exactly one JSON object with action, params, and optional time")
		}
		return
	}
	if request.Action == "" || len(request.Action) > 64 || request.Params == nil {
		reply(400, nil, "invalid_request", "action is required and params must be an array")
		return
	}
	if request.Time != "" {
		if _, err := time.Parse(time.RFC3339Nano, request.Time); err != nil {
			reply(400, nil, "invalid_time", "time must be an RFC3339 timestamp")
			return
		}
	}
	fn, ok := s.actions[request.Action]
	if !ok {
		reply(404, nil, "unknown_action", "action is not registered")
		return
	}
	result, err := fn(r.Context(), request.Params)
	if err != nil {
		// Only explicitly safe validation messages are exposed to callers.
		var invalid *ParamsError
		if errors.As(err, &invalid) {
			reply(400, nil, "invalid_params", invalid.Message)
		} else {
			reply(500, nil, "action_failed", "the action could not be completed")
		}
		return
	}
	reply(200, result, "", "")
}
