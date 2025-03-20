package web_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jimmysawczuk/kit/web"
	"github.com/stretchr/testify/require"
)

func TestHealthCheckHandler(t *testing.T) {
	alwaysHealthy := web.NamedHealthCheckFunc("healthy", func(ctx context.Context) error { return nil })
	alwaysUnhealthy := web.NamedHealthCheckFunc("unhealthy", func(ctx context.Context) error { return fmt.Errorf("unhealthy") })
	slowHealthy := web.NamedHealthCheckFunc("slowHealthy", func(ctx context.Context) error { time.Sleep(400 * time.Millisecond); return nil })

	tests := []struct {
		Name     string
		Expected bool
		HCs      []web.HealthChecker
		Timeout  time.Duration
	}{
		{
			Name:     "HEALTHY",
			Expected: true,
			HCs:      []web.HealthChecker{alwaysHealthy},
			Timeout:  1 * time.Second,
		},
		{
			Name:     "UNHEALTHY",
			Expected: false,
			HCs:      []web.HealthChecker{alwaysUnhealthy},
			Timeout:  1 * time.Second,
		},
		{
			Name:     "TIMEOUT",
			Expected: false,
			HCs:      []web.HealthChecker{slowHealthy},
			Timeout:  300 * time.Millisecond,
		},
		{
			Name:     "HEALTHY_LONGER_TIMEOUT",
			Expected: true,
			HCs:      []web.HealthChecker{slowHealthy},
			Timeout:  500 * time.Millisecond,
		},

		{
			Name:     "MULTIPLE_HEALTHYS",
			Expected: true,
			HCs:      []web.HealthChecker{alwaysHealthy, slowHealthy},
			Timeout:  500 * time.Millisecond,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			hdlr := chi.Chain(func(next http.Handler) http.Handler {
				fn := func(w http.ResponseWriter, r *http.Request) {
					ctx, cancel := context.WithTimeout(r.Context(), test.Timeout)
					defer cancel()

					r = r.WithContext(ctx)
					next.ServeHTTP(w, r)
				}
				return http.HandlerFunc(fn)
			}).Handler(web.HealthCheckHandler(test.HCs...))

			srv := httptest.NewServer(hdlr)

			resp, err := http.Get(srv.URL)
			require.NoError(t, err)

			defer resp.Body.Close()

			var target struct {
				Healthy bool            `json:"healthy"`
				Results map[string]bool `json:"results"`
			}
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&target))
			require.Equal(t, test.Expected, target.Healthy)
		})
	}
}
