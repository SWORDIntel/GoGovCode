package audit

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/NSACodeGov/CodeGov/pkg/models"
)

// Decision represents a policy decision
type Decision string

const (
	DecisionAllow Decision = "allow"
	DecisionDeny  Decision = "deny"
)

// AuditEvent represents a unified audit event
type AuditEvent struct {
	EventID        string           `json:"event_id"`
	Timestamp      time.Time        `json:"timestamp"`
	Actor          string           `json:"actor"`
	Clearance      models.Clearance `json:"clearance"`
	DeviceID       uint16           `json:"device_id"`
	Layer          models.Layer     `json:"layer"`
	Action         string           `json:"action"`
	Method         string           `json:"method"`
	Resource       string           `json:"resource"`
	Decision       Decision         `json:"decision"`
	Reason         string           `json:"reason"`
	RequestID      string           `json:"request_id,omitempty"`
	SourceIP       string           `json:"source_ip,omitempty"`
	StatusCode     int              `json:"status_code,omitempty"`
	AdditionalData map[string]interface{} `json:"additional_data,omitempty"`
}

// Writer defines the interface for audit event writers
type Writer interface {
	Write(event *AuditEvent) error
	Close() error
}

// Logger is the main audit logger
type Logger struct {
	mu      sync.RWMutex
	writers []Writer
	enabled bool
}

// NewLogger creates a new audit logger
func NewLogger() *Logger {
	return &Logger{
		writers: make([]Writer, 0),
		enabled: true,
	}
}

// AddWriter adds a writer to the audit logger
func (l *Logger) AddWriter(w Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.writers = append(l.writers, w)
}

// SetEnabled enables or disables audit logging
func (l *Logger) SetEnabled(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = enabled
}

// Log writes an audit event to all registered writers
func (l *Logger) Log(event *AuditEvent) error {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if !l.enabled {
		return nil
	}

	// Ensure event has an ID and timestamp
	if event.EventID == "" {
		event.EventID = generateEventID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	// Write to all writers
	var lastErr error
	for _, writer := range l.writers {
		if err := writer.Write(event); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// Close closes all writers
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var lastErr error
	for _, writer := range l.writers {
		if err := writer.Close(); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// StdoutWriter writes audit events to stdout
type StdoutWriter struct {
	mu sync.Mutex
}

// NewStdoutWriter creates a new stdout writer
func NewStdoutWriter() *StdoutWriter {
	return &StdoutWriter{}
}

// Write writes an event to stdout
func (w *StdoutWriter) Write(event *AuditEvent) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// Close is a no-op for stdout
func (w *StdoutWriter) Close() error {
	return nil
}

// FileWriter writes audit events to a file
type FileWriter struct {
	mu   sync.Mutex
	file *os.File
}

// NewFileWriter creates a new file writer
func NewFileWriter(path string) (*FileWriter, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit file: %w", err)
	}

	return &FileWriter{
		file: file,
	}, nil
}

// Write writes an event to the file
func (w *FileWriter) Write(event *AuditEvent) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	if _, err := w.file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write audit event: %w", err)
	}

	// Ensure data is flushed to disk
	return w.file.Sync()
}

// Close closes the file
func (w *FileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// MinIOWriter is a stub for MinIO-backed audit logging
// Full implementation will come in Phase 4
type MinIOWriter struct {
	endpoint  string
	bucket    string
	enabled   bool
	batchSize int
	mu        sync.Mutex
	batch     []*AuditEvent
}

// NewMinIOWriter creates a new MinIO writer (stub)
func NewMinIOWriter(endpoint, bucket string) *MinIOWriter {
	return &MinIOWriter{
		endpoint:  endpoint,
		bucket:    bucket,
		enabled:   false, // Disabled until Phase 4
		batchSize: 100,
		batch:     make([]*AuditEvent, 0, 100),
	}
}

// Write writes an event to MinIO (stub - queues for future implementation)
func (w *MinIOWriter) Write(event *AuditEvent) error {
	if !w.enabled {
		// Stub: just ignore for now
		return nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	w.batch = append(w.batch, event)

	// TODO Phase 4: Implement actual MinIO upload with:
	// - Hash chain linking
	// - Batch uploads
	// - Immutable object storage
	// - Merkle tree verification

	if len(w.batch) >= w.batchSize {
		// TODO: Flush batch to MinIO
		w.batch = w.batch[:0]
	}

	return nil
}

// Close flushes any pending events and closes the writer
func (w *MinIOWriter) Close() error {
	if !w.enabled {
		return nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	// TODO Phase 4: Flush remaining batch
	w.batch = w.batch[:0]

	return nil
}

// generateEventID generates a unique event ID
func generateEventID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("evt-%d", time.Now().UnixNano())
	}
	return "evt-" + hex.EncodeToString(b)
}

// NewEvent creates a new audit event with common fields populated
func NewEvent(decision Decision, action, resource, reason string) *AuditEvent {
	return &AuditEvent{
		EventID:   generateEventID(),
		Timestamp: time.Now().UTC(),
		Decision:  decision,
		Action:    action,
		Resource:  resource,
		Reason:    reason,
	}
}
