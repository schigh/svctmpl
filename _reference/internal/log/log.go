package log

import (
	"context"
	"log/slog"
	"os"
)

type fieldsKey struct{}

// New creates a structured logger appropriate for the given environment.
// "prod" uses JSON output; everything else uses human-readable text.
func New(env string) *slog.Logger {
	var handler slog.Handler
	if env == "prod" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}
	return slog.New(handler)
}

// Stash stores structured logging fields in the context for downstream
// propagation. Fields accumulate across multiple Stash calls, each call
// appends to existing fields rather than replacing them.
//
// Typical usage: middleware stashes request_id, trace_id, user_id, etc.
// Downstream code calls Ctx(ctx) to get a logger with all stashed fields.
//
//	// In middleware:
//	ctx = log.Stash(ctx, "request_id", reqID)
//
//	// In another middleware:
//	ctx = log.Stash(ctx, "trace_id", traceID)
//
//	// Anywhere downstream:
//	log.Ctx(ctx).Info("order created", "order_id", id)
//	// Output includes request_id, trace_id, and order_id automatically.
func Stash(ctx context.Context, args ...any) context.Context {
	existing := fromContext(ctx)
	combined := make([]any, 0, len(existing)+len(args))
	combined = append(combined, existing...)
	combined = append(combined, args...)
	return context.WithValue(ctx, fieldsKey{}, combined)
}

// Ctx returns a logger enriched with all fields previously stashed in the
// context. Uses the slog default logger as the base.
func Ctx(ctx context.Context) *slog.Logger {
	logger := slog.Default()
	if args := fromContext(ctx); len(args) > 0 {
		logger = logger.With(args...)
	}
	return logger
}

// SetDefault sets the default slog logger. Call this once during app
// initialization. All subsequent Ctx calls use this logger as the base.
func SetDefault(l *slog.Logger) {
	slog.SetDefault(l)
}

func fromContext(ctx context.Context) []any {
	if args, ok := ctx.Value(fieldsKey{}).([]any); ok {
		return args
	}
	return nil
}
