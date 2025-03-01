package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Middleware = func(http.Handler) http.Handler

type Router struct {
	chi chi.Router
}

func New() Router {
	return Router{
		chi: chi.NewRouter(),
	}
}

func newSubrouter(in chi.Router) Router {
	return Router{
		chi: in,
	}
}

func (ro Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ro.chi.ServeHTTP(w, r)
}

func (ro Router) Get(path string, handler http.Handler, mws ...Middleware) {
	ro.chi.Method(http.MethodGet, path, handler)
}

func (ro Router) Getf(path string, handler http.HandlerFunc, mws ...Middleware) {
	ro.chi.Method(http.MethodGet, path, handler)
}

func (ro Router) Post(path string, handler http.Handler, mws ...Middleware) {
	ro.chi.Method(http.MethodPost, path, handler)
}

func (ro Router) Group(f func(Router), mws ...Middleware) {
	ro.chi.Group(func(inner chi.Router) {
		rr := newSubrouter(inner)
		rr.Use(mws...)
		f(rr)
	})
}

func (ro Router) Route(path string, f func(Router), mws ...Middleware) {
	ro.chi.Route(path, func(inner chi.Router) {
		rr := newSubrouter(inner)
		rr.Use(mws...)
		f(rr)
	})
}

func (r Router) Use(m ...Middleware) {
	r.chi.Use(m...)
}
