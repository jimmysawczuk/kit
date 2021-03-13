package web

import (
	"net/http"
)

// Middleware is a function that wraps a Handler with another.
type Middleware func(Handler) Handler

// compose wraps the provided middlewares around each other in reverse order to consolidate them into
// one middleware.
func compose(mws ...Middleware) Middleware {
	return func(h Handler) Handler {
		for i := len(mws) - 1; i >= 0; i-- {
			h = mws[i](h)
		}
		return h
	}
}

// HandlerFunc returns an http.HandlerFunc that wraps the provided middleware around the provided handler, suitable
// for passing into an http.Server.
func HandlerFunc(h Handler, mws ...Middleware) http.HandlerFunc {
	wr := compose(mws...)

	return func(w http.ResponseWriter, r *http.Request) {
		wr(h)(nil, nil, w, r)
	}
}
