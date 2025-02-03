package middleware

import (
	"context"
	"net/http"

	"github.com/jimmysawczuk/kit/web"
	"github.com/jimmysawczuk/kit/web/requestid"
	"github.com/sirupsen/logrus"
)

// RequestID determines whether a request ID should be created or gleaned from the request, then
// sets it on the context.
func RequestID(h web.Handler) web.Handler {
	return func(ctx context.Context, log logrus.FieldLogger, w http.ResponseWriter, r *http.Request) {
		id := requestid.Next(r)
		ctx = requestid.Set(ctx, id)
		log = log.WithField("@id", id)
		h(ctx, log, w, r.WithContext(ctx))
	}
}
