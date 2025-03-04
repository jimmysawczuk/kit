package web

import (
	"net/http"

	"github.com/jimmysawczuk/kit/web/router"
	"github.com/rs/zerolog"
)

// App holds a router for endpoints as well as Shutdowners and Healthcheckers.
type App struct {
	handler http.Handler
	router  router.Router
	logger  *zerolog.Logger

	sd []Shutdowner
	hc []HealthChecker
}

// NewApp instanciates a new App, with the provided logger and global middleware.
func NewApp() *App {
	return &App{
		router: router.New(),
	}
}

func (a App) Route(f func(router.Router)) *App {
	if a.router == nil {
		a.router = router.New()
	}

	f(a.router)
	return &a
}

func (a App) WithLogger(logger *zerolog.Logger) *App {
	a.logger = logger
	return &a
}

func (a App) WithRouter(r router.Router) *App {
	a.router = r
	return &a
}

func (a App) WithHandler(handler http.Handler) *App {
	a.handler = handler
	return &a
}

func (a App) WithShutdown(s Shutdowner) *App {
	a.sd = append(a.sd, s)
	return &a
}

func (a App) WithHealthCheck(h HealthChecker) *App {
	a.hc = append(a.hc, h)
	return &a
}

// ServeHTTP implements http.Handler, proxying the incoming request to the Router.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if a.logger != nil {
		r = r.WithContext(a.logger.WithContext(r.Context()))
	}

	if a.handler != nil {
		a.handler.ServeHTTP(w, r)
		return
	}

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

	a.router.Group(func(r router.Router) {
		m.Route(r)
	}, mws...)

	return a
}
