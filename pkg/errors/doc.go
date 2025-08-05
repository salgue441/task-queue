// Package errors provides enhanced error handling capabilities for the task
// queue system.
//
// It extends the standard error interface with additional context, error codes,
// stack traces, and error metadata.
//
// The package supports:
//   - Error wrapping with context preservation
//   - HTTP status code mapping
//   - Stack trace capture for debugging
//   - Structured error metadata
//   - Error categorization (validation, not found, conflict, etc.)
//   - gRPC status code conversion
//
// Basic usage:
//
//	err := errors.New("database connection failed").
//	    WithCode(errors.CodeInternal).
//	    WithMetadata("host", "localhost:5432")
//
// Wrapping errors:
//
//	if err := db.Query(); err != nil {
//	    return errors.Wrap(err, "failed to query database").
//	        WithCode(errors.CodeDatabase)
//	}
//
// Checking error types:
//
//	if errors.IsNotFound(err) {
//	    // Handle not found error
//	}
package errors
