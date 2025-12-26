package mcp

import (
	"context"
	"errors"
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

