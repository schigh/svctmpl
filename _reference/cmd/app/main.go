package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/example/myservice/internal/config"
	applog "github.com/example/myservice/internal/log"
	"github.com/example/myservice/internal/server"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// ---- Configuration ----
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// ---- Logger ----
	logger := applog.New(cfg.Service.Environment)
	slog.SetDefault(logger)
	logger.Info("starting service",
		"name", cfg.Service.Name,
		"env", cfg.Service.Environment,
	)

	// ---- Root context: cancelled on SIGINT / SIGTERM ----
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ---- Database ----
	poolCfg, err := pgxpool.ParseConfig(cfg.Database.DSN())
	if err != nil {
		return fmt.Errorf("parsing database DSN: %w", err)
	}
	poolCfg.MaxConns = cfg.Database.MaxConns

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return fmt.Errorf("creating connection pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("pinging database: %w", err)
	}
	logger.Info("database connected", "host", cfg.Database.Host, "db", cfg.Database.DBName)

	// ---- OpenTelemetry ----
	var metricsServer *http.Server
	if cfg.OTel.Enabled {
		shutdown, metricsSrv, otelErr := setupOTel(cfg.OTel)
		if otelErr != nil {
			return fmt.Errorf("setting up otel: %w", otelErr)
		}
		metricsServer = metricsSrv
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if sErr := shutdown(shutdownCtx); sErr != nil {
				logger.Error("otel shutdown error", "error", sErr)
			}
		}()
		logger.Info("otel enabled", "metrics_addr", cfg.OTel.MetricsAddr())
	}

	// ---- HTTP Server ----
	srv := server.New(cfg.HTTP, logger, pool, cfg.OTel.Enabled)
	logger.Info("http server starting", "addr", cfg.HTTP.Addr())

	// Start HTTP server in a goroutine.
	srvErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			srvErr <- err
		}
		close(srvErr)
	}()

	// Start metrics server if OTel is enabled.
	metricsErr := make(chan error, 1)
	if metricsServer != nil {
		go func() {
			if err := metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				metricsErr <- err
			}
			close(metricsErr)
		}()
	}

	// ---- Wait for shutdown signal or fatal error ----
	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-srvErr:
		if err != nil {
			return fmt.Errorf("http server error: %w", err)
		}
	case err := <-metricsErr:
		if err != nil {
			return fmt.Errorf("metrics server error: %w", err)
		}
	}

	// ---- Graceful shutdown: server → database → otel (via defers) ----
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("http server shutdown: %w", err)
	}
	logger.Info("http server stopped")

	if metricsServer != nil {
		if err := metricsServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("metrics server shutdown error", "error", err)
		}
	}

	// pool.Close() and otel shutdown run via defers.
	logger.Info("shutdown complete")
	return nil
}

// setupOTel initialises the OpenTelemetry metrics pipeline with a Prometheus
// exporter and returns a shutdown function plus an HTTP server serving the
// /metrics endpoint.
func setupOTel(cfg config.OTelConfig) (func(context.Context) error, *http.Server, error) {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("creating otel resource: %w", err)
	}

	exporter, err := prometheus.New()
	if err != nil {
		return nil, nil, fmt.Errorf("creating prometheus exporter: %w", err)
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(exporter),
	)
	otel.SetMeterProvider(provider)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	metricsSrv := &http.Server{
		Addr:         cfg.MetricsAddr(),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	shutdown := func(ctx context.Context) error {
		return provider.Shutdown(ctx)
	}

	return shutdown, metricsSrv, nil
}
