package server

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"gofm-server/internal/storage"
)

type Server struct {
	config Config
	client *http.Client
	logs   *logStore
	store  *storage.Store
	mu     sync.RWMutex
}

func New(config Config) *Server {
	if config.Logs.MaxEntries < 1 {
		config.Logs.MaxEntries = 50
	}
	logs, err := newLogStore(config.Logs)
	if err != nil {
		logs, _ = newLogStore(LogConfig{MaxEntries: config.Logs.MaxEntries})
	}
	var store *storage.Store
	if config.Storage.DBFile != "" {
		store, err = storage.Open(config.Storage.DBFile)
		if err != nil {
			log.Printf("storage disabled: %v", err)
		}
	}
	return &Server{config: config, client: &http.Client{Timeout: 30 * time.Second}, logs: logs, store: store}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	publicPath := r.URL.Path
	request := r
	if s.config.BasePath != "" {
		if publicPath != s.config.BasePath && !strings.HasPrefix(publicPath, s.config.BasePath+"/") {
			writeError(w, http.StatusNotFound, "gateway_path_not_found", "request is outside the configured gateway base path")
			return
		}
		request = r.Clone(r.Context())
		requestURL := *r.URL
		requestURL.Path = strings.TrimPrefix(publicPath, s.config.BasePath)
		if requestURL.Path == "" {
			requestURL.Path = "/"
		}
		request.URL = &requestURL
	}
	id := requestID(r)
	hookAfterRequested := requestHookAfterRequested(request)
	w.Header().Set("X-Request-ID", id)
	recorder := &captureWriter{ResponseWriter: w, capture: s.config.Logs.CaptureBodyPreview}
	started := time.Now()
	s.serve(recorder, request)
	status := recorder.status
	if status == 0 {
		status = http.StatusOK
	}
	entry := LogEntry{ID: id, StartedAt: started.UTC().Format(time.RFC3339Nano), DurationMS: time.Since(started).Microseconds() / 1000, Method: r.Method, Path: publicPath, Status: status, Outcome: "success", ResponsePreview: recorder.preview(), HookAfterRequested: hookAfterRequested}
	if route := s.matchAnyMethod(request.URL.Path); route != nil {
		entry.Route = route.Path
		entry.Upstream = route.Target.URL
	}
	if status >= 400 {
		entry.Outcome = "failure"
	}
	s.logs.add(entry)
	if s.store != nil {
		_ = s.store.Append(context.Background(), storage.RequestRecord{ID: entry.ID, StartedAt: started, DurationMS: entry.DurationMS, Method: entry.Method, Path: entry.Path, Status: entry.Status, Route: entry.Route, Upstream: entry.Upstream, Outcome: entry.Outcome, ResponsePreview: entry.ResponsePreview, HookAfterRequested: entry.HookAfterRequested}, s.logs.max)
	}
}

func (s *Server) serve(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if s.allowedOrigin(origin) {
		setCORSHeaders(w.Header(), origin)
	}
	if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
		s.handlePreflight(w, r, origin)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/__gofm/") {
		s.handleAdmin(w, r)
		return
	}
	if r.URL.Path == "/health" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only GET is allowed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	}
	route := s.match(r.Method, r.URL.Path)
	if route == nil {
		writeError(w, http.StatusNotFound, "route_not_found", "no route matches this method and path")
		return
	}
	if len(s.config.Tokens) > 0 && !s.authorized(r.Header.Get("Authorization"), s.config.Tokens) {
		w.Header().Set("WWW-Authenticate", "Bearer")
		writeError(w, http.StatusUnauthorized, "unauthorized", "a valid bearer token is required")
		return
	}
	w.Header().Set("X-GoFM-Auth-Verified", "true")
	s.dispatchHTTP(w, r, route.Target.URL)
}

func (s *Server) handleAdmin(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(r.Header.Get("Authorization"), s.config.AdminTokens) {
		w.Header().Set("WWW-Authenticate", "Bearer")
		writeError(w, http.StatusUnauthorized, "admin_unauthorized", "a valid admin bearer token is required")
		return
	}
	switch r.URL.Path {
	case "/__gofm/overview":
		s.mu.RLock()
		routeCount := len(s.config.Routes)
		s.mu.RUnlock()
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "route_count": routeCount, "log_count": len(s.logs.list()), "log_capacity": s.logs.max, "logs_encrypted": s.logs.box != nil, "persistent_storage": s.store != nil})
	case "/__gofm/routes":
		if r.Method == http.MethodPost {
			s.addRoute(w, r)
			return
		}
		s.mu.RLock()
		configuredRoutes := append([]Route(nil), s.config.Routes...)
		s.mu.RUnlock()
		routes := make([]map[string]any, 0, len(configuredRoutes))
		for _, route := range configuredRoutes {
			routes = append(routes, map[string]any{"id": route.ID, "path": route.Path, "methods": route.Methods, "target_type": route.Target.Type, "target": safeTarget(route.Target.URL), "hooks": route.Hooks})
		}
		writeJSON(w, http.StatusOK, map[string]any{"routes": routes})
	case "/__gofm/logs":
		writeJSON(w, http.StatusOK, map[string]any{"logs": s.logs.list()})
	case "/__gofm/storage/logs":
		if s.store == nil {
			writeError(w, http.StatusNotFound, "storage_disabled", "persistent storage is not configured")
			return
		}
		records, err := s.store.Recent(r.Context(), s.logs.max)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "storage_read_failed", "could not read persistent logs")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"logs": records})
	default:
		writeError(w, http.StatusNotFound, "admin_route_not_found", "unknown admin endpoint")
	}
}

func (s *Server) addRoute(w http.ResponseWriter, r *http.Request) {
	var route Route
	if err := json.NewDecoder(io.LimitReader(r.Body, 64*1024)).Decode(&route); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_route", "route JSON is invalid")
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if route.ID == "" {
		route.ID = "runtime-route"
	}
	if route.Hooks.Before == "" {
		route.Hooks.Before = "fireGoHookBefore"
	}
	if route.Hooks.After == "" {
		route.Hooks.After = "fireGoHookAfter"
	}
	for _, existing := range s.config.Routes {
		if existing.ID == route.ID {
			writeError(w, http.StatusConflict, "route_exists", "a route with this id already exists")
			return
		}
		for _, method := range existing.Methods {
			for _, candidate := range route.Methods {
				if strings.EqualFold(method, candidate) && existing.Path == route.Path {
					writeError(w, http.StatusConflict, "route_exists", "this method and path already exist")
					return
				}
			}
		}
	}
	candidate := Config{Address: s.config.Address, BasePath: s.config.BasePath, Tokens: s.config.Tokens, AdminTokens: s.config.AdminTokens, CORSOrigins: s.config.CORSOrigins, Routes: []Route{route}, Logs: s.config.Logs}
	if err := candidate.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_route", err.Error())
		return
	}
	route = candidate.Routes[0]
	s.config.Routes = append(s.config.Routes, route)
	writeJSON(w, http.StatusCreated, map[string]any{"route": route})
}

func safeTarget(raw string) string {
	target, err := url.Parse(raw)
	if err != nil {
		return "invalid"
	}
	return target.Scheme + "://" + target.Host + target.Path
}

func (s *Server) allowedOrigin(origin string) bool {
	for _, allowed := range s.config.CORSOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
}

func (s *Server) handlePreflight(w http.ResponseWriter, r *http.Request, origin string) {
	if !s.allowedOrigin(origin) {
		writeError(w, http.StatusForbidden, "origin_not_allowed", "the request origin is not allowed")
		return
	}
	route := s.match(r.Header.Get("Access-Control-Request-Method"), r.URL.Path)
	if r.URL.Path != "/health" && !strings.HasPrefix(r.URL.Path, "/__gofm/") && route == nil {
		writeError(w, http.StatusNotFound, "route_not_found", "no route matches this method and path")
		return
	}
	if !allowedCORSHeaders(r.Header.Get("Access-Control-Request-Headers")) {
		writeError(w, http.StatusForbidden, "headers_not_allowed", "the requested CORS headers are not allowed")
		return
	}
	methods := "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	if route != nil {
		methods = strings.Join(route.Methods, ", ")
	}
	w.Header().Set("Access-Control-Allow-Methods", methods)
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	w.Header().Set("Access-Control-Max-Age", "600")
	w.WriteHeader(http.StatusNoContent)
}

func setCORSHeaders(headers http.Header, origin string) {
	headers.Set("Access-Control-Allow-Origin", origin)
	headers.Add("Vary", "Origin")
}

func allowedCORSHeaders(requested string) bool {
	if requested == "" {
		return true
	}
	for _, header := range strings.Split(requested, ",") {
		switch strings.ToLower(strings.TrimSpace(header)) {
		case "authorization", "content-type":
		default:
			return false
		}
	}
	return true
}

func (s *Server) match(method, path string) *Route {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.config.Routes {
		route := &s.config.Routes[i]
		if route.Path != path {
			continue
		}
		for _, allowed := range route.Methods {
			if strings.EqualFold(allowed, method) {
				return route
			}
		}
	}
	return nil
}
func (s *Server) matchAnyMethod(path string) *Route {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.config.Routes {
		if s.config.Routes[i].Path == path {
			return &s.config.Routes[i]
		}
	}
	return nil
}

func (s *Server) authorized(header string, tokens []string) bool {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return false
	}
	presented := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	for _, token := range tokens {
		if subtle.ConstantTimeCompare([]byte(presented), []byte(token)) == 1 {
			return true
		}
	}
	return false
}

func (s *Server) dispatchHTTP(w http.ResponseWriter, r *http.Request, rawTarget string) {
	target, _ := url.Parse(rawTarget)
	forwardURL := *target
	forwardURL.Path = joinPaths(target.Path, r.URL.Path)
	forwardURL.RawQuery = r.URL.RawQuery
	request, err := http.NewRequestWithContext(r.Context(), r.Method, forwardURL.String(), r.Body)
	if err != nil {
		writeError(w, http.StatusBadGateway, "upstream_request_failed", "could not create upstream request")
		return
	}
	copyHeaders(request.Header, r.Header)
	request.Header.Del("Authorization")
	request.Header.Del("Cookie")
	request.Header.Del("Proxy-Authorization")
	request.Host = target.Host
	response, err := s.client.Do(request)
	if err != nil {
		writeError(w, http.StatusBadGateway, "upstream_unavailable", "upstream service is unavailable")
		return
	}
	defer response.Body.Close()
	copyHeaders(w.Header(), response.Header)
	w.WriteHeader(response.StatusCode)
	_, _ = io.Copy(w, response.Body)
}

func joinPaths(base, request string) string {
	if base == "" || base == "/" {
		return request
	}
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(request, "/")
}

var hopHeaders = map[string]struct{}{"Connection": {}, "Keep-Alive": {}, "Proxy-Authenticate": {}, "Proxy-Authorization": {}, "Te": {}, "Trailer": {}, "Transfer-Encoding": {}, "Upgrade": {}}

func copyHeaders(destination, source http.Header) {
	for key, values := range source {
		if _, hop := hopHeaders[http.CanonicalHeaderKey(key)]; hop {
			continue
		}
		for _, value := range values {
			destination.Add(key, value)
		}
	}
}

type captureWriter struct {
	http.ResponseWriter
	status  int
	capture bool
	body    strings.Builder
}

func (w *captureWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
func (w *captureWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	if w.capture && w.body.Len() < 4096 {
		remaining := 4096 - w.body.Len()
		if len(data) > remaining {
			data = data[:remaining]
		}
		w.body.Write(data)
	}
	return w.ResponseWriter.Write(data)
}
func (w *captureWriter) preview() string { return w.body.String() }

func requestHookAfterRequested(r *http.Request) bool {
	if r.Body == nil || !strings.Contains(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
		return false
	}
	data, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return false
	}
	r.Body.Close()
	r.Body = io.NopCloser(strings.NewReader(string(data)))
	var payload map[string]any
	if json.Unmarshal(data, &payload) != nil {
		return false
	}
	value, ok := payload["fireGoHookAfter"]
	return ok && value != nil && value != false && value != ""
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}
