package utils

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestWithRetry(t *testing.T) {
	backoffTime := 10 * time.Millisecond
	backoffMax := 100 * time.Millisecond
	timeout := 2 * time.Second

	actualFailCount := 0
	expectedFailCount := 4
	doFn := func(ctx context.Context) error {
		if actualFailCount < expectedFailCount {
			actualFailCount++
			return fmt.Errorf("try again please")
		}

		return nil
	}

	err := WithRetry(backoffTime, backoffMax, timeout, doFn)
	if err != nil {
		t.Fatal(err)
	}

	if actualFailCount != expectedFailCount {
		t.Errorf(
			"unexpected fail count: actual [%v], expected [%v]",
			actualFailCount,
			expectedFailCount,
		)
	}
}

func TestWithRetryTimeout(t *testing.T) {
	backoffTime := 100 * time.Millisecond
	backoffMax := 500 * time.Millisecond
	timeout := 1 * time.Second

	actualFailCount := 0

	// This function should be executed 4 times and timeout:
	// 100 ms
	// 200 ms
	// 400 ms
	// 500 ms <- here it should exceed 1s timeout
	doFn := func(ctx context.Context) error {
		if actualFailCount < 10 {
			actualFailCount++
			return fmt.Errorf("try again please")
		}

		return nil
	}

	err := WithRetry(backoffTime, backoffMax, timeout, doFn)
	if err == nil {
		t.Fatal("expected a timeout error")
	}

	expectedError := "retry timeout [1s] exceeded"
	if err.Error() != expectedError {
		t.Errorf(
			"unexpected error message\nactual:   [%v]\nexpected: [%v]",
			err.Error(),
			expectedError,
		)
	}

	expectedFailCount := 4
	if actualFailCount != expectedFailCount {
		t.Errorf(
			"unexpected fail count: actual [%v], expected [%v]",
			actualFailCount,
			expectedFailCount,
		)
	}
}
