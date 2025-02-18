package web

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/jimmysawczuk/kit/web/v2/respond"
	"github.com/rs/zerolog"
)

// App holds a router for endpoints as well as Shutdowners and Healthcheckers.
type App struct {
	router chi.Router

	sd []Shutdowner
	hc []HealthChecker
}

// NewApp instanciates a new App, with the provided logger and global middleware.
func NewApp(router chi.Router) *App {
	return &App{
		router: router,
	}
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

// Shutdown closes any resources that we might want to responsibly close before the app is terminated.
func (a *App) Shutdown(ctx context.Context) error {
	panic("not implemented")
	return nil
}

// RouteModule attaches the routes from the provided module to the app, with the provided middleware.
func (a *App) RouteModule(m Module, mws ...Middleware) {
	a.hc = append(a.hc, HealthCheck(m.Healthy))

	a.router.Group(func(r chi.Router) {
		m.Route(r, mws...)
	})
}

// Health is a Handler checks the health of the App, emitting a 503 if not healthy.
func (a *App) Health(ctx context.Context, log *zerolog.Logger, w http.ResponseWriter, r *http.Request) {
	if err := a.checkHealth(); err != nil {
		respond.WithError(ctx, w, r, http.StatusServiceUnavailable, errors.New("not healthy"))
		return
	}

	respond.WithSuccess(ctx, w, r, http.StatusOK, "healthy")
}

// checkHealth iterates through all of its attached health checks, returning nil if they all return nil.
func (a *App) checkHealth() error {
	for _, h := range a.hc {
		if err := h.Healthy(); err != nil {
			return err
		}
	}

	return nil
}
