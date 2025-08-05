package validation

// Validator interface for all validation rules
type Validator interface {
	Validate(value any) error
}

// ValidatorFunc is a function that implements Validator
type ValidatorFunc func(any) error

// Validate implements the Validator interface
func (f ValidatorFunc) Validate(value any) error {
	return f(value)
}

// Rule represents a validation rule with a custom message
type Rule struct {
	Validator Validator
	Message   string
}

// Field represents a field to be validated
type Field struct {
	Name       string
	Value      any
	Validators []Validator
}

// ValidationError represents a validation error with field details
type ValidationError struct {
	Field   string         `json:"field"`
	Message string         `json:"message"`
	Value   any            `json:"value,omitempty"`
	Tag     string         `json:"tag,omitempty"`
	Params  map[string]any `json:"params,omitempty"`
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError
