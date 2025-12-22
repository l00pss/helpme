package o4g_logger

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/sirupsen/logrus"
)

// ContextKey is the type for context keys
type ContextKey string

const (
	// LoggerContextKey is the context key for logger
	LoggerContextKey ContextKey = "logger"
	// RequestIDKey is the context key for wrapper ID
	RequestIDKey ContextKey = "request_id"
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
)

// FromContext extracts logger from context
func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(LoggerContextKey).(*Logger); ok {
		return logger
	}
	return GetDefaultLogger()
}

// ToContext adds logger to context
func ToContext(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, LoggerContextKey, logger)
}

// WithRequestID adds wrapper ID to context and returns logger with wrapper ID field
func WithRequestID(ctx context.Context, requestID string) (*Logger, context.Context) {
	ctx = context.WithValue(ctx, RequestIDKey, requestID)
	logger := FromContext(ctx).WithField("request_id", requestID)
	return &Logger{Logger: logger.Logger, config: GetDefaultLogger().config}, ctx
}

// WithUserID adds user ID to context and returns logger with user ID field
func WithUserID(ctx context.Context, userID string) (*Logger, context.Context) {
	ctx = context.WithValue(ctx, UserIDKey, userID)
	logger := FromContext(ctx).WithField("user_id", userID)
	return &Logger{Logger: logger.Logger, config: GetDefaultLogger().config}, ctx
}

// Timer is a utility for measuring operation duration
type Timer struct {
	start  time.Time
	logger *Logger
	name   string
	fields map[string]interface{}
}

// NewTimer creates a new timer
func NewTimer(logger *Logger, name string, fields ...map[string]interface{}) *Timer {
	timer := &Timer{
		start:  time.Now(),
		logger: logger,
		name:   name,
		fields: make(map[string]interface{}),
	}

	if len(fields) > 0 {
		timer.fields = fields[0]
	}

	return timer
}

// Stop stops the timer and logs the duration
func (t *Timer) Stop() time.Duration {
	duration := time.Since(t.start)

	logFields := map[string]interface{}{
		"operation":   t.name,
		"duration_ms": duration.Milliseconds(),
		"type":        "timer",
	}

	// Merge additional fields
	for k, v := range t.fields {
		logFields[k] = v
	}

	t.logger.WithFields(logFields).Info(fmt.Sprintf("Operation completed: %s", t.name))
	return duration
}

// Stopf stops the timer and logs with formatted message
func (t *Timer) Stopf(format string, args ...interface{}) time.Duration {
	duration := time.Since(t.start)

	logFields := map[string]interface{}{
		"operation":   t.name,
		"duration_ms": duration.Milliseconds(),
		"type":        "timer",
	}

	// Merge additional fields
	for k, v := range t.fields {
		logFields[k] = v
	}

	message := fmt.Sprintf(format, args...)
	t.logger.WithFields(logFields).Info(message)
	return duration
}

// Hook is a custom logrus hook for additional processing
type Hook struct {
	Writer io.Writer
	Level  logrus.Level
}

// NewHook creates a new hook
func NewHook(writer io.Writer, level logrus.Level) *Hook {
	return &Hook{
		Writer: writer,
		Level:  level,
	}
}

// Levels returns the levels this hook should be fired for
func (h *Hook) Levels() []logrus.Level {
	levels := make([]logrus.Level, 0)
	for _, level := range logrus.AllLevels {
		if level <= h.Level {
			levels = append(levels, level)
		}
	}
	return levels
}

// Fire is called when a log entry is fired
func (h *Hook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	_, err = h.Writer.Write([]byte(line))
	return err
}

// AddHook adds a hook to the logger
func (l *Logger) AddHook(hook logrus.Hook) {
	l.Logger.AddHook(hook)
}

// Sampling logger for high-frequency logs
type SamplingLogger struct {
	*Logger
	sampleRate float64
	counter    int64
}

// NewSamplingLogger creates a logger that samples log entries
func NewSamplingLogger(logger *Logger, sampleRate float64) *SamplingLogger {
	return &SamplingLogger{
		Logger:     logger,
		sampleRate: sampleRate,
		counter:    0,
	}
}

// shouldSample determines if this log entry should be sampled
func (s *SamplingLogger) shouldSample() bool {
	s.counter++
	return float64(s.counter)*s.sampleRate >= 1.0
}

// Info logs with sampling
func (s *SamplingLogger) Info(msg string) {
	if s.shouldSample() {
		s.Logger.Info(msg)
		s.counter = 0
	}
}

// Debug logs with sampling
func (s *SamplingLogger) Debug(msg string) {
	if s.shouldSample() {
		s.Logger.Debug(msg)
		s.counter = 0
	}
}

// Utility functions for common logging patterns

// LogStartup logs application startup information
func LogStartup(logger *Logger, serviceName, version, environment string, config map[string]interface{}) {
	logger.WithFields(map[string]interface{}{
		"service":     serviceName,
		"version":     version,
		"environment": environment,
		"config":      config,
		"type":        "startup",
	}).Info("Application starting up")
}

// LogShutdown logs application shutdown information
func LogShutdown(logger *Logger, serviceName string, graceful bool) {
	logger.WithFields(map[string]interface{}{
		"service":  serviceName,
		"graceful": graceful,
		"type":     "shutdown",
	}).Info("Application shutting down")
}

// LogPanic logs panic information with stack trace
func LogPanic(logger *Logger, recovered interface{}, stack []byte) {
	logger.WithFields(map[string]interface{}{
		"panic": recovered,
		"stack": string(stack),
		"type":  "panic",
	}).Error("Panic recovered")
}
