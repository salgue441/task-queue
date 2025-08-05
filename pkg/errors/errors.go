package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
)

// New creates a new error with a message
func New(message string) *Error {
	return &Error{
		Code:     CodeUnknown,
		Message:  message,
		Metadata: make(map[string]any),
		Stack:    captureStack(2),
	}
}

// Newf creates a new error with a formatted message
func Newf(format string, args ...any) *Error {
	return New(fmt.Sprintf(format, args...))
}

// Wrap wraps an existing error with additional context
func Wrap(err error, message string) *Error {
	if err == nil {
		return nil
	}

	var appErr *Error
	if errors.As(err, &appErr) {
		return &Error{
			Code:       appErr.Code,
			Message:    message,
			Cause:      err,
			StatusCode: appErr.StatusCode,
			Metadata:   copyMetadata(appErr.Metadata),
			Stack:      captureStack(2),
		}
	}

	return &Error{
		Code:     CodeUnknown,
		Message:  message,
		Cause:    err,
		Metadata: make(map[string]any),
		Stack:    captureStack(2),
	}
}

// Wrapf wraps an error with a formatted message
func Wrapf(err error, format string, args ...any) *Error {
	return Wrap(err, fmt.Sprintf(format, args...))
}

// WithCode sets the error code
func (e *Error) WithCode(code Code) *Error {
	if e == nil {
		return nil
	}

	e.Code = code
	e.StatusCode = codeToHTTPStatus(code)
	return e
}

// WithStatusCode sets a custom HTTP status code
func (e *Error) WithStatusCode(code int) *Error {
	if e == nil {
		return nil
	}

	e.StatusCode = code
	return e
}

// WithMetadata adds metadata to the error
func (e *Error) WithMetadata(key string, value any) *Error {
	if e == nil {
		return nil
	}

	if e.Metadata == nil {
		e.Metadata = make(map[string]any)
	}

	e.Metadata[key] = value
	return e
}

// WithMetadataMap adds multiple metadata entries
func (e *Error) WithMetadataMap(metadata map[string]any) *Error {
	if e == nil {
		return nil
	}

	if e.Metadata == nil {
		e.Metadata = make(map[string]any)
	}

	for k, v := range metadata {
		e.Metadata[k] = v
	}

	return e
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}

	return e.Message
}

// Unwrap returns the underlying error
func (e *Error) Unwrap() error {
	return e.Cause
}

// Is implements errors.Is interface
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}

	return e.Code == t.Code
}

// HTTPStatus returns the HTTP status code for the error
func (e *Error) HTTPStatus() int {
	if e.StatusCode != 0 {
		return e.StatusCode
	}

	return codeToHTTPStatus(e.Code)
}

// ToJSON converts the error to JSON
func (e *Error) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// captureStack captures the current stack trace
func captureStack(skip int) []Frame {
	const maxStackDepth = 32
	frames := make([]Frame, 0, maxStackDepth)
	pcs := make([]uintptr, maxStackDepth)
	n := runtime.Callers(skip+1, pcs)

	for i := 0; i < n; i++ {
		pc := pcs[i]
		fn := runtime.FuncForPC(pc)

		if fn == nil {
			continue
		}

		file, line := fn.FileLine(pc)
		frames = append(frames, Frame{
			Function: fn.Name(),
			File:     file,
			Line:     line,
		})
	}

	return frames
}

// copyMetadata creates a copy of metadata map
func copyMetadata(m map[string]any) map[string]any {
	if m == nil {
		return make(map[string]any)
	}

	copy := make(map[string]any, len(m))
	for k, v := range m {
		copy[k] = v
	}

	return copy
}

// codeToHTTPStatus maps error codes to HTTP status codes
func codeToHTTPStatus(code Code) int {
	switch code {
	case CodeValidation:
		return http.StatusBadRequest

	case CodeNotFound:
		return http.StatusNotFound

	case CodeAlreadyExists:
		return http.StatusConflict

	case CodePermission:
		return http.StatusForbidden

	case CodeAuthentication:
		return http.StatusUnauthorized

	case CodeRateLimit:
		return http.StatusTooManyRequests

	case CodeTimeout:
		return http.StatusRequestTimeout

	case CodeCanceled:
		return http.StatusRequestTimeout

	case CodeConflict:
		return http.StatusConflict

	case CodeInternal, CodeDatabase, CodeNetwork,
		CodeSerialization, CodeConfiguration:
		return http.StatusInternalServerError

	default:
		return http.StatusInternalServerError
	}
}

// Helper functions for common error types

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Code == CodeNotFound
	}

	return false
}

// IsValidation checks if an error is a validation error
func IsValidation(err error) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Code == CodeValidation
	}

	return false
}

// IsPermission checks if an error is a permission error
func IsPermission(err error) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Code == CodePermission
	}

	return false
}

// IsAuthentication checks if an error is an authentication error
func IsAuthentication(err error) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Code == CodeAuthentication
	}

	return false
}

// IsTimeout checks if an error is a timeout error
func IsTimeout(err error) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Code == CodeTimeout || e.Code == CodeCanceled
	}

	return false
}

// IsConflict checks if an error is a conflict error
func IsConflict(err error) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Code == CodeConflict || e.Code == CodeAlreadyExists
	}

	return false
}

// GetCode returns the error code from an error
func GetCode(err error) Code {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}

	return CodeUnknown
}

// GetHTTPStatus returns the HTTP status code from an error
func GetHTTPStatus(err error) int {
	var e *Error
	if errors.As(err, &e) {
		return e.HTTPStatus()
	}
	
	return http.StatusInternalServerError
}
