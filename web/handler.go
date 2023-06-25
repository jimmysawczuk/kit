package web

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

// Handler is a function that takes an HTTP request and responds appropriately.
type Handler func(context.Context, *zap.Logger, http.ResponseWriter, *http.Request)

func Shim(h http.Handler) Handler {
	initial := func(_ context.Context, _ *zap.Logger, w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}

	return initial
}
