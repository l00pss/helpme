package o4g_logger

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestInit(t *testing.T) {
	// Test initialization
	config := Config{
		Level:       InfoLevel,
		Format:      TextFormat,
		Output:      "stdout",
		ServiceName: "test-service",
		Environment: "test",
	}

	err := Init(config)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Test that default logger is set
	logger := GetDefaultLogger()
	if logger == nil {
		t.Error("Default logger should not be nil after initialization")
	}

	if logger.config.ServiceName != "test-service" {
		t.Errorf("Expected service name 'test-service', got '%s'", logger.config.ServiceName)
	}
}

func TestInitOnce(t *testing.T) {
	// Reset the once for this test
	once.Do(func() {})

	// Test that Init only runs once
	config1 := Config{
		ServiceName: "service1",
		Level:       InfoLevel,
		Format:      TextFormat,
		Output:      "stdout",
	}

	config2 := Config{
		ServiceName: "service2",
		Level:       DebugLevel,
		Format:      JSONFormat,
		Output:      "stderr",
	}

	// First init
	err1 := Init(config1)
	if err1 != nil {
		t.Fatalf("First init failed: %v", err1)
	}

	logger1 := GetDefaultLogger()
	serviceName1 := logger1.config.ServiceName

	// Second init should not change the logger
	err2 := Init(config2)
	if err2 != nil {
		t.Fatalf("Second init failed: %v", err2)
	}

	logger2 := GetDefaultLogger()
	serviceName2 := logger2.config.ServiceName

	if serviceName1 != serviceName2 {
		t.Errorf("Logger should not change on second init. Expected %s, got %s", serviceName1, serviceName2)
	}

	if serviceName2 != "service1" {
		t.Errorf("Service name should remain 'service1', got '%s'", serviceName2)
	}
}

func TestGetDefaultLogger(t *testing.T) {
	// Reset global state
	defaultLogger = nil
	once = sync.Once{}

	// GetDefaultLogger should initialize with default config if not initialized
	logger := GetDefaultLogger()
	if logger == nil {
		t.Error("GetDefaultLogger should not return nil")
	}

	// Should return default config values
	if logger.config.ServiceName != "gatekeeper" {
		t.Errorf("Expected default service name 'gatekeeper', got '%s'", logger.config.ServiceName)
	}

	if logger.config.Environment != "development" {
		t.Errorf("Expected default environment 'development', got '%s'", logger.config.Environment)
	}
}

func TestGlobalConvenienceFunctions(t *testing.T) {
	// Initialize with test config
	config := DefaultConfig()
	config.ServiceName = "test-global"

	err := Init(config)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Test that global functions don't panic
	tests := []struct {
		name string
		fn   func()
	}{
		{"Trace", func() { Trace("test trace") }},
		{"Debug", func() { Debug("test debug") }},
		{"Info", func() { Info("test info") }},
		{"Warn", func() { Warn("test warn") }},
		{"Error", func() { Error("test error") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Global function %s panicked: %v", tt.name, r)
				}
			}()
			tt.fn()
		})
	}
}

func TestGlobalWithFields(t *testing.T) {
	config := DefaultConfig()
	err := Init(config)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Test WithFields
	logger := WithFields(map[string]interface{}{
		"test_field": "test_value",
	})

	if logger == nil {
		t.Error("WithFields should not return nil")
	}

	// Test WithField
	logger2 := WithField("single_field", "single_value")
	if logger2 == nil {
		t.Error("WithField should not return nil")
	}

	// Test WithError
	testErr := errors.New("test error")
	logger3 := WithError(testErr)
	if logger3 == nil {
		t.Error("WithError should not return nil")
	}
}

func TestGlobalLogHTTPRequest(t *testing.T) {
	config := DefaultConfig()
	err := Init(config)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("LogHTTPRequest panicked: %v", r)
		}
	}()

	LogHTTPRequest("GET", "/test", "TestAgent", "127.0.0.1", 200, 100)
}

func TestGlobalLogAuthEvent(t *testing.T) {
	config := DefaultConfig()
	err := Init(config)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("LogAuthEvent panicked: %v", r)
		}
	}()

	LogAuthEvent("login", "user123", "127.0.0.1", true)
}

func TestGlobalLogAuditEvent(t *testing.T) {
	config := DefaultConfig()
	err := Init(config)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("LogAuditEvent panicked: %v", r)
		}
	}()

	details := map[string]interface{}{
		"action": "test",
	}
	LogAuditEvent("create", "user", "user123", details)
}

func TestGlobalLogError(t *testing.T) {
	config := DefaultConfig()
	err := Init(config)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("LogError panicked: %v", r)
		}
	}()

	testErr := errors.New("test error")
	fields := map[string]interface{}{
		"context": "test",
	}
	LogError(testErr, "test_context", fields)
}

// Test panic and fatal functions in separate process to avoid killing test
func TestGlobalPanicFunction(t *testing.T) {
	config := DefaultConfig()
	err := Init(config)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// We can't actually test panic/fatal as they would kill the test
	// But we can verify the function exists and can be called without compilation errors
	defer func() {
		if r := recover(); r != nil {
			// This is expected for Panic function
			if !strings.Contains(fmt.Sprintf("%v", r), "test panic") {
				t.Errorf("Unexpected panic message: %v", r)
			}
		} else {
			t.Error("Panic function should have panicked")
		}
	}()

	Panic("test panic")
}

// Test concurrent access to global logger
func TestConcurrentAccess(t *testing.T) {
	config := DefaultConfig()
	err := Init(config)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	const numGoroutines = 100
	const numLogs = 10

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Goroutine %d panicked: %v", id, r)
				}
				done <- true
			}()

			for j := 0; j < numLogs; j++ {
				Info(fmt.Sprintf("Concurrent log %d-%d", id, j))
				WithField("goroutine_id", id).Info(fmt.Sprintf("Log with field %d-%d", id, j))
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}
