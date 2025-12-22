package o4g_logger

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestColoredFormatterFormat(t *testing.T) {
	formatter := &ColoredFormatter{
		TimestampFormat: time.RFC3339,
		EnableColors:    true,
		ServiceName:     "test-service",
		Environment:     "test",
		EnableCaller:    false,
	}

	entry := &logrus.Entry{
		Time:    time.Now(),
		Level:   logrus.InfoLevel,
		Message: "Test log message",
		Data:    logrus.Fields{},
	}

	formatted, err := formatter.Format(entry)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	output := string(formatted)
	if !strings.Contains(output, "Test log message") {
		t.Error("Formatted output should contain the log message")
	}
	if !strings.Contains(output, "test-service") {
		t.Error("Formatted output should contain service name")
	}
}

func TestColoredFormatterWithoutColors(t *testing.T) {
	formatter := &ColoredFormatter{
		TimestampFormat: time.RFC3339,
		EnableColors:    false,
		ServiceName:     "test-service",
		Environment:     "production",
		EnableCaller:    false,
	}

	entry := &logrus.Entry{
		Time:    time.Now(),
		Level:   logrus.WarnLevel,
		Message: "Warning message",
		Data:    logrus.Fields{},
	}

	formatted, err := formatter.Format(entry)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	output := string(formatted)

	// Should not contain ANSI color codes when colors are disabled
	colorCodes := []string{Reset, Red, Green, Yellow, Blue, Magenta, Cyan}
	for _, code := range colorCodes {
		if strings.Contains(output, code) {
			t.Errorf("Output should not contain color code '%s' when colors disabled", code)
		}
	}

	if !strings.Contains(output, "Warning message") {
		t.Error("Formatted output should contain the log message")
	}
}

func TestColoredFormatterWithCaller(t *testing.T) {
	formatter := &ColoredFormatter{
		TimestampFormat: time.RFC3339,
		EnableColors:    false,
		ServiceName:     "test-service",
		Environment:     "test",
		EnableCaller:    true,
	}

	entry := &logrus.Entry{
		Time:    time.Now(),
		Level:   logrus.ErrorLevel,
		Message: "Error message",
		Data: logrus.Fields{
			"caller_func": "TestFunction",
			"caller_file": "test.go:123",
		},
		Caller: &runtime.Frame{
			Function: "TestFunction",
			File:     "test.go",
			Line:     123,
		},
	}

	formatted, err := formatter.Format(entry)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	output := string(formatted)
	if !strings.Contains(output, "Error message") {
		t.Error("Formatted output should contain the log message")
	}
}

func TestColoredFormatterWithFields(t *testing.T) {
	formatter := &ColoredFormatter{
		TimestampFormat: time.RFC3339,
		EnableColors:    false,
		ServiceName:     "test-service",
		Environment:     "test",
		EnableCaller:    false,
	}

	entry := &logrus.Entry{
		Time:    time.Now(),
		Level:   logrus.InfoLevel,
		Message: "Message with fields",
		Data: logrus.Fields{
			"user_id":    "12345",
			"request_id": "req-678",
			"action":     "login",
		},
	}

	formatted, err := formatter.Format(entry)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	output := string(formatted)
	if !strings.Contains(output, "Message with fields") {
		t.Error("Formatted output should contain the log message")
	}
	if !strings.Contains(output, "user_id") {
		t.Error("Formatted output should contain field names")
	}
}

func TestColoredFormatterDifferentLevels(t *testing.T) {
	formatter := &ColoredFormatter{
		TimestampFormat: time.RFC3339,
		EnableColors:    true,
		ServiceName:     "test-service",
		Environment:     "test",
		EnableCaller:    false,
	}

	levels := []logrus.Level{
		logrus.TraceLevel,
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}

	for _, level := range levels {
		entry := &logrus.Entry{
			Time:    time.Now(),
			Level:   level,
			Message: "Test message for " + level.String(),
			Data:    logrus.Fields{},
		}

		formatted, err := formatter.Format(entry)
		if err != nil {
			t.Errorf("Expected no error for level %s, got: %v", level.String(), err)
		}

		output := string(formatted)
		if !strings.Contains(output, "Test message for "+level.String()) {
			t.Errorf("Formatted output should contain the log message for level %s", level.String())
		}

		// Different levels should have different colors when colors are enabled
		if level == logrus.ErrorLevel && !strings.Contains(output, Red) {
			t.Error("Error level should contain red color code")
		}
	}
}

func TestColoredFormatterTimestamp(t *testing.T) {
	customFormat := "2006-01-02 15:04:05"
	formatter := &ColoredFormatter{
		TimestampFormat: customFormat,
		EnableColors:    false,
		ServiceName:     "test-service",
		Environment:     "test",
		EnableCaller:    false,
	}

	testTime := time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC)
	entry := &logrus.Entry{
		Time:    testTime,
		Level:   logrus.InfoLevel,
		Message: "Test timestamp formatting",
		Data:    logrus.Fields{},
	}

	formatted, err := formatter.Format(entry)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	output := string(formatted)
	expectedTimestamp := testTime.Format(customFormat)
	if !strings.Contains(output, expectedTimestamp) {
		t.Errorf("Expected timestamp '%s' not found in output: %s", expectedTimestamp, output)
	}
}

func TestColorCodes(t *testing.T) {
	// Test that color constants are defined correctly
	colorTests := []struct {
		name  string
		color string
	}{
		{"Reset", Reset},
		{"Red", Red},
		{"Green", Green},
		{"Yellow", Yellow},
		{"Blue", Blue},
		{"Magenta", Magenta},
		{"Cyan", Cyan},
		{"White", White},
		{"Gray", Gray},
		{"BoldRed", BoldRed},
		{"BoldGreen", BoldGreen},
		{"BoldYellow", BoldYellow},
		{"BoldBlue", BoldBlue},
		{"BoldMagenta", BoldMagenta},
		{"BoldCyan", BoldCyan},
		{"BoldWhite", BoldWhite},
		{"HiRed", HiRed},
		{"HiGreen", HiGreen},
		{"HiYellow", HiYellow},
		{"HiBlue", HiBlue},
		{"HiMagenta", HiMagenta},
	}

	for _, tt := range colorTests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.color == "" {
				t.Errorf("Color constant %s should not be empty", tt.name)
			}
			if !strings.HasPrefix(tt.color, "\033[") {
				t.Errorf("Color constant %s should start with ANSI escape sequence", tt.name)
			}
		})
	}
}

// Benchmark tests for formatter
func BenchmarkColoredFormatterWithColors(b *testing.B) {
	formatter := &ColoredFormatter{
		TimestampFormat: time.RFC3339,
		EnableColors:    true,
		ServiceName:     "benchmark-service",
		Environment:     "test",
		EnableCaller:    false,
	}

	entry := &logrus.Entry{
		Time:    time.Now(),
		Level:   logrus.InfoLevel,
		Message: "Benchmark test message",
		Data:    logrus.Fields{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = formatter.Format(entry)
	}
}

func BenchmarkColoredFormatterWithoutColors(b *testing.B) {
	formatter := &ColoredFormatter{
		TimestampFormat: time.RFC3339,
		EnableColors:    false,
		ServiceName:     "benchmark-service",
		Environment:     "test",
		EnableCaller:    false,
	}

	entry := &logrus.Entry{
		Time:    time.Now(),
		Level:   logrus.InfoLevel,
		Message: "Benchmark test message",
		Data:    logrus.Fields{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = formatter.Format(entry)
	}
}

func BenchmarkColoredFormatterWithFields(b *testing.B) {
	formatter := &ColoredFormatter{
		TimestampFormat: time.RFC3339,
		EnableColors:    true,
		ServiceName:     "benchmark-service",
		Environment:     "test",
		EnableCaller:    false,
	}

	entry := &logrus.Entry{
		Time:    time.Now(),
		Level:   logrus.InfoLevel,
		Message: "Benchmark test message with fields",
		Data: logrus.Fields{
			"user_id":    "12345",
			"request_id": "req-678",
			"action":     "benchmark",
			"timestamp":  time.Now(),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = formatter.Format(entry)
	}
}
