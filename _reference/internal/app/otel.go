package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// initOTel sets up OpenTelemetry metrics with a Prometheus exporter when
// enabled in configuration. It starts a dedicated HTTP server for /metrics
// and registers both the provider and server for shutdown.
func (a *App) initOTel() error {
	if !a.Config.OTel.Enabled {
		return nil
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(a.Config.OTel.ServiceName),
		),
	)
	if err != nil {
		return fmt.Errorf("creating otel resource: %w", err)
	}

	exporter, err := prometheus.New()
	if err != nil {
		return fmt.Errorf("creating prometheus exporter: %w", err)
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(exporter),
	)
	otel.SetMeterProvider(provider)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	metricsSrv := &http.Server{
		Addr:         a.Config.OTel.MetricsAddr(),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start metrics server in a goroutine.
	go func() {
		if err := metricsSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.Logger.Error("metrics server error", "error", err)
		}
	}()

	a.Logger.Info("otel enabled", "metrics_addr", a.Config.OTel.MetricsAddr())

	a.onShutdown(func(ctx context.Context) error {
		return errors.Join(
			metricsSrv.Shutdown(ctx),
			provider.Shutdown(ctx),
		)
	})

	return nil
}
