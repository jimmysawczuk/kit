package middleware

import (
	"context"
	"net/http"

	"github.com/jimmysawczuk/kit/web"
	"go.uber.org/zap"
)

func VersionHeader(version string) func(web.Handler) web.Handler {
	return func(h web.Handler) web.Handler {
		return func(ctx context.Context, log *zap.Logger, w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-API-Version", version)

			h(ctx, log, w, r)
		}
	}
}
