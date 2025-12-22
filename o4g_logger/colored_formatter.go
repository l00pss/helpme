package o4g_logger

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	_ "time"

	"github.com/sirupsen/logrus"
)

// ANSI color codes
const (
	Reset = "\033[0m"

	// Regular colors
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	Gray    = "\033[90m"

	// Bold colors
	BoldRed     = "\033[1;31m"
	BoldGreen   = "\033[1;32m"
	BoldYellow  = "\033[1;33m"
	BoldBlue    = "\033[1;34m"
	BoldMagenta = "\033[1;35m"
	BoldCyan    = "\033[1;36m"
	BoldWhite   = "\033[1;37m"

	// High intensity colors
	HiRed     = "\033[91m"
	HiGreen   = "\033[92m"
	HiYellow  = "\033[93m"
	HiBlue    = "\033[94m"
	HiMagenta = "\033[95m"
	HiCyan    = "\033[96m"
	HiWhite   = "\033[97m"
)

// ColoredFormatter is a custom formatter that provides colored output similar to Spring Boot
type ColoredFormatter struct {
	TimestampFormat string
	EnableColors    bool
	ServiceName     string
	Environment     string
	EnableCaller    bool
}

// Format formats the log entry with colors
func (f *ColoredFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	if !f.EnableColors {
		return f.formatPlain(entry)
	}

	return f.formatColored(entry)
}

// formatColored creates a colored log entry similar to Spring Boot
func (f *ColoredFormatter) formatColored(entry *logrus.Entry) ([]byte, error) {
	var b strings.Builder

	// Timestamp with gray color
	timestamp := entry.Time.Format(f.getTimestampFormat())
	b.WriteString(fmt.Sprintf("%s%s%s ", Gray, timestamp, Reset))

	// Log level with appropriate color and padding
	level := f.formatLevelColored(entry.Level)
	b.WriteString(level)
	b.WriteString(" ")

	// Process ID in magenta
	pid := os.Getpid()
	b.WriteString(fmt.Sprintf("%s%d%s ", Magenta, pid, Reset))

	// Separator
	b.WriteString(fmt.Sprintf("%s---%s ", Gray, Reset))

	// Thread/Goroutine info in cyan brackets
	goroutineID := getGoroutineID()
	b.WriteString(fmt.Sprintf("%s[%15s]%s ", Cyan, fmt.Sprintf("goroutine-%d", goroutineID), Reset))

	// Logger name (service.component) in cyan
	loggerName := f.getLoggerName(entry)
	b.WriteString(fmt.Sprintf("%s%-40s%s : ", Cyan, loggerName, Reset))

	// Message with level-appropriate color
	messageColor := f.getMessageColor(entry.Level)
	b.WriteString(fmt.Sprintf("%s%s%s", messageColor, entry.Message, Reset))

	// Add fields if present
	if len(entry.Data) > 0 {
		fieldsStr := f.formatFieldsColored(entry.Data)
		if fieldsStr != "" {
			b.WriteString(" ")
			b.WriteString(fieldsStr)
		}
	}

	// Add caller info if enabled
	if f.EnableCaller && entry.HasCaller() {
		filename := strings.Split(entry.Caller.File, "/")
		shortFile := filename[len(filename)-1]
		b.WriteString(fmt.Sprintf(" %s[%s:%d]%s", Gray, shortFile, entry.Caller.Line, Reset))
	}

	b.WriteString("\n")
	return []byte(b.String()), nil
}

// formatPlain creates a plain text log entry
func (f *ColoredFormatter) formatPlain(entry *logrus.Entry) ([]byte, error) {
	var b strings.Builder

	// Timestamp
	timestamp := entry.Time.Format(f.getTimestampFormat())
	b.WriteString(fmt.Sprintf("%s ", timestamp))

	// Log level
	b.WriteString(fmt.Sprintf("%5s ", strings.ToUpper(entry.Level.String())))

	// Process ID
	pid := os.Getpid()
	b.WriteString(fmt.Sprintf("%d ", pid))

	// Separator
	b.WriteString("--- ")

	// Thread/Goroutine info
	goroutineID := getGoroutineID()
	b.WriteString(fmt.Sprintf("[%15s] ", fmt.Sprintf("goroutine-%d", goroutineID)))

	// Logger name
	loggerName := f.getLoggerName(entry)
	b.WriteString(fmt.Sprintf("%-40s : ", loggerName))

	// Message
	b.WriteString(entry.Message)

	// Add fields if present
	if len(entry.Data) > 0 {
		fieldsStr := f.formatFieldsPlain(entry.Data)
		if fieldsStr != "" {
			b.WriteString(" ")
			b.WriteString(fieldsStr)
		}
	}

	// Add caller info if enabled
	if f.EnableCaller && entry.HasCaller() {
		filename := strings.Split(entry.Caller.File, "/")
		shortFile := filename[len(filename)-1]
		b.WriteString(fmt.Sprintf(" [%s:%d]", shortFile, entry.Caller.Line))
	}

	b.WriteString("\n")
	return []byte(b.String()), nil
}

// formatLevelColored formats the log level with appropriate colors and padding
func (f *ColoredFormatter) formatLevelColored(level logrus.Level) string {
	switch level {
	case logrus.TraceLevel:
		return fmt.Sprintf("%sTRACE%s", BoldMagenta, Reset)
	case logrus.DebugLevel:
		return fmt.Sprintf("%sDEBUG%s", BoldBlue, Reset)
	case logrus.InfoLevel:
		return fmt.Sprintf("%s INFO%s", BoldGreen, Reset)
	case logrus.WarnLevel:
		return fmt.Sprintf("%s WARN%s", BoldYellow, Reset)
	case logrus.ErrorLevel:
		return fmt.Sprintf("%sERROR%s", BoldRed, Reset)
	case logrus.FatalLevel:
		return fmt.Sprintf("%sFATAL%s", HiRed, Reset)
	case logrus.PanicLevel:
		return fmt.Sprintf("%sPANIC%s", HiRed, Reset)
	default:
		return fmt.Sprintf("%5s", strings.ToUpper(level.String()))
	}
}

// getMessageColor returns the appropriate color for the message based on log level
func (f *ColoredFormatter) getMessageColor(level logrus.Level) string {
	switch level {
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		return HiRed
	case logrus.WarnLevel:
		return HiYellow
	case logrus.InfoLevel:
		return HiWhite
	case logrus.DebugLevel:
		return Blue
	case logrus.TraceLevel:
		return Gray
	default:
		return HiWhite
	}
}

// getLoggerName extracts or constructs the logger name
func (f *ColoredFormatter) getLoggerName(entry *logrus.Entry) string {
	if component, ok := entry.Data["component"]; ok {
		return fmt.Sprintf("%s.%v", f.ServiceName, component)
	}
	if module, ok := entry.Data["module"]; ok {
		return fmt.Sprintf("%s.%v", f.ServiceName, module)
	}
	if service, ok := entry.Data["service"]; ok {
		return fmt.Sprintf("%v", service)
	}

	// Try to get caller information for logger name
	if entry.HasCaller() {
		packagePath := entry.Caller.Function
		parts := strings.Split(packagePath, "/")
		if len(parts) > 0 {
			lastPart := parts[len(parts)-1]
			funcParts := strings.Split(lastPart, ".")
			if len(funcParts) > 1 {
				return fmt.Sprintf("%s.%s", f.ServiceName, funcParts[0])
			}
		}
	}

	return f.ServiceName
}

// formatFieldsColored formats the log fields with colors
func (f *ColoredFormatter) formatFieldsColored(fields map[string]interface{}) string {
	if len(fields) == 0 {
		return ""
	}

	var parts []string
	for key, value := range fields {
		// Skip internal fields
		if key == "component" || key == "module" || key == "service" {
			continue
		}

		formattedValue := fmt.Sprintf("%v", value)
		colored := fmt.Sprintf("%s%s%s=%s%v%s",
			HiCyan, key, Reset,
			HiWhite, formattedValue, Reset)
		parts = append(parts, colored)
	}

	if len(parts) == 0 {
		return ""
	}

	return fmt.Sprintf("%s{%s}%s", Gray, strings.Join(parts, ", "), Reset)
}

// formatFieldsPlain formats the log fields without colors
func (f *ColoredFormatter) formatFieldsPlain(fields map[string]interface{}) string {
	if len(fields) == 0 {
		return ""
	}

	var parts []string
	for key, value := range fields {
		// Skip internal fields
		if key == "component" || key == "module" || key == "service" {
			continue
		}

		parts = append(parts, fmt.Sprintf("%s=%v", key, value))
	}

	if len(parts) == 0 {
		return ""
	}

	return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
}

// getTimestampFormat returns the timestamp format to use
func (f *ColoredFormatter) getTimestampFormat() string {
	if f.TimestampFormat != "" {
		return f.TimestampFormat
	}
	return "2006-01-02 15:04:05.000"
}

// getGoroutineID returns a simple goroutine identifier
func getGoroutineID() int {
	return runtime.NumGoroutine()
}
