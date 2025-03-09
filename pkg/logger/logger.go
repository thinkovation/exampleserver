package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	defaultLogger *Logger
	once          sync.Once
)

// LoggerInterface defines the interface for logging operations
type LoggerInterface interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})
	WithFields(fields map[string]interface{}) LoggerInterface
	SetDebug(enabled bool)
	GetLogFile() string
	AddPlugin(plugin LogPlugin) error
}

// Logger is the main logger
type Logger struct {
	logger  *log.Logger
	debug   bool
	logFile string
	writer  *lumberjack.Logger
	plugins []LogPlugin
	mu      sync.RWMutex
}

// Default returns the default logger instance
func Default() LoggerInterface {
	if defaultLogger == nil {
		panic("logger not initialized - call Initialize first")
	}
	return defaultLogger
}

// Debug logs a debug message using the default logger
func Debug(format string, args ...interface{}) {
	Default().Debug(format, args...)
}

// Info logs an info message using the default logger
func Info(format string, args ...interface{}) {
	Default().Info(format, args...)
}

// Warn logs a warning message using the default logger
func Warn(format string, args ...interface{}) {
	Default().Warn(format, args...)
}

// Error logs an error message using the default logger
func Error(format string, args ...interface{}) {
	Default().Error(format, args...)
}

// Fatal logs a fatal message using the default logger and exits
func Fatal(format string, args ...interface{}) {
	Default().Fatal(format, args...)
}

// SetDebug enables or disables debug logging on the default logger
func SetDebug(enabled bool) {
	Default().SetDebug(enabled)
}

// WithFields adds fields to the default logger
func WithFields(fields map[string]interface{}) LoggerInterface {
	return Default().WithFields(fields)
}

// New creates a new logger
func New(config *LogConfig) (*Logger, error) {
	// Set up writers for the logger
	var writers []io.Writer

	// Set up rotating file writer
	rotator := &lumberjack.Logger{
		Filename:   config.LogFile,
		MaxSize:    config.Rotation.MaxSize,
		MaxAge:     config.Rotation.MaxAge,
		MaxBackups: config.Rotation.MaxBackups,
		Compress:   config.Rotation.Compress,
	}
	writers = append(writers, rotator)

	// Add stdout if configured
	if config.LogToStdout {
		writers = append(writers, os.Stdout)
	}

	return &Logger{
		logger:  log.New(io.MultiWriter(writers...), "", log.LstdFlags),
		debug:   config.Debug,
		logFile: config.LogFile,
		writer:  rotator,
	}, nil
}

// Close ensures any buffered logs are written and files are properly closed
func (l *Logger) Close() error {
	if l.writer != nil {
		return l.writer.Close()
	}
	return nil
}

// AddPlugin adds a new log plugin
func (l *Logger) AddPlugin(plugin LogPlugin) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := plugin.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	l.plugins = append(l.plugins, plugin)
	fmt.Println("Added plugin", plugin)
	return nil
}

// RemovePlugin removes a plugin
func (l *Logger) RemovePlugin(plugin LogPlugin) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	for i, p := range l.plugins {
		if p == plugin {
			if err := p.Close(); err != nil {
				return fmt.Errorf("failed to close plugin: %w", err)
			}
			l.plugins = append(l.plugins[:i], l.plugins[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("plugin not found")
}

// Modify logWithSource to handle plugins
func (l *Logger) logWithSource(level, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	var source string
	var line int
	if level == "DEBUG" && l.debug {
		_, file, lineNum, ok := runtime.Caller(2)
		if ok {
			if rel, err := filepath.Rel(os.Getenv("PWD"), file); err == nil {
				file = rel
			}
			source = file
			line = lineNum
		}
	}

	// Create log entry
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
		Source:    source,
		Line:      line,
	}

	// Handle plugins
	l.mu.RLock()
	plugins := l.plugins
	l.mu.RUnlock()

	for _, plugin := range plugins {
		fmt.Println("Checking plugins")
		if plugin.ShouldHandle(entry) {
			fmt.Println("Plugin should handle - So lets go")
			go func(p LogPlugin, e LogEntry) {
				if err := p.Handle(e); err != nil {
					l.logger.Printf("[ERROR] Plugin error: %v", err)
				}
			}(plugin, entry)
		}
	}

	// Log to standard outputs
	if source != "" {
		l.logger.Printf("[%s] %s:%d: %s", level, source, line, msg)
	} else {
		l.logger.Printf("[%s] %s", level, msg)
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	if !l.debug {
		return
	}
	l.logWithSource("DEBUG", format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.logWithSource("INFO", format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.logWithSource("WARN", format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.logWithSource("ERROR", format, args...)
}

func (l *Logger) Fatal(format string, args ...interface{}) {
	l.logWithSource("FATAL", format, args...)
	os.Exit(1)
}

func (l *Logger) WithFields(fields map[string]interface{}) LoggerInterface {
	return l // Fields not supported in basic logger
}

func (l *Logger) SetDebug(enabled bool) {
	l.debug = enabled
}

func (l *Logger) GetLogFile() string {
	return l.logFile
}
