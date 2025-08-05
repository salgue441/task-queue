package retry

import (
	"task-queue/pkg/errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDo_SuccessOnFirstAttempt(t *testing.T) {
	calls := 0
	err := Do(func() error {
		calls++
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestDo_RetryUntilSuccess(t *testing.T) {
	calls := 0
	err := Do(func() error {
		calls++
		if calls < 3 {
			return errors.New("temporary error")
		}

		return nil
	}, WithMaxAttempts(3))

	assert.NoError(t, err)
	assert.Equal(t, 3, calls)
}

func TestDo_MaxAttemptsExceeded(t *testing.T) {
	calls := 0
	err := Do(func() error {
		calls++
		return errors.New("persistent error")
	}, WithMaxAttempts(3))

	assert.Error(t, err)
	assert.Equal(t, 3, calls)
	assert.Contains(t, err.Error(), "operation failed after 3 attempts")
}
