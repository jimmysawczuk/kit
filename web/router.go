package web

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Router wraps a chi.Router.
type Router struct {
	r   chi.Router
	mws []Middleware
}

// Route allows the caller to take over a subset of routes in the global router and attach
// routes (using the provided function) and middleware.
func (r Router) Route(pattern string, f func(Router), rmws ...Middleware) {
	mws := stackMiddleware(r.mws, rmws)

	r.r.Route(pattern, func(cr chi.Router) {
		rr := Router{
			r:   cr,
			mws: mws,
		}

		f(rr)
	})
}

// Get attaches a GET Handler to the provided pattern with the provided middlewares.
func (r Router) Get(pattern string, handler Handler, rmws ...Middleware) {
	mws := append(r.mws, rmws...)
	r.r.MethodFunc(http.MethodGet, pattern, HandlerFunc(handler, mws...))
	r.r.MethodFunc(http.MethodOptions, pattern, HandlerFunc(handler, mws...))
}

// Post attaches a POST Handler to the provided pattern with the provided middlewares.
func (r Router) Post(pattern string, handler Handler, rmws ...Middleware) {
	mws := stackMiddleware(r.mws, rmws)

	r.r.MethodFunc(http.MethodPost, pattern, HandlerFunc(handler, mws...))
	r.r.MethodFunc(http.MethodOptions, pattern, HandlerFunc(handler, mws...))
}

// Put attaches a PUT Handler to the provided pattern with the provided middlewares.
func (r Router) Put(pattern string, handler Handler, rmws ...Middleware) {
	mws := stackMiddleware(r.mws, rmws)

	r.r.MethodFunc(http.MethodPut, pattern, HandlerFunc(handler, mws...))
	r.r.MethodFunc(http.MethodOptions, pattern, HandlerFunc(handler, mws...))
}

// Delete attaches a DELETE handler to the provided pattern with the provided middlewares.
func (r Router) Delete(pattern string, handler Handler, rmws ...Middleware) {
	mws := stackMiddleware(r.mws, rmws)

	r.r.MethodFunc(http.MethodDelete, pattern, HandlerFunc(handler, mws...))
}

// Mount attaches the provided http.Handler directly to the given pattern with the provided middlewares.
// This is useful for serving static files or embedding other APIs directly.
func (r Router) Mount(pattern string, handler Handler, rmws ...Middleware) {
	mws := stackMiddleware(r.mws, rmws)
	r.r.Mount(pattern, HandlerFunc(handler, mws...))
}

// URLParam wraps chi.URLParam.
func URLParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

func stackMiddleware(mws ...[]Middleware) []Middleware {
	tbr := []Middleware{}
	for _, m := range mws {
		tbr = append(tbr, m...)
	}
	return tbr
}
