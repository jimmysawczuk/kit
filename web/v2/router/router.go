package web

import (
	"net/http"

	"github.com/jimmysawczuk/kit/web/v2"
)

type Router interface {
	http.Handler

	Use(...web.Middleware)

	Group(func(Router), ...web.Middleware)
	Mount(string, func(Router), ...web.Middleware)

	Get(string, http.Handler)
	Getf(string, http.HandlerFunc)
	Getc(string, web.HandlerFunc)
}

// type Router struct {
// 	chi.Router
// }

// func NewRouter() *Router {
// 	return &Router{
// 		Router: chi.NewMux(),
// 	}
// }

// func (ro *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	ro.Router.ServeHTTP(w, r)
// }

// func (ro *Router) Use(mws ...Middleware) {
// 	ro.Router.Use(mws...)
// }

// func (ro *Router) Group(fn func(*Router), mws ...Middleware) {
// 	ro.
// 		ro.With(mws...)
// 	fn(ro)
// }

// func (ro *Router) Route(pattern string, fn func(*Router), mws ...Middleware) {
// 	sub := NewRouter()
// 	sub.Use(ro.Middlewares()...)
// 	sub.Use(mws...)
// 	fn(sub)
// 	ro.Mount(pattern, sub)
// }

// func (ro *Router) Mount(pattern string, sub *Router) {
// 	ro.Router.Mount(pattern, sub)
// }

// func (ro *Router) Get(pattern string, handler http.Handler) {
// 	ro.Router.Method(http.MethodGet, pattern, handler)
// }

// func (ro *Router) Getf(pattern string, handler http.HandlerFunc) {
// 	ro.Router.Method(http.MethodGet, pattern, handler)
// }

// func (ro *Router) Getc(pattern string, handler HandlerFunc) {
// 	ro.Router.Method(http.MethodGet, pattern, handler)
// }

// func (ro *Router) Post(pattern string, handler http.Handler) {
// 	ro.Router.Method(http.MethodPost, pattern, handler)
// }

// func (ro *Router) Postc(pattern string, handler HandlerFunc) {
// 	ro.Router.Method(http.MethodPost, pattern, handler)
// }
