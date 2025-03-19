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

// NewApp instanciates a new App with a new router.
func NewApp() *App {
	return &App{
		router: router.New(),
	}
}

func (a App) Routes() []router.Route {
	if a.router == nil {
		return nil
	}

	return a.router.Routes()
}

// Route allow's modifying the App's router using the callback.
func (a App) Route(f func(router.Router)) *App {
	if a.router == nil {
		a.router = router.New()
	}

	a.router.Group(f)
	return &a
}

// WithLogger attaches the provided *zerolog.Logger to the App.
func (a App) WithLogger(logger *zerolog.Logger) *App {
	a.logger = logger
	return &a
}

// WithRouter attaches the provided Router to the app.
func (a App) WithRouter(r router.Router) *App {
	a.router = r
	return &a
}

// WithHandler attaches the provided http.Handler to the app.
func (a App) WithHandler(handler http.Handler) *App {
	a.handler = handler
	return &a
}

// WithShutdown registers the provided Shutdowner to the app.
func (a App) WithShutdown(s Shutdowner) *App {
	a.sd = append(a.sd, s)
	return &a
}

// WithHealthcheck registers the provided Shutdowner to the app.
func (a App) WithHealthCheck(h HealthChecker) *App {
	a.hc = append(a.hc, h)
	return &a
}

// ServeHTTP implements http.Handler. If the app has an attached handler, ServeHTTP proxies
// the requests there. Otherwise, it proxies to the attached Router.
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

// RouteModule attaches the routes from the provided module to the app, with the provided middleware,
// health checks and shutdown funcs.
func (a *App) RouteModule(m Module, mws ...Middleware) *App {
	if ty, ok := m.(HealthChecker); ok {
		a.WithHealthCheck(ty)
	}

	if ty, ok := m.(Shutdowner); ok {
		a.WithShutdown(ty)
	}

	if a.router == nil {
		a.router = router.New()
	}

	a.router.Group(func(r router.Router) {
		m.Route(r)
	}, mws...)

	return a
}

func (a *App) HealthCheckers() []HealthChecker {
	return a.hc
}

func (a *App) Shutdowners() []Shutdowner {
	return a.sd
}
