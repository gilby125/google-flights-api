package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

// Logger provides structured logging with context support
type Logger struct {
	logger *slog.Logger
}

// Config holds logger configuration
type Config struct {
	Level  string `json:"level"`
	Format string `json:"format"` // "json" or "text"
}

// New creates a new structured logger
func New(config Config) *Logger {
	var level slog.Level
	switch strings.ToLower(config.Level) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if config.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return &Logger{
		logger: slog.New(handler),
	}
}

// WithContext returns a logger with context values
func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{
		logger: l.logger,
	}
}

// WithFields returns a logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{
		logger: l.logger.With(args...),
	}
}

// WithField returns a logger with a single additional field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		logger: l.logger.With(key, value),
	}
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.logger.Debug(msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.logger.Warn(msg, args...)
}

// Error logs an error message
func (l *Logger) Error(err error, msg string, args ...interface{}) {
	if err != nil {
		args = append(args, "error", err)
	}
	l.logger.Error(msg, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(err error, msg string, args ...interface{}) {
	if err != nil {
		args = append(args, "error", err)
	}
	l.logger.Error(msg, args...)
	os.Exit(1)
}

// Default logger instance
var defaultLogger *Logger

// Init initializes the default logger
func Init(config Config) {
	defaultLogger = New(config)
}

// Global convenience functions that use the default logger
func Info(msg string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Info(msg, args...)
	}
}

func Debug(msg string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Debug(msg, args...)
	}
}

func Warn(msg string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Warn(msg, args...)
	}
}

func Error(err error, msg string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Error(err, msg, args...)
	}
}

func Fatal(err error, msg string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Fatal(err, msg, args...)
	}
}

func WithFields(fields map[string]interface{}) *Logger {
	if defaultLogger != nil {
		return defaultLogger.WithFields(fields)
	}
	return New(Config{Level: "info", Format: "text"})
}

func WithField(key string, value interface{}) *Logger {
	if defaultLogger != nil {
		return defaultLogger.WithField(key, value)
	}
	return New(Config{Level: "info", Format: "text"})
}

func WithContext(ctx context.Context) *Logger {
	if defaultLogger != nil {
		return defaultLogger.WithContext(ctx)
	}
	return New(Config{Level: "info", Format: "text"})
}