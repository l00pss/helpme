package o4g_logger

import (
	"sync"
)

var (
	defaultLogger *Logger
	once          sync.Once
)

// Init initializes the global logger with the given configuration
func Init(config Config) error {
	var err error
	once.Do(func() {
		defaultLogger, err = NewLogger(config)
	})
	return err
}

// GetDefaultLogger returns the global logger instance
func GetDefaultLogger() *Logger {
	if defaultLogger == nil {
		// Initialize with default config if not already initialized
		_ = Init(DefaultConfig())
	}
	return defaultLogger
}

// Global convenience functions that use the default logger

// Trace logs a message at trace level
func Trace(msg string) {
	GetDefaultLogger().Trace(msg)
}

// Debug logs a message at debug level
func Debug(msg string) {
	GetDefaultLogger().Debug(msg)
}

// Info logs a message at info level
func Info(msg ...string) {
	GetDefaultLogger().Info(msg)
}

// Warn logs a message at warn level
func Warn(msg ...string) {
	GetDefaultLogger().Warn(msg)
}

// Error logs a message at error level
func Error(msg ...string) {
	GetDefaultLogger().Error(msg)
}

// Fatal logs a message at fatal level and exits
func Fatal(msg ...string) {
	GetDefaultLogger().Fatal(msg)
}

// Panic logs a message at panic level and panics
func Panic(msg ...string) {
	GetDefaultLogger().Panic(msg)
}

// WithFields creates a new logger entry with the given fields
func WithFields(fields map[string]interface{}) *Logger {
	return &Logger{
		Logger: GetDefaultLogger().WithFields(fields).Logger,
		config: GetDefaultLogger().config,
	}
}

// WithField creates a new logger entry with a single field
func WithField(key string, value interface{}) *Logger {
	return &Logger{
		Logger: GetDefaultLogger().WithField(key, value).Logger,
		config: GetDefaultLogger().config,
	}
}

// WithError creates a new logger entry with an error field
func WithError(err error) *Logger {
	return &Logger{
		Logger: GetDefaultLogger().WithError(err).Logger,
		config: GetDefaultLogger().config,
	}
}

// LogHTTPRequest logs HTTP wrapper information using the global logger
func LogHTTPRequest(method, path, userAgent, clientIP string, statusCode, responseTime int) {
	GetDefaultLogger().LogHTTPRequest(method, path, userAgent, clientIP, statusCode, responseTime)
}

// LogAuthEvent logs authentication events using the global logger
func LogAuthEvent(event, userID, clientIP string, success bool) {
	GetDefaultLogger().LogAuthEvent(event, userID, clientIP, success)
}

// LogAuditEvent logs audit events using the global logger
func LogAuditEvent(action, resource, userID string, details map[string]interface{}) {
	GetDefaultLogger().LogAuditEvent(action, resource, userID, details)
}

// LogError logs exceptions with context using the global logger
func LogError(err error, context string, fields map[string]interface{}) {
	GetDefaultLogger().LogError(err, context, fields)
}
