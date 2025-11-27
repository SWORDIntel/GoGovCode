package main

import (
	"context"
	"fmt"
	"os"

	"github.com/NSACodeGov/CodeGov/api/routes"
	"github.com/NSACodeGov/CodeGov/config"
	"github.com/NSACodeGov/CodeGov/internal/health"
	"github.com/NSACodeGov/CodeGov/internal/logging"
	"github.com/NSACodeGov/CodeGov/internal/server"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Initialize logger
	logger := logging.New(
		cfg.Service.Name,
		cfg.Service.Version,
		cfg.Logging.Level,
		cfg.Logging.Format,
	)

	logger.Info("initializing gogovcode", map[string]interface{}{
		"version": cfg.Service.Version,
		"profile": cfg.Profile,
	})

	// Initialize health checker
	healthChecker := health.New(cfg.Service.Name, cfg.Service.Version)

	// Register health checks
	healthChecker.RegisterCheck("redis", health.RedisCheck(cfg.Redis.Endpoint, cfg.Redis.Enabled), false)
	healthChecker.RegisterCheck("minio", health.MinIOCheck(cfg.MinIO.Endpoint, cfg.MinIO.Enabled), false)

	// Setup routes
	handler := routes.Setup(logger, healthChecker)

	// Create and start server
	srv := server.New(cfg, logger, healthChecker)
	srv.SetHandler(handler)

	logger.Info("starting server", map[string]interface{}{
		"address": cfg.Addr(),
		"tls":     cfg.TLS.Enabled,
	})

	// Start server (blocks until shutdown)
	if err := srv.Start(context.Background()); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
