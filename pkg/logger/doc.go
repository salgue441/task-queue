// Package logger provides a structured logging interface for the task queue
// system. It wraps uber/zap logger with additional features like context
// awareness, request ID tracking, and standardized log formatting.
//
// The logger supports multiple output formats (JSON, Console) and log levels,
// with built-in support for distributed tracing correlation.
//
// Basic usage:
//
//	log := logger.New(logger.Config{
//	    Level: "info",
//	    Format: "json",
//	})
//	log.Info("server started", "port", 8080)
//
// With context:
//
//	ctx := context.WithValue(context.Background(), logger.RequestIDKey, "req-123")
//	log.WithContext(ctx).Info("processing request")
//
// With fields:
//
//	log.With("user_id", "123").Error("failed to process", "error", err)
package logger
