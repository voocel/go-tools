package retry

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"
)

const (
	DefaultRetryTimes    = 5
	DefaultRetryDuration = time.Second * 3
)

type RetryConfig struct {
	context       context.Context
	retryTimes    uint
	retryDuration time.Duration
}

type RetryFunc func() error

type Option func(*RetryConfig)

func RetryTimes(n uint) Option {
	return func(rc *RetryConfig) {
		rc.retryTimes = n
	}
}

func RetryDuration(d time.Duration) Option {
	return func(rc *RetryConfig) {
		rc.retryDuration = d
	}
}

func Context(ctx context.Context) Option {
	return func(rc *RetryConfig) {
		rc.context = ctx
	}
}

func Retry(retryFunc RetryFunc, opts ...Option) error {
	config := &RetryConfig{
		retryTimes:    DefaultRetryTimes,
		retryDuration: DefaultRetryDuration,
		context:       context.TODO(),
	}

	for _, opt := range opts {
		opt(config)
	}

	var i uint
	for i < config.retryTimes {
		err := retryFunc()
		if err != nil {
			select {
			case <-time.After(config.retryDuration):
			case <-config.context.Done():
				return errors.New("retry is cancelled")
			}
		} else {
			return nil
		}
		i++
	}

	funcPath := runtime.FuncForPC(reflect.ValueOf(retryFunc).Pointer()).Name()
	lastSlash := strings.LastIndex(funcPath, "/")
	funcName := funcPath[lastSlash+1:]

	return fmt.Errorf("function %s run failed after %d times retry", funcName, i)
}
