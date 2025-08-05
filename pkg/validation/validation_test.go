package validation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequired(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"nil value", nil, true},
		{"empty string", "", true},
		{"whitespace string", "  ", true},
		{"valid string", "test", false},
		{"zero int", 0, false},
		{"positive int", 42, false},
		{"empty slice", []byte{}, true},
		{"non-empty slice", []byte{1, 2}, false},
		{"bool false", false, false},
		{"bool true", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Required.Validate(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name    string
		min     float64
		value   interface{}
		wantErr bool
	}{
		{"int below min", 10, 5, true},
		{"int at min", 10, 10, false},
		{"int above min", 10, 15, false},
		{"string too short", 5, "test", true},
		{"string exact length", 4, "test", false},
		{"string longer", 3, "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Min(tt.min).Validate(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name    string
		max     float64
		value   interface{}
		wantErr bool
	}{
		{"int above max", 10, 15, true},
		{"int at max", 10, 10, false},
		{"int below max", 10, 5, false},
		{"string too long", 3, "test", true},
		{"string exact length", 4, "test", false},
		{"string shorter", 5, "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Max(tt.max).Validate(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBetween(t *testing.T) {
	validator := Between(5, 10)

	assert.Error(t, validator.Validate(4))
	assert.NoError(t, validator.Validate(5))
	assert.NoError(t, validator.Validate(7))
	assert.NoError(t, validator.Validate(10))
	assert.Error(t, validator.Validate(11))
}

func TestEmail(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"valid email", "test@example.com", false},
		{"valid with subdomain", "user@mail.example.com", false},
		{"valid with plus", "user+tag@example.com", false},
		{"invalid without @", "test.example.com", true},
		{"invalid without domain", "test@", true},
		{"invalid without local", "@example.com", true},
		{"not a string", 123, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Email.Validate(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestURL(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"valid http", "http://example.com", false},
		{"valid https", "https://example.com", false},
		{"valid with path", "https://example.com/path", false},
		{"valid with query", "https://example.com?q=test", false},
		{"invalid without scheme", "example.com", true},
		{"invalid scheme", "javascript:alert(1)", true},
		{"not a string", 123, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := URL.Validate(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUUID(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"valid uuid", "550e8400-e29b-41d4-a716-446655440000", false},
		{"valid uppercase", "550E8400-E29B-41D4-A716-446655440000", false},
		{"invalid format", "550e8400-e29b-41d4-a716", true},
		{"invalid characters", "550e8400-e29b-41d4-a716-44665544000g", true},
		{"not a string", 123, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := UUID.Validate(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIn(t *testing.T) {
	validator := In("apple", "banana", "orange")

	assert.NoError(t, validator.Validate("apple"))
	assert.NoError(t, validator.Validate("banana"))
	assert.Error(t, validator.Validate("grape"))

	// Test with integers
	intValidator := In(1, 2, 3)
	assert.NoError(t, intValidator.Validate(1))
	assert.Error(t, intValidator.Validate(4))
}

func TestPattern(t *testing.T) {
	validator := Pattern(`^[A-Z][0-9]+$`)

	assert.NoError(t, validator.Validate("A123"))
	assert.NoError(t, validator.Validate("Z999"))
	assert.Error(t, validator.Validate("123A"))
	assert.Error(t, validator.Validate("a123"))
	assert.Error(t, validator.Validate(123))
}

func TestJobType(t *testing.T) {
	validator := JobType()

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"valid simple", "process_image", false},
		{"valid with hyphen", "send-email", false},
		{"valid with underscore", "generate_report", false},
		{"valid alphanumeric", "task123", false},
		{"empty string", "", true},
		{"too long", strings.Repeat("a", 101), true},
		{"with spaces", "process image", true},
		{"with special chars", "process@image", true},
		{"not a string", 123, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	// Test successful validation
	err := Validate(
		NewField("email", "test@example.com", Required, Email),
		NewField("age", 25, Required, Min(18), Max(100)),
		NewField("status", "active", In("active", "inactive")),
	)
	assert.NoError(t, err)

	// Test validation with errors
	err = Validate(
		NewField("email", "invalid-email", Required, Email),
		NewField("age", 150, Required, Min(18), Max(100)),
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}
