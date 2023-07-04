package middleware

import (
	"context"
	"fmt"
	"net/http"
	rdebug "runtime/debug"
	"sync"
	"time"

	"github.com/jimmysawczuk/kit/web"
	"github.com/jimmysawczuk/kit/web/respond"
	"go.uber.org/zap"
)

// WithDefaultTimeout wraps WithTimeout with a timeout of 15 seconds.
var WithDefaultTimeout = WithTimeout(15 * time.Second)

// WithTimeout ensures that the provided handler completes within a set amount of time. If it doesn't, it'll
// write a 503 to the client. It *does not* prevent the handler from completing, but silently swallows any
// additional output.
func WithTimeout(timeout time.Duration) func(web.Handler) web.Handler {
	return func(h web.Handler) web.Handler {
		return func(ctx context.Context, log *zap.Logger, w http.ResponseWriter, r *http.Request) {
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

						log.With(
							zap.Error(err),
							zap.String("mw", "WithTimeout"),
						).Error("recovered from panic")

						rdebug.PrintStack()
						panicCh <- err
					}
				}()

				h(ctx, log, tw, r)
				doneCh <- true
			}()

			select {
			case <-doneCh:
				tw.mx.Lock()
				tw.done = true
				tw.mx.Unlock()
			case <-time.After(timeout):
				tw.mx.Lock()
				respond.WithError(ctx, log, tw, r, http.StatusGatewayTimeout, fmt.Errorf("timed out after %s", timeout))
				tw.done = true
				tw.mx.Unlock()
			case perr := <-panicCh:
				tw.mx.Lock()
				respond.WithError(ctx, log, tw, r, http.StatusInternalServerError, perr)
				tw.done = true
				tw.mx.Unlock()
			}
		}
	}
}

// timeoutWriter wraps http.ResponseWriter with a switch that can activate if the handler times out. If this
// switch is activated, we'll swallow any further calls to the underlying ResponseWriter.
type timeoutWriter struct {
	http.ResponseWriter

	mx   sync.Mutex
	done bool
}

// Header implements http.ResponseWriter. If we haven't timed out yet, proxy this to the normal ResponseWriter.
// Otherwise, fake it.
func (tw *timeoutWriter) Header() http.Header {
	if tw.done {
		return map[string][]string{}
	}

	return tw.ResponseWriter.Header()
}

// WriteHeader implements http.ResponseWriter. If we haven't timed out yet, proxy this to the normal ResponseWriter.
// Otherwise, fake it.
func (tw *timeoutWriter) WriteHeader(status int) {
	if tw.done {
		return
	}

	tw.ResponseWriter.WriteHeader(status)
}

// Write implements http.ResponseWriter. If we haven't timed out yet, proxy this to the normal ResponseWriter.
// Otherwise, this returns an error saying the time for writing has passed.
func (tw *timeoutWriter) Write(b []byte) (int, error) {
	if tw.done {
		return 0, fmt.Errorf("responsewriter already written")
	}

	return tw.ResponseWriter.Write(b)
}
