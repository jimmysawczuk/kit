package web

import (
	"net/http"

	"github.com/go-chi/chi"
)

// App holds a router for endpoints as well as Shutdowners and Healthcheckers.
type App struct {
	router chi.Router

	sd []Shutdowner
	hc []HealthChecker
}

// NewApp instanciates a new App, with the provided logger and global middleware.
func NewApp() *App {
	return &App{
		router: chi.NewMux(),
	}
}

func (a *App) Route(f func(r chi.Router)) *App {
	f(a.router)
	return a
}

func (a *App) WithShutdown(s Shutdowner) *App {
	a.sd = append(a.sd, s)
	return a
}

func (a *App) WithHealthCheck(h HealthChecker) *App {
	a.hc = append(a.hc, h)
	return a
}

// ServeHTTP implements http.Handler, proxying the incoming request to the Router.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}

// RouteModule attaches the routes from the provided module to the app, with the provided middleware.
func (a *App) RouteModule(m Module, name string, mws ...Middleware) *App {
	// TODO: register these with some sort of name
	if ty, ok := m.(HealthChecker); ok {
		a.WithHealthCheck(ty)
	}

	if ty, ok := m.(Shutdowner); ok {
		a.WithShutdown(ty)
	}

	a.Route(func(r chi.Router) {
		m.Route(r, mws...)
	})

	return a
}
