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
	Mount(string, http.Handler, ...Middleware)

	Routes() []Route
}

type Route struct {
	Method  string
	Path    string
	Handler string
}

type chiRouter struct {
	chi chi.Router
}

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
	if len(mws) > 0 {
		ro.chi.Method(method, path, chain(handler, mws...))
	} else {
		ro.chi.Method(method, path, handler)
	}
}

func (ro chiRouter) Connect(path string, handler http.Handler, mws ...Middleware) {
	ro.Method(http.MethodConnect, path, handler, mws...)
}

func (ro chiRouter) Connectf(path string, handler http.HandlerFunc, mws ...Middleware) {
	ro.Method(http.MethodConnect, path, handler, mws...)
}

func (ro chiRouter) Delete(path string, handler http.Handler, mws ...Middleware) {
	ro.Method(http.MethodDelete, path, handler, mws...)
}

func (ro chiRouter) Deletef(path string, handler http.HandlerFunc, mws ...Middleware) {
	ro.Method(http.MethodDelete, path, handler, mws...)
}

func (ro chiRouter) Get(path string, handler http.Handler, mws ...Middleware) {
	ro.Method(http.MethodGet, path, handler, mws...)
}

func (ro chiRouter) Getf(path string, handler http.HandlerFunc, mws ...Middleware) {
	ro.Method(http.MethodGet, path, handler, mws...)
}

func (ro chiRouter) Head(path string, handler http.Handler, mws ...Middleware) {
	ro.Method(http.MethodHead, path, handler, mws...)
}

func (ro chiRouter) Headf(path string, handler http.HandlerFunc, mws ...Middleware) {
	ro.Method(http.MethodHead, path, handler, mws...)
}

func (ro chiRouter) Options(path string, handler http.Handler, mws ...Middleware) {
	ro.Method(http.MethodOptions, path, handler, mws...)
}

func (ro chiRouter) Optionsf(path string, handler http.HandlerFunc, mws ...Middleware) {
	ro.Method(http.MethodOptions, path, handler, mws...)
}

func (ro chiRouter) Patch(path string, handler http.Handler, mws ...Middleware) {
	ro.Method(http.MethodPatch, path, handler, mws...)
}

func (ro chiRouter) Patchf(path string, handler http.HandlerFunc, mws ...Middleware) {
	ro.Method(http.MethodPatch, path, handler, mws...)
}

func (ro chiRouter) Post(path string, handler http.Handler, mws ...Middleware) {
	ro.Method(http.MethodPost, path, handler, mws...)
}

func (ro chiRouter) Postf(path string, handler http.HandlerFunc, mws ...Middleware) {
	ro.Method(http.MethodPost, path, handler, mws...)
}

func (ro chiRouter) Put(path string, handler http.Handler, mws ...Middleware) {
	ro.Method(http.MethodPut, path, handler, mws...)
}

func (ro chiRouter) Putf(path string, handler http.HandlerFunc, mws ...Middleware) {
	ro.Method(http.MethodPut, path, handler, mws...)
}

func (ro chiRouter) Trace(path string, handler http.Handler, mws ...Middleware) {
	ro.Method(http.MethodTrace, path, handler, mws...)
}

func (ro chiRouter) Tracef(path string, handler http.HandlerFunc, mws ...Middleware) {
	ro.Method(http.MethodTrace, path, handler, mws...)
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

func (ro chiRouter) Mount(path string, h http.Handler, mws ...Middleware) {
	ro.chi.Route(path, func(r chi.Router) {
		r.Use(mws...)
		r.Mount("/", h)
	})
}

func (ro chiRouter) Use(m ...Middleware) {
	ro.chi.Use(m...)
}

func (ro chiRouter) Routes() []Route {
	tbr := []Route{}
	chi.Walk(ro.chi, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		tbr = append(tbr, Route{
			Method: method,
			Path:   route,
		})
		return nil
	})
	return tbr
}

// chain applies middlewares to a handler
func chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
