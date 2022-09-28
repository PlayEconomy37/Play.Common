package common

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *App) Serve(router http.Handler, certFile, keyFile string) error {
	// Initialize a tls.Config struct to hold the non-default TLS settings we want
	// the server to use
	tlsConfig := &tls.Config{
		// Go’s favored cipher suites are given preference and we help increase the likelihood that
		// a strong cipher suite which also supports forward secrecy is used
		PreferServerCipherSuites: true,

		// Specify which elliptic curves should be given preference during the TLS handshake.
		// Go supports a few elliptic curves, but as of Go 1.11 only tls.CurveP256 and tls.X25519
		// have assembly implementations. The others are very CPU intensive, so omitting them helps
		// ensure that our server will remain performant under heavy loads.
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	// Declare a HTTP server
	server := &http.Server{
		Addr:         app.Config.Address,
		Handler:      router,
		ErrorLog:     log.New(app.Logger, "", 0), // The "" and 0 indicate that the logger.Logger instance should not use a prefix or any flags
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		TLSConfig:    tlsConfig,
	}

	// Create a shutdownError channel. We will use this to receive any errors returned
	// by the graceful shutdown function.
	shutdownError := make(chan error)

	// Start a background goroutine to catch SIGINT and SIGTERM signals
	go func() {
		// Create a quit channel which carries os.Signal values
		quit := make(chan os.Signal, 1)

		// Use signal.Notify() to listen for incoming SIGINT and SIGTERM signals and
		// relay them to the quit channel. Any other signals will not be caught by
		// signal.Notify() and will retain their default behavior.
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// Read the signal from the quit channel. This code will block until a signal is
		// received.
		sig := <-quit

		// Log a message to say that the signal has been caught. Notice that we also
		// call the String() method on the signal to get the signal name and include it
		// in the log entry properties.
		app.Logger.Info("Shutting down server", map[string]string{
			"signal": sig.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Call Shutdown() on our server, passing in the context we just made.
		// Shutdown() will return nil if the graceful shutdown was successful, or an
		// error (which may happen because of a problem closing the listeners, or
		// because the shutdown didn't complete before the 5-second context deadline is
		// hit). We relay this return value to the shutdownError channel.
		err := server.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		// Log a message to say that we're waiting for any background goroutines to
		// complete their tasks.
		app.Logger.Info("Completing background tasks", map[string]string{
			"addr": server.Addr,
		})

		// Call Wait() to block until our WaitGroup counter is zero --- essentially
		// blocking until the background goroutines have finished. Then we return nil on
		// the shutdownError channel, to indicate that the shutdown completed without
		// any issues.
		app.WaitGroup.Wait()
		shutdownError <- nil
	}()

	app.Logger.Info("Starting server", map[string]string{
		"addr": server.Addr,
	})

	// Calling Shutdown() on our server will cause ListenAndServe() to immediately
	// return a http.ErrServerClosed error. So if we see this error, it is actually a
	// good thing and an indication that the graceful shutdown has started. So we check
	// specifically for this, only returning the error if it is NOT http.ErrServerClosed.
	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Otherwise, we wait to receive the return value from Shutdown() on the
	// shutdownError channel. If return value is an error, we know that there was a
	// problem with the graceful shutdown and we return the error.
	err = <-shutdownError
	if err != nil {
		return err
	}

	// At this point we know that the graceful shutdown completed successfully and we
	// log a "stopped server" message
	app.Logger.Info("Stopped server", map[string]string{
		"addr": server.Addr,
	})

	return nil
}
