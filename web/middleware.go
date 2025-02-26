package web

import (
	"net/http"
)

// Middleware is a function that wraps an http.Handler with another.
type Middleware = func(http.Handler) http.Handler

// compose wraps the provided middlewares around each other in reverse order to consolidate them into
// one middleware.
func compose(mws ...Middleware) Middleware {
	return func(h http.Handler) http.Handler {
		for i := len(mws) - 1; i >= 0; i-- {
			h = mws[i](h)
		}
		return h
	}
}

// collapse converts a slice of slices to a single slice of Middleware.
func collapse(mws ...[]Middleware) []Middleware {
	tbr := []Middleware{}
	for _, m := range mws {
		tbr = append(tbr, m...)
	}
	return tbr
}

// // HandlerFunc returns an http.HandlerFunc that wraps the provided middleware around the provided handler, suitable
// // for passing into an http.Server.
// func HandlerFunc(h Handler, mws ...Middleware) http.HandlerFunc {
// 	wr := compose(mws...)

// 	return func(w http.ResponseWriter, r *http.Request) {
// 		wr(h)(nil, nil, w, r)
// 	}
// }
