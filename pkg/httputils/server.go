package httputils

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gurch101/gowebutils/pkg/parser"
)

const defaultPort = 8080

const readTimeout = 5 * time.Second

const writeTimeout = 10 * time.Second

const shutdownTimeout = 5 * time.Second

func ServeHTTP(handler http.Handler, logger *slog.Logger) error {
	port, err := parser.ParseEnvInt("SERVER_PORT", defaultPort)
	if err != nil {
		return fmt.Errorf("invalid server port: %w", err)
	}

	//nolint: exhaustruct
	tlsConfig := &tls.Config{
		MinVersion:       tls.VersionTLS13,
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	//nolint: exhaustruct
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           handler,
		TLSConfig:         tlsConfig,
		IdleTimeout:       time.Minute,
		ReadHeaderTimeout: readTimeout,
		WriteTimeout:      writeTimeout,
		ErrorLog:          NewSlogErrorWriter(logger),
	}

	shutdownError := make(chan error)

	go func() {
		// Create a quit channel which carries os.Signal values.
		quit := make(chan os.Signal, 1)
		// Use signal.Notify() to listen for incoming SIGINT and SIGTERM signals and relay them to the quit channel.
		// Any other signals will not be caught by signal.Notify() and will retain their default behavior.
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		// Read the signal from the quit channel. This code will block until a signal is received.
		s := <-quit

		slog.Info("shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		shutdownError <- server.Shutdown(ctx)
	}()

	slog.Info("starting server", "port", port)

	err = server.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")

	if !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error %w", err)
	}

	err = <-shutdownError
	if err != nil {
		return fmt.Errorf("server shutdown error %w", err)
	}

	slog.Info("server stopped", "port", port)

	return nil
}
