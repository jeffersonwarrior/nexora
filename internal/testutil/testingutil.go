package testutil

import (
	"context"
	"os"
	"testing"
	"time"
)

// RunWithTimeout runs a test function with a timeout, preventing tests from running forever
// Example usage:
//   func TestLongRunningOperation(t *testing.T) {
//       testingutil.RunWithTimeout(t, 10*time.Second, func(ctx context.Context) {
//           // Your test code here, using ctx.Context() for cancellation support
//           result := doLongOperation(ctx.Context())
//           if result != expected {
//               t.Error("unexpected result")
//           }
//       })
//   }
func RunWithTimeout(t *testing.T, timeout time.Duration, testFunc func(ctx context.Context)) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan struct{})
	
	// Run the test in a goroutine
	go func() {
		defer close(done)
		testFunc(ctx)
	}()

	select {
	case <-done:
		// Test completed normally
	case <-ctx.Done():
		// Test timed out
		t.Fatalf("test timed out after %v", timeout)
	}
}

// RunWithTimeoutAndDeadline runs a test with both a timeout and an explicit deadline
// This is useful for tests that should complete within a specific time window
func RunWithTimeoutAndDeadline(t *testing.T, timeout time.Duration, deadline time.Time, testFunc func(ctx context.Context)) {
	t.Helper()

	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	if timeout > 0 {
		var cancelTimeout context.CancelFunc
		ctx, cancelTimeout = context.WithTimeout(ctx, timeout)
		defer cancelTimeout()
	}

	done := make(chan struct{})
	
	// Run the test in a goroutine
	go func() {
		defer close(done)
		testFunc(ctx)
	}()

	select {
	case <-done:
		// Test completed normally
	case <-ctx.Done():
		// Test timed out or missed deadline
		if ctx.Err() == context.DeadlineExceeded {
			t.Fatalf("test missed deadline %v", deadline)
		} else {
			t.Fatalf("test timed out after %v", timeout)
		}
	}
}

// TimeoutEnvVar returns the timeout duration from an environment variable
// This allows customizing test timeouts without changing code
func TimeoutEnvVar(envVar string, defaultTimeout time.Duration) time.Duration {
	if s := os.Getenv(envVar); s != "" {
		if d, err := time.ParseDuration(s); err == nil {
			return d
		}
	}
	return defaultTimeout
}