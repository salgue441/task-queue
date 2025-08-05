package logger

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new logger instance with the given configuration
func New(cfg Config) Logger {
	zapConfig := zap.Config{
		Level:             parseLevel(cfg.Level),
		Development:       cfg.Development,
		DisableCaller:     !cfg.Caller,
		DisableStacktrace: !cfg.Stacktrace,
		Encoding:          cfg.Format,
		EncoderConfig:     getEncoderConfig(cfg.Format),
		OutputPaths:       cfg.OutputPaths,
		ErrorOutputPaths:  cfg.ErrorPaths,
		InitialFields: map[string]any{
			"app": "task-queue",
			"pid": os.Getpid(),
		},
	}

	logger, err := zapConfig.Build(
		zap.AddCallerSkip(1),
	)

	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}

	return &zapLogger{
		sugar: logger.Sugar(),
	}
}

// NewNop returns a no-op logger for testing
func NewNop() Logger {
	return &zapLogger{
		sugar: zap.NewNop().Sugar(),
	}
}

// getEncoderConfig returns the appropriate encoder config based on format
func getEncoderConfig(format string) zapcore.EncoderConfig {
	base := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if format == "console" {
		base.EncodeLevel = zapcore.CapitalColorLevelEncoder
		base.EncodeTime = localTimeEncoder
		base.ConsoleSeparator = " | "
	}

	return base
}

// localTimeEncoder encodes time in local format for console output
func localTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// parseLevel parses the log level string
func parseLevel(level string) zap.AtomicLevel {
	switch level {
	case "debug":
		return zap.NewAtomicLevelAt(zap.DebugLevel)

	case "info":
		return zap.NewAtomicLevelAt(zap.InfoLevel)

	case "warn":
		return zap.NewAtomicLevelAt(zap.WarnLevel)

	case "error":
		return zap.NewAtomicLevelAt(zap.ErrorLevel)

	case "fatal":
		return zap.NewAtomicLevelAt(zap.FatalLevel)

	default:
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	}
}

// Debug logs a debug message
func (l *zapLogger) Debug(msg string, keysAndValues ...any) {
	l.sugar.Debugw(msg, keysAndValues...)
}

// Info logs an info message
func (l *zapLogger) Info(msg string, keysAndValues ...any) {
	l.sugar.Infow(msg, keysAndValues...)
}

// Warn logs a warning message
func (l *zapLogger) Warn(msg string, keysAndValues ...any) {
	l.sugar.Warnw(msg, keysAndValues...)
}

// Error logs an error message
func (l *zapLogger) Error(msg string, keysAndValues ...any) {
	l.sugar.Errorw(msg, keysAndValues...)
}

// Fatal logs a fatal message and exits the program
func (l *zapLogger) Fatal(msg string, keysAndValues ...any) {
	l.sugar.Fatalw(msg, keysAndValues...)
}

// WithContext returns a logger with context values
func (l *zapLogger) WithContext(ctx context.Context) Logger {
	if ctx == nil {
		return l
	}

	fields := make([]interface{}, 0, 6)
	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		fields = append(fields, "request_id", requestID)
	}

	if traceID := ctx.Value(TraceIDKey); traceID != nil {
		fields = append(fields, "trace_id", traceID)
	}

	if userID := ctx.Value(UserIDKey); userID != nil {
		fields = append(fields, "user_id", userID)
	}

	if len(fields) > 0 {
		return &zapLogger{
			sugar: l.sugar.With(fields...),
		}
	}

	return l
}

// With returns a logger with additional fields
func (l *zapLogger) With(keysAndValues ...interface{}) Logger {
	return &zapLogger{
		sugar: l.sugar.With(keysAndValues...),
	}
}

// WithError returns a logger with an error field
func (l *zapLogger) WithError(err error) Logger {
	if err == nil {
		return l
	}

	return &zapLogger{
		sugar: l.sugar.With("error", err.Error()),
	}
}

// Named returns a named logger
func (l *zapLogger) Named(name string) Logger {
	return &zapLogger{
		sugar: l.sugar.Named(name),
	}
}

// Sync flushes any buffered log entries
func (l *zapLogger) Sync() error {
	return l.sugar.Sync()
}

// Global logger instance
var globalLogger Logger

func init() {
	globalLogger = New(DefaultConfig())
}

// SetGlobal sets the global logger instance
func SetGlobal(l Logger) {
	globalLogger = l
}

// Global returns the global logger instance
func Global() Logger {
	return globalLogger
}

// Helper functions for global logger

// Debug logs a debug message using the global logger
func Debug(msg string, keysAndValues ...interface{}) {
	globalLogger.Debug(msg, keysAndValues...)
}

// Info logs an info message using the global logger
func Info(msg string, keysAndValues ...interface{}) {
	globalLogger.Info(msg, keysAndValues...)
}

// Warn logs a warning message using the global logger
func Warn(msg string, keysAndValues ...interface{}) {
	globalLogger.Warn(msg, keysAndValues...)
}

// Error logs an error message using the global logger
func Error(msg string, keysAndValues ...interface{}) {
	globalLogger.Error(msg, keysAndValues...)
}

// Fatal logs a fatal message using the global logger and exits
func Fatal(msg string, keysAndValues ...interface{}) {
	globalLogger.Fatal(msg, keysAndValues...)
}
