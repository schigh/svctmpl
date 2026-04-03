package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/example/myservice/internal/errs"
)

// errorResponse is the standard JSON envelope for error replies.
type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// respondError writes a JSON error response with the given HTTP status,
// machine-readable code, and human-readable message.
func respondError(w http.ResponseWriter, status int, code, message string) {
	respondJSON(w, status, errorResponse{
		Error: message,
		Code:  code,
	})
}

// respondJSON marshals data as JSON and writes it with the given HTTP status.
// On marshal failure it falls back to a 500 plain-text response.
func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	body, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal server error","code":"INTERNAL"}`))
		return
	}

	w.WriteHeader(status)
	_, _ = w.Write(body)
}

// mapServiceError translates a service-layer error into an HTTP status code
// and machine-readable error code.
func mapServiceError(err error) (int, string) {
	switch {
	case errors.Is(err, errs.ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND"
	case errors.Is(err, errs.ErrInvalid):
		return http.StatusBadRequest, "INVALID"
	case errors.Is(err, errs.ErrConflict):
		return http.StatusConflict, "CONFLICT"
	default:
		return http.StatusInternalServerError, "INTERNAL"
	}
}

// respondServiceError writes a JSON error response for the given service-layer
// error, mapping it to the appropriate HTTP status code.
func respondServiceError(w http.ResponseWriter, err error) {
	status, code := mapServiceError(err)
	respondError(w, status, code, err.Error())
}
