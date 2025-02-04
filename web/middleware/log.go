package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/jimmysawczuk/kit/web"
	"github.com/sirupsen/logrus"
)

type loggableResponseWriter struct {
	http.ResponseWriter

	start        time.Time
	contentType  string
	statusCode   int
	bytesWritten int
}

// InspectRequest adds logging about how long the request took to execute, the amount of bytes written and the
// content type.
func InspectRequest(h web.Handler) web.Handler {
	return func(ctx context.Context, log logrus.FieldLogger, w http.ResponseWriter, r *http.Request) {
		lrw := &loggableResponseWriter{
			ResponseWriter: w,
			start:          time.Now(),
			statusCode:     http.StatusOK, // default this to 200, because that's what the stdlib does
		}

		log = log.WithField("started", lrw.start)
		log = log.WithField("path", r.URL.Path)

		log.Info("request: started")
		h(ctx, log, lrw, r)
		log.WithFields(logrus.Fields{
			"dur":          time.Since(lrw.start).String(),
			"ns":           time.Since(lrw.start),
			"bytesWritten": lrw.bytesWritten,
			"status":       lrw.statusCode,
			"statusText":   http.StatusText(lrw.statusCode),
			"contentType":  lrw.contentType,
		}).Info("request: finished")
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
