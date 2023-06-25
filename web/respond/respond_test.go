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
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var log *zap.Logger

func TestMain(m *testing.M) {
	log, _ = zap.NewDevelopment()

	os.Exit(m.Run())
}

func bootstrap(ctx context.Context, log *zap.Logger, h web.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h(ctx, log, w, r)
	})
}

func TestRespondWithSuccess(t *testing.T) {
	tests := []struct {
		name              string
		handler           web.Handler
		expectedStatus    int
		expectedRequestID string
		expectedOutput    string
	}{
		{
			name: "200_RESPONSE",
			handler: web.Handler(func(ctx context.Context, log *zap.Logger, w http.ResponseWriter, r *http.Request) {
				respond.WithSuccess(ctx, log, w, r, http.StatusOK, struct {
					Success bool `json:"success"`
				}{
					Success: true,
				})
			}),
			expectedStatus: http.StatusOK,
			expectedOutput: `{"success":true}`,
		},
		{
			name: "201_RESPONSE",
			handler: web.Handler(func(ctx context.Context, log *zap.Logger, w http.ResponseWriter, r *http.Request) {
				ctx = requestid.Set(ctx, "FakeID")
				respond.WithSuccess(ctx, log, w, r, http.StatusCreated, struct {
					Status string `json:"status"`
				}{
					Status: "created",
				})
			}),
			expectedStatus:    http.StatusCreated,
			expectedRequestID: "FakeID",
			expectedOutput:    `{"status":"created"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hdlr := bootstrap(context.Background(), log, test.handler)

			srv := httptest.NewServer(hdlr)
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
			handler: web.Handler(func(ctx context.Context, log *zap.Logger, w http.ResponseWriter, r *http.Request) {
				ctx = requestid.Set(ctx, "FakeID")
				respond.WithError(ctx, log, w, r, http.StatusBadRequest, errors.New("bad request"))
			}),
			expectedStatus:    http.StatusBadRequest,
			expectedRequestID: "FakeID",
			expectedOutput:    `{"error":"bad request","requestID":"FakeID","status":400}`,
		},
		{
			name: "400_RESPONSE_EXTRA_INFO",
			handler: web.Handler(func(ctx context.Context, log *zap.Logger, w http.ResponseWriter, r *http.Request) {
				ctx = requestid.Set(ctx, "FakeID")
				err := respond.ErrWithInfo(errors.New("bad request"), struct {
					Problem string `json:"problem"`
				}{
					Problem: "Bad user ID",
				})

				respond.WithError(ctx, log, w, r, http.StatusBadRequest, err)
			}),
			expectedStatus:    http.StatusBadRequest,
			expectedRequestID: "FakeID",
			expectedOutput:    `{"error":"bad request","requestID":"FakeID","status":400,"info":{"problem":"Bad user ID"}}`,
		},
		{
			name: "400_RESPONSE_ERROR_CODE_INFO",
			handler: web.Handler(func(ctx context.Context, log *zap.Logger, w http.ResponseWriter, r *http.Request) {
				ctx = requestid.Set(ctx, "FakeID")
				err := respond.ErrWithInfo(errors.New("bad request"), struct {
					Problem string `json:"problem"`
				}{
					Problem: "Bad user ID",
				})

				respond.WithCodedError(ctx, log, w, r, http.StatusBadRequest, "ERR_1234", err)
			}),
			expectedStatus:    http.StatusBadRequest,
			expectedRequestID: "FakeID",
			expectedOutput:    `{"error":"bad request","requestID":"FakeID","code":"ERR_1234","status":400,"info":{"problem":"Bad user ID"}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hdlr := bootstrap(context.Background(), log, test.handler)

			srv := httptest.NewServer(hdlr)
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
			handler: web.Handler(func(ctx context.Context, log *zap.Logger, w http.ResponseWriter, r *http.Request) {
				ctx = requestid.Set(ctx, "FakeID")
				respond.WithError(ctx, log, w, r, http.StatusBadRequest, errors.New("bad request"))
			}),
			expectedStatus:    http.StatusBadRequest,
			expectedRequestID: "FakeID",
			expectedOutput:    `{"requestID":"FakeID","status":400}`,
		},
		{
			name: "400_RESPONSE_EXTRA_INFO",
			handler: web.Handler(func(ctx context.Context, log *zap.Logger, w http.ResponseWriter, r *http.Request) {
				ctx = requestid.Set(ctx, "FakeID")
				err := respond.ErrWithInfo(errors.New("bad request"), struct {
					Problem string `json:"problem"`
				}{
					Problem: "Bad user ID",
				})

				respond.WithError(ctx, log, w, r, http.StatusBadRequest, err)
			}),
			expectedStatus:    http.StatusBadRequest,
			expectedRequestID: "FakeID",
			expectedOutput:    `{"requestID":"FakeID","status":400,"info":{"problem":"Bad user ID"}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			respond.DefaultResponder.SuppressErrors = true
			hdlr := bootstrap(context.Background(), log, test.handler)

			srv := httptest.NewServer(hdlr)
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
