package web_test

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/jimmysawczuk/kit/web/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

type ctxKey struct{}

var appendCtxKey ctxKey

func appendStr(s string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			v, ok := ctx.Value(appendCtxKey).(string)
			if !ok {
				v = ""
			}

			v = v + s

			ctx = context.WithValue(ctx, appendCtxKey, v)

			r = r.WithContext(ctx)

			h.ServeHTTP(w, r)
		})
	}
}

func middlewareResult(ctx context.Context, l *zerolog.Logger, w http.ResponseWriter, r *http.Request) {
	v, _ := ctx.Value(appendCtxKey).(string)

	w.Write([]byte(v))
}

func TestMiddlewareOrder(t *testing.T) {
	router := chi.NewRouter()
	router.Group(func(r chi.Router) {
		r.Use(appendStr("A"), appendStr("B"), appendStr("C"))
		r.Method(http.MethodGet, "/hello", web.HandlerFunc(middlewareResult))
	})

	log.Printf("%+v", router.Middlewares())

	router.Group(func(r chi.Router) {
		r.Use(appendStr("D"), appendStr("E"), appendStr("F"))
		r.Method(http.MethodGet, "/world", web.HandlerFunc(middlewareResult))

		r.Group(func(r chi.Router) {
			r.Use(appendStr("G"), appendStr("H"))

			r.Method(http.MethodGet, "/world/v2", web.HandlerFunc(middlewareResult))
		})
	})

	a := web.NewApp(router)

	srv := httptest.NewServer(a)

	{
		resp, err := http.Get(srv.URL + "/hello")
		require.NoError(t, err)
		require.Equal(t, `ABC`, strings.TrimSpace(getBody(resp.Body)))
	}

	{
		resp, err := http.Get(srv.URL + "/world")
		require.NoError(t, err)
		require.Equal(t, `DEF`, strings.TrimSpace(getBody(resp.Body)))
	}

	{
		resp, err := http.Get(srv.URL + "/world/v2")
		require.NoError(t, err)
		require.Equal(t, `DEFGH`, strings.TrimSpace(getBody(resp.Body)))
	}
}
