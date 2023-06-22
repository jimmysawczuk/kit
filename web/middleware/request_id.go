package middleware

import (
	"context"
	"net/http"

	"github.com/jimmysawczuk/kit/web"
	"github.com/jimmysawczuk/kit/web/requestid"
	"golang.org/x/exp/slog"
)

// RequestID determines whether a request ID should be created or gleaned from the request, then
// sets it on the context.
func RequestID(h web.Handler) web.Handler {
	return func(ctx context.Context, log *slog.Logger, w http.ResponseWriter, r *http.Request) {
		ctx = requestid.Set(ctx, requestid.Next(r))
		h(ctx, log, w, r.WithContext(ctx))
	}
}
