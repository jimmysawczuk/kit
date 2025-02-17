package web

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jimmysawczuk/kit/web/respond"
	"github.com/sirupsen/logrus"
)

// App holds a Router for endpoints as well as a contextual logger.
type App struct {
	r Router

	sd           []Shutdowner
	healthChecks []func() error
}

// NewApp instanciates a new App, with the provided logger and global middleware.
func NewApp(mws ...Middleware) *App {
	a := &App{}

	a.r = Router{
		r:   chi.NewMux(),
		mws: mws,
	}

	return a
}

// checkHealth iterates through all of its attached health checks, returning nil if they all return nil.
func (a *App) checkHealth() error {
	for _, h := range a.healthChecks {
		if err := h(); err != nil {
			return err
		}
	}

	return nil
}

// ServeHTTP implements http.Handler, proxying the incoming request to the Router.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.r.r.ServeHTTP(w, r)
}

// Shutdown closes any resources that we might want to responsibly close before the app is terminated.
func (a *App) Shutdown(ctx context.Context) error {
	return nil
}

// Route allows the caller to set custom routes on the global scope, with the provided middleware.
func (a *App) Route(rs func(r Router), mws ...Middleware) {
	a.r.Route("/", rs, mws...)
}

// RouteModule attaches the routes from the provided module to the app, with the provided middleware.
func (a *App) RouteModule(m Module, mws ...Middleware) {
	a.healthChecks = append(a.healthChecks, m.Healthy)
	m.Route(a.r, mws...)
}

// Health is a Handler checks the health of the App, emitting a 503 if not healthy.
func (a *App) Health(ctx context.Context, log logrus.FieldLogger, w http.ResponseWriter, r *http.Request) {
	if err := a.checkHealth(); err != nil {
		respond.WithError(ctx, log, w, r, http.StatusServiceUnavailable, errors.New("not healthy"))
		return
	}

	respond.WithSuccess(ctx, log, w, r, http.StatusOK, "healthy")
}
