package web_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jimmysawczuk/kit/web"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestShutdown(t *testing.T) {
	shutdowner := web.NamedShutdownFunc("shutdowner", func(ctx context.Context) error { return nil })
	slowShutdowner := web.NamedShutdownFunc("slowShutdowner", func(ctx context.Context) error { time.Sleep(500 * time.Millisecond); return nil })
	_ = slowShutdowner

	tests := []struct {
		Name        string
		Shutdowners []web.Shutdowner
		Timeout     time.Duration
		ErrExpected error
	}{
		{
			Name:        "HAPPY_PATH",
			Shutdowners: []web.Shutdowner{shutdowner},
			Timeout:     1 * time.Second,
			ErrExpected: nil,
		},

		{
			Name:        "TWO_SHUTDOWNS",
			Shutdowners: []web.Shutdowner{shutdowner, slowShutdowner},
			Timeout:     1 * time.Second,
			ErrExpected: nil,
		},

		{
			Name:        "TIMEOUT",
			Shutdowners: []web.Shutdowner{slowShutdowner},
			Timeout:     100 * time.Millisecond,
			ErrExpected: context.DeadlineExceeded,
		},
	}

	log := zerolog.New(os.Stderr)
	sig := make(chan os.Signal)
	stop := make(chan bool)
	done := make(chan error)

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), test.Timeout)
			defer cancel()

			go web.Shutdown(ctx, &log, sig, stop, done, test.Shutdowners...)

			stop <- true

			err := <-done
			if test.ErrExpected == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, test.ErrExpected)
			}
		})
	}
}
