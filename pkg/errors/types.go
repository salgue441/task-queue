package errors

// Code represents an error code
type Code string

// Predefined error codes
const (
	CodeUnknown        Code = "UNKNOWN"
	CodeInternal       Code = "INTERNAL"
	CodeValidation     Code = "VALIDATION"
	CodeNotFound       Code = "NOT_FOUND"
	CodeAlreadyExists  Code = "ALREADY_EXISTS"
	CodePermission     Code = "PERMISSION_DENIED"
	CodeAuthentication Code = "UNAUTHENTICATED"
	CodeRateLimit      Code = "RATE_LIMITED"
	CodeTimeout        Code = "TIMEOUT"
	CodeCanceled       Code = "CANCELED"
	CodeConflict       Code = "CONFLICT"
	CodeDatabase       Code = "DATABASE_ERROR"
	CodeNetwork        Code = "NETWORK_ERROR"
	CodeSerialization  Code = "SERIALIZATION_ERROR"
	CodeConfiguration  Code = "CONFIGURATION_ERROR"
)

// Error represents an enhanced error with additional context
type Error struct {
	Code       Code           `json:"code"`
	Message    string         `json:"message"`
	Cause      error          `json:"-"`
	StatusCode int            `json:"status_code,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	Stack      []Frame        `json:"stack,omitempty"`
}

// Frame represents a stack frame
type Frame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}
