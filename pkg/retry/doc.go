// Package retry provides retry mechanisms with configurable backoff strategies
// for the task queue system. It supports exponential backoff, jitter, and
// maximum retry attempts.
//
// The package is designed to handle transient failures in distributed systems
// with customizable retry policies per operation type.
//
// Basic usage:
//
//	err := retry.Do(
//	    func() error {
//	        return someOperation()
//	    },
//	    retry.WithMaxAttempts(3),
//	    retry.WithBackoff(retry.ExponentialBackoff),
//	)
//
// Custom backoff:
//
//	backoff := retry.NewExponentialBackoff(
//	    retry.WithInitialDelay(100*time.Millisecond),
//	    retry.WithMaxDelay(10*time.Second),
//	    retry.WithMultiplier(2.0),
//	)
//	err := retry.Do(operation, retry.WithBackoffStrategy(backoff))
package retry
