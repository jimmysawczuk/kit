package web

import (
	"context"
	"net/http"

	"github.com/rs/zerolog"
)

// Handler is a function that takes an HTTP request and responds appropriately.
type Handler func(context.Context, *zerolog.Logger, http.ResponseWriter, *http.Request)

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h(r.Context(), zerolog.Ctx(r.Context()), w, r)
}
