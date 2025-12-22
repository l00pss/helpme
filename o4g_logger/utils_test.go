package o4g_logger

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestFromContext(t *testing.T) {
	// Test with context that has no logger
	ctx := context.Background()
	logger := FromContext(ctx)

	if logger == nil {
		t.Error("FromContext should return default logger when no logger in context")
	}

	// Should return the default logger
	defaultLogger := GetDefaultLogger()
	if logger != defaultLogger {
		t.Error("FromContext should return default logger when no logger in context")
	}
}

func TestFromContextWithLogger(t *testing.T) {
	// Initialize a test logger
	config := Config{
		Level:       DebugLevel,
		Format:      JSONFormat,
		Output:      "stdout",
		ServiceName: "test-context-service",
		Environment: "test",
	}

	testLogger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create test logger: %v", err)
	}

	// Add logger to context
	ctx := context.WithValue(context.Background(), LoggerContextKey, testLogger)

	// Retrieve logger from context
	retrievedLogger := FromContext(ctx)

	if retrievedLogger == nil {
		t.Error("FromContext should return the logger from context")
	}

	if retrievedLogger.config.ServiceName != "test-context-service" {
		t.Errorf("Expected service name 'test-context-service', got '%s'",
			retrievedLogger.config.ServiceName)
	}
}

func TestToContext(t *testing.T) {
	// Create a test logger
	config := Config{
		Level:       WarnLevel,
		Format:      TextFormat,
		Output:      "stderr",
		ServiceName: "context-test-service",
		Environment: "test",
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Add logger to context
	ctx := context.Background()
	newCtx := ToContext(ctx, logger)

	// Retrieve logger from new context
	retrievedLogger := FromContext(newCtx)

	if retrievedLogger == nil {
		t.Error("Logger should be retrievable from context")
	}

	if retrievedLogger.config.ServiceName != "context-test-service" {
		t.Errorf("Expected service name 'context-test-service', got '%s'",
			retrievedLogger.config.ServiceName)
	}
}

func TestWithRequestID(t *testing.T) {
	// Initialize default logger
	err := Init(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	ctx := context.Background()
	requestID := "req-12345"

	logger, newCtx := WithRequestID(ctx, requestID)

	if logger == nil {
		t.Error("WithRequestID should return a logger")
	}

	// Check that request ID was added to context
	retrievedID := newCtx.Value(RequestIDKey)
	if retrievedID != requestID {
		t.Errorf("Expected request ID '%s', got '%v'", requestID, retrievedID)
	}

	// The returned logger should be configured properly
	if logger.config.ServiceName == "" {
		t.Error("Logger should inherit config from default logger")
	}
}

func TestWithUserID(t *testing.T) {
	// Initialize default logger
	err := Init(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	ctx := context.Background()
	userID := "user-67890"

	logger, newCtx := WithUserID(ctx, userID)

	if logger == nil {
		t.Error("WithUserID should return a logger")
	}

	// Check that user ID was added to context
	retrievedID := newCtx.Value(UserIDKey)
	if retrievedID != userID {
		t.Errorf("Expected user ID '%s', got '%v'", userID, retrievedID)
	}

	// The returned logger should be configured properly
	if logger.config.ServiceName == "" {
		t.Error("Logger should inherit config from default logger")
	}
}

func TestContextKeys(t *testing.T) {
	// Test that context keys are properly defined
	tests := []struct {
		name string
		key  ContextKey
	}{
		{"LoggerContextKey", LoggerContextKey},
		{"RequestIDKey", RequestIDKey},
		{"UserIDKey", UserIDKey},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.key) == "" {
				t.Errorf("Context key %s should not be empty", tt.name)
			}
		})
	}

	// Test that keys are unique
	keys := []ContextKey{LoggerContextKey, RequestIDKey, UserIDKey}
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] == keys[j] {
				t.Errorf("Context keys should be unique: %s == %s", keys[i], keys[j])
			}
		}
	}
}

func TestContextChaining(t *testing.T) {
	// Initialize default logger
	err := Init(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Create a chain of context operations
	ctx := context.Background()
	requestID := "req-123"
	userID := "user-456"

	// Add request ID
	_, ctx = WithRequestID(ctx, requestID)

	// Add user ID
	logger, ctx := WithUserID(ctx, userID)

	// Verify both values are in context
	retrievedRequestID := ctx.Value(RequestIDKey)
	if retrievedRequestID != requestID {
		t.Errorf("Expected request ID '%s', got '%v'", requestID, retrievedRequestID)
	}

	retrievedUserID := ctx.Value(UserIDKey)
	if retrievedUserID != userID {
		t.Errorf("Expected user ID '%s', got '%v'", userID, retrievedUserID)
	}

	// Logger should be properly configured
	if logger == nil {
		t.Error("Final logger should not be nil")
	}
}

func TestFromContextWithWrongType(t *testing.T) {
	// Add a non-logger value with the logger key
	ctx := context.WithValue(context.Background(), LoggerContextKey, "not a logger")

	logger := FromContext(ctx)

	// Should return default logger when wrong type is stored
	defaultLogger := GetDefaultLogger()
	if logger != defaultLogger {
		t.Error("FromContext should return default logger when wrong type in context")
	}
}

func TestNewTimer(t *testing.T) {
	logger, err := NewLogger(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test timer without fields
	timer1 := NewTimer(logger, "test_operation")
	if timer1 == nil {
		t.Error("NewTimer should not return nil")
	}
	if timer1.name != "test_operation" {
		t.Errorf("Expected operation name 'test_operation', got '%s'", timer1.name)
	}
	if timer1.logger != logger {
		t.Error("Timer should reference the provided logger")
	}

	// Test timer with fields
	fields := map[string]interface{}{
		"test_field": "test_value",
	}
	timer2 := NewTimer(logger, "test_with_fields", fields)
	if timer2 == nil {
		t.Error("NewTimer with fields should not return nil")
	}
	if timer2.fields["test_field"] != "test_value" {
		t.Error("Timer should store provided fields")
	}
}

func TestTimerStop(t *testing.T) {
	var buf bytes.Buffer

	config := DefaultConfig()
	config.EnableColors = false

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	logger.SetOutput(&buf)

	timer := NewTimer(logger, "test_timer_operation")

	// Wait a small amount to ensure some duration
	time.Sleep(1 * time.Millisecond)

	duration := timer.Stop()

	if duration <= 0 {
		t.Error("Timer duration should be positive")
	}

	output := buf.String()
	if !strings.Contains(output, "test_timer_operation") {
		t.Error("Timer output should contain operation name")
	}
	if !strings.Contains(output, "duration_ms") {
		t.Error("Timer output should contain duration_ms field")
	}
}

func TestTimerStopf(t *testing.T) {
	var buf bytes.Buffer

	config := DefaultConfig()
	config.EnableColors = false

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	logger.SetOutput(&buf)

	timer := NewTimer(logger, "formatted_timer")

	time.Sleep(1 * time.Millisecond)

	duration := timer.Stopf("Custom message with value: %d", 42)

	if duration <= 0 {
		t.Error("Timer duration should be positive")
	}

	output := buf.String()
	if !strings.Contains(output, "Custom message with value: 42") {
		t.Error("Timer output should contain formatted message")
	}
	if !strings.Contains(output, "duration_ms") {
		t.Error("Timer output should contain duration_ms field")
	}
}

func TestTimerWithFields(t *testing.T) {
	var buf bytes.Buffer

	config := DefaultConfig()
	config.EnableColors = false
	config.Format = JSONFormat

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	logger.SetOutput(&buf)

	fields := map[string]interface{}{
		"request_id": "req-123",
		"user_id":    "user-456",
	}
	timer := NewTimer(logger, "operation_with_fields", fields)

	time.Sleep(1 * time.Millisecond)
	timer.Stop()

	output := buf.String()
	if !strings.Contains(output, "request_id") {
		t.Error("Timer output should contain request_id field")
	}
	if !strings.Contains(output, "user_id") {
		t.Error("Timer output should contain user_id field")
	}
	if !strings.Contains(output, "req-123") {
		t.Error("Timer output should contain request_id value")
	}
}

func TestNewHook(t *testing.T) {
	var buf bytes.Buffer

	hook := NewHook(&buf, logrus.InfoLevel)
	if hook == nil {
		t.Error("NewHook should not return nil")
	}
	if hook.Writer != &buf {
		t.Error("Hook should reference the provided writer")
	}
	if hook.Level != logrus.InfoLevel {
		t.Error("Hook should store the provided level")
	}
}

// Benchmark tests
func BenchmarkFromContext(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FromContext(ctx)
	}
}

func BenchmarkFromContextWithLogger(b *testing.B) {
	logger, _ := NewLogger(DefaultConfig())
	ctx := ToContext(context.Background(), logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FromContext(ctx)
	}
}

func BenchmarkWithRequestID(b *testing.B) {
	_ = Init(DefaultConfig())
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = WithRequestID(ctx, "req-123")
	}
}

func BenchmarkNewTimer(b *testing.B) {
	logger, _ := NewLogger(DefaultConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewTimer(logger, "benchmark_operation")
	}
}

func BenchmarkTimerStop(b *testing.B) {
	logger, _ := NewLogger(DefaultConfig())
	logger.SetOutput(io.Discard) // Don't actually write output

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		timer := NewTimer(logger, "benchmark_timer")
		timer.Stop()
	}
}

func BenchmarkContextOperations(b *testing.B) {
	_ = Init(DefaultConfig())
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, newCtx := WithRequestID(ctx, "req-123")
		_, _ = WithUserID(newCtx, "user-456")
	}
}
