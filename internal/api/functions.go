package api

import (
	"context"
	"encoding/json"
)

// ParamsError is safe to show to a caller. Other errors stay private.
type ParamsError struct{ Message string }

func (e *ParamsError) Error() string { return e.Message }

func noParams(params []json.RawMessage) error {
	if len(params) != 0 {
		return &ParamsError{Message: "this action expects an empty params array"}
	}
	return nil
}

func hello(_ context.Context, params []json.RawMessage) (any, error) {
	return "hello world", noParams(params)
}

func success(_ context.Context, params []json.RawMessage) (any, error) {
	return "successful", noParams(params)
}

func boolean(_ context.Context, params []json.RawMessage) (any, error) {
	return true, noParams(params)
}

// Echo demonstrates dynamic parameters without converting numbers to floats.
func echo(_ context.Context, params []json.RawMessage) (any, error) {
	return params, nil
}

// Sample returns a fixed fictional job without reading a database.
func sample(_ context.Context, params []json.RawMessage) (any, error) {
	if err := noParams(params); err != nil {
		return nil, err
	}

	return map[string]string{
		"id":     "WORKSHOP-001",
		"title":  "Inspect demo equipment",
		"status": "scheduled",
	}, nil
}
