package middleware

import (
	"context"
	"net/http"

	"github.com/jimmysawczuk/kit/web"
	"go.uber.org/zap"
)

// Bootstrap creates an initial context and log entry, and is therefore the first middleware that should be applied.
// You may choose to use your own app-specific Bootstrap implementation to attach a custom logger or context.
func Bootstrap(h web.Handler) web.Handler {
	return func(_ context.Context, _ *zap.Logger, w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		entry := zap.L()

		h(ctx, entry, w, r)
	}
}
