package o4g_logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

// LogLevel represents different log levels
type LogLevel string

const (
	TraceLevel LogLevel = "trace"
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
	PanicLevel LogLevel = "panic"
)

// OutputFormat represents different output formats
type OutputFormat string

const (
	TextFormat OutputFormat = "text"
	JSONFormat OutputFormat = "json"
)

// Config holds the logger configuration
type Config struct {
	Level           LogLevel     `yaml:"level" json:"level"`
	Format          OutputFormat `yaml:"format" json:"format"`
	Output          string       `yaml:"output" json:"output"` // "stdout", "stderr", or file path
	EnableCaller    bool         `yaml:"enable_caller" json:"enable_caller"`
	EnableColors    bool         `yaml:"enable_colors" json:"enable_colors"`
	ServiceName     string       `yaml:"service_name" json:"service_name"`
	Environment     string       `yaml:"environment" json:"environment"`
	TimestampFormat string       `yaml:"timestamp_format" json:"timestamp_format"`
}

// Logger wraps logrus with additional functionality
type Logger struct {
	*logrus.Logger
	config Config
}

// DefaultConfig returns a default logger configuration
func DefaultConfig() Config {
	return Config{
		Level:           InfoLevel,
		Format:          TextFormat,
		Output:          "stdout",
		EnableCaller:    true,
		EnableColors:    true,
		ServiceName:     "gatekeeper",
		Environment:     "development",
		TimestampFormat: time.RFC3339,
	}
}

// NewLogger creates a new logger instance with the given configuration
func NewLogger(config Config) (*Logger, error) {
	log := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(string(config.Level))
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %v", err)
	}
	log.SetLevel(level)

	// Set output
	switch config.Output {
	case "stdout":
		log.SetOutput(os.Stdout)
	case "stderr":
		log.SetOutput(os.Stderr)
	default:
		// Assume it's a file path
		file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %v", err)
		}
		log.SetOutput(file)
	}

	// Set formatter
	switch config.Format {
	case JSONFormat:
		formatter := &logrus.JSONFormatter{
			TimestampFormat: config.TimestampFormat,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "function",
				logrus.FieldKeyFile:  "file",
			},
		}
		log.SetFormatter(formatter)
	default:
		// Use our custom colored formatter for text output
		formatter := &ColoredFormatter{
			TimestampFormat: config.TimestampFormat,
			EnableColors:    config.EnableColors,
			ServiceName:     config.ServiceName,
			Environment:     config.Environment,
			EnableCaller:    config.EnableCaller,
		}
		log.SetFormatter(formatter)
	}

	// Enable caller info if requested
	log.SetReportCaller(config.EnableCaller)

	logger := &Logger{
		Logger: log,
		config: config,
	}

	return logger, nil
}

// WithFields creates a new logger entry with the given fields
func (l *Logger) WithFields(fields map[string]interface{}) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields(fields))
}

// WithField creates a new logger entry with a single field
func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.Logger.WithField(key, value)
}

// WithError creates a new logger entry with an error field
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}

// WithContext creates contextual logger entries
func (l *Logger) WithContext() *logrus.Entry {
	entry := l.Logger.WithFields(logrus.Fields{
		"service":     l.config.ServiceName,
		"environment": l.config.Environment,
	})

	if l.config.EnableCaller {
		if pc, file, line, ok := runtime.Caller(1); ok {
			funcName := runtime.FuncForPC(pc).Name()
			entry = entry.WithFields(logrus.Fields{
				"caller_func": filepath.Base(funcName),
				"caller_file": fmt.Sprintf("%s:%d", filepath.Base(file), line),
			})
		}
	}

	return entry
}

// HTTP wrapper logging helpers
func (l *Logger) LogHTTPRequest(method, path, userAgent, clientIP string, statusCode, responseTime int) {
	l.WithFields(map[string]interface{}{
		"method":        method,
		"path":          path,
		"user_agent":    userAgent,
		"client_ip":     clientIP,
		"status_code":   statusCode,
		"response_time": responseTime,
		"type":          "http_request",
	}).Info("HTTP wrapper processed")
}

// Database operation logging helpers
func (l *Logger) LogDBOperation(operation, table string, duration time.Duration, rowsAffected int64) {
	l.WithFields(map[string]interface{}{
		"operation":     operation,
		"table":         table,
		"duration_ms":   duration.Milliseconds(),
		"rows_affected": rowsAffected,
		"type":          "db_operation",
	}).Debug("Database operation completed")
}

// Authentication logging helpers
func (l *Logger) LogAuthEvent(event, userID, clientIP string, success bool) {
	level := logrus.InfoLevel
	if !success {
		level = logrus.WarnLevel
	}

	l.WithFields(map[string]interface{}{
		"event":     event,
		"user_id":   userID,
		"client_ip": clientIP,
		"success":   success,
		"type":      "auth_event",
	}).Log(level, fmt.Sprintf("Authentication event: %s", event))
}

// Audit logging helpers
func (l *Logger) LogAuditEvent(action, resource, userID string, details map[string]interface{}) {
	fields := map[string]interface{}{
		"action":    action,
		"resource":  resource,
		"user_id":   userID,
		"type":      "audit_event",
		"timestamp": time.Now().UTC(),
	}

	// Merge additional details
	for k, v := range details {
		fields[k] = v
	}

	l.WithFields(fields).Info("Audit event")
}

// Error logging with stack trace
func (l *Logger) LogError(err error, context string, fields map[string]interface{}) {
	logFields := map[string]interface{}{
		"error":   err.Error(),
		"context": context,
		"type":    "error",
	}

	// Merge additional fields
	for k, v := range fields {
		logFields[k] = v
	}

	l.WithFields(logFields).Error("GetError occurred")
}

// Performance logging
func (l *Logger) LogPerformance(operation string, duration time.Duration, fields map[string]interface{}) {
	logFields := map[string]interface{}{
		"operation":   operation,
		"duration_ms": duration.Milliseconds(),
		"type":        "performance",
	}

	// Merge additional fields
	for k, v := range fields {
		logFields[k] = v
	}

	var level logrus.Level
	switch {
	case duration > 5*time.Second:
		level = logrus.WarnLevel
	case duration > 1*time.Second:
		level = logrus.InfoLevel
	default:
		level = logrus.DebugLevel
	}

	l.WithFields(logFields).Log(level, fmt.Sprintf("Operation completed: %s", operation))
}

// SetOutput changes the output destination
func (l *Logger) SetOutput(output io.Writer) {
	l.Logger.SetOutput(output)
}

// SetLevel changes the log level
func (l *Logger) SetLevel(level LogLevel) error {
	logrusLevel, err := logrus.ParseLevel(string(level))
	if err != nil {
		return err
	}
	l.Logger.SetLevel(logrusLevel)
	l.config.Level = level
	return nil
}

// GetLevel returns current log level
func (l *Logger) GetLevel() LogLevel {
	return l.config.Level
}

// IsLevelEnabled checks if a log level is enabled
func (l *Logger) IsLevelEnabled(level LogLevel) bool {
	logrusLevel, err := logrus.ParseLevel(string(level))
	if err != nil {
		return false
	}
	return l.Logger.IsLevelEnabled(logrusLevel)
}

// Structured logging methods with context
func (l *Logger) TraceWithContext(msg string, fields ...map[string]interface{}) {
	entry := l.WithContext()
	if len(fields) > 0 {
		entry = entry.WithFields(logrus.Fields(fields[0]))
	}
	entry.Trace(msg)
}

func (l *Logger) DebugWithContext(msg string, fields ...map[string]interface{}) {
	entry := l.WithContext()
	if len(fields) > 0 {
		entry = entry.WithFields(logrus.Fields(fields[0]))
	}
	entry.Debug(msg)
}

func (l *Logger) InfoWithContext(msg string, fields ...map[string]interface{}) {
	entry := l.WithContext()
	if len(fields) > 0 {
		entry = entry.WithFields(logrus.Fields(fields[0]))
	}
	entry.Info(msg)
}

func (l *Logger) WarnWithContext(msg string, fields ...map[string]interface{}) {
	entry := l.WithContext()
	if len(fields) > 0 {
		entry = entry.WithFields(logrus.Fields(fields[0]))
	}
	entry.Warn(msg)
}

func (l *Logger) ErrorWithContext(msg string, fields ...map[string]interface{}) {
	entry := l.WithContext()
	if len(fields) > 0 {
		entry = entry.WithFields(logrus.Fields(fields[0]))
	}
	entry.Error(msg)
}

func (l *Logger) FatalWithContext(msg string, fields ...map[string]interface{}) {
	entry := l.WithContext()
	if len(fields) > 0 {
		entry = entry.WithFields(logrus.Fields(fields[0]))
	}
	entry.Fatal(msg)
}

func (l *Logger) PanicWithContext(msg string, fields ...map[string]interface{}) {
	entry := l.WithContext()
	if len(fields) > 0 {
		entry = entry.WithFields(logrus.Fields(fields[0]))
	}
	entry.Panic(msg)
}
