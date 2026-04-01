package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// DBPinger is satisfied by any type that can verify database connectivity.
type DBPinger interface {
	Ping(ctx context.Context) error
}

// HealthHandler provides liveness and readiness probes.
type HealthHandler struct {
	db DBPinger
}

// NewHealthHandler creates a HealthHandler. db may be nil, in which case
// the readiness probe always reports unhealthy.
func NewHealthHandler(db DBPinger) *HealthHandler {
	return &HealthHandler{db: db}
}

// Routes registers the health endpoints on the given router.
func (h *HealthHandler) Routes(r chi.Router) {
	r.Get("/healthz", h.Healthz)
	r.Get("/readyz", h.Readyz)
}

// Healthz is a liveness probe that always returns 200.
func (h *HealthHandler) Healthz(w http.ResponseWriter, _ *http.Request) {
	RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Readyz is a readiness probe. It pings the database and returns 200 if
// reachable, 503 otherwise.
func (h *HealthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		RespondError(w, http.StatusServiceUnavailable, "UNAVAILABLE", "database not configured")
		return
	}
	if err := h.db.Ping(r.Context()); err != nil {
		RespondError(w, http.StatusServiceUnavailable, "UNAVAILABLE", "database ping failed")
		return
	}
	RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
