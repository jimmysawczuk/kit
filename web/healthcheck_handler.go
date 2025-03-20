package web

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/jimmysawczuk/kit/web/respond"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Health is a Handler checks the health of the App by checking all of the app's HealthCheckers sequentially,
// emitting a 503 if any return an error or the check cannot complete within the specified duration. If timeout <= 0,
// there is no timeout.
func HealthCheckHandler(hc ...HealthChecker) Handler {
	return func(ctx context.Context, l *zerolog.Logger, w http.ResponseWriter, r *http.Request) {
		m := sync.Map{}

		wg := sync.WaitGroup{}
		wg.Add(len(hc))

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		for _, h := range hc {
			go func() {
				err := h.HealthCheck(ctx)
				if err != nil {
					log.Error().
						Err(err).
						Type("type", h).
						Str("name", h.Name()).
						Msgf("health check failed")
				}

				m.Store(h.Name(), err)

				wg.Done()
			}()
		}

		wgDone := make(chan struct{})

		go func() {
			wg.Wait()
			wgDone <- struct{}{}
		}()

		healthy := true

		select {
		case <-ctx.Done():
		case <-wgDone:
		}

		if ctx.Err() == context.DeadlineExceeded {
			healthy = false
		}

		results := map[string]bool{}
		m.Range(func(key, value any) bool {
			name := key.(string)
			err, _ := value.(error)

			results[name] = err == nil
			healthy = healthy && (err == nil)
			return true
		})

		code := http.StatusOK
		if !healthy {
			code = http.StatusServiceUnavailable
		}

		respond.Success(ctx, code, struct {
			Healthy bool            `json:"healthy"`
			Code    int             `json:"code"`
			Time    time.Time       `json:"time"`
			Results map[string]bool `json:"results"`
		}{
			Healthy: healthy,
			Code:    code,
			Time:    time.Now().UTC(),
			Results: results,
		}).Write(w)
	}
}
