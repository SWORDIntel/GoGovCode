package routes

import (
	"net/http"

	"github.com/NSACodeGov/CodeGov/api/middleware"
	"github.com/NSACodeGov/CodeGov/internal/health"
	"github.com/NSACodeGov/CodeGov/internal/logging"
)

// Setup configures all HTTP routes
func Setup(logger *logging.Logger, healthChecker *health.Checker) http.Handler {
	mux := http.NewServeMux()

	// Health endpoints
	mux.HandleFunc("/healthz", healthChecker.LivenessHandler())
	mux.HandleFunc("/readyz", healthChecker.ReadinessHandler())

	// Root endpoint
	mux.HandleFunc("/", rootHandler(logger))

	// Apply middleware chain
	handler := middleware.Chain(
		middleware.RequestID,
		middleware.Recovery(logger),
		middleware.Logging(logger),
	)(mux)

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
		w.Write([]byte(`{"service":"gogovcode","status":"running","phase":"1"}`))
	}
}
