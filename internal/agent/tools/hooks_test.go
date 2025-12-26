package tools

import (
	"context"
	"testing"

	"charm.land/fantasy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHookChain(t *testing.T) {
	hook1 := NewSecurityHook([]string{"bash"}, []string{})
	hook2 := NewMetricsHook()

	chain := NewHookChain(hook1, hook2)
	assert.NotNil(t, chain)
	assert.Len(t, chain.hooks, 2)
}

func TestHookChain_BeforeCall(t *testing.T) {
	hook := NewSecurityHook([]string{"bash"}, []string{})
	chain := NewHookChain(hook)

	ctx := context.Background()

	// Allowed tool
	err := chain.BeforeCall(ctx, "bash", map[string]interface{}{})
	assert.NoError(t, err)

	// Disallowed tool
	err = chain.BeforeCall(ctx, "forbidden", map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed")
}

func TestHookChain_AfterCall(t *testing.T) {
	hook := NewMetricsHook()
	chain := NewHookChain(hook)

	ctx := context.Background()

	// Must call BeforeCall first to initialize metrics
	chain.BeforeCall(ctx, "test-tool", nil)

	resp := fantasy.NewTextResponse("test")

	result := chain.AfterCall(ctx, "test-tool", resp)
	assert.Equal(t, resp, result)
}

func TestHookChain_OnError(t *testing.T) {
	hook := NewMetricsHook()
	chain := NewHookChain(hook)

	ctx := context.Background()

	// Must call BeforeCall first to initialize metrics
	chain.BeforeCall(ctx, "test-tool", nil)

	testErr := assert.AnError

	err := chain.OnError(ctx, "test-tool", testErr)
	assert.Error(t, err)
}

func TestSecurityHook_AllowedTools(t *testing.T) {
	hook := NewSecurityHook([]string{"bash", "read", "write"}, []string{})

	tests := []struct {
		name      string
		toolName  string
		wantError bool
	}{
		{
			name:      "allowed bash",
			toolName:  "bash",
			wantError: false,
		},
		{
			name:      "allowed read",
			toolName:  "read",
			wantError: false,
		},
		{
			name:      "disallowed tool",
			toolName:  "dangerous",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hook.BeforeCall(context.Background(), tt.toolName, nil)
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not allowed")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecurityHook_EmptyAllowList(t *testing.T) {
	// Empty allow list means all tools allowed
	hook := NewSecurityHook([]string{}, []string{})

	err := hook.BeforeCall(context.Background(), "any-tool", nil)
	assert.NoError(t, err)
}

func TestSecurityHook_AfterCallPassthrough(t *testing.T) {
	hook := NewSecurityHook([]string{"bash"}, []string{})
	resp := fantasy.NewTextResponse("test")

	result := hook.AfterCall(context.Background(), "bash", resp)
	assert.Equal(t, resp, result)
}

func TestSecurityHook_OnErrorPassthrough(t *testing.T) {
	hook := NewSecurityHook([]string{"bash"}, []string{})
	testErr := assert.AnError

	err := hook.OnError(context.Background(), "bash", testErr)
	assert.Equal(t, testErr, err)
}

func TestMetricsHook_TracksCalls(t *testing.T) {
	hook := NewMetricsHook()
	ctx := context.Background()

	// Make multiple calls
	err := hook.BeforeCall(ctx, "bash", nil)
	require.NoError(t, err)

	err = hook.BeforeCall(ctx, "bash", nil)
	require.NoError(t, err)

	err = hook.BeforeCall(ctx, "read", nil)
	require.NoError(t, err)

	// Check metrics
	bashMetrics := hook.GetMetrics("bash")
	assert.NotNil(t, bashMetrics)
	assert.Equal(t, int64(2), bashMetrics.CallCount)

	readMetrics := hook.GetMetrics("read")
	assert.NotNil(t, readMetrics)
	assert.Equal(t, int64(1), readMetrics.CallCount)
}

func TestMetricsHook_TracksSuccess(t *testing.T) {
	hook := NewMetricsHook()
	ctx := context.Background()

	// Setup tool
	hook.BeforeCall(ctx, "test", nil)

	// Successful response
	resp := fantasy.NewTextResponse("success")
	hook.AfterCall(ctx, "test", resp)

	metrics := hook.GetMetrics("test")
	assert.Equal(t, int64(1), metrics.SuccessCount)
	assert.Equal(t, int64(0), metrics.ErrorCount)
}

func TestMetricsHook_TracksErrors(t *testing.T) {
	hook := NewMetricsHook()
	ctx := context.Background()

	// Setup tool
	hook.BeforeCall(ctx, "test", nil)

	// Error
	hook.OnError(ctx, "test", assert.AnError)

	metrics := hook.GetMetrics("test")
	assert.Equal(t, int64(1), metrics.ErrorCount)
}

func TestMetricsHook_AllMetrics(t *testing.T) {
	hook := NewMetricsHook()
	ctx := context.Background()

	// Track multiple tools
	hook.BeforeCall(ctx, "bash", nil)
	hook.BeforeCall(ctx, "read", nil)
	hook.BeforeCall(ctx, "write", nil)

	allMetrics := hook.AllMetrics()
	assert.Len(t, allMetrics, 3)
	assert.Contains(t, allMetrics, "bash")
	assert.Contains(t, allMetrics, "read")
	assert.Contains(t, allMetrics, "write")
}

func TestMetricsHook_GetMetricsReturnsNil(t *testing.T) {
	hook := NewMetricsHook()

	metrics := hook.GetMetrics("nonexistent")
	assert.Nil(t, metrics)
}

func TestNameOfHook(t *testing.T) {
	tests := []struct {
		name     string
		hook     ToolHook
		expected string
	}{
		{
			name:     "security hook",
			hook:     NewSecurityHook([]string{}, []string{}),
			expected: "SecurityHook",
		},
		{
			name:     "metrics hook",
			hook:     NewMetricsHook(),
			expected: "MetricsHook",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nameOfHook(tt.hook)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrToolNotAllowed(t *testing.T) {
	err := ErrToolNotAllowed("dangerous-tool")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dangerous-tool")
	assert.Contains(t, err.Error(), "not allowed")
}

func TestHookChain_MultipleHooksSequential(t *testing.T) {
	security := NewSecurityHook([]string{"allowed"}, []string{})
	metrics := NewMetricsHook()
	chain := NewHookChain(security, metrics)

	ctx := context.Background()

	// Allowed tool - both hooks should execute
	err := chain.BeforeCall(ctx, "allowed", nil)
	assert.NoError(t, err)

	// Verify metrics hook was called
	m := metrics.GetMetrics("allowed")
	assert.NotNil(t, m)
	assert.Equal(t, int64(1), m.CallCount)

	// Disallowed tool - should stop at security hook
	err = chain.BeforeCall(ctx, "disallowed", nil)
	assert.Error(t, err)

	// Metrics should still be zero for disallowed
	m = metrics.GetMetrics("disallowed")
	assert.Nil(t, m)
}

func TestHookChain_EmptyChain(t *testing.T) {
	chain := NewHookChain()

	ctx := context.Background()

	// Should not error with empty chain
	err := chain.BeforeCall(ctx, "any", nil)
	assert.NoError(t, err)

	resp := fantasy.NewTextResponse("test")
	result := chain.AfterCall(ctx, "any", resp)
	assert.Equal(t, resp, result)

	err = chain.OnError(ctx, "any", assert.AnError)
	assert.Error(t, err)
}
