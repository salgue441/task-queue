package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	err := New("test error")

	assert.NotNil(t, err)
	assert.Equal(t, "test error", err.Message)
	assert.Equal(t, CodeUnknown, err.Code)
	assert.NotEmpty(t, err.Stack)
	assert.NotNil(t, err.Metadata)
}

func TestNewf(t *testing.T) {
	err := Newf("error: %s", "test")

	assert.Equal(t, "error: test", err.Message)
}

func TestWrap(t *testing.T) {
	// Test wrapping standard error
	stdErr := errors.New("standard error")
	wrapped := Wrap(stdErr, "wrapped message")

	assert.NotNil(t, wrapped)
	assert.Equal(t, "wrapped message", wrapped.Message)
	assert.Equal(t, stdErr, wrapped.Cause)
	assert.NotEmpty(t, wrapped.Stack)

	// Test wrapping nil
	assert.Nil(t, Wrap(nil, "message"))

	// Test wrapping our error type
	appErr := New("app error").WithCode(CodeValidation)
	wrapped2 := Wrap(appErr, "wrapped app error")

	assert.Equal(t, CodeValidation, wrapped2.Code)
	assert.Equal(t, "wrapped app error", wrapped2.Message)
}

func TestWithCode(t *testing.T) {
	err := New("test").WithCode(CodeNotFound)

	assert.Equal(t, CodeNotFound, err.Code)
	assert.Equal(t, http.StatusNotFound, err.StatusCode)
}

func TestWithMetadata(t *testing.T) {
	err := New("test").
		WithMetadata("key1", "value1").
		WithMetadata("key2", 123)

	assert.Equal(t, "value1", err.Metadata["key1"])
	assert.Equal(t, 123, err.Metadata["key2"])

	// Test WithMetadataMap
	err2 := New("test").WithMetadataMap(map[string]interface{}{
		"key3": "value3",
		"key4": true,
	})

	assert.Equal(t, "value3", err2.Metadata["key3"])
	assert.Equal(t, true, err2.Metadata["key4"])
}

func TestErrorInterface(t *testing.T) {
	// Test simple error
	err := New("test error")
	assert.Equal(t, "test error", err.Error())

	// Test error with cause
	cause := errors.New("cause")
	wrapped := Wrap(cause, "wrapped")
	assert.Contains(t, wrapped.Error(), "wrapped")
	assert.Contains(t, wrapped.Error(), "cause")
}

func TestIs(t *testing.T) {
	err1 := New("error1").WithCode(CodeNotFound)
	err2 := New("error2").WithCode(CodeNotFound)
	err3 := New("error3").WithCode(CodeValidation)

	assert.True(t, errors.Is(err1, err2))
	assert.False(t, errors.Is(err1, err3))
}

func TestHTTPStatus(t *testing.T) {
	tests := []struct {
		code     Code
		expected int
	}{
		{CodeValidation, http.StatusBadRequest},
		{CodeNotFound, http.StatusNotFound},
		{CodeAlreadyExists, http.StatusConflict},
		{CodePermission, http.StatusForbidden},
		{CodeAuthentication, http.StatusUnauthorized},
		{CodeRateLimit, http.StatusTooManyRequests},
		{CodeTimeout, http.StatusRequestTimeout},
		{CodeInternal, http.StatusInternalServerError},
		{CodeUnknown, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			err := New("test").WithCode(tt.code)
			assert.Equal(t, tt.expected, err.HTTPStatus())
		})
	}

	// Test custom status code
	err := New("test").WithStatusCode(http.StatusTeapot)
	assert.Equal(t, http.StatusTeapot, err.HTTPStatus())
}

func TestToJSON(t *testing.T) {
	err := New("test error").
		WithCode(CodeValidation).
		WithMetadata("field", "email")

	data, jsonErr := err.ToJSON()
	require.NoError(t, jsonErr)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &result))

	assert.Equal(t, "VALIDATION", result["code"])
	assert.Equal(t, "test error", result["message"])
	assert.NotNil(t, result["metadata"])
}

func TestHelperFunctions(t *testing.T) {
	// Test IsNotFound
	notFoundErr := New("not found").WithCode(CodeNotFound)
	assert.True(t, IsNotFound(notFoundErr))
	assert.False(t, IsNotFound(New("other")))

	// Test IsValidation
	validationErr := New("invalid").WithCode(CodeValidation)
	assert.True(t, IsValidation(validationErr))

	// Test IsPermission
	permErr := New("forbidden").WithCode(CodePermission)
	assert.True(t, IsPermission(permErr))

	// Test IsAuthentication
	authErr := New("unauthorized").WithCode(CodeAuthentication)
	assert.True(t, IsAuthentication(authErr))

	// Test IsTimeout
	timeoutErr := New("timeout").WithCode(CodeTimeout)
	assert.True(t, IsTimeout(timeoutErr))

	canceledErr := New("canceled").WithCode(CodeCanceled)
	assert.True(t, IsTimeout(canceledErr))

	// Test IsConflict
	conflictErr := New("conflict").WithCode(CodeConflict)
	assert.True(t, IsConflict(conflictErr))

	existsErr := New("exists").WithCode(CodeAlreadyExists)
	assert.True(t, IsConflict(existsErr))

	// Test GetCode
	assert.Equal(t, CodeNotFound, GetCode(notFoundErr))
	assert.Equal(t, CodeUnknown, GetCode(errors.New("standard")))

	// Test GetHTTPStatus
	assert.Equal(t, http.StatusNotFound, GetHTTPStatus(notFoundErr))
	assert.Equal(t, http.StatusInternalServerError, GetHTTPStatus(errors.New("standard")))
}

func TestNilError(t *testing.T) {
	var err *Error

	// Test that nil methods don't panic
	assert.Nil(t, err.WithCode(CodeNotFound))
	assert.Nil(t, err.WithStatusCode(404))
	assert.Nil(t, err.WithMetadata("key", "value"))
	assert.Nil(t, err.WithMetadataMap(map[string]interface{}{"key": "value"}))
}
