package handler

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse is the standard JSON envelope for error replies.
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// RespondError writes a JSON error response with the given HTTP status,
// machine-readable code, and human-readable message.
func RespondError(w http.ResponseWriter, status int, code, message string) {
	RespondJSON(w, status, ErrorResponse{
		Error: message,
		Code:  code,
	})
}

// RespondJSON marshals data as JSON and writes it with the given HTTP status.
// On marshal failure it falls back to a 500 plain-text response.
func RespondJSON(w http.ResponseWriter, status int, data any) {
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
