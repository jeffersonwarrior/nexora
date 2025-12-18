package state

import (
	"context"
	"time"
)

// AgentExecutionContext tracks the execution context of an agent session
type AgentExecutionContext struct {
	// Core state tracking
	State           AgentState
	SessionID       string
	CancelFunc      context.CancelFunc
	StartTime       time.Time
	
	// Operation counters
	ToolCallCount   int
	ErrorCount      int
	RetryCount      int
	LastError       error
	
	// Resource tracking
	CPUUsage        float64
	MemoryUsage     uint64
	DiskFree        uint64
	
	// Progress tracking
	LastProgress    *ProgressStats
	StuckCount      int
	CurrentPhase    string
	
	// Recovery context
	InRecovery      bool
	RecoveryAttempt int
}

// ResetCounters resets the operational counters
func (ctx *AgentExecutionContext) ResetCounters() {
	ctx.ToolCallCount = 0
	ctx.ErrorCount = 0
	ctx.RetryCount = 0
	ctx.StuckCount = 0
	ctx.InRecovery = false
	ctx.RecoveryAttempt = 0
}

// MarkError increments error count and records the last error
func (ctx *AgentExecutionContext) MarkError(err error) {
	ctx.ErrorCount++
	ctx.LastError = err
}

// GetDuration returns the duration since start time
func (ctx *AgentExecutionContext) GetDuration() time.Duration {
	return time.Since(ctx.StartTime)
}

// Clone creates a shallow copy of the execution context
func (ctx *AgentExecutionContext) Clone() *AgentExecutionContext {
	clone := *ctx
	return &clone
}

// NewAgentExecutionContext creates a new execution context
func NewAgentExecutionContext(sessionID string) *AgentExecutionContext {
	return &AgentExecutionContext{
		State:        StateIdle,
		SessionID:    sessionID,
		StartTime:    time.Now(),
		CurrentPhase: "initialization",
	}
}