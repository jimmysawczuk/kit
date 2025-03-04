package middleware

import (
	"net/http"

	"github.com/rs/zerolog"
)

// RequestInfo attaches the http method, path, and URL to the logger in the context.
func RequestInfo(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		zerolog.Ctx(ctx).UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Str("@req.method", r.Method).
				Str("@req.url", r.URL.String()).
				Str("@req.path", r.URL.Path)
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
