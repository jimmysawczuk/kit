package web

import (
	"context"
	"net/http"

	"github.com/sirupsen/logrus"
)

// Handler is a function that takes an HTTP request and responds appropriately.
type Handler func(context.Context, logrus.FieldLogger, http.ResponseWriter, *http.Request)

func Shim(h http.Handler) Handler {
	initial := func(_ context.Context, _ logrus.FieldLogger, w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}

	return initial
}
