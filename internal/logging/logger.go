package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Level represents log severity level
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// Context keys for log fields
type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	DeviceIDKey  contextKey = "device_id"
	LayerKey     contextKey = "layer"
)

// Logger provides structured logging with correlation IDs
type Logger struct {
	mu           sync.Mutex
	output       io.Writer
	level        Level
	serviceName  string
	serviceVer   string
	format       string // "json" or "text"
	defaultFields map[string]interface{}
}

// Entry represents a single log entry
type Entry struct {
	Timestamp  string                 `json:"timestamp"`
	Level      string                 `json:"level"`
	Message    string                 `json:"msg"`
	Service    string                 `json:"service"`
	Version    string                 `json:"version"`
	RequestID  string                 `json:"request_id,omitempty"`
	DeviceID   string                 `json:"device_id,omitempty"`
	Layer      string                 `json:"layer,omitempty"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
}

// New creates a new Logger
func New(serviceName, serviceVersion, level, format string) *Logger {
	return &Logger{
		output:        os.Stdout,
		level:         Level(level),
		serviceName:   serviceName,
		serviceVer:    serviceVersion,
		format:        format,
		defaultFields: make(map[string]interface{}),
	}
}

// WithField adds a default field to all log entries
func (l *Logger) WithField(key string, value interface{}) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.defaultFields[key] = value
	return l
}

// WithFields adds multiple default fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	for k, v := range fields {
		l.defaultFields[k] = v
	}
	return l
}

// SetOutput sets the output writer
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = w
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	l.log(context.Background(), LevelDebug, msg, fields...)
}

// DebugContext logs a debug message with context
func (l *Logger) DebugContext(ctx context.Context, msg string, fields ...map[string]interface{}) {
	l.log(ctx, LevelDebug, msg, fields...)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	l.log(context.Background(), LevelInfo, msg, fields...)
}

// InfoContext logs an info message with context
func (l *Logger) InfoContext(ctx context.Context, msg string, fields ...map[string]interface{}) {
	l.log(ctx, LevelInfo, msg, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...map[string]interface{}) {
	l.log(context.Background(), LevelWarn, msg, fields...)
}

// WarnContext logs a warning message with context
func (l *Logger) WarnContext(ctx context.Context, msg string, fields ...map[string]interface{}) {
	l.log(ctx, LevelWarn, msg, fields...)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...map[string]interface{}) {
	l.log(context.Background(), LevelError, msg, fields...)
}

// ErrorContext logs an error message with context
func (l *Logger) ErrorContext(ctx context.Context, msg string, fields ...map[string]interface{}) {
	l.log(ctx, LevelError, msg, fields...)
}

// log is the internal logging function
func (l *Logger) log(ctx context.Context, level Level, msg string, fields ...map[string]interface{}) {
	if !l.shouldLog(level) {
		return
	}

	entry := Entry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     string(level),
		Message:   msg,
		Service:   l.serviceName,
		Version:   l.serviceVer,
		Fields:    make(map[string]interface{}),
	}

	// Extract context values
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		entry.RequestID = requestID
	}
	if deviceID, ok := ctx.Value(DeviceIDKey).(string); ok && deviceID != "" {
		entry.DeviceID = deviceID
	}
	if layer, ok := ctx.Value(LayerKey).(string); ok && layer != "" {
		entry.Layer = layer
	}

	// Add default fields
	l.mu.Lock()
	for k, v := range l.defaultFields {
		entry.Fields[k] = v
	}
	l.mu.Unlock()

	// Add provided fields
	if len(fields) > 0 {
		for k, v := range fields[0] {
			entry.Fields[k] = v
		}
	}

	// Remove empty fields map if no fields
	if len(entry.Fields) == 0 {
		entry.Fields = nil
	}

	l.write(entry)
}

// shouldLog checks if a message at the given level should be logged
func (l *Logger) shouldLog(level Level) bool {
	levelOrder := map[Level]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelWarn:  2,
		LevelError: 3,
	}

	return levelOrder[level] >= levelOrder[l.level]
}

// write outputs the log entry
func (l *Logger) write(entry Entry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var output string

	if l.format == "text" {
		// Simple text format for development
		fieldsStr := ""
		if entry.Fields != nil {
			data, _ := json.Marshal(entry.Fields)
			fieldsStr = " " + string(data)
		}

		output = fmt.Sprintf("[%s] %s %s/%s: %s%s",
			entry.Timestamp,
			entry.Level,
			entry.Service,
			entry.Version,
			entry.Message,
			fieldsStr,
		)

		if entry.RequestID != "" {
			output += fmt.Sprintf(" [req=%s]", entry.RequestID)
		}
	} else {
		// JSON format (default)
		data, err := json.Marshal(entry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to marshal log entry: %v\n", err)
			return
		}
		output = string(data)
	}

	fmt.Fprintln(l.output, output)
}

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// WithDeviceID adds a device ID to the context
func WithDeviceID(ctx context.Context, deviceID string) context.Context {
	return context.WithValue(ctx, DeviceIDKey, deviceID)
}

// WithLayer adds a layer to the context
func WithLayer(ctx context.Context, layer string) context.Context {
	return context.WithValue(ctx, LayerKey, layer)
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}
