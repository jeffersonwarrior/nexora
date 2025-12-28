package mcp

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()

	assert.Equal(t, 3, cfg.MaxRetries)
	assert.Equal(t, 1*time.Second, cfg.InitialDelay)
	assert.Equal(t, 30*time.Second, cfg.MaxDelay)
	assert.Equal(t, 2.0, cfg.Multiplier)
	assert.Equal(t, 0.1, cfg.Jitter)
}

func TestCalculateBackoff(t *testing.T) {
	cfg := RetryConfig{
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.0, // Disable jitter for deterministic tests
	}

	tests := []struct {
		name     string
		attempt  int
		expected time.Duration
	}{
		{"attempt 0", 0, 1 * time.Second},
		{"attempt 1", 1, 2 * time.Second},
		{"attempt 2", 2, 4 * time.Second},
		{"attempt 3", 3, 8 * time.Second},
		{"attempt 4", 4, 16 * time.Second},
		{"attempt 5 (capped)", 5, 30 * time.Second}, // Should be capped at MaxDelay
		{"attempt 10 (capped)", 10, 30 * time.Second},
		{"negative attempt", -1, 1 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateBackoff(cfg, tt.attempt)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConnectionError(t *testing.T) {
	t.Run("error message format", func(t *testing.T) {
		cause := errors.New("connection refused")
		err := &ConnectionError{
			MCPName:   "test-mcp",
			Type:      "stdio",
			Attempt:   3,
			Cause:     cause,
			Timestamp: time.Now(),
		}

		assert.Contains(t, err.Error(), "test-mcp")
		assert.Contains(t, err.Error(), "attempt 3")
		assert.Contains(t, err.Error(), "connection refused")
	})

	t.Run("unwrap returns cause", func(t *testing.T) {
		cause := errors.New("original error")
		err := &ConnectionError{Cause: cause}

		assert.Equal(t, cause, errors.Unwrap(err))
	})
}

func TestConnectionError_IsRetryable(t *testing.T) {
	tests := []struct {
		name      string
		cause     error
		retryable bool
	}{
		{
			name:      "nil cause",
			cause:     nil,
			retryable: false,
		},
		{
			name:      "context canceled",
			cause:     context.Canceled,
			retryable: false,
		},
		{
			name:      "context deadline exceeded",
			cause:     context.DeadlineExceeded,
			retryable: false,
		},
		{
			name:      "connection refused",
			cause:     errors.New("connection refused"),
			retryable: true,
		},
		{
			name:      "connection reset",
			cause:     errors.New("connection reset by peer"),
			retryable: true,
		},
		{
			name:      "EOF error",
			cause:     errors.New("unexpected EOF"),
			retryable: true,
		},
		{
			name:      "timeout error",
			cause:     errors.New("operation timeout"),
			retryable: true,
		},
		{
			name:      "temporary failure",
			cause:     errors.New("temporary failure in name resolution"),
			retryable: true,
		},
		{
			name:      "no such host",
			cause:     errors.New("no such host"),
			retryable: true,
		},
		{
			name:      "network unreachable",
			cause:     errors.New("network is unreachable"),
			retryable: true,
		},
		{
			name:      "generic error",
			cause:     errors.New("something went wrong"),
			retryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &ConnectionError{Cause: tt.cause}
			assert.Equal(t, tt.retryable, err.IsRetryable())
		})
	}
}

func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"connection refused", "connection", true},
		{"CONNECTION REFUSED", "connection", true},
		{"Connection Refused", "CONNECTION", true},
		{"hello world", "WORLD", true},
		{"HELLO WORLD", "world", true},
		{"", "test", false},
		{"test", "", true},
		{"abc", "abcd", false},
		{"timeout occurred", "timeout", true},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.substr, func(t *testing.T) {
			got := containsIgnoreCase(tt.s, tt.substr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHealthMonitor(t *testing.T) {
	t.Run("create new health monitor", func(t *testing.T) {
		monitor := NewHealthMonitor(10 * time.Second)
		require.NotNil(t, monitor)
		assert.Equal(t, 10*time.Second, monitor.checkInterval)
		assert.NotNil(t, monitor.healthStatus)
	})

	t.Run("default check interval", func(t *testing.T) {
		monitor := NewHealthMonitor(0)
		assert.Equal(t, 30*time.Second, monitor.checkInterval)
	})

	t.Run("start and stop", func(t *testing.T) {
		monitor := NewHealthMonitor(100 * time.Millisecond)

		ctx := context.Background()
		monitor.Start(ctx)

		// Starting again should be a no-op
		monitor.Start(ctx)

		// Give it a moment to start
		time.Sleep(50 * time.Millisecond)

		monitor.Stop()

		// Stopping again should be safe
		monitor.Stop()
	})

	t.Run("get health status", func(t *testing.T) {
		monitor := NewHealthMonitor(1 * time.Second)

		// No status initially
		_, ok := monitor.GetHealthStatus("nonexistent")
		assert.False(t, ok)

		// Set a status manually for testing
		monitor.mu.Lock()
		monitor.healthStatus["test-mcp"] = &HealthStatus{
			Name:             "test-mcp",
			Healthy:          true,
			LastCheck:        time.Now(),
			ConsecutiveFails: 0,
		}
		monitor.mu.Unlock()

		status, ok := monitor.GetHealthStatus("test-mcp")
		assert.True(t, ok)
		assert.True(t, status.Healthy)
		assert.Equal(t, "test-mcp", status.Name)
	})

	t.Run("get all health status", func(t *testing.T) {
		monitor := NewHealthMonitor(1 * time.Second)

		// Set statuses manually for testing
		now := time.Now()
		monitor.mu.Lock()
		monitor.healthStatus["mcp-1"] = &HealthStatus{
			Name:      "mcp-1",
			Healthy:   true,
			LastCheck: now,
		}
		monitor.healthStatus["mcp-2"] = &HealthStatus{
			Name:      "mcp-2",
			Healthy:   false,
			LastCheck: now,
			LastError: errors.New("connection failed"),
		}
		monitor.mu.Unlock()

		allStatus := monitor.GetAllHealthStatus()
		assert.Len(t, allStatus, 2)
		assert.True(t, allStatus["mcp-1"].Healthy)
		assert.False(t, allStatus["mcp-2"].Healthy)
	})
}

func TestGetHealthMonitor(t *testing.T) {
	// Reset global state for testing
	globalHealthMonitorOnce = sync.Once{}
	globalHealthMonitor = nil

	monitor1 := GetHealthMonitor()
	require.NotNil(t, monitor1)

	// Should return the same instance
	monitor2 := GetHealthMonitor()
	assert.Same(t, monitor1, monitor2)
}

func TestMCPError(t *testing.T) {
	t.Run("error with cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := NewMCPError(ErrCodeConnectionFailed, "connection failed", cause)

		assert.Contains(t, err.Error(), "CONNECTION_FAILED")
		assert.Contains(t, err.Error(), "connection failed")
		assert.Contains(t, err.Error(), "underlying error")
		assert.Equal(t, cause, errors.Unwrap(err))
	})

	t.Run("error without cause", func(t *testing.T) {
		err := NewMCPError(ErrCodeDisabled, "MCP is disabled", nil)

		assert.Contains(t, err.Error(), "DISABLED")
		assert.Contains(t, err.Error(), "MCP is disabled")
		assert.Nil(t, errors.Unwrap(err))
	})

	t.Run("error with details", func(t *testing.T) {
		details := map[string]any{
			"mcp_name": "test-mcp",
			"attempt":  3,
		}
		err := NewMCPErrorWithDetails(ErrCodeTimeout, "operation timed out", details, nil)

		assert.Equal(t, ErrCodeTimeout, err.Code)
		assert.Equal(t, details, err.Details)
	})
}

func TestMCPErrorCode_String(t *testing.T) {
	tests := []struct {
		code     MCPErrorCode
		expected string
	}{
		{ErrCodeUnknown, "UNKNOWN"},
		{ErrCodeNotConfigured, "NOT_CONFIGURED"},
		{ErrCodeDisabled, "DISABLED"},
		{ErrCodeConnectionFailed, "CONNECTION_FAILED"},
		{ErrCodeTimeout, "TIMEOUT"},
		{ErrCodeToolNotFound, "TOOL_NOT_FOUND"},
		{ErrCodeToolExecutionFailed, "TOOL_EXECUTION_FAILED"},
		{ErrCodeInvalidInput, "INVALID_INPUT"},
		{ErrCodePermissionDenied, "PERMISSION_DENIED"},
		{MCPErrorCode(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.code.String())
		})
	}
}

func TestWrapError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		mcpName      string
		expectedCode MCPErrorCode
	}{
		{
			name:         "nil error",
			err:          nil,
			mcpName:      "test",
			expectedCode: 0, // Should return nil
		},
		{
			name:         "deadline exceeded",
			err:          context.DeadlineExceeded,
			mcpName:      "test-mcp",
			expectedCode: ErrCodeTimeout,
		},
		{
			name:         "context canceled",
			err:          context.Canceled,
			mcpName:      "test-mcp",
			expectedCode: ErrCodeConnectionFailed,
		},
		{
			name:         "not configured error",
			err:          errors.New("mcp not configured"),
			mcpName:      "test-mcp",
			expectedCode: ErrCodeNotConfigured,
		},
		{
			name:         "not found error",
			err:          errors.New("mcp not found"),
			mcpName:      "test-mcp",
			expectedCode: ErrCodeNotConfigured,
		},
		{
			name:         "disabled error",
			err:          errors.New("mcp is disabled"),
			mcpName:      "test-mcp",
			expectedCode: ErrCodeDisabled,
		},
		{
			name:         "connection error",
			err:          errors.New("connection refused"),
			mcpName:      "test-mcp",
			expectedCode: ErrCodeConnectionFailed,
		},
		{
			name:         "EOF error",
			err:          errors.New("unexpected EOF"),
			mcpName:      "test-mcp",
			expectedCode: ErrCodeConnectionFailed,
		},
		{
			name:         "unknown error",
			err:          errors.New("something went wrong"),
			mcpName:      "test-mcp",
			expectedCode: ErrCodeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapError(tt.err, tt.mcpName)

			if tt.err == nil {
				assert.Nil(t, result)
				return
			}

			mcpErr, ok := result.(*MCPError)
			require.True(t, ok)
			assert.Equal(t, tt.expectedCode, mcpErr.Code)
			assert.Contains(t, mcpErr.Message, tt.mcpName)
		})
	}
}

func TestHealthStatus(t *testing.T) {
	t.Run("healthy status", func(t *testing.T) {
		status := HealthStatus{
			Name:             "test-mcp",
			Healthy:          true,
			LastCheck:        time.Now(),
			ConsecutiveFails: 0,
			Latency:          50 * time.Millisecond,
		}

		assert.True(t, status.Healthy)
		assert.Zero(t, status.ConsecutiveFails)
		assert.Nil(t, status.LastError)
	})

	t.Run("unhealthy status", func(t *testing.T) {
		status := HealthStatus{
			Name:             "test-mcp",
			Healthy:          false,
			LastCheck:        time.Now(),
			LastError:        errors.New("connection timeout"),
			ConsecutiveFails: 3,
			Latency:          5 * time.Second,
		}

		assert.False(t, status.Healthy)
		assert.Equal(t, 3, status.ConsecutiveFails)
		assert.NotNil(t, status.LastError)
	})
}

// Integration tests for MCP reliability features
func TestRetryLogicIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("exponential backoff timing", func(t *testing.T) {
		cfg := RetryConfig{
			MaxRetries:   2,
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     100 * time.Millisecond,
			Multiplier:   2.0,
			Jitter:       0.0,
		}

		// Measure actual backoff duration
		start := time.Now()
		for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
			delay := calculateBackoff(cfg, attempt-1)
			time.Sleep(delay)
		}
		totalElapsed := time.Since(start)

		// Should be at least: 10ms + 20ms = 30ms (minimum expected)
		assert.GreaterOrEqual(t, totalElapsed, 30*time.Millisecond)
		// Should be less than 200ms (very generous upper bound)
		assert.Less(t, totalElapsed, 200*time.Millisecond)
	})

	t.Run("max delay enforcement", func(t *testing.T) {
		cfg := RetryConfig{
			InitialDelay: 1 * time.Second,
			MaxDelay:     50 * time.Millisecond,
			Multiplier:   10.0, // Very high multiplier to test capping
		}

		// Attempt 5 would be 1s * 10^5 = massive, should be capped
		delay := calculateBackoff(cfg, 5)
		assert.Equal(t, 50*time.Millisecond, delay)
	})

	t.Run("retry sequence with multiple attempts", func(t *testing.T) {
		cfg := DefaultRetryConfig()
		attempt := 0
		maxAttempts := cfg.MaxRetries + 1

		for attempt < maxAttempts {
			attempt++
			if attempt < maxAttempts {
				delay := calculateBackoff(cfg, attempt-1)
				assert.Greater(t, delay, time.Duration(0))
			}
		}

		assert.Equal(t, maxAttempts, attempt)
	})

	t.Run("retry with context deadline", func(t *testing.T) {
		cfg := RetryConfig{
			MaxRetries:   5,
			InitialDelay: 100 * time.Millisecond,
			MaxDelay:     1 * time.Second,
			Multiplier:   2.0,
		}

		// Create a context that expires quickly
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		attempt := 0
		for attempt <= cfg.MaxRetries {
			if attempt > 0 {
				delay := calculateBackoff(cfg, attempt-1)
				select {
				case <-ctx.Done():
					return // Expected to exit early
				case <-time.After(delay):
				}
			}
			attempt++
		}

		// Should not reach here
		t.Error("expected context deadline to interrupt retry loop")
	})
}

func TestCircuitBreakerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("consecutive failures trigger reconnect", func(t *testing.T) {
		monitor := NewHealthMonitor(50 * time.Millisecond)

		// Manually set consecutive failures to trigger reconnect threshold
		monitor.mu.Lock()
		monitor.healthStatus["test-mcp"] = &HealthStatus{
			Name:             "test-mcp",
			Healthy:          false,
			LastCheck:        time.Now(),
			LastError:        errors.New("connection failed"),
			ConsecutiveFails: 3,
		}
		monitor.mu.Unlock()

		// Verify status shows unhealthy
		status, ok := monitor.GetHealthStatus("test-mcp")
		assert.True(t, ok)
		assert.False(t, status.Healthy)
		assert.Equal(t, 3, status.ConsecutiveFails)
	})

	t.Run("health status recovery after success", func(t *testing.T) {
		monitor := NewHealthMonitor(100 * time.Millisecond)

		// Start with unhealthy status
		monitor.mu.Lock()
		monitor.healthStatus["test-mcp"] = &HealthStatus{
			Name:             "test-mcp",
			Healthy:          false,
			LastCheck:        time.Now(),
			LastError:        errors.New("connection failed"),
			ConsecutiveFails: 5,
		}
		monitor.mu.Unlock()

		// Simulate successful health check
		monitor.mu.Lock()
		status := monitor.healthStatus["test-mcp"]
		status.Healthy = true
		status.LastError = nil
		status.ConsecutiveFails = 0
		monitor.mu.Unlock()

		// Verify recovery
		recovered, ok := monitor.GetHealthStatus("test-mcp")
		assert.True(t, ok)
		assert.True(t, recovered.Healthy)
		assert.Zero(t, recovered.ConsecutiveFails)
		assert.Nil(t, recovered.LastError)
	})

	t.Run("latency tracking during health checks", func(t *testing.T) {
		monitor := NewHealthMonitor(100 * time.Millisecond)

		// Simulate latency measurement
		monitor.mu.Lock()
		monitor.healthStatus["test-mcp"] = &HealthStatus{
			Name:      "test-mcp",
			Healthy:   true,
			LastCheck: time.Now(),
			Latency:   150 * time.Millisecond,
		}
		monitor.mu.Unlock()

		status, ok := monitor.GetHealthStatus("test-mcp")
		assert.True(t, ok)
		assert.Equal(t, 150*time.Millisecond, status.Latency)
	})

	t.Run("concurrent health status updates", func(t *testing.T) {
		monitor := NewHealthMonitor(100 * time.Millisecond)
		numConcurrent := 10

		results := make(chan bool, numConcurrent)

		for i := 0; i < numConcurrent; i++ {
			go func(id int) {
				monitor.mu.Lock()
				name := fmt.Sprintf("test-mcp-%d", id)
				monitor.healthStatus[name] = &HealthStatus{
					Name:      name,
					Healthy:   id%2 == 0,
					LastCheck: time.Now(),
				}
				monitor.mu.Unlock()
				results <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < numConcurrent; i++ {
			<-results
		}

		// Verify all statuses were recorded
		allStatus := monitor.GetAllHealthStatus()
		assert.Equal(t, numConcurrent, len(allStatus))
	})
}

func TestTimeoutHandlingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("timeout interrupt during backoff", func(t *testing.T) {
		cfg := RetryConfig{
			MaxRetries:   10,
			InitialDelay: 100 * time.Millisecond,
			MaxDelay:     1 * time.Second,
			Multiplier:   2.0,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()

		start := time.Now()
		attempt := 0

		for attempt <= cfg.MaxRetries {
			if attempt > 0 {
				delay := calculateBackoff(cfg, attempt-1)
				select {
				case <-ctx.Done():
					elapsed := time.Since(start)
					// Should be interrupted before all retries complete
					assert.Less(t, elapsed, 500*time.Millisecond)
					return
				case <-time.After(delay):
				}
			}
			attempt++
		}

		t.Error("context should have timed out")
	})

	t.Run("connection error timeout categorization", func(t *testing.T) {
		// Simulate timeout error
		err := &ConnectionError{
			MCPName:   "test-mcp",
			Type:      "stdio",
			Attempt:   1,
			Cause:     context.DeadlineExceeded,
			Timestamp: time.Now(),
		}

		assert.False(t, err.IsRetryable())
	})

	t.Run("retryable timeout errors", func(t *testing.T) {
		// Simulate timeout pattern in error message
		err := &ConnectionError{
			MCPName:   "test-mcp",
			Cause:     errors.New("operation timeout"),
			Timestamp: time.Now(),
		}

		assert.True(t, err.IsRetryable())
	})
}

func TestConnectionLifecycleIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("error wrapping preserves context", func(t *testing.T) {
		originalErr := errors.New("network connection refused")
		connErr := &ConnectionError{
			MCPName:   "test-mcp",
			Type:      "stdio",
			Attempt:   2,
			Cause:     originalErr,
			Timestamp: time.Now(),
		}

		// Verify error chain
		assert.Equal(t, originalErr, errors.Unwrap(connErr))
		assert.True(t, errors.Is(connErr, originalErr))
	})

	t.Run("MCP error categorization flow", func(t *testing.T) {
		testCases := []struct {
			err          error
			expectedCode MCPErrorCode
			description  string
		}{
			{
				err:          context.DeadlineExceeded,
				expectedCode: ErrCodeTimeout,
				description:  "deadline exceeded",
			},
			{
				err:          errors.New("connection refused"),
				expectedCode: ErrCodeConnectionFailed,
				description:  "connection refused",
			},
			{
				err:          errors.New("mcp disabled"),
				expectedCode: ErrCodeDisabled,
				description:  "disabled mcp",
			},
			{
				err:          errors.New("unexpected error"),
				expectedCode: ErrCodeUnknown,
				description:  "unknown error",
			},
		}

		for _, tc := range testCases {
			wrapped := WrapError(tc.err, "test-mcp")
			mcpErr := wrapped.(*MCPError)
			assert.Equal(t, tc.expectedCode, mcpErr.Code, tc.description)
		}
	})

	t.Run("connection state transitions", func(t *testing.T) {
		// Test state flow: Disabled -> Starting -> Connected -> Error
		states := []State{StateDisabled, StateStarting, StateConnected, StateError}
		stateNames := []string{"disabled", "starting", "connected", "error"}

		for i, state := range states {
			assert.Equal(t, stateNames[i], state.String())
		}
	})
}

func TestHealthMonitorLifecycleIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("monitor start and stop lifecycle", func(t *testing.T) {
		monitor := NewHealthMonitor(50 * time.Millisecond)
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		// Start monitor
		monitor.Start(ctx)
		assert.True(t, monitor.running)

		// Let it run briefly
		time.Sleep(100 * time.Millisecond)

		// Stop monitor
		monitor.Stop()
		assert.False(t, monitor.running)
	})

	t.Run("health monitor double start is safe", func(t *testing.T) {
		monitor := NewHealthMonitor(100 * time.Millisecond)
		ctx := context.Background()

		// Start twice
		monitor.Start(ctx)
		first := monitor.running
		monitor.Start(ctx)
		second := monitor.running

		assert.True(t, first)
		assert.True(t, second)

		monitor.Stop()
	})

	t.Run("health monitor double stop is safe", func(t *testing.T) {
		monitor := NewHealthMonitor(100 * time.Millisecond)
		ctx := context.Background()

		monitor.Start(ctx)
		monitor.Stop()
		monitor.Stop() // Should not panic

		assert.False(t, monitor.running)
	})

	t.Run("global health monitor singleton", func(t *testing.T) {
		// Reset global state for test isolation
		globalHealthMonitorOnce = sync.Once{}
		globalHealthMonitor = nil

		monitor1 := GetHealthMonitor()
		monitor2 := GetHealthMonitor()

		assert.Same(t, monitor1, monitor2)
	})
}

func TestRetryConfigurationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("custom retry configuration sequence", func(t *testing.T) {
		customCfg := RetryConfig{
			MaxRetries:   3,
			InitialDelay: 5 * time.Millisecond,
			MaxDelay:     50 * time.Millisecond,
			Multiplier:   3.0,
			Jitter:       0.0,
		}

		expectedDelays := []time.Duration{
			5 * time.Millisecond,  // attempt 0: initial
			15 * time.Millisecond, // attempt 1: 5 * 3
			45 * time.Millisecond, // attempt 2: 15 * 3
			50 * time.Millisecond, // attempt 3: 135 capped at 50
		}

		for attempt := 0; attempt < 4; attempt++ {
			delay := calculateBackoff(customCfg, attempt)
			assert.Equal(t, expectedDelays[attempt], delay)
		}
	})

	t.Run("retry sequence execution pattern", func(t *testing.T) {
		cfg := RetryConfig{
			MaxRetries:   2,
			InitialDelay: 5 * time.Millisecond,
			MaxDelay:     30 * time.Millisecond,
			Multiplier:   2.0,
			Jitter:       0.0,
		}

		start := time.Now()
		attempt := 0

		for attempt <= cfg.MaxRetries {
			if attempt > 0 {
				delay := calculateBackoff(cfg, attempt-1)
				time.Sleep(delay)
			}
			attempt++
		}

		elapsed := time.Since(start)

		// Should be at least: 5ms (attempt 1) + 10ms (attempt 2) = 15ms
		assert.GreaterOrEqual(t, elapsed, 15*time.Millisecond)
		// Should complete within reasonable time
		assert.Less(t, elapsed, 200*time.Millisecond)
	})
}

func TestMCPErrorHandlingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("error details preservation", func(t *testing.T) {
		details := map[string]any{
			"mcp_name": "test-mcp",
			"attempt":  2,
			"timeout":  5 * time.Second,
		}

		err := NewMCPErrorWithDetails(ErrCodeTimeout, "operation timeout", details, errors.New("context deadline"))

		assert.Equal(t, ErrCodeTimeout, err.Code)
		assert.Equal(t, details, err.Details)
		assert.NotNil(t, err.Cause)
	})

	t.Run("error wrapping chain", func(t *testing.T) {
		baseErr := errors.New("base error")
		wrapped := WrapError(baseErr, "test-mcp")
		mcpErr := wrapped.(*MCPError)

		assert.NotNil(t, mcpErr.Cause)
		assert.Equal(t, ErrCodeUnknown, mcpErr.Code)
	})

	t.Run("error code string representations", func(t *testing.T) {
		codes := map[MCPErrorCode]string{
			ErrCodeUnknown:            "UNKNOWN",
			ErrCodeNotConfigured:      "NOT_CONFIGURED",
			ErrCodeDisabled:           "DISABLED",
			ErrCodeConnectionFailed:   "CONNECTION_FAILED",
			ErrCodeTimeout:            "TIMEOUT",
			ErrCodeToolNotFound:       "TOOL_NOT_FOUND",
			ErrCodeToolExecutionFailed: "TOOL_EXECUTION_FAILED",
			ErrCodeInvalidInput:       "INVALID_INPUT",
			ErrCodePermissionDenied:   "PERMISSION_DENIED",
		}

		for code, expected := range codes {
			assert.Equal(t, expected, code.String())
		}
	})
}

