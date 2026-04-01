package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const headerXRequestID = "X-Request-ID"

type requestIDKey struct{}

// RequestID is HTTP middleware that ensures every request has a unique
// identifier. If the incoming request carries an X-Request-ID header its value
// is reused; otherwise a new UUID is generated. The ID is stored in the
// request context and echoed back in the response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(headerXRequestID)
		if id == "" {
			id = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), requestIDKey{}, id)
		w.Header().Set(headerXRequestID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID extracts the request ID from the context. Returns an empty
// string when no ID is present.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}
