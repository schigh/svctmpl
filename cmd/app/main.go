package main

import (
	"context"
	"os"
	"os/signal"
	"slices"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/schigh/svctmpl/internal/config"
	"github.com/schigh/svctmpl/internal/log"
)

func main() {
	// register for syscalls
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// load env
	// If environment variables for this service have a common prefix, set it here.
	// Prefixes are a common practice in K8s environments
	if prefix := os.Getenv("ENV_PREFIX"); prefix != "" {
		config.SetPrefix(prefix)
	}
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// bootstrap dependencies
	// Each of these bootstrapping functions returns a closure that will be invoked when the service is shutting down.
	// Any unrecoverable errors encountered during bootstrapping will panic
	closers := []func(ctx context.Context){
		bootstrapLogging(ctx, cfg.Service),   // logging
		bootstrapDatabase(ctx, cfg.Database), // postgres
		bootstrapGRPC(ctx, cfg.GRPC),         // gRPC server
		bootstrapHTTP(ctx, cfg.HTTP),         // HTTP server
		bootstrapService(ctx, cfg.Service),   // service
	}
	// reverse the closers so that things allocated/initialized first are deallocated/terminated last
	slices.Reverse(closers)

	// TODO: context logger
	log.Ctx(ctx).Info("starting service")

	// block until shutdown signal
	<-ctx.Done()
	log.Ctx(ctx).Info("shutting down")
	for _, closer := range closers {
		closer(ctx)
	}
}

func shutdown(ctx context.Context) {
	log.Ctx(ctx).Info("shutting down")
}

// Lightweight bootstrapping functions.
// These might be a little too lightweight for some service implementations, but the gist of these functions is the same
// no matter where they live. Each function returns a closer function that can be invoked when the application is
// shutting down.
// Any long-lived dependencies must be declared outside of these functions to maintain proper ownership and lifecycle.

// main service
func bootstrapService(ctx context.Context, cfg config.Service) func(ctx context.Context) {
	return func(ctx context.Context) {}
}

// logging
func bootstrapLogging(_ context.Context, cfg config.Service) func(context.Context) {
	zCfg := zap.NewProductionConfig()
	ll, err := zapcore.ParseLevel(cfg.LogLevel)
	if err != nil {
		panic(err)
	}
	zCfg.Level = zap.NewAtomicLevelAt(ll)
	logger, err := zap.NewProductionConfig().Build()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
	return func(_ context.Context) {
		_ = logger.Sync()
	}
}

// database
func bootstrapDatabase(_ context.Context, cfg config.Database) func(ctx context.Context) {
	return func(ctx context.Context) {}
}

// gRPC server
func bootstrapGRPC(_ context.Context, cfg config.GRPC) func(ctx context.Context) {
	return func(ctx context.Context) {}
}

// HTTP server
func bootstrapHTTP(_ context.Context, cfg config.HTTP) func(ctx context.Context) {
	return func(ctx context.Context) {}
}
