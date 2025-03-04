package middleware

import (
	"context"
	"net/http"
)

// WithContext attaches a log field with the provided name and value to the logger that's passed through the
// request.
func WithContext(mod func(context.Context) context.Context) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := mod(r.Context())

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
