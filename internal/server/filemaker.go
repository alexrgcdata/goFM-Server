package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"gofm-server/internal/filemaker"
)

func (s *Server) handleFileMaker(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only POST is allowed")
		return
	}
	if len(s.config.Tokens) == 0 || !s.authorized(r.Header.Get("Authorization"), s.config.Tokens) {
		w.Header().Set("WWW-Authenticate", "Bearer")
		writeError(w, http.StatusUnauthorized, "unauthorized", "a valid application bearer token is required")
		return
	}
	if !strings.Contains(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
		writeError(w, http.StatusUnsupportedMediaType, "content_type_required", "Content-Type must be application/json")
		return
	}
	var request filemaker.Request
	decoder := json.NewDecoder(io.LimitReader(r.Body, s.config.Security.MaxRequestBodyBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_filemaker_request", "request JSON is invalid")
		return
	}
	s.mu.RLock()
	gateway := s.fm
	s.mu.RUnlock()
	response, err := gateway.Execute(r.Context(), request)
	if err != nil {
		var apiError *filemaker.APIError
		if errors.As(err, &apiError) {
			writeError(w, apiError.Status, apiError.Code, apiError.Message)
			return
		}
		writeError(w, http.StatusBadRequest, "filemaker_request_rejected", err.Error())
		return
	}
	if response.Meta == nil {
		response.Meta = make(map[string]any)
	}
	response.Meta["request_id"] = w.Header().Get("X-Request-ID")
	w.Header().Set("X-GoFM-Auth-Verified", "true")
	writeJSON(w, http.StatusOK, response)
}
