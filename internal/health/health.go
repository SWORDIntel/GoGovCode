package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Status represents the health status
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// CheckFunc is a function that performs a health check
type CheckFunc func(ctx context.Context) error

// Check represents a single health check
type Check struct {
	Name     string
	Checker  CheckFunc
	Critical bool // If true, failure marks overall status as unhealthy
}

// Response represents a health check response
type Response struct {
	Status    Status              `json:"status"`
	Timestamp string              `json:"timestamp"`
	Service   string              `json:"service"`
	Version   string              `json:"version"`
	Checks    map[string]CheckResult `json:"checks,omitempty"`
}

// CheckResult represents the result of a single check
type CheckResult struct {
	Status    Status `json:"status"`
	Message   string `json:"message,omitempty"`
	Duration  string `json:"duration"`
}

// Checker manages health checks
type Checker struct {
	mu          sync.RWMutex
	checks      map[string]Check
	serviceName string
	serviceVer  string
}

// New creates a new health checker
func New(serviceName, serviceVersion string) *Checker {
	return &Checker{
		checks:      make(map[string]Check),
		serviceName: serviceName,
		serviceVer:  serviceVersion,
	}
}

// RegisterCheck adds a health check
func (c *Checker) RegisterCheck(name string, checker CheckFunc, critical bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.checks[name] = Check{
		Name:     name,
		Checker:  checker,
		Critical: critical,
	}
}

// RunChecks executes all registered health checks
func (c *Checker) RunChecks(ctx context.Context) Response {
	c.mu.RLock()
	checks := make(map[string]Check, len(c.checks))
	for k, v := range c.checks {
		checks[k] = v
	}
	c.mu.RUnlock()

	response := Response{
		Status:    StatusHealthy,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Service:   c.serviceName,
		Version:   c.serviceVer,
		Checks:    make(map[string]CheckResult),
	}

	// Run all checks in parallel
	type result struct {
		name     string
		err      error
		duration time.Duration
	}

	resultCh := make(chan result, len(checks))
	var wg sync.WaitGroup

	for name, check := range checks {
		wg.Add(1)
		go func(n string, ch Check) {
			defer wg.Done()

			start := time.Now()
			checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			err := ch.Checker(checkCtx)
			duration := time.Since(start)

			resultCh <- result{
				name:     n,
				err:      err,
				duration: duration,
			}
		}(name, check)
	}

	// Wait for all checks to complete
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results
	hasDegraded := false
	hasUnhealthy := false

	for res := range resultCh {
		check := checks[res.name]

		checkResult := CheckResult{
			Status:   StatusHealthy,
			Duration: res.duration.String(),
		}

		if res.err != nil {
			checkResult.Message = res.err.Error()

			if check.Critical {
				checkResult.Status = StatusUnhealthy
				hasUnhealthy = true
			} else {
				checkResult.Status = StatusDegraded
				hasDegraded = true
			}
		}

		response.Checks[res.name] = checkResult
	}

	// Determine overall status
	if hasUnhealthy {
		response.Status = StatusUnhealthy
	} else if hasDegraded {
		response.Status = StatusDegraded
	}

	return response
}

// LivenessHandler returns a simple liveness check handler (always returns 200)
func (c *Checker) LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := Response{
			Status:    StatusHealthy,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Service:   c.serviceName,
			Version:   c.serviceVer,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// ReadinessHandler returns a readiness check handler
func (c *Checker) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := c.RunChecks(r.Context())

		w.Header().Set("Content-Type", "application/json")

		statusCode := http.StatusOK
		if response.Status == StatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		}

		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(response)
	}
}

// RedisCheck creates a health check for Redis connectivity
// This is a stub for Phase 1 - will be implemented in later phases
func RedisCheck(endpoint string, enabled bool) CheckFunc {
	return func(ctx context.Context) error {
		if !enabled {
			return nil // Skip if not enabled
		}
		// Placeholder: actual Redis check will be implemented in Phase 3
		return nil
	}
}

// MinIOCheck creates a health check for MinIO connectivity
// This is a stub for Phase 1 - will be implemented in later phases
func MinIOCheck(endpoint string, enabled bool) CheckFunc {
	return func(ctx context.Context) error {
		if !enabled {
			return nil // Skip if not enabled
		}
		// Placeholder: actual MinIO check will be implemented in Phase 4
		return nil
	}
}
