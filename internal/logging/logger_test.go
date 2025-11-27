package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	logger := New("test-service", "1.0.0", "info", "json")

	if logger == nil {
		t.Fatal("expected non-nil logger")
	}

	if logger.serviceName != "test-service" {
		t.Errorf("expected service name 'test-service', got %s", logger.serviceName)
	}

	if logger.serviceVer != "1.0.0" {
		t.Errorf("expected service version '1.0.0', got %s", logger.serviceVer)
	}
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name       string
		logLevel   Level
		logFunc    func(*Logger, string)
		shouldLog  bool
	}{
		{
			name:      "debug logs at debug level",
			logLevel:  LevelDebug,
			logFunc:   func(l *Logger, msg string) { l.Debug(msg) },
			shouldLog: true,
		},
		{
			name:      "debug does not log at info level",
			logLevel:  LevelInfo,
			logFunc:   func(l *Logger, msg string) { l.Debug(msg) },
			shouldLog: false,
		},
		{
			name:      "info logs at info level",
			logLevel:  LevelInfo,
			logFunc:   func(l *Logger, msg string) { l.Info(msg) },
			shouldLog: true,
		},
		{
			name:      "info logs at debug level",
			logLevel:  LevelDebug,
			logFunc:   func(l *Logger, msg string) { l.Info(msg) },
			shouldLog: true,
		},
		{
			name:      "error always logs",
			logLevel:  LevelError,
			logFunc:   func(l *Logger, msg string) { l.Error(msg) },
			shouldLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New("test", "1.0.0", string(tt.logLevel), "json")
			logger.SetOutput(&buf)

			tt.logFunc(logger, "test message")

			logged := buf.Len() > 0
			if logged != tt.shouldLog {
				t.Errorf("expected log=%v, got log=%v", tt.shouldLog, logged)
			}
		})
	}
}

func TestJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := New("test-service", "1.0.0", "info", "json")
	logger.SetOutput(&buf)

	logger.Info("test message")

	var entry Entry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON log: %v", err)
	}

	if entry.Message != "test message" {
		t.Errorf("expected message 'test message', got %s", entry.Message)
	}

	if entry.Service != "test-service" {
		t.Errorf("expected service 'test-service', got %s", entry.Service)
	}

	if entry.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", entry.Version)
	}

	if entry.Level != "info" {
		t.Errorf("expected level 'info', got %s", entry.Level)
	}
}

func TestWithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := New("test", "1.0.0", "info", "json")
	logger.SetOutput(&buf)

	logger.Info("test", map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	})

	var entry Entry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON log: %v", err)
	}

	if entry.Fields["key1"] != "value1" {
		t.Errorf("expected field key1='value1', got %v", entry.Fields["key1"])
	}

	if int(entry.Fields["key2"].(float64)) != 42 {
		t.Errorf("expected field key2=42, got %v", entry.Fields["key2"])
	}
}

func TestContextValues(t *testing.T) {
	var buf bytes.Buffer
	logger := New("test", "1.0.0", "info", "json")
	logger.SetOutput(&buf)

	ctx := WithRequestID(context.Background(), "req-123")
	ctx = WithDeviceID(ctx, "device-456")
	ctx = WithLayer(ctx, "api")

	logger.InfoContext(ctx, "test")

	var entry Entry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON log: %v", err)
	}

	if entry.RequestID != "req-123" {
		t.Errorf("expected request_id 'req-123', got %s", entry.RequestID)
	}

	if entry.DeviceID != "device-456" {
		t.Errorf("expected device_id 'device-456', got %s", entry.DeviceID)
	}

	if entry.Layer != "api" {
		t.Errorf("expected layer 'api', got %s", entry.Layer)
	}
}

func TestTextFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := New("test", "1.0.0", "info", "text")
	logger.SetOutput(&buf)

	logger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("expected output to contain 'test message', got: %s", output)
	}

	if !strings.Contains(output, "info") {
		t.Errorf("expected output to contain 'info', got: %s", output)
	}
}

func TestGetRequestID(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req-123")

	requestID := GetRequestID(ctx)
	if requestID != "req-123" {
		t.Errorf("expected request ID 'req-123', got %s", requestID)
	}

	// Test empty context
	emptyCtx := context.Background()
	if id := GetRequestID(emptyCtx); id != "" {
		t.Errorf("expected empty request ID, got %s", id)
	}
}
