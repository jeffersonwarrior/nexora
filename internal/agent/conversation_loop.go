package agent

import (
	"charm.land/fantasy"
	"github.com/nexora/nexora/internal/message"
)

// ConversationFlowState tracks the flow of a multi-step conversation
type ConversationFlowState struct {
	sessionID      string
	pendingTool    string
	toolResults    []fantasy.ToolResultContent
	shouldContinue bool
	lastAction     string
}

// ConversationLoopManager manages automatic continuation of agent workflows
type ConversationLoopManager struct {
	states map[string]*ConversationFlowState
}

func NewConversationLoopManager() *ConversationLoopManager {
	return &ConversationLoopManager{
		states: make(map[string]*ConversationFlowState),
	}
}

// ShouldAutoContinue determines if the conversation should automatically continue
func (c *ConversationLoopManager) ShouldAutoContinue(sessionID string, msg message.Message) bool {
	// Don't use auto-continue based on tool results alone
	// Let the conversation manager handle state-based continuation
	return false
}

func (c *ConversationLoopManager) RecordToolResult(sessionID string, toolName string, result fantasy.ToolResultContent) {
	if _, exists := c.states[sessionID]; !exists {
		c.states[sessionID] = &ConversationFlowState{
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
		c.states[sessionID] = &ConversationFlowState{
			sessionID: sessionID,
		}
	}
	c.states[sessionID].pendingTool = toolName
}

// Reset clears the conversation state
func (c *ConversationLoopManager) Reset(sessionID string) {
	delete(c.states, sessionID)
}