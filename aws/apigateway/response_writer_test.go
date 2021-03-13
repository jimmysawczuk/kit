package apigateway

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResponseWriter(t *testing.T) {
	build := func() (*responseWriter, *http.Request) {
		wr := &responseWriter{
			header: http.Header{
				"Content-Type": []string{"text/plain"},
			},
			buf: &bytes.Buffer{},
		}
		r, _ := http.NewRequest(http.MethodGet, "http://localhost", nil)

		return wr, r
	}

	wr, r := build()
	okHandler(wr, r)

	require.Equal(t, "application/json; charset=utf-8", wr.contentType)
	require.Equal(t, "application/json; charset=utf-8", wr.Header().Get("content-type"))
	require.Equal(t, 200, wr.statusCode)

	wr, r = build()
	imgHandler(wr, r)

	require.Equal(t, "image/jpeg", wr.contentType)
	require.Equal(t, "image/jpeg", wr.Header().Get("content-type"))
	require.Equal(t, 200, wr.statusCode)
}
