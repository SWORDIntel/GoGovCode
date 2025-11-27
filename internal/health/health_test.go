package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	checker := New("test-service", "1.0.0")

	if checker == nil {
		t.Fatal("expected non-nil checker")
	}

	if checker.serviceName != "test-service" {
		t.Errorf("expected service name 'test-service', got %s", checker.serviceName)
	}
}

func TestRegisterCheck(t *testing.T) {
	checker := New("test", "1.0.0")

	checkCalled := false
	checkFunc := func(ctx context.Context) error {
		checkCalled = true
		return nil
	}

	checker.RegisterCheck("test-check", checkFunc, false)

	// Verify check was registered
	if len(checker.checks) != 1 {
		t.Errorf("expected 1 check, got %d", len(checker.checks))
	}

	// Run checks to verify it works
	response := checker.RunChecks(context.Background())

	if !checkCalled {
		t.Error("expected check to be called")
	}

	if response.Status != StatusHealthy {
		t.Errorf("expected status healthy, got %s", response.Status)
	}
}

func TestRunChecks_AllHealthy(t *testing.T) {
	checker := New("test", "1.0.0")

	checker.RegisterCheck("check1", func(ctx context.Context) error {
		return nil
	}, true)

	checker.RegisterCheck("check2", func(ctx context.Context) error {
		return nil
	}, false)

	response := checker.RunChecks(context.Background())

	if response.Status != StatusHealthy {
		t.Errorf("expected status healthy, got %s", response.Status)
	}

	if len(response.Checks) != 2 {
		t.Errorf("expected 2 check results, got %d", len(response.Checks))
	}
}

func TestRunChecks_CriticalFailure(t *testing.T) {
	checker := New("test", "1.0.0")

	checker.RegisterCheck("critical-check", func(ctx context.Context) error {
		return errors.New("critical failure")
	}, true)

	response := checker.RunChecks(context.Background())

	if response.Status != StatusUnhealthy {
		t.Errorf("expected status unhealthy, got %s", response.Status)
	}

	checkResult := response.Checks["critical-check"]
	if checkResult.Status != StatusUnhealthy {
		t.Errorf("expected check status unhealthy, got %s", checkResult.Status)
	}

	if checkResult.Message != "critical failure" {
		t.Errorf("expected error message 'critical failure', got %s", checkResult.Message)
	}
}

func TestRunChecks_NonCriticalFailure(t *testing.T) {
	checker := New("test", "1.0.0")

	checker.RegisterCheck("non-critical-check", func(ctx context.Context) error {
		return errors.New("minor issue")
	}, false)

	response := checker.RunChecks(context.Background())

	if response.Status != StatusDegraded {
		t.Errorf("expected status degraded, got %s", response.Status)
	}

	checkResult := response.Checks["non-critical-check"]
	if checkResult.Status != StatusDegraded {
		t.Errorf("expected check status degraded, got %s", checkResult.Status)
	}
}

func TestRunChecks_MixedFailures(t *testing.T) {
	checker := New("test", "1.0.0")

	checker.RegisterCheck("critical", func(ctx context.Context) error {
		return errors.New("critical error")
	}, true)

	checker.RegisterCheck("non-critical", func(ctx context.Context) error {
		return errors.New("minor error")
	}, false)

	checker.RegisterCheck("healthy", func(ctx context.Context) error {
		return nil
	}, false)

	response := checker.RunChecks(context.Background())

	// Critical failure should make overall status unhealthy
	if response.Status != StatusUnhealthy {
		t.Errorf("expected status unhealthy, got %s", response.Status)
	}

	if len(response.Checks) != 3 {
		t.Errorf("expected 3 check results, got %d", len(response.Checks))
	}
}

func TestLivenessHandler(t *testing.T) {
	checker := New("test", "1.0.0")
	handler := checker.LivenessHandler()

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected content-type application/json, got %s", contentType)
	}
}

func TestReadinessHandler_Healthy(t *testing.T) {
	checker := New("test", "1.0.0")

	checker.RegisterCheck("test", func(ctx context.Context) error {
		return nil
	}, true)

	handler := checker.ReadinessHandler()

	req := httptest.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestReadinessHandler_Unhealthy(t *testing.T) {
	checker := New("test", "1.0.0")

	checker.RegisterCheck("test", func(ctx context.Context) error {
		return errors.New("service unavailable")
	}, true)

	handler := checker.ReadinessHandler()

	req := httptest.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", w.Code)
	}
}

func TestRedisCheck_Disabled(t *testing.T) {
	check := RedisCheck("localhost:6379", false)

	err := check(context.Background())
	if err != nil {
		t.Errorf("expected no error when disabled, got %v", err)
	}
}

func TestMinIOCheck_Disabled(t *testing.T) {
	check := MinIOCheck("localhost:9000", false)

	err := check(context.Background())
	if err != nil {
		t.Errorf("expected no error when disabled, got %v", err)
	}
}
