package server

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type Server struct {
	config       Config
	client       *http.Client
	tokensMu     sync.RWMutex
	issuedTokens map[string]struct{}
}

func New(config Config) *Server {
	return &Server{config: config, client: http.DefaultClient, issuedTokens: make(map[string]struct{})}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if s.allowedOrigin(origin) {
		setCORSHeaders(w.Header(), origin)
	}
	if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
		s.handlePreflight(w, r, origin)
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
	if r.URL.Path == "/auth/token" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only POST is allowed")
			return
		}
		s.issueToken(w, r)
		return
	}

	route := s.match(r.Method, r.URL.Path)
	if route == nil {
		writeError(w, http.StatusNotFound, "route_not_found", "no route matches this method and path")
		return
	}
	if len(s.config.Tokens) > 0 && !s.authorized(r.Header.Get("Authorization")) {
		w.Header().Set("WWW-Authenticate", "Bearer")
		writeError(w, http.StatusUnauthorized, "unauthorized", "a valid bearer token is required")
		return
	}
	s.dispatchHTTP(w, r, route.Target.URL)
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
	if route == nil {
		writeError(w, http.StatusNotFound, "route_not_found", "no route matches this method and path")
		return
	}
	requestedHeaders := r.Header.Get("Access-Control-Request-Headers")
	if !allowedCORSHeaders(requestedHeaders) {
		writeError(w, http.StatusForbidden, "headers_not_allowed", "the requested CORS headers are not allowed")
		return
	}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(route.Methods, ", "))
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

func (s *Server) authorized(header string) bool {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return false
	}
	presented := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	for _, token := range s.config.Tokens {
		if subtle.ConstantTimeCompare([]byte(presented), []byte(token)) == 1 {
			return true
		}
	}
	s.tokensMu.RLock()
	_, issued := s.issuedTokens[presented]
	s.tokensMu.RUnlock()
	if issued {
		return true
	}
	return false
}

func (s *Server) issueToken(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	if err := decoder.Decode(&credentials); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "body must be a JSON object with username and password")
		return
	}
	valid := false
	for _, configured := range s.config.Credentials {
		if subtle.ConstantTimeCompare([]byte(credentials.Username), []byte(configured.Username)) == 1 && subtle.ConstantTimeCompare([]byte(credentials.Password), []byte(configured.Password)) == 1 {
			valid = true
			break
		}
	}
	if !valid {
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "username or password is incorrect")
		return
	}
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		writeError(w, http.StatusInternalServerError, "token_generation_failed", "could not create an access token")
		return
	}
	token := base64.RawURLEncoding.EncodeToString(bytes)
	s.tokensMu.Lock()
	s.issuedTokens[token] = struct{}{}
	s.tokensMu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{"access_token": token, "token_type": "Bearer"})
}

func (s *Server) dispatchHTTP(w http.ResponseWriter, r *http.Request, rawTarget string) {
	target, _ := url.Parse(rawTarget) // Config validation has already verified this.
	forwardURL := *target
	forwardURL.Path = joinPaths(target.Path, r.URL.Path)
	forwardURL.RawQuery = r.URL.RawQuery

	request, err := http.NewRequestWithContext(r.Context(), r.Method, forwardURL.String(), r.Body)
	if err != nil {
		writeError(w, http.StatusBadGateway, "upstream_request_failed", "could not create upstream request")
		return
	}
	copyHeaders(request.Header, r.Header)
	request.Header.Del("Authorization") // Do not leak this server's bearer token upstream.
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

var hopHeaders = map[string]struct{}{
	"Connection": {}, "Keep-Alive": {}, "Proxy-Authenticate": {}, "Proxy-Authorization": {},
	"Te": {}, "Trailer": {}, "Transfer-Encoding": {}, "Upgrade": {},
}

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

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}
