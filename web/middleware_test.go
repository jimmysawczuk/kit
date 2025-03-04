package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jimmysawczuk/kit/web"
	"github.com/jimmysawczuk/kit/web/router"
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
	a := web.NewApp().Route(func(r router.Router) {
		r.Group(func(r router.Router) {
			r.Use(appendStr("A"), appendStr("B"), appendStr("C"))
			r.Get("/hello", web.Handler(middlewareResult))
		})

		r.Group(func(r router.Router) {
			r.Get("/world", web.Handler(middlewareResult), appendStr("I"))

			r.Group(func(r router.Router) {
				r.Use(appendStr("G"), appendStr("H"))

				r.Get("/world/v2", web.Handler(middlewareResult), appendStr("J"))
				r.Get("/world/v3", web.Handler(middlewareResult), appendStr("K"))
			})
		}, appendStr("D"), appendStr("E"), appendStr("F"))
	})

	srv := httptest.NewServer(a)

	tests := []struct {
		path     string
		expected string
	}{
		{
			path:     "/hello",
			expected: "ABC",
		},

		{
			path:     "/world",
			expected: "DEFI",
		},

		{
			path:     "/world/v2",
			expected: "DEFGHJ",
		},

		{
			path:     "/world/v3",
			expected: "DEFGHK",
		},
	}

	for _, test := range tests {
		resp, err := http.Get(srv.URL + test.path)
		require.NoError(t, err)
		require.Equal(t, test.expected, strings.TrimSpace(getBody(resp.Body)))
	}
}
