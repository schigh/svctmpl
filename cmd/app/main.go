package main

import (
	"context"
	"go.uber.org/zap"
	"os/signal"
	"slices"
	"syscall"
)

func main() {
	// register for syscalls
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// bootstrap dependencies
	closers := []func(ctx context.Context){
		bootstrapLogging(ctx),  // logging
		bootstrapDatabase(ctx), // postgres
		bootstrapService(ctx),
		bootstrapGRPC(ctx), // gRPC server
		bootstrapHTTP(ctx), // HTTP server
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

func bootstrapService(ctx context.Context) func(ctx context.Context) {
	return func(ctx context.Context) {}
}

// Lightweight bootstrapping functions.
// These might be a little too lightweight for some service implementations, but the gist of these functions is the same no matter where they live. Each function returns
func bootstrapLogging(_ context.Context) func(context.Context) {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
	return func(_ context.Context) {
		_ = logger.Sync()
	}
}

func bootstrapDatabase(_ context.Context) func(ctx context.Context) {
	return func(ctx context.Context) {}
}

func bootstrapGRPC(_ context.Context) func(ctx context.Context) {
	return func(ctx context.Context) {}
}

func bootstrapHTTP(_ context.Context) func(ctx context.Context) {
	return func(ctx context.Context) {}
}
