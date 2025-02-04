package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jimmysawczuk/kit/web"
	"github.com/sirupsen/logrus"
)

// WithField attaches a log field with the provided name and value to the logger that's passed through the
// request.
func WithField(name string, value interface{}) func(web.Handler) web.Handler {
	return func(h web.Handler) web.Handler {
		return func(ctx context.Context, log logrus.FieldLogger, w http.ResponseWriter, r *http.Request) {
			log = log.WithField(name, value)
			h(ctx, log, w, r)
		}
	}
}

// LogRequestInfo attaches a set of default log fields to the logger that's passed through the request.
func LogRequestInfo(h web.Handler) web.Handler {
	return func(ctx context.Context, log logrus.FieldLogger, w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(ctx)

		log = log.WithFields(logrus.Fields{
			"method": r.Method,
			"url":    r.URL.String(),
			"route": map[string]interface{}{
				"method": rctx.RouteMethod,
				"path":   rctx.RoutePattern(),
			},
		})

		h(ctx, log, w, r)
	}
}

// LogRequestIP logs the remote IP address attached to the request. Requires RealIP to be present before this middleware.
func LogRequestIP(h web.Handler) web.Handler {
	return func(ctx context.Context, log logrus.FieldLogger, w http.ResponseWriter, r *http.Request) {
		if ip, ok := ctx.Value(ipAddressKey).(string); ok && ip != "" {
			log = log.WithField("@ip", ip)
		}

		h(ctx, log, w, r)
	}
}
