package agents

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
)

// AgentExecutor handles trigger-based agent execution
type AgentExecutor struct {
	registry *AgentRegistry
}

// NewAgentExecutor creates a new agent executor
func NewAgentExecutor(registry *AgentRegistry) *AgentExecutor {
	return &AgentExecutor{
		registry: registry,
	}
}

// ExecuteForInput finds and returns the best agent for given input
func (e *AgentExecutor) ExecuteForInput(ctx context.Context, input string, tools []string) (*AgentConfig, string, error) {
	// Find matching agent
	agent := e.registry.MatchAgent(input)
	if agent == nil {
		return nil, "", fmt.Errorf("no agent found for input: %s", input)
	}

	// Validate agent has access to requested tools
	if err := e.validateToolAccess(agent, tools); err != nil {
		return nil, "", fmt.Errorf("tool access validation failed: %w", err)
	}

	// Return agent with formatted prompt
	prompt := e.formatAgentPrompt(agent, input)
	return agent, prompt, nil
}

// ExecuteAgent executes a specific agent
func (e *AgentExecutor) ExecuteAgent(ctx context.Context, agent *AgentConfig, input string, tools []string) (*AgentConfig, string, error) {
	// Validate agent has access to requested tools
	if err := e.validateToolAccess(agent, tools); err != nil {
		return nil, "", fmt.Errorf("tool access validation failed: %w", err)
	}

	// Create prompt with agent's system prompt
	prompt := e.formatAgentPrompt(agent, input)
	return agent, prompt, nil
}

// formatAgentPrompt creates the full prompt for an agent
func (e *AgentExecutor) formatAgentPrompt(agent *AgentConfig, input string) string {
	return fmt.Sprintf("%s\n\nUser: %s", agent.SystemPrompt, input)
}

// validateToolAccess checks if agent is allowed to use requested tools
func (e *AgentExecutor) validateToolAccess(agent *AgentConfig, requestedTools []string) error {
	// If agent has no tool restrictions, allow all
	if len(agent.Tools) == 0 {
		return nil
	}

	// Check each requested tool is in agent's allowed list
	for _, tool := range requestedTools {
		if !slices.Contains(agent.Tools, tool) {
			return fmt.Errorf("agent '%s' not allowed to use tool '%s'", agent.Name, tool)
		}
	}

	return nil
}

// colorResponse applies visual styling to agent responses
func (e *AgentExecutor) colorResponse(response string, color string) string {
	style := lipgloss.NewStyle()

	switch color {
	case "green":
		style = style.Foreground(lipgloss.Color("10"))
	case "blue":
		style = style.Foreground(lipgloss.Color("12"))
	case "red":
		style = style.Foreground(lipgloss.Color("9"))
	case "yellow":
		style = style.Foreground(lipgloss.Color("11"))
	case "cyan":
		style = style.Foreground(lipgloss.Color("14"))
	case "magenta":
		style = style.Foreground(lipgloss.Color("13"))
	default:
		// No coloring for unknown colors
		return response
	}

	return style.Render(response)
}

// GetAgent returns an agent by name
func (e *AgentExecutor) GetAgent(name string) (*AgentConfig, bool) {
	return e.registry.GetAgent(name)
}

// ListAgents returns all available agent names
func (e *AgentExecutor) ListAgents() []string {
	return e.registry.ListAgents()
}

// CreateAgent creates a new agent from the given configuration
func (e *AgentExecutor) CreateAgent(config *AgentConfig) error {
	// Validate the agent configuration
	if err := ValidateAgent(config); err != nil {
		return fmt.Errorf("invalid agent configuration: %w", err)
	}

	// Add to registry
	e.registry.agents[config.Name] = config
	return nil
}

// RemoveAgent removes an agent from the registry
func (e *AgentExecutor) RemoveAgent(name string) error {
	if _, exists := e.registry.GetAgent(name); !exists {
		return fmt.Errorf("agent '%s' not found", name)
	}

	delete(e.registry.agents, name)
	return nil
}

// ExplainTriggering explains why an agent was selected
func (e *AgentExecutor) ExplainTriggering(input string) string {
	agent := e.registry.MatchAgent(input)
	if agent == nil {
		return "No agent matched"
	}

	input = strings.ToLower(input)
	matchedTriggers := make([]string, 0)

	for _, trigger := range agent.Triggers {
		if strings.Contains(input, strings.ToLower(trigger)) {
			matchedTriggers = append(matchedTriggers, trigger)
		}
	}

	if len(matchedTriggers) == 0 {
		return fmt.Sprintf("Agent '%s' selected but no clear triggers matched", agent.Name)
	}

	return fmt.Sprintf("Agent '%s' selected due to triggers: %s", agent.Name, strings.Join(matchedTriggers, ", "))
}
