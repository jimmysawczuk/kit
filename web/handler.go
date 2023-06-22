package web

import (
	"context"
	"net/http"

	"golang.org/x/exp/slog"
)

// Handler is a function that takes an HTTP request and responds appropriately.
type Handler func(context.Context, *slog.Logger, http.ResponseWriter, *http.Request)

func Shim(h http.Handler) Handler {
	initial := func(_ context.Context, _ *slog.Logger, w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}

	return initial
}
