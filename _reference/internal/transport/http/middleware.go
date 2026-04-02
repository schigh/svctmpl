package http

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	applog "github.com/example/myservice/internal/log"
)

const headerXRequestID = "X-Request-ID"

// requestID is HTTP middleware that ensures every request has a unique
// identifier. If the incoming request carries an X-Request-ID header its value
// is reused; otherwise a new UUID is generated. The ID is stashed into the
// logging context and echoed back in the response header.
func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(headerXRequestID)
		if id == "" {
			id = uuid.New().String()
		}

		// Stash into log context so every downstream log.Ctx(ctx) call
		// automatically includes request_id.
		ctx := applog.Stash(r.Context(), "request_id", id)
		w.Header().Set(headerXRequestID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

// logging is HTTP middleware that records each request's method, path, status
// code, and duration. Fields stashed by upstream middleware (request_id,
// trace_id, etc.) are included automatically via log.Ctx.
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		applog.Ctx(r.Context()).Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration", time.Since(start),
		)
	})
}
