package routes

import (
	"net/http"

	"github.com/NSACodeGov/CodeGov/api/handlers"
	"github.com/NSACodeGov/CodeGov/api/middleware"
	"github.com/NSACodeGov/CodeGov/internal/health"
	"github.com/NSACodeGov/CodeGov/internal/logging"
)

// Config holds route configuration
type Config struct {
	Logger             *logging.Logger
	HealthChecker      *health.Checker
	ClearanceConfig    *middleware.ClearanceConfig
}

// Setup configures all HTTP routes
func Setup(config *Config) http.Handler {
	mux := http.NewServeMux()

	// Health endpoints (no auth required)
	mux.HandleFunc("/healthz", config.HealthChecker.LivenessHandler())
	mux.HandleFunc("/readyz", config.HealthChecker.ReadinessHandler())

	// Root endpoint (no auth required)
	mux.HandleFunc("/", rootHandler(config.Logger))

	// Public API endpoints
	mux.HandleFunc("/api/public", handlers.PublicHandler(config.Logger))

	// Protected API endpoints (require clearance)
	mux.HandleFunc("/api/restricted", handlers.RestrictedHandler(config.Logger))
	mux.HandleFunc("/api/device-only", handlers.DeviceOnlyHandler(config.Logger))
	mux.HandleFunc("/api/device/status", handlers.DeviceStatusHandler(config.Logger))
	mux.HandleFunc("/api/high-security", handlers.HighSecurityHandler(config.Logger))

	// Apply middleware chain
	middlewares := []func(http.Handler) http.Handler{
		middleware.RequestID,
		middleware.Recovery(config.Logger),
		middleware.Logging(config.Logger),
	}

	// Add clearance middleware if configured
	if config.ClearanceConfig != nil && config.ClearanceConfig.Enabled {
		middlewares = append(middlewares, middleware.Clearance(config.ClearanceConfig))
	}

	handler := middleware.Chain(middlewares...)(mux)

	return handler
}

// rootHandler returns a simple root handler
func rootHandler(logger *logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"service":"gogovcode","status":"running","phase":"2"}`))
	}
}
