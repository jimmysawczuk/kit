package respond_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jimmysawczuk/kit/web"
	"github.com/jimmysawczuk/kit/web/requestid"
	"github.com/jimmysawczuk/kit/web/respond"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

var log = zerolog.New(os.Stderr)

func assignFakeRequestID(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := requestid.Set(r.Context(), "FakeID")
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	})
}

func TestRespondWithSuccess(t *testing.T) {
	tests := []struct {
		name              string
		handler           http.Handler
		expectedStatus    int
		expectedRequestID string
		expectedOutput    string
	}{
		{
			name: "200_RESPONSE",
			handler: web.Handler(func(ctx context.Context, log *zerolog.Logger, w http.ResponseWriter, r *http.Request) {
				respond.Success(ctx, http.StatusOK, struct {
					Success bool `json:"success"`
				}{
					Success: true,
				}).Write(w)
			}),
			expectedStatus: http.StatusOK,
			expectedOutput: `{"success":true}`,
		},
		{
			name: "201_RESPONSE",
			handler: web.Handler(func(ctx context.Context, log *zerolog.Logger, w http.ResponseWriter, r *http.Request) {
				ctx = requestid.Set(ctx, "FakeID")
				respond.Success(ctx, http.StatusCreated, struct {
					Status string `json:"status"`
				}{
					Status: "created",
				}).Write(w)
			}),
			expectedStatus:    http.StatusCreated,
			expectedRequestID: "FakeID",
			expectedOutput:    `{"status":"created"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := httptest.NewServer(test.handler)
			defer srv.Close()

			req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.Equal(t, test.expectedStatus, resp.StatusCode)
			require.Equal(t, test.expectedRequestID, resp.Header.Get("X-Request-Id"))

			buf := bytes.Buffer{}
			_, err = io.Copy(&buf, resp.Body)
			require.NoError(t, err)

			require.Equal(t, test.expectedOutput, strings.TrimSpace(buf.String()))
		})
	}
}

func TestRespondWithError(t *testing.T) {
	tests := []struct {
		name              string
		handler           web.Handler
		expectedStatus    int
		expectedRequestID string
		expectedOutput    string
	}{
		{
			name: "400_RESPONSE",
			handler: web.Handler(func(ctx context.Context, log *zerolog.Logger, w http.ResponseWriter, r *http.Request) {
				ctx = requestid.Set(ctx, "FakeID")
				respond.Error(ctx, http.StatusBadRequest, errors.New("bad request")).Write(w)
			}),
			expectedStatus:    http.StatusBadRequest,
			expectedRequestID: "FakeID",
			expectedOutput:    `{"error":"bad request","requestID":"FakeID","status":400}`,
		},
		{
			name: "400_RESPONSE_EXTRA_INFO",
			handler: web.Handler(func(ctx context.Context, log *zerolog.Logger, w http.ResponseWriter, r *http.Request) {
				ctx = requestid.Set(ctx, "FakeID")
				err := respond.ErrWithInfo(errors.New("bad request"), struct {
					Problem string `json:"problem"`
				}{
					Problem: "Bad user ID",
				})

				respond.Error(ctx, http.StatusBadRequest, err).Write(w)
			}),
			expectedStatus:    http.StatusBadRequest,
			expectedRequestID: "FakeID",
			expectedOutput:    `{"error":"bad request","requestID":"FakeID","status":400,"info":{"problem":"Bad user ID"}}`,
		},
		{
			name: "400_RESPONSE_ERROR_CODE_INFO",
			handler: web.Handler(func(ctx context.Context, log *zerolog.Logger, w http.ResponseWriter, r *http.Request) {
				ctx = requestid.Set(ctx, "FakeID")
				err := respond.ErrWithInfo(errors.New("bad request"), struct {
					Problem string `json:"problem"`
				}{
					Problem: "Bad user ID",
				})

				respond.CodedError(ctx, http.StatusBadRequest, "ERR_1234", err).Write(w)
			}),
			expectedStatus:    http.StatusBadRequest,
			expectedRequestID: "FakeID",
			expectedOutput:    `{"error":"bad request","requestID":"FakeID","code":"ERR_1234","status":400,"info":{"problem":"Bad user ID"}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := httptest.NewServer(test.handler)
			defer srv.Close()

			req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.Equal(t, test.expectedStatus, resp.StatusCode)
			require.Equal(t, test.expectedRequestID, resp.Header.Get("X-Request-Id"))

			buf := bytes.Buffer{}
			_, err = io.Copy(&buf, resp.Body)
			require.NoError(t, err)

			require.Equal(t, test.expectedOutput, strings.TrimSpace(buf.String()))
		})
	}
}

func TestRespondWithErrorSuppressed(t *testing.T) {
	tests := []struct {
		name              string
		handler           web.Handler
		expectedStatus    int
		expectedRequestID string
		expectedOutput    string
	}{
		{
			name: "400_RESPONSE",
			handler: web.Handler(func(ctx context.Context, log *zerolog.Logger, w http.ResponseWriter, r *http.Request) {
				ctx = requestid.Set(ctx, "FakeID")
				respond.Error(ctx, http.StatusBadRequest, errors.New("bad request")).Write(w)
			}),
			expectedStatus:    http.StatusBadRequest,
			expectedRequestID: "FakeID",
			expectedOutput:    `{"requestID":"FakeID","status":400}`,
		},
		{
			name: "400_RESPONSE_EXTRA_INFO",
			handler: web.Handler(func(ctx context.Context, log *zerolog.Logger, w http.ResponseWriter, r *http.Request) {
				ctx = requestid.Set(ctx, "FakeID")
				err := respond.ErrWithInfo(errors.New("bad request"), struct {
					Problem string `json:"problem"`
				}{
					Problem: "Bad user ID",
				})

				respond.Error(ctx, http.StatusBadRequest, err).Write(w)
			}),
			expectedStatus:    http.StatusBadRequest,
			expectedRequestID: "FakeID",
			expectedOutput:    `{"requestID":"FakeID","status":400,"info":{"problem":"Bad user ID"}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			respond.DefaultResponder = &respond.JSONResponder{SuppressErrors: true}

			srv := httptest.NewServer(test.handler)
			defer srv.Close()

			req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.Equal(t, test.expectedStatus, resp.StatusCode)
			require.Equal(t, test.expectedRequestID, resp.Header.Get("X-Request-Id"))

			buf := bytes.Buffer{}
			_, err = io.Copy(&buf, resp.Body)
			require.NoError(t, err)

			require.Equal(t, test.expectedOutput, strings.TrimSpace(buf.String()))
		})
	}
}

func TestWithHeader(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.Handler
		expectedStatus int
		expectedHeader string
		expectedValue  string
	}{
		{
			name: "SUCCESS_WITH_CUSTOM_HEADER",
			handler: web.Handler(func(ctx context.Context, log *zerolog.Logger, w http.ResponseWriter, r *http.Request) {
				resp := respond.Success(ctx, http.StatusOK, struct {
					Success bool `json:"success"`
				}{
					Success: true,
				})
				resp.WithHeader(func(h http.Header) http.Header {
					h.Set("X-Custom-Header", "custom-value")
					return h
				})
				resp.Write(w)
			}),
			expectedStatus: http.StatusOK,
			expectedHeader: "X-Custom-Header",
			expectedValue:  "custom-value",
		},
		{
			name: "ERROR_WITH_CUSTOM_HEADER",
			handler: web.Handler(func(ctx context.Context, log *zerolog.Logger, w http.ResponseWriter, r *http.Request) {
				resp := respond.Error(ctx, http.StatusBadRequest, errors.New("bad request"))
				resp.WithHeader(func(h http.Header) http.Header {
					h.Set("X-Error-Code", "ERR_400")
					return h
				})
				resp.Write(w)
			}),
			expectedStatus: http.StatusBadRequest,
			expectedHeader: "X-Error-Code",
			expectedValue:  "ERR_400",
		},
		{
			name: "SUCCESS_WITH_MULTIPLE_HEADERS",
			handler: web.Handler(func(ctx context.Context, log *zerolog.Logger, w http.ResponseWriter, r *http.Request) {
				resp := respond.Success(ctx, http.StatusOK, struct {
					Data string `json:"data"`
				}{
					Data: "test",
				})
				resp.WithHeader(func(h http.Header) http.Header {
					h.Set("X-Header-1", "value1")
					h.Set("X-Header-2", "value2")
					h.Add("X-Header-3", "value3a")
					h.Add("X-Header-3", "value3b")
					return h
				})
				resp.Write(w)
			}),
			expectedStatus: http.StatusOK,
			expectedHeader: "X-Header-1",
			expectedValue:  "value1",
		},
		{
			name: "SUCCESS_WITH_CACHE_CONTROL",
			handler: web.Handler(func(ctx context.Context, log *zerolog.Logger, w http.ResponseWriter, r *http.Request) {
				resp := respond.Success(ctx, http.StatusOK, struct {
					Data string `json:"data"`
				}{
					Data: "cached",
				})
				resp.WithHeader(func(h http.Header) http.Header {
					h.Set("Cache-Control", "max-age=3600")
					return h
				})
				resp.Write(w)
			}),
			expectedStatus: http.StatusOK,
			expectedHeader: "Cache-Control",
			expectedValue:  "max-age=3600",
		},
		{
			name: "SUCCESS_PRESERVES_CONTENT_TYPE",
			handler: web.Handler(func(ctx context.Context, log *zerolog.Logger, w http.ResponseWriter, r *http.Request) {
				resp := respond.Success(ctx, http.StatusOK, struct {
					Data string `json:"data"`
				}{
					Data: "test",
				})
				resp.WithHeader(func(h http.Header) http.Header {
					h.Set("X-Custom", "value")
					return h
				})
				resp.Write(w)
			}),
			expectedStatus: http.StatusOK,
			expectedHeader: "Content-Type",
			expectedValue:  "application/json; charset=utf-8",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := httptest.NewServer(test.handler)
			defer srv.Close()

			req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.Equal(t, test.expectedStatus, resp.StatusCode)
			require.Equal(t, test.expectedValue, resp.Header.Get(test.expectedHeader))
		})
	}
}

func TestWithHeaderMultipleValues(t *testing.T) {
	handler := web.Handler(func(ctx context.Context, log *zerolog.Logger, w http.ResponseWriter, r *http.Request) {
		resp := respond.Success(ctx, http.StatusOK, struct {
			Data string `json:"data"`
		}{
			Data: "test",
		})
		resp.WithHeader(func(h http.Header) http.Header {
			h.Add("X-Multiple", "value1")
			h.Add("X-Multiple", "value2")
			return h
		})
		resp.Write(w)
	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	values := resp.Header.Values("X-Multiple")
	require.Len(t, values, 2)
	require.Contains(t, values, "value1")
	require.Contains(t, values, "value2")
}
