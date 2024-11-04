package loginject

import (
	"context"
	"net/http"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/schigh/svctmpl/internal/log"
	"github.com/schigh/svctmpl/internal/middleware/requestid"
)

// HTTP is a middleware that injects the request ID into the log context if available in the request context.
func HTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if requestID := requestid.FromContext(ctx); requestID != "" {
			ctx = log.Stash(ctx, zap.String("request_id", requestID))
		}
		next.ServeHTTP(rw, r.WithContext(ctx))
	})
}

// GRPC is a middleware function for gRPC that adds a request ID to the context if it exists, then calls the next handler.
func GRPC(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if requestID := requestid.FromContext(ctx); requestID != "" {
		ctx = log.Stash(ctx, zap.String("request_id", requestID))
	}
	return handler(ctx, req)
}
