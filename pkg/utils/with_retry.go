package utils

import (
	"context"
	"fmt"
	"time"
)

// DefaultBackoffTime is the default value of backoff time used by
// WithDefaultRetry function.
const DefaultBackoffTime = 1 * time.Second

// DefaultMaxBackoffTime is the default value of max backoff time used by
// WithDefaultRetry function.
const DefaultMaxBackoffTime = 120 * time.Second

// WithDefaultRetry executes the provided doFn as long as it returns an error or
// until a timeout is hit. It applies exponential backoff wait of
// DefaultBackoffTime * 2^n before n retry of doFn. In case the calculated
// backoff is longer than DefaultMaxBackoffTime, the DefaultMaxBackoffTime is
// applied.
func WithDefaultRetry(
	timeout time.Duration,
	doFn func(ctx context.Context) error,
) error {
	return WithRetry(DefaultBackoffTime, DefaultMaxBackoffTime, timeout, doFn)
}

// WithRetry executes the provided doFn as long as it returns an error or until
// a timeout is hit. It applies exponential backoff wait of backoffTime * 2^n
// before n retry of doFn. In case the calculated backoff is longer than
// backoffMax, the backoffMax wait is applied.
func WithRetry(
	backoffTime time.Duration,
	backoffMax time.Duration,
	timeout time.Duration,
	doFn func(ctx context.Context) error,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry timeout [%v] exceeded", timeout)
		default:
			err := doFn(ctx)
			if err == nil {
				return nil
			}

			ok := backoffWait(ctx, backoffTime)
			if !ok {
				return fmt.Errorf("retry timeout [%v] exceeded", timeout)
			}

			backoffTime *= 2
			if backoffTime > backoffMax {
				backoffTime = backoffMax
			}
		}
	}
}

func backoffWait(ctx context.Context, waitTime time.Duration) bool {
	timer := time.NewTimer(waitTime)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
