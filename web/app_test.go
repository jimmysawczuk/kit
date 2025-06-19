package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jimmysawczuk/kit/web"
	"github.com/jimmysawczuk/kit/web/respond"
	"github.com/jimmysawczuk/kit/web/router"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestBasicApp(t *testing.T) {
	a := web.NewApp().Route(func(r router.Router) {
		r.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("hello, world"))
		}))
	})

	srv := httptest.NewServer(a)

	resp, err := http.Get(srv.URL + "/")
	require.NoError(t, err)
	require.Equal(t, "hello, world", getBody(resp.Body))
}

type module struct{}

var (
	_ web.Module        = module{}
	_ web.HealthChecker = module{}
	_ web.Shutdowner    = module{}
)

// Healthy implements web.Module.
func (m module) HealthCheck(_ context.Context) error { return nil }
func (m module) Shutdown(_ context.Context) error    { return nil }
func (m module) Name() string                        { return "test" }

// Route implements web.Module.
func (m module) Route(r router.Router) {
	r.Get("/hello", web.Handler(func(ctx context.Context, l *zerolog.Logger, w http.ResponseWriter, r *http.Request) {
		respond.Success(ctx, http.StatusOK, struct {
			Hello string `json:"hello"`
		}{
			Hello: "world",
		}).Write(w)
	}))

	r.Post("/world", web.Handler(func(ctx context.Context, l *zerolog.Logger, w http.ResponseWriter, r *http.Request) {
		respond.Success(ctx, http.StatusOK, struct {
			Hello string `json:"hello"`
		}{
			Hello: "universe",
		}).Write(w)
	}))
}

func TestModule(t *testing.T) {
	mod := module{}

	a := web.NewApp().
		WithModule(mod).
		WithHealthCheckHandler("/health")

	srv := httptest.NewServer(a)

	{
		resp, err := http.Get(srv.URL + "/hello")
		require.NoError(t, err)
		require.Equal(t, `{"hello":"world"}`, strings.TrimSpace(getBody(resp.Body)))
	}

	{
		resp, err := http.Post(srv.URL+"/world", "", nil)
		require.NoError(t, err)
		require.Equal(t, `{"hello":"universe"}`, strings.TrimSpace(getBody(resp.Body)))
	}

	require.Len(t, a.HealthCheckers(), 1)
	require.Len(t, a.Shutdowners(), 1)

	{
		resp, err := http.Get(srv.URL + "/health")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}
}

func TestRouteCtx(t *testing.T) {
	a := web.NewApp().Route(func(r router.Router) {
		r.Get("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rctx := chi.RouteContext(r.Context())

			v := rctx.URLParam("*")

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(v))
		}))
	})

	srv := httptest.NewServer(a)

	resp, err := http.Get(srv.URL + "/index.html")
	require.NoError(t, err)
	require.Equal(t, "index.html", getBody(resp.Body))
}
