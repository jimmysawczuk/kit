package middleware

import (
	"context"
	"net/http"

	"github.com/jimmysawczuk/kit/web"
	"github.com/sirupsen/logrus"
)

func VersionHeader(version string) func(web.Handler) web.Handler {
	return func(h web.Handler) web.Handler {
		return func(ctx context.Context, log logrus.FieldLogger, w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-API-Version", version)

			h(ctx, log, w, r)
		}
	}
}
