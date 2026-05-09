package middleware

import (
	"fmt"
	"net/http"
	rdebug "runtime/debug"
	"sync/atomic"
	"time"

	"github.com/jimmysawczuk/kit/web/respond"
	"github.com/rs/zerolog"
)

// WithDefaultTimeout wraps WithTimeout with a timeout of 15 seconds.
var WithDefaultTimeout = WithTimeout(15 * time.Second)

// WithTimeout ensures that the provided handler completes within a set amount of time. If it doesn't, it'll
// write a 503 to the client. It *does not* prevent the handler from completing, but silently swallows any
// additional output.
func WithTimeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			log := zerolog.Ctx(ctx)

			// http.ResponseWriters don't like when we try to read/write from the header, or call Write after
			// the connection closes, so we'll wrap our actual writer with something that can swallow any writes
			// that are made after the timeout writer interrupts.
			tw := &timeoutWriter{ResponseWriter: w}

			doneCh := make(chan bool, 1)
			panicCh := make(chan error, 1)
			go func() {
				// If our handler panics, print the stack from here so we get a nice clean stack trace. Then
				// write an Internal Server Error to the ResponseWriter.
				defer func() {
					if p := recover(); p != nil {
						err := fmt.Errorf("panic: %v", p)
						log.Error().Err(err).Msg("with timeout: recovered from panic")
						rdebug.PrintStack()
						panicCh <- err
					}
				}()

				h.ServeHTTP(tw, r)
				doneCh <- true
			}()

			select {
			case <-doneCh:
				tw.done.Store(true)
			case <-time.After(timeout):
				respond.Error(ctx, http.StatusGatewayTimeout, fmt.Errorf("timed out after %s", timeout)).Write(w)
				tw.done.Store(true)
			case perr := <-panicCh:
				respond.Error(ctx, http.StatusInternalServerError, perr).Write(w)
				tw.done.Store(true)
			}
		})
	}
}

// timeoutWriter wraps http.ResponseWriter with a switch that can activate if the handler times out. If this
// switch is activated, we'll swallow any further calls to the underlying ResponseWriter.
type timeoutWriter struct {
	http.ResponseWriter

	done atomic.Bool
}

// Header implements http.ResponseWriter. If we haven't timed out yet, proxy this to the normal ResponseWriter.
// Otherwise, fake it.
func (tw *timeoutWriter) Header() http.Header {
	if tw.done.Load() {
		return map[string][]string{}
	}

	return tw.ResponseWriter.Header()
}

// WriteHeader implements http.ResponseWriter. If we haven't timed out yet, proxy this to the normal ResponseWriter.
// Otherwise, fake it.
func (tw *timeoutWriter) WriteHeader(status int) {
	if tw.done.Load() {
		return
	}

	tw.ResponseWriter.WriteHeader(status)
}

// Write implements http.ResponseWriter. If we haven't timed out yet, proxy this to the normal ResponseWriter.
// Otherwise, this is a no-op.
func (tw *timeoutWriter) Write(b []byte) (int, error) {
	if tw.done.Load() {
		return 0, nil
	}

	return tw.ResponseWriter.Write(b)
}
