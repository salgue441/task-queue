package validation

import (
	"fmt"
	"regexp"
	"strings"
	"task-queue/pkg/errors"
)

// Error implements the error interface
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}

	var msgs []string
	for _, e := range ve {
		msgs = append(msgs, fmt.Sprintf("%s: %s", e.Field, e.Message))
	}

	return fmt.Sprintf("validation failed: %s", strings.Join(msgs, "; "))
}

// Validate validates multiple fields
func Validate(fields ...*Field) error {
	var validationErrors ValidationErrors
	for _, field := range fields {
		for _, validator := range field.Validators {
			if err := validator.Validate(field.Value); err != nil {
				validationErrors = append(validationErrors, ValidationError{
					Field:   field.Name,
					Message: err.Error(),
					Value:   field.Value,
				})

				break
			}
		}
	}

	if len(validationErrors) > 0 {
		return errors.New(validationErrors.Error()).
			WithCode(errors.CodeValidation).
			WithMetadata("fields", validationErrors)
	}

	return nil
}

// NewField creates a new field for validation
func NewField(name string, value any, validators ...Validator) *Field {
	return &Field{
		Name:       name,
		Value:      value,
		Validators: validators,
	}
}

// Common validators

// Required validates that a value is not empty
var Required = ValidatorFunc(func(value any) error {
	if value == nil {
		return fmt.Errorf("is required")
	}

	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("is required")
		}

	case []byte:
		if len(v) == 0 {
			return fmt.Errorf("is required")
		}

	// Numeric and boolean types are always considered present
	case int, int8, int16, int32, int64:
	case uint, uint8, uint16, uint32, uint64:
	case float32, float64:
	case bool:
	default:
		if fmt.Sprintf("%v", v) == "" {
			return fmt.Errorf("is required")
		}
	}

	return nil
})

// Min returns a validator that checks if a value is at least min
func Min(min float64) Validator {
	return ValidatorFunc(func(value any) error {
		switch v := value.(type) {
		case int:
			if float64(v) < min {
				return fmt.Errorf("must be at least %v", min)
			}

		case int64:
			if float64(v) < min {
				return fmt.Errorf("must be at least %v", min)
			}

		case float64:
			if v < min {
				return fmt.Errorf("must be at least %v", min)
			}

		case string:
			if float64(len(v)) < min {
				return fmt.Errorf("must be at least %v characters", min)
			}

		case []byte:
			if float64(len(v)) < min {
				return fmt.Errorf("must be at least %v bytes", min)
			}

		default:
			return fmt.Errorf("cannot apply min validation to type %T", value)
		}

		return nil
	})
}

// Max returns a validator that checks if a value is at most max
func Max(max float64) Validator {
	return ValidatorFunc(func(value any) error {
		switch v := value.(type) {
		case int:
			if float64(v) > max {
				return fmt.Errorf("must be at most %v", max)
			}

		case int64:
			if float64(v) > max {
				return fmt.Errorf("must be at most %v", max)
			}

		case float64:
			if v > max {
				return fmt.Errorf("must be at most %v", max)
			}

		case string:
			if float64(len(v)) > max {
				return fmt.Errorf("must be at most %v characters", max)
			}

		case []byte:
			if float64(len(v)) > max {
				return fmt.Errorf("must be at most %v bytes", max)
			}

		default:
			return fmt.Errorf("cannot apply max validation to type %T", value)
		}

		return nil
	})
}

// Between returns a validator that checks if a value is between min and max
func Between(min, max float64) Validator {
	return ValidatorFunc(func(value any) error {
		if err := Min(min).Validate(value); err != nil {
			return err
		}

		return Max(max).Validate(value)
	})
}

// Email validates that a string is a valid email address
var Email = ValidatorFunc(func(value any) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("must be a string")
	}

	emailRegex := regexp.MustCompile(
		`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(str) {
		return fmt.Errorf("must be a valid email address")
	}

	return nil
})

// URL validates that a string is a valid URL
var URL = ValidatorFunc(func(value any) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("must be a string")
	}

	urlRegex := regexp.MustCompile(`^(https?|ftp)://[^\s/$.?#].[^\s]*$`)
	if !urlRegex.MatchString(str) {
		return fmt.Errorf("must be a valid URL")
	}

	return nil
})

// UUID validates that a string is a valid UUID
var UUID = ValidatorFunc(func(value any) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("must be a string")
	}

	uuidRegex := regexp.MustCompile(
		`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !uuidRegex.MatchString(strings.ToLower(str)) {
		return fmt.Errorf("must be a valid UUID")
	}

	return nil
})

// In returns a validator that checks if a value is in a list
func In(values ...any) Validator {
	return ValidatorFunc(func(value any) error {
		for _, v := range values {
			if value == v {
				return nil
			}
		}

		return fmt.Errorf("must be one of %v", values)
	})
}

// Pattern returns a validator that checks if a string matches a regex pattern
func Pattern(pattern string) Validator {
	regex := regexp.MustCompile(pattern)
	return ValidatorFunc(func(value any) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("must be a string")
		}

		if !regex.MatchString(str) {
			return fmt.Errorf("must match pattern %s", pattern)
		}

		return nil
	})
}

// Custom validator for job type
func JobType() Validator {
	return ValidatorFunc(func(value any) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("must be a string")
		}

		if len(str) < 1 || len(str) > 100 {
			return fmt.Errorf("must be between 1 and 100 characters")
		}

		validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
		if !validPattern.MatchString(str) {
			return fmt.Errorf(
				"can only contain letters, numbers, underscore, and hyphen")
		}

		return nil
	})
}

// Custom validator for job priority
func JobPriority() Validator {
	return In(0, 1, 2, 3)
}
