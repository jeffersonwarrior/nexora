package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nexora/nexora/internal/agent"
	"github.com/nexora/nexora/internal/session"
)

// DebugAgentWrapper wraps the session agent to add debugging for conversation loops
type DebugAgentWrapper struct {
	agent.SessionAgent
	lastAction map[string]time.Time
}

func NewDebugAgentWrapper(wrapped agent.SessionAgent) *DebugAgentWrapper {
	return &DebugAgentWrapper{
		SessionAgent: wrapped,
		lastAction:   make(map[string]time.Time),
	}
}

func (d *DebugAgentWrapper) Run(ctx context.Context, call agent.SessionAgentCall) (*fantasy.AgentResult, error) {
	sessionID := call.SessionID
	now := time.Now()
	
	// Check for rapid successive calls
	if lastTime, exists := d.lastAction[sessionID]; exists {
		if now.Sub(lastTime) < 2*time.Second {
			slog.Warn("RAPID SUCCESSIVE CALLS DETECTED",
				"session_id", sessionID,
				"time_since_last", now.Sub(lastTime),
				"prompt", call.Prompt[:min(100, len(call.Prompt))])
		}
	}
	d.lastAction[sessionID] = now
	
	// Log continuation prompts
	if call.Prompt == "CONTINUE_AFTER_TOOL_EXECUTION" {
		slog.Debug("TOOL EXECUTION CONTINUATION",
			"session_id", sessionID,
			"agent_busy_before", d.IsSessionBusy(sessionID))
	}
	
	// Call the original agent
	result, err := d.SessionAgent.Run(ctx, call)
	
	// Log after execution
	slog.Debug("AGENT RUN COMPLETED",
		"session_id", sessionID,
		"error", err != nil,
		"queued_prompts", d.QueuedPrompts(sessionID),
		"agent_busy_after", d.IsSessionBusy(sessionID))
	
	// Check if this might trigger another continuation
	if err == nil && d.IsSessionBusy(sessionID) {
		slog.Warn("AGENT STILL BUSY AFTER RUN - POTENTIAL LOOP",
			"session_id", sessionID,
			"queued_prompts", d.QueuedPrompts(sessionID))
	}
	
	return result, err
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// InstallDebugWrapper can be called from app initialization to wrap the agent
func InstallDebugWrapper(app *App) {
	if app.AgentCoordinator != nil {
		// This would need to be adapted based on the actual app structure
		slog.Info("Installing debug agent wrapper for conversation loop detection")
		// app.AgentCoordinator = NewDebugAgentWrapper(app.AgentCoordinator)
	}
}