package main

import (
	"context"
	"os"
	"os/signal"
	"slices"
	"syscall"

	"go.uber.org/zap"

	"github.com/schigh/svctmpl/internal/config"
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
	closers := []func(ctx context.Context){
		bootstrapLogging(ctx, cfg.Log),       // logging
		bootstrapDatabase(ctx, cfg.Database), // postgres
		bootstrapGRPC(ctx, cfg.GRPC),         // gRPC server
		bootstrapHTTP(ctx, cfg.HTTP),         // HTTP server
		bootstrapService(ctx, cfg.Service),   // service
	}
	// reverse the closers so that things allocated/initialized first are deallocated/terminated last
	slices.Reverse(closers)

	// TODO: context logger
	zap.L().Info("starting server")

	// block until shutdown signal
	<-ctx.Done()
	zap.L().Info("shutting down server")
	for _, closer := range closers {
		closer(ctx)
	}
}

func bootstrapService(ctx context.Context, cfg config.Service) func(ctx context.Context) {
	return func(ctx context.Context) {}
}

// Lightweight bootstrapping functions.
// These might be a little too lightweight for some service implementations, but the gist of these functions is the same no matter where they live. Each function returns
func bootstrapLogging(_ context.Context, cfg config.Log) func(context.Context) {
	zCfg := zap.NewProductionConfig()
	zCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	logger, err := zap.NewProductionConfig().Build()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
	return func(_ context.Context) {
		_ = logger.Sync()
	}
}

func bootstrapDatabase(_ context.Context, cfg config.Database) func(ctx context.Context) {
	return func(ctx context.Context) {}
}

func bootstrapGRPC(_ context.Context, cfg config.GRPC) func(ctx context.Context) {
	return func(ctx context.Context) {}
}

func bootstrapHTTP(_ context.Context, cfg config.HTTP) func(ctx context.Context) {
	return func(ctx context.Context) {}
}
