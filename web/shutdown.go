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

type namedShutdownFunc struct {
	fn   ShutdownerFunc
	name string
}

var _ Shutdowner = namedShutdownFunc{}

// Name implements Shutdowner.
func (n namedShutdownFunc) Name() string {
	return n.name
}

// Shutdown implements Shutdowner.
func (n namedShutdownFunc) Shutdown(ctx context.Context) error {
	return n.fn(ctx)
}

func NamedShutdownFunc(name string, fn ShutdownerFunc) Shutdowner {
	return namedShutdownFunc{
		name: name,
		fn:   fn,
	}
}

// Shutdown gracefully executes the provided Shutdowners in parallel. It will log any errors that are returned.
func Shutdown(timeout time.Duration, log *zerolog.Logger, sig chan os.Signal, stopped chan bool, done chan error, sd ...Shutdowner) {
	select {
	case v := <-sig:
		log.Info().Msgf("signal received: %s", v)
	case <-stopped:
		log.Info().Msg("stop signal received")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(len(sd))

	for _, v := range sd {
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

	wgDone := make(chan struct{})

	go func() {
		wg.Wait()
		close(wgDone)
	}()

	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			done <- context.DeadlineExceeded
			return
		}
	case <-wgDone:
	}

	done <- nil
}
