package middleware

import (
	"context"
	"fmt"
	"net/http"
	rdebug "runtime/debug"

	"github.com/jimmysawczuk/kit/web"
	"github.com/jimmysawczuk/kit/web/respond"
	"github.com/sirupsen/logrus"
)

// Recoverer catches any panics that the wrapped Handler might cause.
func Recoverer(h web.Handler) web.Handler {
	return func(ctx context.Context, log logrus.FieldLogger, w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				err := fmt.Errorf("panic: %v", p)
				log.WithError(err).WithField("mw", "Recoverer").Error("recovered from panic")
				rdebug.PrintStack()
				respond.WithError(ctx, log, w, r, http.StatusInternalServerError, err)
			}
		}()

		h(ctx, log, w, r)
	}
}
