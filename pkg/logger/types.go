package logger

import (
	"context"

	"go.uber.org/zap"
)

// ContextKey is a type for context keys used by the logger
type ContextKey string

const (
	// RequestIDKey is the context key for request IDs
	RequestIDKey ContextKey = "request_id"

	// TraceIDKey is the context key for trace IDs
	TraceIDKey ContextKey = "trace_id"

	// UserIDKey is the context key for user IDs
	UserIDKey ContextKey = "user_id"
)

// Logger is the interface for structured logging
type Logger interface {
	// Core logging methods
	Debug(msg string, keysAndValues ...any)
	Info(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
	Fatal(msg string, keysAndValues ...any)

	// Context-Aware logging
	WithContext(ctx context.Context) Logger
	With(keysAndValues ...any) Logger
	WithError(err error) Logger

	// Named logger for component identification
	Named(name string) Logger

	// Sync flushes any buffered log entries
	Sync() error
}

// zapLogger wraps zap.SugaredLogger to implement the Logger interface
type zapLogger struct {
	sugar *zap.SugaredLogger
}

// Config holds logger configuration
type Config struct {
	Level       string   `json:"level" yaml:"level"`
	Format      string   `json:"format" yaml:"format"`
	OutputPaths []string `json:"output_paths" yaml:"output_paths"`
	ErrorPaths  []string `json:"error_paths" yaml:"error_paths"`
	Development bool     `json:"development" yaml:"development"`
	Caller      bool     `json:"caller" yaml:"caller"`
	Stacktrace  bool     `json:"stacktrace" yaml:"stacktrace"`
}

// DefaultConfig returns a default logger configuration
func DefaultConfig() Config {
	return Config{
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{"stdout"},
		ErrorPaths:  []string{"stderr"},
		Development: false,
		Caller:      true,
		Stacktrace:  true,
	}
}
