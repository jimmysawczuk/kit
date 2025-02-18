package middleware

import (
	"net/http"

	"github.com/jimmysawczuk/kit/web/requestid"
	"github.com/rs/zerolog"
)

// RequestID determines whether a request ID should be created or gleaned from the request, then
// sets it on the context.
func RequestID(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := requestid.Next(r)

		ctx := r.Context()
		ctx = requestid.Set(ctx, id)

		zerolog.Ctx(ctx).UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Str("@id", id)
		})

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
