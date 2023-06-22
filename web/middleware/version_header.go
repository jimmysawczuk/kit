package middleware

import (
	"context"
	"net/http"

	"github.com/jimmysawczuk/kit/web"
	"golang.org/x/exp/slog"
)

func VersionHeader(version string) func(web.Handler) web.Handler {
	return func(h web.Handler) web.Handler {
		return func(ctx context.Context, log *slog.Logger, w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-API-Version", version)

			h(ctx, log, w, r)
		}
	}
}
