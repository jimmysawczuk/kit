package middleware

import (
	"context"
	"net/http"

	"github.com/jimmysawczuk/kit/web"
	"golang.org/x/exp/slog"
)

// WithField attaches a log field with the provided name and value to the logger that's passed through the
// request.
func WithField(name string, value interface{}) func(web.Handler) web.Handler {
	return func(h web.Handler) web.Handler {
		return func(ctx context.Context, log *slog.Logger, w http.ResponseWriter, r *http.Request) {
			log = log.With(name, value)
			h(ctx, log, w, r)
		}
	}
}

// DefaultLogFields attaches a set of default log fields to the logger that's passed through the request.
func DefaultLogFields(h web.Handler) web.Handler {
	return func(ctx context.Context, log *slog.Logger, w http.ResponseWriter, r *http.Request) {
		// rctx := chi.RouteContext(ctx)

		// log = log.With(logrus.Fields{
		// 	"@requestID": ctx.Value(RequestIDKey).(string),
		// 	"method":     r.Method,
		// 	"url":        r.URL.String(),
		// 	"route": map[string]interface{}{
		// 		"method": rctx.RouteMethod,
		// 		"path":   rctx.RoutePattern(),
		// 	},
		// })

		h(ctx, log, w, r)
	}
}
