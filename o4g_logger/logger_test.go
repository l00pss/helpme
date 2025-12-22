package o4g_logger

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Level != InfoLevel {
		t.Errorf("Expected default level to be %v, got %v", InfoLevel, config.Level)
	}
	if config.Format != TextFormat {
		t.Errorf("Expected default format to be %v, got %v", TextFormat, config.Format)
	}
	if config.Output != "stdout" {
		t.Errorf("Expected default output to be stdout, got %v", config.Output)
	}
	if !config.EnableCaller {
		t.Error("Expected default EnableCaller to be true")
	}
	if !config.EnableColors {
		t.Error("Expected default EnableColors to be true")
	}
	if config.ServiceName != "gatekeeper" {
		t.Errorf("Expected default ServiceName to be gatekeeper, got %v", config.ServiceName)
	}
	if config.Environment != "development" {
		t.Errorf("Expected default Environment to be development, got %v", config.Environment)
	}
}

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "Valid config with default values",
			config: Config{
				Level:           InfoLevel,
				Format:          TextFormat,
				Output:          "stdout",
				EnableCaller:    true,
				EnableColors:    true,
				ServiceName:     "test-service",
				Environment:     "test",
				TimestampFormat: time.RFC3339,
			},
			wantErr: false,
		},
		{
			name: "JSON format",
			config: Config{
				Level:           DebugLevel,
				Format:          JSONFormat,
				Output:          "stderr",
				EnableCaller:    false,
				EnableColors:    false,
				ServiceName:     "test-service",
				Environment:     "production",
				TimestampFormat: time.RFC3339,
			},
			wantErr: false,
		},
		{
			name: "Invalid log level",
			config: Config{
				Level:  "invalid",
				Format: TextFormat,
				Output: "stdout",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if logger == nil {
				t.Error("Expected logger but got nil")
				return
			}

			// Verify config is stored
			if logger.config.ServiceName != tt.config.ServiceName {
				t.Errorf("Expected service name %v, got %v", tt.config.ServiceName, logger.config.ServiceName)
			}
		})
	}
}

func TestNewLoggerFileOutput(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-log-*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	config := Config{
		Level:           InfoLevel,
		Format:          TextFormat,
		Output:          tmpFile.Name(),
		EnableCaller:    false,
		EnableColors:    false,
		ServiceName:     "test-service",
		Environment:     "test",
		TimestampFormat: time.RFC3339,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test writing to file
	logger.Info("Test log message")

	// Read the file content
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "Test log message") {
		t.Error("Log message not found in file")
	}
}

func TestLoggerMethods(t *testing.T) {
	var buf bytes.Buffer

	config := DefaultConfig()
	config.Output = "stdout"
	config.EnableColors = false

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.SetOutput(&buf)

	tests := []struct {
		name     string
		logFunc  func()
		expected string
	}{
		{
			name: "Info log",
			logFunc: func() {
				logger.Info("Test info message")
			},
			expected: "Test info message",
		},
		{
			name: "Debug log",
			logFunc: func() {
				logger.Debug("Test debug message")
			},
			expected: "Test debug message",
		},
		{
			name: "Warn log",
			logFunc: func() {
				logger.Warn("Test warn message")
			},
			expected: "Test warn message",
		},
		{
			name: "Error log",
			logFunc: func() {
				logger.Error("Test error message")
			},
			expected: "Test error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()

			output := buf.String()
			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected output to contain '%s', got '%s'", tt.expected, output)
			}
		})
	}
}

func TestLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer

	config := DefaultConfig()
	config.EnableColors = false

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.SetOutput(&buf)

	logger.WithFields(map[string]interface{}{
		"user_id": "12345",
		"action":  "login",
	}).Info("User login")

	output := buf.String()
	if !strings.Contains(output, "User login") {
		t.Error("Expected log message not found")
	}
	if !strings.Contains(output, "user_id") {
		t.Error("Expected field user_id not found")
	}
	if !strings.Contains(output, "action") {
		t.Error("Expected field action not found")
	}
}

func TestLoggerWithField(t *testing.T) {
	var buf bytes.Buffer

	config := DefaultConfig()
	config.EnableColors = false

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.SetOutput(&buf)

	logger.WithField("request_id", "req-123").Info("Processing request")

	output := buf.String()
	if !strings.Contains(output, "Processing request") {
		t.Error("Expected log message not found")
	}
	if !strings.Contains(output, "request_id") {
		t.Error("Expected field request_id not found")
	}
}

func TestLoggerWithError(t *testing.T) {
	var buf bytes.Buffer

	config := DefaultConfig()
	config.EnableColors = false

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.SetOutput(&buf)

	testErr := errors.New("test error")
	logger.WithError(testErr).Error("An error occurred")

	output := buf.String()
	if !strings.Contains(output, "An error occurred") {
		t.Error("Expected log message not found")
	}
	if !strings.Contains(output, "test error") {
		t.Error("Expected error message not found")
	}
}

func TestSetLevel(t *testing.T) {
	logger, err := NewLogger(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test valid level change
	err = logger.SetLevel(ErrorLevel)
	if err != nil {
		t.Errorf("Unexpected error setting level: %v", err)
	}

	if logger.GetLevel() != ErrorLevel {
		t.Errorf("Expected level %v, got %v", ErrorLevel, logger.GetLevel())
	}

	// Test invalid level
	err = logger.SetLevel("invalid")
	if err == nil {
		t.Error("Expected error for invalid level")
	}
}

func TestIsLevelEnabled(t *testing.T) {
	logger, err := NewLogger(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Set to Info level
	logger.SetLevel(InfoLevel)

	tests := []struct {
		level   LogLevel
		enabled bool
	}{
		{TraceLevel, false},
		{DebugLevel, false},
		{InfoLevel, true},
		{WarnLevel, true},
		{ErrorLevel, true},
		{FatalLevel, true},
		{PanicLevel, true},
	}

	for _, tt := range tests {
		if logger.IsLevelEnabled(tt.level) != tt.enabled {
			t.Errorf("Level %v enabled status expected %v, got %v",
				tt.level, tt.enabled, logger.IsLevelEnabled(tt.level))
		}
	}
}

func TestLogHTTPRequest(t *testing.T) {
	var buf bytes.Buffer

	config := DefaultConfig()
	config.Format = JSONFormat

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.SetOutput(&buf)

	logger.LogHTTPRequest("GET", "/api/users", "Mozilla/5.0", "192.168.1.1", 200, 150)

	output := buf.String()
	expectedFields := []string{
		"GET", "/api/users", "Mozilla/5.0", "192.168.1.1", "200", "150", "http_request",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected field '%s' not found in output: %s", field, output)
		}
	}
}

func TestLogAuthEvent(t *testing.T) {
	var buf bytes.Buffer

	config := DefaultConfig()
	config.Format = JSONFormat

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.SetOutput(&buf)

	logger.LogAuthEvent("login", "user123", "192.168.1.1", true)

	output := buf.String()
	expectedFields := []string{
		"login", "user123", "192.168.1.1", "true", "auth_event",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected field '%s' not found in output: %s", field, output)
		}
	}
}

func TestLogAuditEvent(t *testing.T) {
	var buf bytes.Buffer

	config := DefaultConfig()
	config.Format = JSONFormat

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.SetOutput(&buf)

	details := map[string]interface{}{
		"ip":     "192.168.1.1",
		"reason": "password_change",
	}
	logger.LogAuditEvent("update", "user", "user123", details)

	output := buf.String()
	expectedFields := []string{
		"update", "user", "user123", "audit_event", "ip", "reason",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected field '%s' not found in output: %s", field, output)
		}
	}
}

func TestLogError(t *testing.T) {
	var buf bytes.Buffer

	config := DefaultConfig()
	config.Format = JSONFormat

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.SetOutput(&buf)

	testErr := errors.New("database connection failed")
	fields := map[string]interface{}{
		"database": "users",
		"retry":    3,
	}
	logger.LogError(testErr, "user_service", fields)

	output := buf.String()
	expectedFields := []string{
		"database connection failed", "user_service", "database", "retry", "error",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected field '%s' not found in output: %s", field, output)
		}
	}
}

func TestLogPerformance(t *testing.T) {
	var buf bytes.Buffer

	config := DefaultConfig()
	config.Format = JSONFormat

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.SetOutput(&buf)

	fields := map[string]interface{}{
		"query": "SELECT * FROM users",
	}
	logger.LogPerformance("database_query", 500*time.Millisecond, fields)

	output := buf.String()
	expectedFields := []string{
		"database_query", "500", "performance", "query",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected field '%s' not found in output: %s", field, output)
		}
	}
}

// Benchmark tests
func BenchmarkLoggerInfo(b *testing.B) {
	logger, _ := NewLogger(DefaultConfig())
	logger.SetOutput(os.Stdout) // Use stdout to avoid file I/O overhead

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark test message")
	}
}

func BenchmarkLoggerWithFields(b *testing.B) {
	logger, _ := NewLogger(DefaultConfig())
	logger.SetOutput(os.Stdout)

	fields := map[string]interface{}{
		"user_id":    "12345",
		"action":     "test",
		"timestamp":  time.Now(),
		"request_id": "req-123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithFields(fields).Info("Benchmark test with fields")
	}
}

func BenchmarkLoggerJSON(b *testing.B) {
	config := DefaultConfig()
	config.Format = JSONFormat
	logger, _ := NewLogger(config)
	logger.SetOutput(os.Stdout)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark JSON test message")
	}
}
