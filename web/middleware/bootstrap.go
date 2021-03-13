package middleware

import (
	"context"
	"net/http"

	"github.com/jimmysawczuk/kit/web"
	"github.com/sirupsen/logrus"
)

// Bootstrap creates an initial context and log entry, and is therefore the first middleware that should be applied.
// You may choose to use your own app-specific Bootstrap implementation to attach a custom logger or context.
func Bootstrap(h web.Handler) web.Handler {
	return func(_ context.Context, _ logrus.FieldLogger, w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		entry := logrus.NewEntry(logrus.StandardLogger())

		h(ctx, entry, w, r)
	}
}
