package web

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type Shutdowner interface {
	Name() string
	Shutdown(context.Context) error
}

type ShutdownerFunc func(context.Context) error

func (s ShutdownerFunc) Shutdown(ctx context.Context) error {
	return s(ctx)
}

// Shutdown gracefully shuts down an HTTP server and app.
func (a *App) Shutdown(log *zerolog.Logger, sig chan os.Signal, stopped chan bool, done chan bool, timeout time.Duration) {
	// We're waiting for either of these signals to fire before exiting, but the behavior
	// is exactly the same afterwards.
	select {
	case v := <-sig:
		log.Info().Msgf("signal received: %s", v)
	case <-stopped:
		log.Info().Msgf("stop signal received")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(len(a.sd))

	for _, v := range a.sd {
		go func(ctx context.Context, s Shutdowner) {
			if err := s.Shutdown(ctx); err != nil {
				log.Error().
					Err(err).
					Type("type", s).
					Str("name", s.Name()).
					Msgf("shutdown failed")
			}

			wg.Done()
		}(ctx, v)
	}

	wg.Wait()

	done <- true
}
