package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/NSACodeGov/CodeGov/api/middleware"
	"github.com/NSACodeGov/CodeGov/api/routes"
	"github.com/NSACodeGov/CodeGov/config"
	"github.com/NSACodeGov/CodeGov/internal/audit"
	"github.com/NSACodeGov/CodeGov/internal/health"
	"github.com/NSACodeGov/CodeGov/internal/logging"
	"github.com/NSACodeGov/CodeGov/internal/policy"
	"github.com/NSACodeGov/CodeGov/internal/server"
	"github.com/NSACodeGov/CodeGov/pkg/models"
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

	// Initialize device registry
	deviceRegistry := models.NewDeviceRegistry()

	// Register example devices for testing
	registerExampleDevices(deviceRegistry, logger)

	// Initialize audit logger
	auditLogger := audit.NewLogger()
	auditLogger.AddWriter(audit.NewStdoutWriter())

	// Initialize policy engine
	policyEngine := policy.NewEngine(deviceRegistry)

	// Load default policy (or from file if specified)
	loadDefaultPolicy(policyEngine, logger)

	// Initialize health checker
	healthChecker := health.New(cfg.Service.Name, cfg.Service.Version)

	// Register health checks
	healthChecker.RegisterCheck("redis", health.RedisCheck(cfg.Redis.Endpoint, cfg.Redis.Enabled), false)
	healthChecker.RegisterCheck("minio", health.MinIOCheck(cfg.MinIO.Endpoint, cfg.MinIO.Enabled), false)

	// Configure clearance middleware
	clearanceConfig := &middleware.ClearanceConfig{
		PolicyEngine:   policyEngine,
		AuditLogger:    auditLogger,
		Logger:         logger,
		DeviceRegistry: deviceRegistry,
		Enabled:        true, // Enable clearance enforcement
	}

	// Setup routes
	routeConfig := &routes.Config{
		Logger:          logger,
		HealthChecker:   healthChecker,
		ClearanceConfig: clearanceConfig,
	}
	handler := routes.Setup(routeConfig)

	// Create and start server
	srv := server.New(cfg, logger, healthChecker)
	srv.SetHandler(handler)

	logger.Info("starting server", map[string]interface{}{
		"address": cfg.Addr(),
		"tls":     cfg.TLS.Enabled,
		"phase":   "2",
	})

	// Start server (blocks until shutdown)
	if err := srv.Start(context.Background()); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	// Cleanup
	auditLogger.Close()

	return nil
}

// registerExampleDevices registers example devices for testing
func registerExampleDevices(registry *models.DeviceRegistry, logger *logging.Logger) {
	devices := []*models.Device{
		{
			ID:        1,
			Name:      "sensor-001",
			Layer:     models.LayerData,
			Class:     models.DeviceClassSensor,
			Clearance: models.ClearanceLevel3,
		},
		{
			ID:        2,
			Name:      "gateway-001",
			Layer:     models.LayerTransport,
			Class:     models.DeviceClassGateway,
			Clearance: models.ClearanceLevel5,
		},
		{
			ID:        3,
			Name:      "controller-001",
			Layer:     models.LayerControl,
			Class:     models.DeviceClassController,
			Clearance: models.ClearanceLevel7,
		},
		{
			ID:        4,
			Name:      "app-server-001",
			Layer:     models.LayerApplication,
			Class:     models.DeviceClassController,
			Clearance: models.ClearanceLevel9,
		},
	}

	for _, device := range devices {
		if err := registry.Register(device); err != nil {
			logger.Error("failed to register device", map[string]interface{}{
				"device": device.Name,
				"error":  err.Error(),
			})
		} else {
			logger.Info("registered device", map[string]interface{}{
				"device_id": device.ID,
				"name":      device.Name,
				"layer":     device.Layer,
				"clearance": device.Clearance.String(),
			})
		}
	}
}

// loadDefaultPolicy loads a default policy for testing
func loadDefaultPolicy(engine *policy.Engine, logger *logging.Logger) {
	defaultPolicy := &policy.Policy{
		Version: "1.0",
		Rules: []*policy.Rule{
			{
				ID:       "allow-public",
				Name:     "Allow public endpoints",
				Effect:   policy.EffectAllow,
				Routes:   []string{"/", "/healthz", "/readyz", "/api/public"},
				Methods:  []string{"*"},
				Priority: 100,
			},
			{
				ID:                "allow-restricted",
				Name:              "Allow restricted with clearance level 3+",
				Effect:            policy.EffectAllow,
				Routes:            []string{"/api/restricted"},
				Methods:           []string{"GET", "POST"},
				RequiredClearance: models.ClearanceLevel3,
				Priority:          50,
			},
			{
				ID:                "allow-device-only",
				Name:              "Allow device endpoints for registered devices",
				Effect:            policy.EffectAllow,
				Routes:            []string{"/api/device-only", "/api/device/status"},
				Methods:           []string{"GET"},
				RequiredClearance: models.ClearanceLevel3,
				AllowedDevices:    []uint16{1, 2, 3, 4},
				Priority:          60,
			},
			{
				ID:                "allow-high-security",
				Name:              "Allow high security endpoints for level 7+",
				Effect:            policy.EffectAllow,
				Routes:            []string{"/api/high-security"},
				Methods:           []string{"GET", "POST"},
				RequiredClearance: models.ClearanceLevel7,
				Priority:          70,
			},
			{
				ID:       "deny-default",
				Name:     "Deny all other requests",
				Effect:   policy.EffectDeny,
				Routes:   []string{"*"},
				Methods:  []string{"*"},
				Priority: 0,
			},
		},
	}

	if err := engine.Validate(defaultPolicy); err != nil {
		logger.Error("failed to validate default policy", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	data, _ := json.Marshal(defaultPolicy)
	if err := engine.LoadFromJSON(data); err != nil {
		logger.Error("failed to load default policy", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		logger.Info("loaded default policy", map[string]interface{}{
			"rules": len(defaultPolicy.Rules),
		})
	}
}
