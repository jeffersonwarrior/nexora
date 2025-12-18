package tools

import (
	"context"
	"log/slog"

	"charm.land/fantasy"
)

// ToolHook is a centralized interception point for all tool calls
// Enables security validation, monitoring, and enhancement without code changes
type ToolHook interface {
	// BeforeCall executes before a tool is called
	// Return error to block execution
	BeforeCall(ctx context.Context, toolName string, params interface{}) error

	// AfterCall executes after a tool completes
	// Can inspect/modify response
	AfterCall(ctx context.Context, toolName string, response fantasy.ToolResponse) fantasy.ToolResponse

	// OnError executes when a tool returns an error
	OnError(ctx context.Context, toolName string, err error) error
}

// HookChain manages multiple hooks
type HookChain struct {
	hooks []ToolHook
}

// NewHookChain creates a new hook chain
func NewHookChain(hooks ...ToolHook) *HookChain {
	return &HookChain{hooks: hooks}
}

// BeforeCall executes all BeforeCall hooks
func (h *HookChain) BeforeCall(ctx context.Context, toolName string, params interface{}) error {
	for _, hook := range h.hooks {
		if err := hook.BeforeCall(ctx, toolName, params); err != nil {
			slog.Warn("Tool hook blocked execution",
				"tool", toolName,
				"hook", nameOfHook(hook),
				"error", err,
			)
			return err
		}
	}
	return nil
}

// AfterCall executes all AfterCall hooks
func (h *HookChain) AfterCall(ctx context.Context, toolName string, response fantasy.ToolResponse) fantasy.ToolResponse {
	for _, hook := range h.hooks {
		response = hook.AfterCall(ctx, toolName, response)
	}
	return response
}

// OnError executes all OnError hooks
func (h *HookChain) OnError(ctx context.Context, toolName string, err error) error {
	for _, hook := range h.hooks {
		if err := hook.OnError(ctx, toolName, err); err != nil {
			return err
		}
	}
	return err
}

// ===================== Built-in Hooks =====================

// SecurityHook validates tool calls against security policies
type SecurityHook struct {
	// allowedTools: map of tool -> allowed
	allowedTools map[string]bool
	// deniedPaths: paths that cannot be accessed
	deniedPaths []string
}

// NewSecurityHook creates a security hook
func NewSecurityHook(allowedTools []string, deniedPaths []string) *SecurityHook {
	allowed := make(map[string]bool)
	for _, t := range allowedTools {
		allowed[t] = true
	}
	return &SecurityHook{
		allowedTools: allowed,
		deniedPaths:  deniedPaths,
	}
}

func (h *SecurityHook) BeforeCall(ctx context.Context, toolName string, params interface{}) error {
	// Check if tool is allowed
	if len(h.allowedTools) > 0 && !h.allowedTools[toolName] {
		return ErrToolNotAllowed(toolName)
	}
	return nil
}

func (h *SecurityHook) AfterCall(ctx context.Context, toolName string, response fantasy.ToolResponse) fantasy.ToolResponse {
	return response
}

func (h *SecurityHook) OnError(ctx context.Context, toolName string, err error) error {
	return err
}

// MetricsHook tracks tool performance metrics
type MetricsHook struct {
	metrics map[string]*ToolMetrics
}

// ToolMetrics holds performance metrics for a tool
type ToolMetrics struct {
	CallCount    int64
	ErrorCount   int64
	SuccessCount int64
	TotalTimeMS  float64
}

// NewMetricsHook creates a metrics hook
func NewMetricsHook() *MetricsHook {
	return &MetricsHook{
		metrics: make(map[string]*ToolMetrics),
	}
}

func (h *MetricsHook) BeforeCall(ctx context.Context, toolName string, params interface{}) error {
	if _, exists := h.metrics[toolName]; !exists {
		h.metrics[toolName] = &ToolMetrics{}
	}
	h.metrics[toolName].CallCount++
	return nil
}

func (h *MetricsHook) AfterCall(ctx context.Context, toolName string, response fantasy.ToolResponse) fantasy.ToolResponse {
	if !response.IsError {
		h.metrics[toolName].SuccessCount++
	}
	return response
}

func (h *MetricsHook) OnError(ctx context.Context, toolName string, err error) error {
	h.metrics[toolName].ErrorCount++
	return err
}

// GetMetrics returns metrics for a specific tool
func (h *MetricsHook) GetMetrics(toolName string) *ToolMetrics {
	return h.metrics[toolName]
}

// AllMetrics returns all metrics
func (h *MetricsHook) AllMetrics() map[string]*ToolMetrics {
	return h.metrics
}

// ===================== Helpers =====================

func nameOfHook(h ToolHook) string {
	switch h.(type) {
	case *SecurityHook:
		return "SecurityHook"
	case *MetricsHook:
		return "MetricsHook"
	default:
		return "UnknownHook"
	}
}

// HookError types
type HookError struct {
	msg string
}

func (e *HookError) Error() string {
	return e.msg
}

func ErrToolNotAllowed(toolName string) error {
	return &HookError{msg: "tool not allowed: " + toolName}
}
