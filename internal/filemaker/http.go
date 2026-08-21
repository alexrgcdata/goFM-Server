package filemaker

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type APIError struct {
	Status  int
	Code    string
	Message string
}

func (e *APIError) Error() string { return e.Message }

func decodeJSON(response *http.Response, destination any) error {
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, 8<<20))
	if err != nil {
		return &APIError{Status: http.StatusBadGateway, Code: "filemaker_response_read_failed", Message: "could not read FileMaker response"}
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return &APIError{Status: http.StatusBadGateway, Code: "filemaker_request_failed", Message: fmt.Sprintf("FileMaker returned HTTP %d", response.StatusCode)}
	}
	if len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, destination); err != nil {
		return &APIError{Status: http.StatusBadGateway, Code: "filemaker_invalid_json", Message: "FileMaker returned invalid JSON"}
	}
	return nil
}
