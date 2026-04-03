package http

import "net/http"

// Healthz is a liveness probe that always returns 200.
func (h *Handler) Healthz(w http.ResponseWriter, _ *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Readyz is a readiness probe. It pings the database and returns
// 200 if reachable, 503 otherwise.
func (h *Handler) Readyz(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		respondError(w, http.StatusServiceUnavailable, "UNAVAILABLE", "database not configured")
		return
	}
	if err := h.db.Ping(r.Context()); err != nil {
		respondError(w, http.StatusServiceUnavailable, "UNAVAILABLE", "database ping failed")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
