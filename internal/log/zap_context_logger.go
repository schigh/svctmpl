package log

import (
	"context"
	"slices"

	"go.uber.org/zap"
)

type ctxKey struct{}

var key = ctxKey{}

// Logger decorates zap.Logger with context-aware functionality
type Logger struct {
	z    *zap.Logger
	meta []zap.Field
}

// Stash updates a context with variables that will be logged in all calls downstream from this context
func Stash(parent context.Context, fields ...zap.Field) context.Context {
	if v := parent.Value(key); v != nil {
		oldMeta, _ := v.([]zap.Field)
		oldMeta = slices.Grow(oldMeta, len(fields))
		return context.WithValue(parent, key, append(oldMeta, fields...))
	}

	return context.WithValue(parent, key, fields)
}

// Ctx returns a logger with any context-scoped metadata attached
func Ctx(parent context.Context) Logger {
	out := Logger{z: zap.L()}
	if v := parent.Value(key); v != nil {
		meta, ok := v.([]zap.Field)
		if ok {
			out.meta = meta

			return out
		}
	}
	out.meta = make([]zap.Field, 0)

	return out
}

// Debug logs debug-level messages
func (logger Logger) Debug(msg string, fields ...zap.Field) {
	logger.z.Debug(msg, append(logger.meta, fields...)...)
}

// Info logs info-level messages
func (logger Logger) Info(msg string, fields ...zap.Field) {
	logger.z.Info(msg, append(logger.meta, fields...)...)
}

// Warn logs warning-level messages
func (logger Logger) Warn(msg string, fields ...zap.Field) {
	logger.z.Warn(msg, append(logger.meta, fields...)...)
}

// Error logs error-level messages
func (logger Logger) Error(msg string, fields ...zap.Field) {
	logger.z.Error(msg, append(logger.meta, fields...)...)
}

// DPanic logs development-panic-level messages
// If the application logger is in development mode, this will propagate the panic
func (logger Logger) DPanic(msg string, fields ...zap.Field) {
	logger.z.DPanic(msg, append(logger.meta, fields...)...)
}

// Panic logs panic-level messages
// This will cause the application to halt
func (logger Logger) Panic(msg string, fields ...zap.Field) {
	logger.z.Panic(msg, append(logger.meta, fields...)...)
}

// Fatal logs fatal-level messages
// This will cause the application to halt and ignore any stacked defers
func (logger Logger) Fatal(msg string, fields ...zap.Field) {
	logger.z.Fatal(msg, append(logger.meta, fields...)...)
}

// Sync causes the logger to sync with the output buffer
// This only needs to be called upon application shutdown to dump any queued logs
func (logger Logger) Sync() error {
	return logger.z.Sync()
}

// With will decorate the underlying zap logger with the provided fields
func (logger Logger) With(fields ...zap.Field) Logger {
	z := logger.z.With(fields...)
	return Logger{z: z}
}
