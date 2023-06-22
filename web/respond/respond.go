package respond

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jimmysawczuk/kit/web/requestid"
	"github.com/pkg/errors"
	"golang.org/x/exp/slog"
)

type errResponse struct {
	Error     string      `json:"error,omitempty"`
	RequestID string      `json:"requestID,omitempty"`
	ErrorCode string      `json:"code,omitempty"`
	Status    int         `json:"status,omitempty"`
	Info      interface{} `json:"info,omitempty"`
}

// ErrorInfoer is an optional interface that errors can optionally implement to provide
// additional context in an error response.
type ErrorInfoer interface {
	error
	ErrorInfo() interface{}
}

type errWithInfo struct {
	err  error
	info interface{}
}

func (ei errWithInfo) Error() string {
	return ei.err.Error()
}

func (ei errWithInfo) ErrorInfo() interface{} {
	return ei.info
}

func ErrWithInfo(err error, info interface{}) errWithInfo {
	return errWithInfo{
		err:  err,
		info: info,
	}
}

// Responder is a set of config for responding to requests.
type Responder struct {
	SuppressErrors bool
}

// DefaultResponder is a default Responder.
var DefaultResponder = Responder{}

// WithError is a shortcut for WithCodedError(ctx, log, w, httpStatus, "", err).
func (re Responder) WithError(ctx context.Context, log *slog.Logger, w http.ResponseWriter, r *http.Request, httpStatus int, err error) {
	re.WithCodedError(ctx, log, w, r, httpStatus, "", err)
}

// WithCodedError writes the provided error to the ResponseWriter, as well as the HTTP status code.
// An enum-style code (i.e. INVALID_TOKEN) may also be provided.
func (re Responder) WithCodedError(ctx context.Context, log *slog.Logger, w http.ResponseWriter, r *http.Request, httpStatus int, code string, err error) {
	serr := http.StatusText(httpStatus)
	if err != nil {
		serr = err.Error()
	}

	reqID := requestid.Get(ctx)
	if reqID != "" {
		w.Header().Set("X-Request-Id", reqID)
	}

	if ct := w.Header().Get("Content-Type"); ct == "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	}
	w.WriteHeader(httpStatus)

	resp := errResponse{
		Error:     serr,
		RequestID: reqID,
		Status:    httpStatus,
		ErrorCode: code,
	}

	if ty, ok := errors.Cause(err).(ErrorInfoer); ok {
		resp.Info = ty.ErrorInfo()
	}

	if re.SuppressErrors {
		msg := log.With("error", err).With("statusCode", httpStatus)
		// msg := log.WithError(err).WithField("statusCode", httpStatus)
		if resp.Info != nil {
			msg = msg.With("info", resp.Info)
			// msg = msg.WithField("info", resp.Info)
		}

		resp.Error = ""

		msg.Error("error suppressed")
	}

	json.NewEncoder(w).Encode(resp)
}

// WithSuccess writes the provided response to the ResponseWriter (unwrapped) and sets the provided HTTP response status.
func (re Responder) WithSuccess(ctx context.Context, log *slog.Logger, w http.ResponseWriter, r *http.Request, httpStatus int, v interface{}) {
	if ct := w.Header().Get("Content-Type"); ct == "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	}

	reqID := requestid.Get(ctx)
	if reqID != "" {
		w.Header().Set("X-Request-Id", reqID)
	}

	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(v)
}

// WithError is a shortcut for DefaultResponder.WithError.
func WithError(ctx context.Context, log *slog.Logger, w http.ResponseWriter, r *http.Request, httpStatus int, err error) {
	DefaultResponder.WithError(ctx, log, w, r, httpStatus, err)
}

// WithCodedError is a shortcut for DefaultResponder.WithCodedError.
func WithCodedError(ctx context.Context, log *slog.Logger, w http.ResponseWriter, r *http.Request, httpStatus int, code string, err error) {
	DefaultResponder.WithCodedError(ctx, log, w, r, httpStatus, code, err)
}

// WithSuccess is a shortcut for DefaultResponder.WithSuccess.
func WithSuccess(ctx context.Context, log *slog.Logger, w http.ResponseWriter, r *http.Request, httpStatus int, v interface{}) {
	DefaultResponder.WithSuccess(ctx, log, w, r, httpStatus, v)
}
