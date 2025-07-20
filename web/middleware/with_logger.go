package middleware

import (
	"net/http"

	"github.com/rs/zerolog"
)

func WithLogger(log *zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := log.With().Logger().WithContext(r.Context())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
