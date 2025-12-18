//go:build ignore
// +build ignore

// TODO: This file is incomplete and needs to be finished
// See internal/agent/coordinator.go for reference implementation
// Required fixes:
// 1. Import fantasy package
// 2. Implement buildAgent method
// 3. Fix message.NewAssistantMessage usage
// 4. Fix AgentConfig references

package agent

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/nexora/nexora/internal/config"
	"github.com/nexora/nexora/internal/message"
	"github.com/nexora/nexora/internal/session"
)

type MultiSessionCoordinator struct {
	cfg          *config.Config
	sessions     session.Service
	messages     message.Service
	windows      map[string]*SessionWindow
	activeID     string
	windowMu     sync.RWMutex
	conversation *ConversationLoopManager
}

func NewMultiSessionCoordinator(
	cfg *config.Config,
	sessions session.Service,
	messages message.Service,
) *MultiSessionCoordinator {
	return &MultiSessionCoordinator{
		cfg:          cfg,
		sessions:     sessions,
		messages:     messages,
		windows:      make(map[string]*SessionWindow),
		activeID:     "",
		conversation: NewConversationLoopManager(),
	}
}

func (m *MultiSessionCoordinator) CreateWindow(ctx context.Context, title, prompt string) (string, error) {
	id := uuid.New().String()

	// Build agent for this window (Claude-style isolation)
	agentCfg, ok := m.cfg.Agents[config.AgentCoder]
	if !ok {
		return "", errors.New("coder agent not configured")
	}

	agent, err := m.buildWindowAgent(ctx, agentCfg)
	if err != nil {
		return "", err
	}

	window := NewSessionWindow(id, title, agent)

	m.windowMu.Lock()
	m.windows[id] = window
	if m.activeID == "" {
		m.activeID = id
		window.IsActive = true
	}
	m.windowMu.Unlock()

	return id, nil
}

func (m *MultiSessionCoordinator) SwitchWindow(sessionID string) error {
	m.windowMu.Lock()
	defer m.windowMu.Unlock()

	oldWindow, oldExists := m.windows[m.activeID]
	newWindow, newExists := m.windows[sessionID]

	if !newExists {
		return fmt.Errorf("window %s not found", sessionID)
	}

	// Deactivate old
	if oldExists {
		oldWindow.IsActive = false
	}

	// Activate new
	newWindow.IsActive = true
	m.activeID = sessionID

	return nil
}

func (m *MultiSessionCoordinator) ListWindows() []SessionWindow {
	m.windowMu.RLock()
	defer m.windowMu.RUnlock()

	var windows []SessionWindow
	for _, w := range m.windows {
		// Copy without agent to avoid serialization issues
		copy := *w
		copy.Agent = nil
		windows = append(windows, copy)
	}
	return windows
}

func (m *MultiSessionCoordinator) ForkWindow(parentID, prompt string) (string, error) {
	parent, exists := m.windows[parentID]
	if !exists {
		return "", fmt.Errorf("parent window %s not found", parentID)
	}

	// Create fork with parent history as context
	newID, err := m.CreateWindow(context.Background(), "Fork: "+parent.Title, prompt)
	if err != nil {
		return "", err
	}

	newWindow := m.windows[newID]
	newWindow.ForkParent = parentID
	// Copy recent history for context
	newWindow.History = append([]message.Message(nil), parent.History[len(parent.History)-10:]...)

	return newID, nil
}

func (m *MultiSessionCoordinator) Run(ctx context.Context, windowID, prompt string, attachments ...message.Attachment) (*fantasy.AgentResult, error) {
	m.windowMu.RLock()
	window, exists := m.windows[windowID]
	m.windowMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("window %s not found", windowID)
	}

	// Record tool use for auto-continue
	m.conversation.RecordToolUse(windowID, "user_prompt")

	result, err := window.Agent.Run(ctx, SessionAgentCall{
		SessionID:   windowID,
		Prompt:      prompt,
		Attachments: attachments,
	})

	if err == nil {
		// Add result to window history
		window.AddMessage(message.NewAssistantMessage(result.Content))
		m.conversation.RecordToolResult(windowID, "agent_response", result.Content)
	}

	return result, err
}

func (m *MultiSessionCoordinator) ActiveWindow() string {
	m.windowMu.RLock()
	defer m.windowMu.RUnlock()
	return m.activeID
}

func (m *MultiSessionCoordinator) buildWindowAgent(ctx context.Context, agentCfg config.AgentConfig) (SessionAgent, error) {
	// Delegate to existing coordinator logic for agent creation
	prompt, err := coderPrompt(prompt.WithWorkingDir(m.cfg.WorkingDir()))
	if err != nil {
		return nil, err
	}

	// Use existing agent building pattern from coordinator.go
	agent, err := m.buildAgent(ctx, prompt, agentCfg)
	if err != nil {
		return nil, err
	}

	return agent, nil
}

func (m *MultiSessionCoordinator) buildAgent(ctx context.Context, prompt string, agentCfg config.AgentConfig) (SessionAgent, error) {
	// Copy from existing coordinator.buildAgent logic
	// This needs to be implemented based on coordinator.go patterns
	return nil, fmt.Errorf("buildAgent implementation required - copy from coordinator")
}
