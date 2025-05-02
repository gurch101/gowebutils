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
	go gracefulShutdown(server, logger, shutdownError)

	slog.Info("server started", "port", port)

	serverErr := startServer(server)

	if serverErr != nil && !errors.Is(serverErr, http.ErrServerClosed) {
		return fmt.Errorf("server error %w", serverErr)
	}

	if err := <-shutdownError; err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	slog.Info("server stopped", "port", port)

	return nil
}

func gracefulShutdown(server *http.Server, logger *slog.Logger, shutdownErr chan<- error) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	s := <-quit

	logger.Info("shutting down server", "signal", s.String())

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	shutdownErr <- server.Shutdown(ctx)
}

func startServer(server *http.Server) error {
	if _, err := os.Stat("./tls/cert.pem"); err == nil {
		//nolint: wrapcheck
		return server.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	}

	//nolint: wrapcheck
	return server.ListenAndServe()
}
