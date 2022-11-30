package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetryFailed(t *testing.T) {
	var number int
	increaseNumber := func() error {
		number++
		return errors.New("error occurs")
	}

	err := Retry(increaseNumber, RetryDuration(time.Microsecond*50))

	assert.NotNil(t, err)
	assert.Equal(t, DefaultRetryTimes, number)
}

func TestRetrySucceeded(t *testing.T) {
	var number int
	increaseNumber := func() error {
		number++
		if number == DefaultRetryTimes {
			return nil
		}
		return errors.New("error occurs")
	}

	err := Retry(increaseNumber, RetryDuration(time.Microsecond*50))

	assert.Nil(t, err)
	assert.Equal(t, DefaultRetryTimes, number)
}

func TestSetRetryTimes(t *testing.T) {
	var number int
	increaseNumber := func() error {
		number++
		return errors.New("error occurs")
	}

	err := Retry(increaseNumber, RetryDuration(time.Microsecond*50), RetryTimes(3))

	assert.NotNil(t, err)
	assert.Equal(t, 3, number)
}

func TestCancelRetry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	var number int
	increaseNumber := func() error {
		number++
		if number > 3 {
			cancel()
		}
		return errors.New("error occurs")
	}

	err := Retry(increaseNumber,
		RetryDuration(time.Microsecond*50),
		Context(ctx),
	)

	assert.NotNil(t, err)
	assert.Equal(t, 4, number)
}
