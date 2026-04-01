package server

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/example/myservice/internal/config"
	"github.com/example/myservice/internal/handler"
	applog "github.com/example/myservice/internal/log"
	"github.com/example/myservice/internal/middleware"
	"github.com/example/myservice/internal/repository"
)

// New creates a fully-wired *http.Server ready to ListenAndServe.
func New(cfg config.HTTPConfig, logger *slog.Logger, pool *pgxpool.Pool, otelEnabled bool) *http.Server {
	r := chi.NewRouter()

	// --- Middleware chain ---
	// Request ID first so downstream middleware and handlers can reference it.
	r.Use(middleware.RequestID)
	// Inject the logger (with request ID) into every request context.
	r.Use(loggerInjector(logger))
	// Structured request logging.
	r.Use(middleware.Logging)
	// Panic recovery — returns 500 instead of killing the process.
	r.Use(chimw.Recoverer)
	// OpenTelemetry HTTP instrumentation (only when enabled).
	if otelEnabled {
		r.Use(func(next http.Handler) http.Handler {
			return otelhttp.NewHandler(next, "http.request")
		})
	}

	// --- Routes ---
	healthHandler := handler.NewHealthHandler(pool)
	healthHandler.Routes(r)

	resourceRepo := repository.NewPgxResourceRepository(pool)
	resourceHandler := handler.NewResourceHandler(resourceRepo)
	resourceHandler.Routes(r)

	return &http.Server{
		Addr:         cfg.Addr(),
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}

// loggerInjector returns middleware that stores the application logger (with
// the current request ID) in the request context.
func loggerInjector(base *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqLogger := base.With("request_id", middleware.GetRequestID(r.Context()))
			ctx := applog.WithLogger(r.Context(), reqLogger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
