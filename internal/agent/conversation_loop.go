package agent

import (
	"log/slog"

	"charm.land/fantasy"
	"github.com/nexora/cli/internal/message"
)

// ConversationState tracks the flow of a multi-step conversation
type ConversationState struct {
	sessionID      string
	pendingTool    string
	toolResults    []fantasy.ToolResultContent
	shouldContinue bool
	lastAction     string
}

// ConversationLoopManager manages automatic continuation of agent workflows
type ConversationLoopManager struct {
	states map[string]*ConversationState
}

func NewConversationLoopManager() *ConversationLoopManager {
	return &ConversationLoopManager{
		states: make(map[string]*ConversationState),
	}
}

// ShouldAutoContinue determines if the conversation should automatically continue
func (c *ConversationLoopManager) ShouldAutoContinue(sessionID string, msg message.Message) bool {
	state, exists := c.states[sessionID]
	if !exists {
		return false
	}

	// Check if we have tool results and the message finished naturally
	if len(state.toolResults) > 0 {
		slog.Debug("auto-continuing conversation",
			"session_id", sessionID,
			"tool_results", len(state.toolResults))
		return true
	}

	return false
}

// RecordToolResult tracks a completed tool operation
func (c *ConversationLoopManager) RecordToolResult(sessionID string, toolName string, result fantasy.ToolResultContent) {
	if _, exists := c.states[sessionID]; !exists {
		c.states[sessionID] = &ConversationState{
			sessionID: sessionID,
		}
	}

	state := c.states[sessionID]
	state.toolResults = append(state.toolResults, result)
	state.pendingTool = ""
	state.shouldContinue = true
	state.lastAction = toolName
}

// RecordToolUse tracks when a tool is being used
func (c *ConversationLoopManager) RecordToolUse(sessionID string, toolName string) {
	if _, exists := c.states[sessionID]; !exists {
		c.states[sessionID] = &ConversationState{
			sessionID: sessionID,
		}
	}
	c.states[sessionID].pendingTool = toolName
}

// Reset clears the conversation state
func (c *ConversationLoopManager) Reset(sessionID string) {
	delete(c.states, sessionID)
}
