package audit

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/NSACodeGov/CodeGov/pkg/models"
)

func TestNewLogger(t *testing.T) {
	logger := NewLogger()

	if logger == nil {
		t.Fatal("expected non-nil logger")
	}

	if !logger.enabled {
		t.Error("expected logger to be enabled by default")
	}
}

func TestAddWriter(t *testing.T) {
	logger := NewLogger()
	writer := NewStdoutWriter()

	logger.AddWriter(writer)

	if len(logger.writers) != 1 {
		t.Errorf("expected 1 writer, got %d", len(logger.writers))
	}
}

func TestSetEnabled(t *testing.T) {
	logger := NewLogger()

	logger.SetEnabled(false)
	if logger.enabled {
		t.Error("expected logger to be disabled")
	}

	logger.SetEnabled(true)
	if !logger.enabled {
		t.Error("expected logger to be enabled")
	}
}

func TestLog(t *testing.T) {
	logger := NewLogger()

	// Use a buffer to capture output
	var buf bytes.Buffer
	testWriter := &bufferWriter{buf: &buf}
	logger.AddWriter(testWriter)

	event := &AuditEvent{
		Actor:    "test-user",
		Action:   "/test",
		Method:   "GET",
		Decision: DecisionAllow,
		Reason:   "test reason",
	}

	if err := logger.Log(event); err != nil {
		t.Fatalf("failed to log event: %v", err)
	}

	// Verify event was logged
	if testWriter.callCount != 1 {
		t.Errorf("expected 1 write call, got %d", testWriter.callCount)
	}

	// Verify event has ID and timestamp
	if event.EventID == "" {
		t.Error("expected event to have ID")
	}

	if event.Timestamp.IsZero() {
		t.Error("expected event to have timestamp")
	}
}

func TestLogDisabled(t *testing.T) {
	logger := NewLogger()
	logger.SetEnabled(false)

	testWriter := &bufferWriter{}
	logger.AddWriter(testWriter)

	event := &AuditEvent{
		Action: "/test",
	}

	if err := logger.Log(event); err != nil {
		t.Fatalf("failed to log event: %v", err)
	}

	// Should not write when disabled
	if testWriter.callCount != 0 {
		t.Errorf("expected 0 write calls when disabled, got %d", testWriter.callCount)
	}
}

func TestStdoutWriter(t *testing.T) {
	writer := NewStdoutWriter()

	event := &AuditEvent{
		EventID:   "test-event",
		Timestamp: time.Now(),
		Actor:     "test-user",
		Action:    "/test",
		Method:    "GET",
		Decision:  DecisionAllow,
		Reason:    "test",
	}

	// Should not error
	if err := writer.Write(event); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Close should not error
	if err := writer.Close(); err != nil {
		t.Errorf("unexpected error on close: %v", err)
	}
}

func TestFileWriter(t *testing.T) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "audit-test-*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	writer, err := NewFileWriter(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to create file writer: %v", err)
	}
	defer writer.Close()

	event := &AuditEvent{
		EventID:   "test-event",
		Timestamp: time.Now(),
		Actor:     "test-user",
		Action:    "/test",
		Method:    "GET",
		Decision:  DecisionAllow,
		Reason:    "test",
	}

	if err := writer.Write(event); err != nil {
		t.Fatalf("failed to write event: %v", err)
	}

	// Close writer
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	// Read file and verify content
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read audit file: %v", err)
	}

	var written AuditEvent
	if err := json.Unmarshal(data, &written); err != nil {
		t.Fatalf("failed to parse audit event: %v", err)
	}

	if written.EventID != event.EventID {
		t.Errorf("expected event ID %s, got %s", event.EventID, written.EventID)
	}

	if written.Actor != event.Actor {
		t.Errorf("expected actor %s, got %s", event.Actor, written.Actor)
	}
}

func TestMinIOWriter(t *testing.T) {
	writer := NewMinIOWriter("localhost:9000", "audit")

	// Should not error even though it's a stub
	event := &AuditEvent{
		EventID:  "test-event",
		Decision: DecisionAllow,
	}

	if err := writer.Write(event); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Errorf("unexpected error on close: %v", err)
	}
}

func TestNewEvent(t *testing.T) {
	event := NewEvent(DecisionAllow, "/test", "/test/resource", "test reason")

	if event == nil {
		t.Fatal("expected non-nil event")
	}

	if event.EventID == "" {
		t.Error("expected event to have ID")
	}

	if event.Timestamp.IsZero() {
		t.Error("expected event to have timestamp")
	}

	if event.Decision != DecisionAllow {
		t.Errorf("expected decision allow, got %s", event.Decision)
	}

	if event.Action != "/test" {
		t.Errorf("expected action '/test', got %s", event.Action)
	}

	if event.Resource != "/test/resource" {
		t.Errorf("expected resource '/test/resource', got %s", event.Resource)
	}

	if event.Reason != "test reason" {
		t.Errorf("expected reason 'test reason', got %s", event.Reason)
	}
}

func TestAuditEventJSON(t *testing.T) {
	event := &AuditEvent{
		EventID:   "evt-123",
		Timestamp: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Actor:     "device-1",
		Clearance: models.ClearanceLevel5,
		DeviceID:  1,
		Layer:     models.LayerControl,
		Action:    "/api/test",
		Method:    "GET",
		Resource:  "/api/test?foo=bar",
		Decision:  DecisionAllow,
		Reason:    "policy allows",
		RequestID: "req-456",
		SourceIP:  "192.168.1.1",
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}

	var decoded AuditEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal event: %v", err)
	}

	if decoded.EventID != event.EventID {
		t.Errorf("event ID mismatch: expected %s, got %s", event.EventID, decoded.EventID)
	}

	if decoded.Actor != event.Actor {
		t.Errorf("actor mismatch: expected %s, got %s", event.Actor, decoded.Actor)
	}

	if decoded.Decision != event.Decision {
		t.Errorf("decision mismatch: expected %s, got %s", event.Decision, decoded.Decision)
	}
}

// bufferWriter is a test writer that captures writes
type bufferWriter struct {
	buf       *bytes.Buffer
	callCount int
}

func (w *bufferWriter) Write(event *AuditEvent) error {
	w.callCount++
	if w.buf != nil {
		data, _ := json.Marshal(event)
		w.buf.Write(data)
	}
	return nil
}

func (w *bufferWriter) Close() error {
	return nil
}
