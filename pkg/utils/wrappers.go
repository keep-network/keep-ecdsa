package utils

import (
	"context"
	"fmt"
	"time"
)

// DoWithRetry executes the provided doFn as long as it returns an error or until
// a timeout is hit. It applies exponential backoff wait of backoffTime * 2^n
// before n retry of doFn. In case the calculated backoff is longer than
// backoffMax, the backoffMax wait is applied.
func DoWithRetry(
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

			timedOut := backoffWait(ctx, backoffTime)
			if timedOut {
				return fmt.Errorf("retry timeout [%v] exceeded", timeout)
			}

			backoffTime *= 2
			if backoffTime > backoffMax {
				backoffTime = backoffMax
			}
		}
	}
}

const (
	// DefaultDoBackoffTime is the default value of backoff time used by
	// DoWithDefaultRetry function.
	DefaultDoBackoffTime = 1 * time.Second

	// DefaultDoMaxBackoffTime is the default value of max backoff time used by
	// DoWithDefaultRetry function.
	DefaultDoMaxBackoffTime = 120 * time.Second
)

// DoWithDefaultRetry executes the provided doFn as long as it returns an error or
// until a timeout is hit. It applies exponential backoff wait of
// DefaultBackoffTime * 2^n before n retry of doFn. In case the calculated
// backoff is longer than DefaultMaxBackoffTime, the DefaultMaxBackoffTime is
// applied.
func DoWithDefaultRetry(
	timeout time.Duration,
	doFn func(ctx context.Context) error,
) error {
	return DoWithRetry(
		DefaultDoBackoffTime,
		DefaultDoMaxBackoffTime,
		timeout,
		doFn,
	)
}

// ConfirmWithTimeout executes the provided confirmFn until it returns true or
// until it fails or until a timeout is hit. It applies exponential backoff wait
// of backoffTime * 2^n before n retry of confirmFn. In case the calculated
// backoff is longer than backoffMax, the backoffMax is applied.
// In case confirmFn returns an error, ConfirmWithTimeout exits with the same
// error immediately. This is different from DoWithRetry behavior as the use
// case for this function is different. ConfirmWithTimeout is intended to be
// used to confirm a chain state and not to try to enforce a successful
// execution of some function.
func ConfirmWithTimeout(
	backoffTime time.Duration,
	backoffMax time.Duration,
	timeout time.Duration,
	confirmFn func(ctx context.Context) (bool, error),
) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return false, nil
		default:
			ok, err := confirmFn(ctx)
			if err == nil && ok {
				return true, nil
			}
			if err != nil {
				return false, err
			}

			timedOut := backoffWait(ctx, backoffTime)
			if timedOut {
				return false, nil
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
		return true
	case <-timer.C:
		return false
	}
}