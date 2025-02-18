package web_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/jimmysawczuk/kit/web/v2"
	"github.com/jimmysawczuk/kit/web/v2/respond"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestBasicApp(t *testing.T) {
	router := chi.NewRouter()

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello, world"))
	})

	a := web.NewApp(router)

	srv := httptest.NewServer(a)

	resp, err := http.Get(srv.URL + "/")
	require.NoError(t, err)
	require.Equal(t, "hello, world", getBody(resp.Body))
}

type module struct{}

var _ web.Module = module{}

// Healthy implements web.Module.
func (m module) Healthy() error {
	return nil
}

// Route implements web.Module.
func (m module) Route(router chi.Router, mws ...web.Middleware) {
	router.Method(http.MethodGet, "/hello", web.HandlerFunc(func(ctx context.Context, l *zerolog.Logger, w http.ResponseWriter, r *http.Request) {
		respond.WithSuccess(ctx, w, r, http.StatusOK, struct {
			Hello string `json:"hello"`
		}{
			Hello: "world",
		})
	}))

	router.Method(http.MethodPost, "/world", web.HandlerFunc(func(ctx context.Context, l *zerolog.Logger, w http.ResponseWriter, r *http.Request) {
		respond.WithSuccess(ctx, w, r, http.StatusOK, struct {
			Hello string `json:"hello"`
		}{
			Hello: "universe",
		})
	}))
}

func TestModule(t *testing.T) {
	router := chi.NewRouter()

	mod := module{}

	a := web.NewApp(router)
	a.RouteModule(mod)

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
}

func getBody(r io.ReadCloser) string {
	buf := bytes.Buffer{}
	io.Copy(&buf, r)
	r.Close()
	return buf.String()
}
