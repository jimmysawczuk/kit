package web

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type ShutdownFunc func() error

func (f ShutdownFunc) Shutdown() error {
	return f()
}

func (f ShutdownFunc) sealed() {}

type ShutdownCtxFunc func(context.Context) error

func (f ShutdownCtxFunc) sealed() {}

type Shutdowner interface {
	sealed()
}

// Shutdown gracefully shuts down an HTTP server and app.
func Shutdown(log *zerolog.Logger, sig chan os.Signal, stopped chan bool, done chan bool, timeout time.Duration, shutdowners ...Shutdowner) {
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
	wg.Add(len(shutdowners))

	for _, v := range shutdowners {
		go func(ctx context.Context, s Shutdowner) {
			switch fn := s.(type) {
			case ShutdownCtxFunc:
				if err := fn(ctx); err != nil {
					log.Error().Err(err).Msgf("could not shutdown %T", fn)
				}

			case ShutdownFunc:
				if err := fn(); err != nil {
					log.Error().Err(err).Msgf("could not shutdown %T", fn)
				}
			}

			wg.Done()
		}(ctx, v)
	}

	wg.Wait()

	done <- true
}
