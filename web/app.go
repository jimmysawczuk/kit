package web

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/jimmysawczuk/kit/web/respond"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// App holds a Router for endpoints as well as a contextual logger.
type App struct {
	r   Router
	log *zap.Logger
	sd  []Shutdowner

	healthChecks []func() error
}

// NewApp instanciates a new App, with the provided logger and global middleware.
func NewApp(log *zap.Logger, mws ...Middleware) *App {
	a := &App{
		log: log,
	}

	a.r = Router{
		r:   chi.NewRouter(),
		mws: append([]Middleware{a.Bootstrap}, mws...),
	}

	return a
}

// Bootstrap is middleware that initializes a new context.Context from the request context, and creates a new log
// entry for passing through the request. It should be the *first* middleware invoked.
func (a *App) Bootstrap(h Handler) Handler {
	return func(_ context.Context, _ *zap.Logger, w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		entry := a.log.With()

		h(ctx, entry, w, r)
	}
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
func (a *App) Health(ctx context.Context, log *zap.Logger, w http.ResponseWriter, r *http.Request) {
	if err := a.checkHealth(); err != nil {
		respond.WithError(ctx, log, w, r, http.StatusServiceUnavailable, errors.New("not healthy"))
		return
	}

	respond.WithSuccess(ctx, log, w, r, http.StatusOK, "healthy")
}
