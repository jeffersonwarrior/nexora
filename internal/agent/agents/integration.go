package agents

import (
	"context"

	"github.com/nexora/nexora/internal/config"
)

// AgentIntegration integrates the new trigger-based agent system with the existing nexora framework
type AgentIntegration struct {
	executor *AgentExecutor
}

// NewAgentIntegration creates a new integration instance
func NewAgentIntegration() *AgentIntegration {
	registry := NewAgentRegistry()

	// Load agents from the agents directory
	_ = registry.LoadAgentsFromDirectory(".")

	executor := NewAgentExecutor(registry)

	return &AgentIntegration{
		executor: executor,
	}
}

// SelectAgentForInput chooses the best agent for a given input using the new trigger system
// This can be called from the existing agent coordinator to decide which agent to use
func (ai *AgentIntegration) SelectAgentForInput(ctx context.Context, input string, tools []string) (*config.Agent, error) {
	// Use the new trigger system to find the best agent
	agent, _, err := ai.executor.ExecuteForInput(ctx, input, tools)
	if err != nil {
		return nil, err
	}

	// Convert to the existing Agent format
	// This bridges the new agents with the old system
	nexoraAgent := &config.Agent{
		ID:           agent.Name,
		Name:         agent.Name,
		Description:  agent.Description,
		Model:        config.SelectedModelTypeLarge, // Default to large model
		AllowedTools: agent.Tools,
	}

	return nexoraAgent, nil
}

// ListAvailableAgents returns all available agents in a format compatible with the existing system
func (ai *AgentIntegration) ListAvailableAgents() []*config.Agent {
	agents := ai.executor.ListAgents()
	var nexoraAgents []*config.Agent

	for _, agentName := range agents {
		agent, exists := ai.executor.GetAgent(agentName)
		if !exists {
			continue
		}

		// Convert to nexora Agent format
		nexoraAgent := &config.Agent{
			ID:           agent.Name,
			Name:         agent.Name,
			Description:  agent.Description,
			Model:        config.SelectedModelTypeLarge,
			AllowedTools: agent.Tools,
		}

		nexoraAgents = append(nexoraAgents, nexoraAgent)
	}

	return nexoraAgents
}

// ExplainCurrentSelection explains why the current agent was selected
func (ai *AgentIntegration) ExplainCurrentSelection(input string) string {
	return ai.executor.ExplainTriggering(input)
}

// GetAgentColor returns the visual color for an agent
func (ai *AgentIntegration) GetAgentColor(agentName string) string {
	agent, exists := ai.executor.GetAgent(agentName)
	if !exists {
		return ""
	}
	return agent.Color
}

// ShouldUseLegacyAgent determines if we should fall back to the legacy agent system
// This provides backward compatibility while we migrate to the new system
func (ai *AgentIntegration) ShouldUseLegacyAgent(input string) bool {
	// If no new agents match, use legacy
	agent := ai.executor.registry.MatchAgent(input)
	return agent == nil
}

// AddNewAgent allows dynamic creation of agents at runtime
func (ai *AgentIntegration) AddNewAgent(config *AgentConfig) error {
	return ai.executor.CreateAgent(config)
}

// RemoveAgent allows removal of agents at runtime
func (ai *AgentIntegration) RemoveAgent(agentName string) error {
	return ai.executor.RemoveAgent(agentName)
}

// GetAgentTriggers returns the triggers for a specific agent
func (ai *AgentIntegration) GetAgentTriggers(agentName string) []string {
	agent, exists := ai.executor.GetAgent(agentName)
	if !exists {
		return nil
	}
	return agent.Triggers
}

// GetAgentSystemPrompt returns the system prompt for a specific agent
func (ai *AgentIntegration) GetAgentSystemPrompt(agentName string) string {
	agent, exists := ai.executor.GetAgent(agentName)
	if !exists {
		return ""
	}
	return agent.SystemPrompt
}
