package respond

import (
	"context"
	"net/http"
)

type Responder interface {
	Error(context.Context, int, error) Response
	CodedError(context.Context, int, string, error) Response
	Success(context.Context, int, any) Response
}

type Response interface {
	Write(w http.ResponseWriter) error
}

// DefaultResponder is a default Responder.
var DefaultResponder Responder = &JSONResponder{}

// Error is a shortcut for DefaultResponder.Error.
func Error(ctx context.Context, httpStatus int, err error) Response {
	return DefaultResponder.Error(ctx, httpStatus, err)
}

// CodedError is a shortcut for DefaultResponder.CodedError.
func CodedError(ctx context.Context, httpStatus int, code string, err error) Response {
	return DefaultResponder.CodedError(ctx, httpStatus, code, err)
}

// Success is a shortcut for DefaultResponder.Success.
func Success(ctx context.Context, httpStatus int, v interface{}) Response {
	return DefaultResponder.Success(ctx, httpStatus, v)
}
