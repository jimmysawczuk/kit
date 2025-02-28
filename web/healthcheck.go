package web

import (
	"context"
	"net/http"
	"time"

	"github.com/jimmysawczuk/kit/web/respond"
	"github.com/rs/zerolog/log"
)

type HealthChecker interface {
	Name() string
	HealthCheck(context.Context) error
}

type HealthCheckerFunc func(context.Context) error

func (h HealthCheckerFunc) HealthCheck(ctx context.Context) error {
	return h(ctx)
}

// Health is a Handler checks the health of the App, emitting a 503 if not healthy.
func (a *App) Health(dur time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), dur)
		defer cancel()

		healthy := true
		results := map[string]bool{}

		for _, f := range a.hc {
			err := f.HealthCheck(ctx)
			if err != nil {
				log.Error().
					Err(err).
					Type("type", f).
					Str("name", f.Name()).
					Msgf("health check failed")
			}

			res := err == nil
			healthy = healthy && res
			results[f.Name()] = res
		}

		code := http.StatusOK
		if !healthy {
			code = http.StatusServiceUnavailable
		}

		respond.Success(r.Context(), code, struct {
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
	})
}
