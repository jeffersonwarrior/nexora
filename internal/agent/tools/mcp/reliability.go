// Package mcp provides functionality for managing Model Context Protocol (MCP)
// clients within the Nexora application.
package mcp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nexora/nexora/internal/config"
)

// RetryConfig configures the exponential backoff retry behavior
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts (0 = no retries)
	MaxRetries int

	// InitialDelay is the initial delay between retries
	InitialDelay time.Duration

	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration

	// Multiplier is the factor by which delay increases each retry
	Multiplier float64

	// Jitter adds randomness to prevent thundering herd (0.0 to 1.0)
	Jitter float64
}

// DefaultRetryConfig returns sensible defaults for MCP retry behavior
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:   3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.1,
	}
}

// calculateBackoff calculates the delay for a given retry attempt using exponential backoff
func calculateBackoff(cfg RetryConfig, attempt int) time.Duration {
	if attempt <= 0 {
		return cfg.InitialDelay
	}

	delay := float64(cfg.InitialDelay) * math.Pow(cfg.Multiplier, float64(attempt))
	if delay > float64(cfg.MaxDelay) {
		delay = float64(cfg.MaxDelay)
	}

	// Note: jitter would require randomness but we keep it deterministic for testing
	// In production, you might add: delay += delay * cfg.Jitter * (rand.Float64()*2 - 1)

	return time.Duration(delay)
}

// ConnectionError wraps MCP connection errors with additional context
type ConnectionError struct {
	MCPName   string
	Type      string
	Attempt   int
	Cause     error
	Timestamp time.Time
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("MCP '%s' connection error (attempt %d): %v", e.MCPName, e.Attempt, e.Cause)
}

func (e *ConnectionError) Unwrap() error {
	return e.Cause
}

// IsRetryable checks if the error is retryable
func (e *ConnectionError) IsRetryable() bool {
	if e.Cause == nil {
		return false
	}

	// Context cancelled/deadline exceeded are not retryable
	if errors.Is(e.Cause, context.Canceled) || errors.Is(e.Cause, context.DeadlineExceeded) {
		return false
	}

	// Check for common retryable conditions
	errStr := e.Cause.Error()
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"EOF",
		"timeout",
		"temporary failure",
		"no such host",
		"network is unreachable",
	}

	for _, pattern := range retryablePatterns {
		if containsIgnoreCase(errStr, pattern) {
			return true
		}
	}

	return false
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findIgnoreCase(s, substr))
}

func findIgnoreCase(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			sc, subc := s[i+j], substr[j]
			// Simple ASCII case-insensitive comparison
			if sc != subc && (sc^32) != subc && sc != (subc^32) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// HealthStatus represents the health status of an MCP connection
type HealthStatus struct {
	Name          string
	Healthy       bool
	LastCheck     time.Time
	LastError     error
	ConsecutiveFails int
	Latency       time.Duration
}

// HealthMonitor monitors the health of MCP connections
type HealthMonitor struct {
	mu           sync.RWMutex
	healthStatus map[string]*HealthStatus
	checkInterval time.Duration
	cancel       context.CancelFunc
	running      bool
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(checkInterval time.Duration) *HealthMonitor {
	if checkInterval <= 0 {
		checkInterval = 30 * time.Second
	}
	return &HealthMonitor{
		healthStatus: make(map[string]*HealthStatus),
		checkInterval: checkInterval,
	}
}

// Start begins periodic health checks for all MCP connections
func (m *HealthMonitor) Start(ctx context.Context) {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true

	monitorCtx, cancel := context.WithCancel(ctx)
	m.cancel = cancel
	m.mu.Unlock()

	go m.runHealthChecks(monitorCtx)
}

// Stop stops the health monitor
func (m *HealthMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cancel != nil {
		m.cancel()
	}
	m.running = false
}

// runHealthChecks performs periodic health checks
func (m *HealthMonitor) runHealthChecks(ctx context.Context) {
	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.checkAllConnections(ctx)
		}
	}
}

// checkAllConnections checks the health of all active MCP connections
func (m *HealthMonitor) checkAllConnections(ctx context.Context) {
	for name := range sessions.Seq2() {
		go m.checkConnection(ctx, name)
	}
}

// checkConnection checks the health of a single MCP connection
func (m *HealthMonitor) checkConnection(ctx context.Context, name string) {
	session, ok := sessions.Get(name)
	if !ok {
		return
	}

	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	start := time.Now()
	err := session.Ping(checkCtx, nil)
	latency := time.Since(start)

	m.mu.Lock()
	defer m.mu.Unlock()

	status, exists := m.healthStatus[name]
	if !exists {
		status = &HealthStatus{Name: name}
		m.healthStatus[name] = status
	}

	status.LastCheck = time.Now()
	status.Latency = latency

	if err != nil {
		status.Healthy = false
		status.LastError = err
		status.ConsecutiveFails++

		slog.Warn("MCP health check failed",
			"name", name,
			"error", err,
			"consecutive_fails", status.ConsecutiveFails,
		)

		// Trigger reconnection if too many consecutive failures
		if status.ConsecutiveFails >= 3 {
			go m.triggerReconnect(context.Background(), name)
		}
	} else {
		status.Healthy = true
		status.LastError = nil
		status.ConsecutiveFails = 0
	}
}

// triggerReconnect attempts to reconnect an unhealthy MCP connection
func (m *HealthMonitor) triggerReconnect(ctx context.Context, name string) {
	cfg := config.Get()
	mcpConfig, ok := cfg.MCP[name]
	if !ok || mcpConfig.Disabled {
		return
	}

	slog.Info("Attempting to reconnect MCP", "name", name)

	retryCfg := DefaultRetryConfig()
	var lastErr error

	for attempt := 0; attempt <= retryCfg.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := calculateBackoff(retryCfg, attempt-1)
			slog.Debug("Waiting before retry", "name", name, "delay", delay, "attempt", attempt)

			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
			}
		}

		updateState(name, StateStarting, nil, nil, Counts{})

		session, err := createSession(ctx, name, mcpConfig, cfg.Resolver())
		if err != nil {
			lastErr = &ConnectionError{
				MCPName:   name,
				Type:      string(mcpConfig.Type),
				Attempt:   attempt + 1,
				Cause:     err,
				Timestamp: time.Now(),
			}

			connErr, ok := lastErr.(*ConnectionError)
			if !ok || !connErr.IsRetryable() {
				slog.Error("MCP reconnection failed (non-retryable)",
					"name", name,
					"error", err,
				)
				updateState(name, StateError, lastErr, nil, Counts{})
				return
			}

			slog.Warn("MCP reconnection attempt failed",
				"name", name,
				"attempt", attempt+1,
				"error", err,
			)
			continue
		}

		// Get tools and prompts for the new session
		tools, err := getTools(ctx, session)
		if err != nil {
			slog.Error("Error listing tools after reconnection", "name", name, "error", err)
			session.Close()
			continue
		}

		prompts, err := getPrompts(ctx, session)
		if err != nil {
			slog.Error("Error listing prompts after reconnection", "name", name, "error", err)
			session.Close()
			continue
		}

		updateTools(name, tools)
		updatePrompts(name, prompts)
		sessions.Set(name, session)

		updateState(name, StateConnected, nil, session, Counts{
			Tools:   len(tools),
			Prompts: len(prompts),
		})

		slog.Info("MCP reconnected successfully",
			"name", name,
			"tools", len(tools),
			"prompts", len(prompts),
		)

		// Reset health status
		m.mu.Lock()
		if status, exists := m.healthStatus[name]; exists {
			status.Healthy = true
			status.LastError = nil
			status.ConsecutiveFails = 0
		}
		m.mu.Unlock()

		return
	}

	// All retries exhausted
	updateState(name, StateError, lastErr, nil, Counts{})
	slog.Error("MCP reconnection failed after all retries",
		"name", name,
		"max_retries", retryCfg.MaxRetries,
	)
}

// GetHealthStatus returns the health status of an MCP connection
func (m *HealthMonitor) GetHealthStatus(name string) (HealthStatus, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status, ok := m.healthStatus[name]
	if !ok {
		return HealthStatus{}, false
	}
	return *status, true
}

// GetAllHealthStatus returns the health status of all MCP connections
func (m *HealthMonitor) GetAllHealthStatus() map[string]HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]HealthStatus, len(m.healthStatus))
	for name, status := range m.healthStatus {
		result[name] = *status
	}
	return result
}

// Global health monitor instance
var (
	globalHealthMonitor     *HealthMonitor
	globalHealthMonitorOnce sync.Once
)

// GetHealthMonitor returns the global health monitor instance
func GetHealthMonitor() *HealthMonitor {
	globalHealthMonitorOnce.Do(func() {
		globalHealthMonitor = NewHealthMonitor(30 * time.Second)
	})
	return globalHealthMonitor
}

// StartHealthMonitoring starts the global health monitor
func StartHealthMonitoring(ctx context.Context) {
	GetHealthMonitor().Start(ctx)
}

// StopHealthMonitoring stops the global health monitor
func StopHealthMonitoring() {
	if globalHealthMonitor != nil {
		globalHealthMonitor.Stop()
	}
}

// getOrRenewClientWithRetry is a drop-in replacement for getOrRenewClient with retry logic
func getOrRenewClientWithRetry(ctx context.Context, name string, retryCfg RetryConfig) (*mcp.ClientSession, error) {
	var lastErr error

	for attempt := 0; attempt <= retryCfg.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := calculateBackoff(retryCfg, attempt-1)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		session, err := getOrRenewClient(ctx, name)
		if err == nil {
			return session, nil
		}

		lastErr = &ConnectionError{
			MCPName:   name,
			Attempt:   attempt + 1,
			Cause:     err,
			Timestamp: time.Now(),
		}

		connErr := lastErr.(*ConnectionError)
		if !connErr.IsRetryable() {
			return nil, lastErr
		}

		slog.Debug("MCP client operation failed, retrying",
			"name", name,
			"attempt", attempt+1,
			"error", err,
		)
	}

	return nil, lastErr
}

// EnsureConnection ensures an MCP connection is available, attempting to establish one if needed
func EnsureConnection(ctx context.Context, name string) error {
	state, ok := states.Get(name)
	if !ok {
		return fmt.Errorf("MCP '%s' not configured", name)
	}

	if state.State == StateConnected {
		// Verify connection is still valid
		if session, sessionOk := sessions.Get(name); sessionOk {
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			if err := session.Ping(pingCtx, nil); err == nil {
				return nil
			}
		}
	}

	if state.State == StateDisabled {
		return fmt.Errorf("MCP '%s' is disabled", name)
	}

	// Try to reconnect
	cfg := config.Get()
	mcpConfig, ok := cfg.MCP[name]
	if !ok {
		return fmt.Errorf("MCP '%s' configuration not found", name)
	}

	retryCfg := DefaultRetryConfig()
	var lastErr error

	for attempt := 0; attempt <= retryCfg.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := calculateBackoff(retryCfg, attempt-1)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		session, err := createSession(ctx, name, mcpConfig, cfg.Resolver())
		if err != nil {
			lastErr = &ConnectionError{
				MCPName:   name,
				Type:      string(mcpConfig.Type),
				Attempt:   attempt + 1,
				Cause:     err,
				Timestamp: time.Now(),
			}
			continue
		}

		tools, err := getTools(ctx, session)
		if err != nil {
			session.Close()
			continue
		}

		prompts, err := getPrompts(ctx, session)
		if err != nil {
			session.Close()
			continue
		}

		updateTools(name, tools)
		updatePrompts(name, prompts)
		sessions.Set(name, session)
		updateState(name, StateConnected, nil, session, Counts{
			Tools:   len(tools),
			Prompts: len(prompts),
		})

		return nil
	}

	return lastErr
}

// MCPError represents a categorized MCP error
type MCPError struct {
	Code    MCPErrorCode
	Message string
	Details map[string]any
	Cause   error
}

// MCPErrorCode categorizes MCP errors
type MCPErrorCode int

const (
	ErrCodeUnknown MCPErrorCode = iota
	ErrCodeNotConfigured
	ErrCodeDisabled
	ErrCodeConnectionFailed
	ErrCodeTimeout
	ErrCodeToolNotFound
	ErrCodeToolExecutionFailed
	ErrCodeInvalidInput
	ErrCodePermissionDenied
)

func (e *MCPError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code.String(), e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code.String(), e.Message)
}

func (e *MCPError) Unwrap() error {
	return e.Cause
}

func (c MCPErrorCode) String() string {
	switch c {
	case ErrCodeNotConfigured:
		return "NOT_CONFIGURED"
	case ErrCodeDisabled:
		return "DISABLED"
	case ErrCodeConnectionFailed:
		return "CONNECTION_FAILED"
	case ErrCodeTimeout:
		return "TIMEOUT"
	case ErrCodeToolNotFound:
		return "TOOL_NOT_FOUND"
	case ErrCodeToolExecutionFailed:
		return "TOOL_EXECUTION_FAILED"
	case ErrCodeInvalidInput:
		return "INVALID_INPUT"
	case ErrCodePermissionDenied:
		return "PERMISSION_DENIED"
	default:
		return "UNKNOWN"
	}
}

// NewMCPError creates a new categorized MCP error
func NewMCPError(code MCPErrorCode, message string, cause error) *MCPError {
	return &MCPError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// NewMCPErrorWithDetails creates a new categorized MCP error with additional details
func NewMCPErrorWithDetails(code MCPErrorCode, message string, details map[string]any, cause error) *MCPError {
	return &MCPError{
		Code:    code,
		Message: message,
		Details: details,
		Cause:   cause,
	}
}

// WrapError wraps an error with MCP context
func WrapError(err error, mcpName string) error {
	if err == nil {
		return nil
	}

	// Detect error type and categorize
	if errors.Is(err, context.DeadlineExceeded) {
		return NewMCPError(ErrCodeTimeout, fmt.Sprintf("MCP '%s' operation timed out", mcpName), err)
	}

	if errors.Is(err, context.Canceled) {
		return NewMCPError(ErrCodeConnectionFailed, fmt.Sprintf("MCP '%s' operation was canceled", mcpName), err)
	}

	errStr := err.Error()
	if containsIgnoreCase(errStr, "not configured") || containsIgnoreCase(errStr, "not found") {
		return NewMCPError(ErrCodeNotConfigured, fmt.Sprintf("MCP '%s' is not configured", mcpName), err)
	}

	if containsIgnoreCase(errStr, "disabled") {
		return NewMCPError(ErrCodeDisabled, fmt.Sprintf("MCP '%s' is disabled", mcpName), err)
	}

	if containsIgnoreCase(errStr, "connection") || containsIgnoreCase(errStr, "EOF") {
		return NewMCPError(ErrCodeConnectionFailed, fmt.Sprintf("MCP '%s' connection failed", mcpName), err)
	}

	return NewMCPError(ErrCodeUnknown, fmt.Sprintf("MCP '%s' error", mcpName), err)
}
