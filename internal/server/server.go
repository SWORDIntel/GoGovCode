package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NSACodeGov/CodeGov/config"
	"github.com/NSACodeGov/CodeGov/internal/health"
	"github.com/NSACodeGov/CodeGov/internal/logging"
)

// Server represents the HTTP server
type Server struct {
	config  *config.Config
	logger  *logging.Logger
	health  *health.Checker
	handler http.Handler
	server  *http.Server
}

// New creates a new server instance
func New(cfg *config.Config, logger *logging.Logger, healthChecker *health.Checker) *Server {
	return &Server{
		config: cfg,
		logger: logger,
		health: healthChecker,
	}
}

// SetHandler sets the HTTP handler
func (s *Server) SetHandler(h http.Handler) {
	s.handler = h
}

// Start starts the HTTP server with graceful shutdown
func (s *Server) Start(ctx context.Context) error {
	// Create HTTP server
	s.server = &http.Server{
		Addr:         s.config.Addr(),
		Handler:      s.handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Configure TLS if enabled
	if s.config.TLS.Enabled {
		cert, err := tls.LoadX509KeyPair(s.config.TLS.CertFile, s.config.TLS.KeyFile)
		if err != nil {
			return fmt.Errorf("failed to load TLS certificates: %w", err)
		}

		s.server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			},
		}
	}

	// Channel to listen for errors from the server
	serverErrors := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		s.logger.Info("starting server", map[string]interface{}{
			"addr":       s.config.Addr(),
			"tls":        s.config.TLS.Enabled,
			"profile":    s.config.Profile,
		})

		if s.config.TLS.Enabled {
			serverErrors <- s.server.ListenAndServeTLS("", "")
		} else {
			serverErrors <- s.server.ListenAndServe()
		}
	}()

	// Create channel to listen for interrupt signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or an error
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		s.logger.Info("shutdown signal received", map[string]interface{}{
			"signal": sig.String(),
		})

		// Give outstanding requests a deadline for completion
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Ask the server to shutdown gracefully
		if err := s.server.Shutdown(ctx); err != nil {
			s.logger.Error("graceful shutdown failed", map[string]interface{}{
				"error": err.Error(),
			})

			// Force close if graceful shutdown fails
			if err := s.server.Close(); err != nil {
				return fmt.Errorf("failed to close server: %w", err)
			}
		}

		s.logger.Info("server stopped")
	}

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	s.logger.Info("shutting down server")

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	return nil
}
