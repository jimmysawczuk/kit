package middleware

import (
	"context"
	"net/http"
	rdebug "runtime/debug"

	"github.com/jimmysawczuk/kit/web"
	"github.com/jimmysawczuk/kit/web/respond"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Recoverer catches any panics that the wrapped Handler might cause.
func Recoverer(h web.Handler) web.Handler {
	return func(ctx context.Context, log *zap.Logger, w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				err := errors.Errorf("panic: %v", p)

				log.With(
					zap.Error(err),
					zap.String("mw", "Recoverer"),
				).Error("recovered from panic")

				rdebug.PrintStack()
				respond.WithError(ctx, log, w, r, http.StatusInternalServerError, err)
			}
		}()

		h(ctx, log, w, r)
	}
}
