package retry

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"task-queue/pkg/errors"
	"time"
)

// defaultRetryIf determines if an error should trigger a retry
func defaultRetryIf(err error) bool {
	if err == nil {
		return false
	}

	if errors.IsValidation(err) {
		return false
	}

	if errors.IsNotFound(err) {
		return false
	}

	if errors.IsPermission(err) || errors.IsAuthentication(err) {
		return false
	}

	return true
}

// WithMaxAttempts sets the maximum number of retry attempts
func WithMaxAttempts(attempts int) Option {
	return func(c *Config) {
		c.MaxAttempts = attempts
	}
}

// WithMaxDelay sets the maximum delay between retries
func WithMaxDelay(delay time.Duration) Option {
	return func(c *Config) {
		c.MaxDelay = delay
	}
}

// WithBackoffStrategy sets a custom backoff strategy
func WithBackoffStrategy(strategy BackoffStrategy) Option {
	return func(c *Config) {
		c.BackoffStrategy = strategy
	}
}

// WithRetryIf sets a custom function to determine if retry should occur
func WithRetryIf(fn func(error) bool) Option {
	return func(c *Config) {
		c.RetryIf = fn
	}
}

// WithOnRetry sets a callback function called on each retry
func WithOnRetry(fn func(int, error)) Option {
	return func(c *Config) {
		c.OnRetry = fn
	}
}

// WithContext sets the context for the retry operation
func WithContext(ctx context.Context) Option {
	return func(c *Config) {
		c.Context = ctx
	}
}

// Do executes the operation with retry logic
func Do(operation Operation, opts ...Option) error {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	return DoWithConfig(operation, config)
}

// DoWithConfig executes the operation with the given configuration
func DoWithConfig(operation Operation, config *Config) error {
	if config.MaxAttempts <= 0 {
		return fmt.Errorf("max attempts must be greater than 0")
	}

	var lastErr error
	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		select {
		case <-config.Context.Done():
			return errors.Wrap(config.Context.Err(), "retry cancelled").
				WithCode(errors.CodeCanceled)

		default:
		}

		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err
		if !config.RetryIf(err) {
			return err
		}

		if attempt >= config.MaxAttempts {
			break
		}

		if config.OnRetry != nil {
			config.OnRetry(attempt, err)
		}

		delay := config.BackoffStrategy.Next(attempt)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}

		timer := time.NewTimer(delay)
		select {
		case <-timer.C:
		case <-config.Context.Done():
			timer.Stop()

			return errors.Wrap(
				config.Context.Err(),
				"retry cancelled during backoff",
			).WithCode(errors.CodeCanceled)
		}
	}

	return errors.Wrapf(lastErr,
		"operation failed after %d attempts", config.MaxAttempts,
	).WithCode(errors.CodeInternal).
		WithMetadata("attempts", config.MaxAttempts)
}

// DoWithContext executes a context-aware operation with retry logic
func DoWithContext(ctx context.Context, operation ContextOperation,
	opts ...Option) error {
	op := func() error {
		return operation(ctx)
	}

	opts = append(opts, WithContext(ctx))
	return Do(op, opts...)
}

// WithInitialDelay sets the initial delay
func WithInitialDelay(delay time.Duration) BackoffOption {
	return func(b *ExponentialBackoff) {
		b.InitialDelay = delay
	}
}

// WithMultiplier sets the backoff multiplier
func WithMultiplier(multiplier float64) BackoffOption {
	return func(b *ExponentialBackoff) {
		b.Multiplier = multiplier
	}
}

// WithJitter sets the jitter factor (0.0 to 1.0)
func WithJitter(jitter float64) BackoffOption {
	return func(b *ExponentialBackoff) {
		b.Jitter = jitter
	}
}

// Next returns the next delay with exponential backoff
func (e *ExponentialBackoff) Next(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}

	delay := float64(e.InitialDelay) * math.Pow(e.Multiplier, float64(attempt-1))
	if e.Jitter > 0 {
		jitter := delay * e.Jitter
		delay = delay - jitter + (rand.Float64() * jitter * 2)
	}

	if delay > float64(e.MaxDelay) {
		delay = float64(e.MaxDelay)
	}

	return time.Duration(delay)
}

// Reset resets the backoff strategy
func (e *ExponentialBackoff) Reset() {
	// No state to reset for exponential backoff
}

// Next returns the next delay with linear backoff
func (l *LinearBackoff) Next(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}

	delay := l.InitialDelay + time.Duration(attempt-1)*l.Increment
	if delay > l.MaxDelay {
		delay = l.MaxDelay
	}

	return delay
}

// Reset resets the backoff strategy
func (l *LinearBackoff) Reset() {
	// No state to reset for linear backoff
}

// Next returns the fixed delay
func (f *FixedBackoff) Next(attempt int) time.Duration {
	return f.Delay
}

// Reset resets the backoff strategy
func (f *FixedBackoff) Reset() {
	// No state to reset for fixed backoff
}
