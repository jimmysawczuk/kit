package web

import (
	"context"
	"net/http"

	"github.com/jimmysawczuk/kit/web/respond"
	"github.com/rs/zerolog"
)

// Handler is a function that takes an HTTP request and responds appropriately.
type Handler func(context.Context, *zerolog.Logger, http.ResponseWriter, *http.Request)

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h(r.Context(), zerolog.Ctx(r.Context()), w, r)
}

// SyncHandler is a function that takes an HTTP request and responds exactly once.
type SyncHandler func(context.Context, *zerolog.Logger, *http.Request) respond.Response

func (sh SyncHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := sh(r.Context(), zerolog.Ctx(r.Context()), r)
	resp.Write(w)
}

// Middleware is a function that wraps an http.Handler with another.
type Middleware = func(http.Handler) http.Handler
