package respond

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/jimmysawczuk/kit/web/requestid"
	"github.com/rs/zerolog"
)

// Responder is a set of config for responding to requests.
type JSONResponder struct {
	SuppressErrors bool
}

type JSONResponse struct {
	ctx    context.Context
	status int
	header http.Header
	body   any
}

type ErrorResponse struct {
	Error     string `json:"error,omitempty"`
	RequestID string `json:"requestID,omitempty"`
	ErrorCode string `json:"code,omitempty"`
	Status    int    `json:"status,omitempty"`
	Info      any    `json:"info,omitempty"`
}

// WithError is a shortcut for WithCodedError(ctx, log, w, httpStatus, "", err).
func (jr *JSONResponder) Error(ctx context.Context, httpStatus int, err error) Response {
	return jr.CodedError(ctx, httpStatus, "", err)
}

// WithCodedError writes the provided error to the ResponseWriter, as well as the HTTP status code.
// An enum-style code (i.e. INVALID_TOKEN) may also be provided.
func (jr *JSONResponder) CodedError(ctx context.Context, httpStatus int, code string, err error) Response {
	resp := &JSONResponse{
		ctx:    ctx,
		status: httpStatus,
		header: http.Header{},
	}

	serr := http.StatusText(httpStatus)
	if err != nil {
		serr = err.Error()
	}

	reqID := requestid.Get(ctx)
	if reqID != "" {
		resp.header.Set("X-Request-Id", reqID)
	}

	if ct := resp.header.Get("Content-Type"); ct == "" {
		resp.header.Set("Content-Type", "application/json; charset=utf-8")
	}

	body := ErrorResponse{
		Error:     serr,
		RequestID: reqID,
		Status:    httpStatus,
		ErrorCode: code,
	}

	var ei ErrorInfoer
	if errors.As(err, &ei) {
		body.Info = ei.Info()
	}

	if jr.SuppressErrors {
		log := zerolog.Ctx(ctx)

		msg := log.Error().
			Err(err).
			Int("statusCode", httpStatus)

		if body.Info != nil {
			msg = msg.Any("info", body.Info)
		}

		body.Error = ""

		msg.Msg("error suppressed")
	}

	resp.body = body
	return resp
}

// WithSuccess writes the provided response to the ResponseWriter (unwrapped) and sets the provided HTTP response status.
func (jr *JSONResponder) Success(ctx context.Context, httpStatus int, body any) Response {
	resp := JSONResponse{
		ctx:    ctx,
		header: http.Header{},
		status: httpStatus,
		body:   body,
	}

	if ct := resp.header.Get("Content-Type"); ct == "" {
		resp.header.Set("Content-Type", "application/json; charset=utf-8")
	}

	reqID := requestid.Get(ctx)
	if reqID != "" {
		resp.header.Set("X-Request-Id", reqID)
	}

	return resp
}

func (r JSONResponse) Write(w http.ResponseWriter) error {
	log := zerolog.Ctx(r.ctx)

	for h := range r.header {
		for _, v := range r.header[h] {
			w.Header().Add(h, v)
		}
	}

	by, err := json.Marshal(r.body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		werr := fmt.Errorf("json: marshal: %w", err)

		log.Err(werr).
			Msg("json response: write: couldn't marshal response")

			// ignoring this error intentionally; we can't do anything about it anyway
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			ErrorCode: "JSON_MARSHAL",
			Error:     "failed to marshal json response",
			RequestID: requestid.Get(r.ctx),
			Status:    http.StatusInternalServerError,
		})
		return werr
	}

	w.WriteHeader(r.status)
	if _, err := w.Write(by); err != nil {
		werr := fmt.Errorf("response writer: write: %w", err)
		log.Err(werr).Msg("json response: couldn't write")
		return werr
	}

	return nil
}
