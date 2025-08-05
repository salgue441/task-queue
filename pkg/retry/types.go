package retry

import (
	"context"
	"time"
)

// Operation is a function that can be retried
type Operation func() error

// ContextOperation is a function that can be retried with context
type ContextOperation func(context.Context) error

// BackoffStrategy defines the interface for backoff strategies
type BackoffStrategy interface {
	// Next returns the next delay duration
	Next(attempt int) time.Duration

	// Reset resets the backoff strategy
	Reset()
}

// Config holds retry configuration
type Config struct {
	MaxAttempts     int
	MaxDelay        time.Duration
	BackoffStrategy BackoffStrategy
	RetryIf         func(error) bool
	OnRetry         func(attempt int, err error)
	Context         context.Context
}

// Option is a function that modifies retry configuration
type Option func(*Config)

// DefaultConfig returns default retry configuration
func DefaultConfig() *Config {
	return &Config{
		MaxAttempts:     3,
		MaxDelay:        30 * time.Second,
		BackoffStrategy: NewExponentialBackoff(),
		RetryIf:         defaultRetryIf,
		OnRetry:         nil,
		Context:         context.Background(),
	}
}

// ExponentialBackoff implements exponential backoff with jitter
type ExponentialBackoff struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	Jitter       float64
}

// NewExponentialBackoff creates a new exponential backoff strategy
func NewExponentialBackoff(opts ...BackoffOption) *ExponentialBackoff {
	b := &ExponentialBackoff{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.1,
	}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

// BackoffOption modifies exponential backoff configuration
type BackoffOption func(*ExponentialBackoff)

// LinearBackoff implements linear backoff
type LinearBackoff struct {
	InitialDelay time.Duration
	Increment    time.Duration
	MaxDelay     time.Duration
}

// NewLinearBackoff creates a new linear backoff strategy
func NewLinearBackoff(initial, increment, max time.Duration) *LinearBackoff {
	return &LinearBackoff{
		InitialDelay: initial,
		Increment:    increment,
		MaxDelay:     max,
	}
}

// FixedBackoff implements fixed delay backoff
type FixedBackoff struct {
	Delay time.Duration
}

// NewFixedBackoff creates a new fixed backoff strategy
func NewFixedBackoff(delay time.Duration) *FixedBackoff {
	return &FixedBackoff{Delay: delay}
}
