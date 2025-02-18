package middleware

import (
	"fmt"
	"net/http"
	rdebug "runtime/debug"

	"github.com/jimmysawczuk/kit/web/v2/respond"
	"github.com/rs/zerolog"
)

// Recoverer catches any panics that the wrapped Handler might cause.
func Recoverer(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				log := zerolog.Ctx(r.Context())

				err := fmt.Errorf("panic: %v", p)

				log.Error().
					Err(err).
					Msg("recovered from panic")

				rdebug.PrintStack()
				respond.WithError(r.Context(), w, r, http.StatusInternalServerError, err)
			}
		}()

		h.ServeHTTP(w, r)
	})
}
