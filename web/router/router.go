package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Middleware = func(http.Handler) http.Handler

type Router interface {
	http.Handler

	Method(string, string, http.Handler, ...Middleware)
	Connect(string, http.Handler, ...Middleware)
	Connectf(string, http.HandlerFunc, ...Middleware)
	Delete(string, http.Handler, ...Middleware)
	Deletef(string, http.HandlerFunc, ...Middleware)
	Get(string, http.Handler, ...Middleware)
	Getf(string, http.HandlerFunc, ...Middleware)
	Head(string, http.Handler, ...Middleware)
	Headf(string, http.HandlerFunc, ...Middleware)
	Options(string, http.Handler, ...Middleware)
	Optionsf(string, http.HandlerFunc, ...Middleware)
	Patch(string, http.Handler, ...Middleware)
	Patchf(string, http.HandlerFunc, ...Middleware)
	Post(string, http.Handler, ...Middleware)
	Postf(string, http.HandlerFunc, ...Middleware)
	Put(string, http.Handler, ...Middleware)
	Putf(string, http.HandlerFunc, ...Middleware)
	Trace(string, http.Handler, ...Middleware)
	Tracef(string, http.HandlerFunc, ...Middleware)

	Use(...Middleware)
	Group(func(Router), ...Middleware)
	Route(string, func(Router), ...Middleware)
}

type chiRouter struct {
	chi chi.Router
}

var _ Router = chiRouter{}

func New() Router {
	return chiRouter{
		chi: chi.NewRouter(),
	}
}

func newSubrouter(in chi.Router) Router {
	return chiRouter{
		chi: in,
	}
}

func (ro chiRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ro.chi.ServeHTTP(w, r)
}

func (ro chiRouter) Method(method string, path string, handler http.Handler, mws ...Middleware) {
	ro.chi.With(mws...).Method(method, path, handler)
}

func (ro chiRouter) Connect(path string, handler http.Handler, mws ...Middleware) {
	ro.chi.With(mws...).Method(http.MethodConnect, path, handler)
}

func (ro chiRouter) Connectf(path string, handler http.HandlerFunc, mws ...Middleware) {
	ro.chi.With(mws...).Method(http.MethodConnect, path, handler)
}

func (ro chiRouter) Delete(path string, handler http.Handler, mws ...Middleware) {
	ro.chi.With(mws...).Method(http.MethodDelete, path, handler)
}

func (ro chiRouter) Deletef(path string, handler http.HandlerFunc, mws ...Middleware) {
	ro.chi.With(mws...).Method(http.MethodDelete, path, handler)
}

func (ro chiRouter) Get(path string, handler http.Handler, mws ...Middleware) {
	ro.chi.With(mws...).Method(http.MethodGet, path, handler)
}

func (ro chiRouter) Getf(path string, handler http.HandlerFunc, mws ...Middleware) {
	ro.chi.With(mws...).Method(http.MethodGet, path, handler)
}

func (ro chiRouter) Head(path string, handler http.Handler, mws ...Middleware) {
	ro.chi.With(mws...).Method(http.MethodHead, path, handler)
}

func (ro chiRouter) Headf(path string, handler http.HandlerFunc, mws ...Middleware) {
	ro.chi.With(mws...).Method(http.MethodHead, path, handler)
}

func (ro chiRouter) Options(path string, handler http.Handler, mws ...Middleware) {
	ro.chi.With(mws...).Method(http.MethodOptions, path, handler)
}

func (ro chiRouter) Optionsf(path string, handler http.HandlerFunc, mws ...Middleware) {
	ro.chi.With(mws...).Method(http.MethodOptions, path, handler)
}

func (ro chiRouter) Patch(path string, handler http.Handler, mws ...Middleware) {
	ro.chi.With(mws...).Method(http.MethodPatch, path, handler)
}

func (ro chiRouter) Patchf(path string, handler http.HandlerFunc, mws ...Middleware) {
	ro.chi.With(mws...).Method(http.MethodPatch, path, handler)
}

func (ro chiRouter) Post(path string, handler http.Handler, mws ...Middleware) {
	ro.chi.With(mws...).Method(http.MethodPost, path, handler)
}

func (ro chiRouter) Postf(path string, handler http.HandlerFunc, mws ...Middleware) {
	ro.chi.With(mws...).Method(http.MethodPost, path, handler)
}

func (ro chiRouter) Put(path string, handler http.Handler, mws ...Middleware) {
	ro.chi.With(mws...).Method(http.MethodPut, path, handler)
}

func (ro chiRouter) Putf(path string, handler http.HandlerFunc, mws ...Middleware) {
	ro.chi.With(mws...).Method(http.MethodPut, path, handler)
}

func (ro chiRouter) Trace(path string, handler http.Handler, mws ...Middleware) {
	ro.chi.With(mws...).Method(http.MethodTrace, path, handler)
}

func (ro chiRouter) Tracef(path string, handler http.HandlerFunc, mws ...Middleware) {
	ro.chi.With(mws...).Method(http.MethodTrace, path, handler)
}

func (ro chiRouter) Group(f func(Router), mws ...Middleware) {
	ro.chi.Group(func(inner chi.Router) {
		rr := newSubrouter(inner)
		rr.Use(mws...)
		f(rr)
	})
}

func (ro chiRouter) Route(path string, f func(Router), mws ...Middleware) {
	ro.chi.Route(path, func(inner chi.Router) {
		rr := newSubrouter(inner)
		rr.Use(mws...)
		f(rr)
	})
}

func (ro chiRouter) Use(m ...Middleware) {
	ro.chi.Use(m...)
}
