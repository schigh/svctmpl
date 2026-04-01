package http

import "net/http"

// healthz is a liveness probe that always returns 200.
func healthz(w http.ResponseWriter, _ *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// readyz returns a readiness probe handler. It pings the database and returns
// 200 if reachable, 503 otherwise.
func readyz(db DBPinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			respondError(w, http.StatusServiceUnavailable, "UNAVAILABLE", "database not configured")
			return
		}
		if err := db.Ping(r.Context()); err != nil {
			respondError(w, http.StatusServiceUnavailable, "UNAVAILABLE", "database ping failed")
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
