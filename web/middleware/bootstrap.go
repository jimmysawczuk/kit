package middleware

import (
	"context"
	"net/http"

	"github.com/jimmysawczuk/kit/web"
	"golang.org/x/exp/slog"
)

// Bootstrap creates an initial context and log entry, and is therefore the first middleware that should be applied.
// You may choose to use your own app-specific Bootstrap implementation to attach a custom logger or context.
func Bootstrap(h web.Handler) web.Handler {
	return func(_ context.Context, _ *slog.Logger, w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		entry := slog.Default()

		h(ctx, entry, w, r)
	}
}
