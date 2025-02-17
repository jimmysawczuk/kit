package respond

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jimmysawczuk/kit/web/v2/requestid"
	"github.com/rs/zerolog"
)

type Responder interface {
	WithError(context.Context, http.ResponseWriter, *http.Request, int, error)
	WithCodedError(context.Context, http.ResponseWriter, *http.Request, string, int, error)
	WithSuccess(context.Context, http.ResponseWriter, *http.Request, int, any)
}

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
type JSONResponder struct {
	SuppressErrors bool
}

// DefaultResponder is a default Responder.
var DefaultResponder = &JSONResponder{}

// WithError is a shortcut for WithCodedError(ctx, log, w, httpStatus, "", err).
func (jr *JSONResponder) WithError(ctx context.Context, w http.ResponseWriter, r *http.Request, httpStatus int, err error) {
	jr.WithCodedError(ctx, w, r, httpStatus, "", err)
}

// WithCodedError writes the provided error to the ResponseWriter, as well as the HTTP status code.
// An enum-style code (i.e. INVALID_TOKEN) may also be provided.
func (jr *JSONResponder) WithCodedError(ctx context.Context, w http.ResponseWriter, r *http.Request, httpStatus int, code string, err error) {
	serr := http.StatusText(httpStatus)
	if err != nil {
		serr = err.Error()
	}

	reqID := requestid.Get(r.Context())
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

	var ei ErrorInfoer
	if errors.As(err, &ei) {
		resp.Info = ei.ErrorInfo()
	}

	if jr.SuppressErrors {
		log := zerolog.Ctx(ctx)

		msg := log.Error().
			Err(err).
			Int("statusCode", httpStatus)

		if resp.Info != nil {
			msg = msg.Any("info", resp.Info)
		}

		resp.Error = ""

		msg.Msg("error suppressed")
	}

	json.NewEncoder(w).Encode(resp)
}

// WithSuccess writes the provided response to the ResponseWriter (unwrapped) and sets the provided HTTP response status.
func (jr *JSONResponder) WithSuccess(ctx context.Context, w http.ResponseWriter, r *http.Request, httpStatus int, v interface{}) {
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
func WithError(ctx context.Context, w http.ResponseWriter, r *http.Request, httpStatus int, err error) {
	DefaultResponder.WithError(ctx, w, r, httpStatus, err)
}

// WithCodedError is a shortcut for DefaultResponder.WithCodedError.
func WithCodedError(ctx context.Context, w http.ResponseWriter, r *http.Request, httpStatus int, code string, err error) {
	DefaultResponder.WithCodedError(ctx, w, r, httpStatus, code, err)
}

// WithSuccess is a shortcut for DefaultResponder.WithSuccess.
func WithSuccess(ctx context.Context, w http.ResponseWriter, r *http.Request, httpStatus int, v interface{}) {
	DefaultResponder.WithSuccess(ctx, w, r, httpStatus, v)
}
