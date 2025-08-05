package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name:   "default config",
			config: DefaultConfig(),
		},
		{
			name: "json format",
			config: Config{
				Level:  "debug",
				Format: "json",
			},
		},
		{
			name: "console format",
			config: Config{
				Level:  "info",
				Format: "console",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := New(tt.config)
			require.NotNil(t, log)

			// Test basic logging doesn't panic
			assert.NotPanics(t, func() {
				log.Debug("debug message")
				log.Info("info message")
				log.Warn("warn message")
				log.Error("error message")
			})
		})
	}
}

func TestLoggerWithContext(t *testing.T) {
	log := NewNop() // Use nop logger for testing

	ctx := context.Background()
	ctx = context.WithValue(ctx, RequestIDKey, "test-request-123")
	ctx = context.WithValue(ctx, TraceIDKey, "trace-456")
	ctx = context.WithValue(ctx, UserIDKey, "user-789")

	ctxLogger := log.WithContext(ctx)
	assert.NotNil(t, ctxLogger)

	// Test that WithContext doesn't panic with nil context
	nilLogger := log.WithContext(context.TODO())
	assert.NotNil(t, nilLogger)
}

func TestLoggerWith(t *testing.T) {
	log := NewNop()

	// Test With method
	withLogger := log.With("key1", "value1", "key2", 123)
	assert.NotNil(t, withLogger)

	// Test WithError
	err := assert.AnError
	errorLogger := log.WithError(err)
	assert.NotNil(t, errorLogger)

	// Test WithError with nil
	nilErrorLogger := log.WithError(nil)
	assert.NotNil(t, nilErrorLogger)
}

func TestLoggerNamed(t *testing.T) {
	log := NewNop()

	namedLogger := log.Named("test-component")
	assert.NotNil(t, namedLogger)
}

func TestGlobalLogger(t *testing.T) {
	// Test global logger exists
	assert.NotNil(t, Global())

	// Test SetGlobal
	newLogger := NewNop()
	SetGlobal(newLogger)
	assert.Equal(t, newLogger, Global())

	// Test global helper functions don't panic
	assert.NotPanics(t, func() {
		Debug("debug")
		Info("info")
		Warn("warn")
		Error("error")
	})
}
