package requestid

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/schigh/svctmpl/internal/log"
)

// HeaderName is the HTTP/gRPC header key used to identify the unique request ID.
const HeaderName = "X-Request-ID"

type key struct{}

var ctxKey = key{}

func genID(ctx context.Context) string {
	uid, err := uuid.NewRandom()
	if err != nil {
		log.Ctx(ctx).Error("generate request id failed", zap.Error(err))
		return strconv.Itoa(time.Now().Nanosecond()) + strconv.Itoa(time.Now().Nanosecond())
	}
	return uid.String()
}

// NewContext returns a new context based on a parent context with an added request ID value.
func NewContext(parent context.Context, id string) context.Context {
	return context.WithValue(parent, ctxKey, id)
}

// FromContext retrieves a string ID from the given context using a predefined key.
func FromContext(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey).(string)
	return id
}

// HTTP is a middleware that injects a request ID into the context of each request.
// If the request does not already contain a request ID, one will be generated.
func HTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		var id string
		if id = r.Header.Get(HeaderName); id == "" {
			id = genID(r.Context())
		}
		r = r.WithContext(NewContext(r.Context(), id))
		next.ServeHTTP(rw, r)
	})
}

// GRPC is a gRPC middleware that ensures every request is assigned a unique request ID.
// It retrieves the request ID from incoming metadata or generates a new one if not found.
// The request ID is then added to the context, ensuring traceability across the system.
func GRPC(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	var id string
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		id = genID(ctx)
		ctx = NewContext(ctx, id)
		return handler(ctx, req)
	}
	hd := md.Get(HeaderName)
	if len(hd) == 0 {
		id = genID(ctx)
		ctx = NewContext(ctx, id)
		return handler(ctx, req)
	}
	id = hd[0]
	ctx = NewContext(ctx, id)
	return handler(ctx, req)
}
