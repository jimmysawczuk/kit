package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/jimmysawczuk/kit/web"
	"go.uber.org/zap"
)

type loggableResponseWriter struct {
	http.ResponseWriter

	start        time.Time
	contentType  string
	statusCode   int
	bytesWritten int
}

// LogRequest adds logging about how long the request took to execute.
func LogRequest(h web.Handler) web.Handler {
	return func(ctx context.Context, log *zap.Logger, w http.ResponseWriter, r *http.Request) {
		lrw := &loggableResponseWriter{
			ResponseWriter: w,
			start:          time.Now(),
			statusCode:     http.StatusOK, // default this to 200, because that's what the stdlib does
		}

		log = log.With(
			zap.Time("started", lrw.start),
			zap.String("path", r.URL.Path),
		)

		ctx = context.WithValue(ctx, StartTimeKey, lrw.start)

		log.Info("request: started")
		h(ctx, log, lrw, r)
		log.With(
			zap.Duration("dur", time.Now().Sub(lrw.start)),
			zap.Int("bytesWritten", lrw.bytesWritten),
			zap.Int("status", lrw.statusCode),
			zap.String("statusText", http.StatusText(lrw.statusCode)),
			zap.String("contentType", lrw.contentType),
		).Info("request: finished")
	}
}

func (l *loggableResponseWriter) Write(b []byte) (int, error) {
	written, err := l.ResponseWriter.Write(b)
	l.bytesWritten += written
	return written, err
}

func (l *loggableResponseWriter) Header() http.Header {
	return l.ResponseWriter.Header()
}

func (l *loggableResponseWriter) WriteHeader(code int) {
	l.ResponseWriter.WriteHeader(code)
	l.contentType = l.Header().Get("Content-Type")
	l.statusCode = code
}
