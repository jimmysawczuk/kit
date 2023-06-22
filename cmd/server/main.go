package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jimmysawczuk/kit/web"
	"golang.org/x/exp/slog"
)

const port = 3000

func main() {
	log := slog.Default()

	app := web.NewApp(slog.Default())

	// Attach the app and the cert to a server, setting some sensible default timeout values.
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           app,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      120 * time.Second,
		// TLSConfig: &tls.Config{
		// 	Certificates: []tls.Certificate{cert},
		// },
	}

	// Setup the shutdown handler.
	done := make(chan bool)
	stopped := make(chan bool, 1)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM)
	signal.Notify(sig, syscall.SIGINT)

	var shutdowns []web.Shutdowner = []web.Shutdowner{
		web.ShutdownCtxFunc(srv.Shutdown),
		web.ShutdownCtxFunc(app.Shutdown),
	}

	go web.Shutdown(log, sig, stopped, done, 30*time.Second, shutdowns...)

	// Start the server!
	go func() {
		log.With(
			"host", "127.0.0.1",
			"port", port,
			"url", fmt.Sprintf("http://%s:%d", "127.0.0.1", port),
		).Info("starting server")

		err := srv.ListenAndServe()

		// The previous call blocks until the server is terminated, so we know when we get here, the server
		// was stopped.
		log.With("error", err).Info("stopped")
		stopped <- true
	}()
	<-done
	log.Info("terminating")
}
