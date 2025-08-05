// Package validation provides input validation utilities for the task queue
// system. It offers both struct-based validation using tags and programmatic
// validation with a fluent API.
//
// The package integrates with the errors package to provide consistent error
// reporting and supports custom validation rules for domain-specific
// requirements.
//
// Struct validation example:
//
//	type CreateJobRequest struct {
//	    Type     string `validate:"required,min=1,max=100"`
//	    Priority int    `validate:"min=0,max=3"`
//	}
//
//	v := validation.New()
//	if err := v.Struct(req); err != nil {
//	    // Handle validation error
//	}
//
// Fluent validation example:
//
//	err := validation.Validate(
//	    validation.Field("email", email, validation.Required, validation.Email),
//	    validation.Field("age", age, validation.Min(18), validation.Max(100)),
//	)
package validation
