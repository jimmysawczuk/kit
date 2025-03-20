package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type loggableResponseWriter struct {
	http.ResponseWriter

	start time.Time
	end   time.Time

	bytesWritten int
	contentType  string
	status       int
}

// ProfileRequest adds logging about how long the request took to execute.
func ProfileRequest(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := &loggableResponseWriter{
			ResponseWriter: w,
			start:          time.Now(),
		}

		log := zerolog.Ctx(r.Context())
		log.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Time("@req.start", lrw.start)
		})

		log.Info().Msg("request: started")

		h.ServeHTTP(lrw, r)

		lrw.end = time.Now()

		log.Info().
			Time("@req.end", lrw.end).
			Dur("@req.dur", lrw.end.Sub(lrw.start)).
			Str("@resp.type", lrw.contentType).
			Int("@resp.size", lrw.bytesWritten).
			Int("@resp.status", lrw.status).
			Msg("request: finished")
	})
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
	l.status = code
}
