package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// AgentConfig represents a Claude Code style agent definition
type AgentConfig struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	Model        string   `yaml:"model"`    // "inherit" or specific model
	Color        string   `yaml:"color"`    // Visual indicator
	Tools        []string `yaml:"tools"`    // Explicit tool access
	Triggers     []string `yaml:"triggers"` // Natural language triggers
	SystemPrompt string   // Markdown content after frontmatter
}

// AgentRegistry holds all loaded agents
type AgentRegistry struct {
	agents map[string]*AgentConfig
}

// NewAgentRegistry creates a new agent registry
func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{
		agents: make(map[string]*AgentConfig),
	}
}

// LoadAgentFromFile loads an agent from a markdown file with YAML frontmatter
func (r *AgentRegistry) LoadAgentFromFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read agent file: %w", err)
	}

	// Split frontmatter from content
	parts := strings.SplitN(string(content), "---", 3)
	if len(parts) != 3 {
		return fmt.Errorf("invalid agent file format: missing YAML frontmatter")
	}

	// Parse YAML frontmatter
	var config AgentConfig
	if err := yaml.Unmarshal([]byte(parts[1]), &config); err != nil {
		return fmt.Errorf("failed to parse agent frontmatter: %w", err)
	}

	// Set system prompt from markdown content
	config.SystemPrompt = strings.TrimSpace(parts[2])

	// Store in registry
	r.agents[config.Name] = &config
	return nil
}

// LoadAgentsFromDirectory loads all agents from a directory
func (r *AgentRegistry) LoadAgentsFromDirectory(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read agents directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())
		if err := r.LoadAgentFromFile(filePath); err != nil {
			return fmt.Errorf("failed to load agent %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// GetAgent retrieves an agent by name
func (r *AgentRegistry) GetAgent(name string) (*AgentConfig, bool) {
	agent, exists := r.agents[name]
	return agent, exists
}

// ListAgents returns all agent names
func (r *AgentRegistry) ListAgents() []string {
	names := make([]string, 0, len(r.agents))
	for name := range r.agents {
		names = append(names, name)
	}
	return names
}

// MatchAgent finds the best agent for a given input using triggers
func (r *AgentRegistry) MatchAgent(input string) *AgentConfig {
	input = strings.ToLower(input)

	var bestMatch *AgentConfig
	var bestScore int

	for _, agent := range r.agents {
		score := 0
		for _, trigger := range agent.Triggers {
			if strings.Contains(input, strings.ToLower(trigger)) {
				score += len(trigger) // Prefer longer, more specific triggers
			}
		}
		if score > bestScore {
			bestScore = score
			bestMatch = agent
		}
	}

	// If no agent matched, return the first agent as default
	if bestMatch == nil && len(r.agents) > 0 {
		for _, agent := range r.agents {
			return agent // Return first agent found
		}
	}

	return bestMatch
}

// ValidateAgent checks if an agent config is valid
func ValidateAgent(config *AgentConfig) error {
	if config.Name == "" {
		return fmt.Errorf("agent name is required")
	}
	if config.Description == "" {
		return fmt.Errorf("agent description is required")
	}
	if config.SystemPrompt == "" {
		return fmt.Errorf("agent system prompt is required")
	}
	if len(config.Triggers) == 0 {
		return fmt.Errorf("agent must have at least one trigger")
	}
	return nil
}
