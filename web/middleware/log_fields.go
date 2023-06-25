package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/jimmysawczuk/kit/web"
	"go.uber.org/zap"
)

// WithField attaches a log field with the provided name and value to the logger that's passed through the
// request.
func WithField(field zap.Field) func(web.Handler) web.Handler {
	return func(h web.Handler) web.Handler {
		return func(ctx context.Context, log *zap.Logger, w http.ResponseWriter, r *http.Request) {
			log = log.With(field)
			h(ctx, log, w, r)
		}
	}
}

// DefaultLogFields attaches a set of default log fields to the logger that's passed through the request.
func DefaultLogFields(h web.Handler) web.Handler {
	return func(ctx context.Context, log *zap.Logger, w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(ctx)

		log = log.With(
			zap.String("@requestID", ctx.Value(RequestIDKey).(string)),
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.String("route.method", rctx.RouteMethod),
			zap.String("route.path", rctx.RoutePattern()),
		)

		h(ctx, log, w, r)
	}
}
