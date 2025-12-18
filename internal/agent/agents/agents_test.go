package agents

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAgentParser(t *testing.T) {
	t.Parallel()

	registry := NewAgentRegistry()

	// Test loading from directory
	err := registry.LoadAgentsFromDirectory(".")
	require.NoError(t, err)

	// Check that agents were loaded
	agents := registry.ListAgents()
	require.Greater(t, len(agents), 0)

	// Test getting specific agents
	codeArchitect, exists := registry.GetAgent("code-architect")
	require.True(t, exists)
	require.Equal(t, "code-architect", codeArchitect.Name)
	require.Equal(t, "blue", codeArchitect.Color)
	require.Contains(t, codeArchitect.Triggers, "architecture")

	bugSpecialist, exists := registry.GetAgent("bug-specialist")
	require.True(t, exists)
	require.Equal(t, "bug-specialist", bugSpecialist.Name)
	require.Equal(t, "red", bugSpecialist.Color)
	require.Contains(t, bugSpecialist.Triggers, "bug")
}

func TestAgentMatching(t *testing.T) {
	t.Parallel()

	registry := NewAgentRegistry()
	err := registry.LoadAgentsFromDirectory(".")
	require.NoError(t, err)

	// Test trigger matching
	testCases := []struct {
		input       string
		expected    string
		description string
	}{
		{
			input:       "I need help with system architecture",
			expected:    "code-architect",
			description: "Should match architecture trigger",
		},
		{
			input:       "There's a bug in my code",
			expected:    "bug-specialist",
			description: "Should match bug trigger",
		},
		{
			input:       "This is running very slow",
			expected:    "performance-optimizer",
			description: "Should match performance trigger",
		},
		{
			input:       "hello world",
			expected:    "", // Any agent is fine for no clear triggers
			description: "Input with no clear triggers",
		},
	}

	for _, tc := range testCases {
		agent := registry.MatchAgent(tc.input)
		require.NotNil(t, agent, tc.description+": no agent matched")
		if tc.expected != "" {
			require.Equal(t, tc.expected, agent.Name, tc.description)
		}
	}
}

func TestAgentExecution(t *testing.T) {
	t.Parallel()

	registry := NewAgentRegistry()
	err := registry.LoadAgentsFromDirectory(".")
	require.NoError(t, err)

	executor := NewAgentExecutor(registry)

	// Test executing with architecture query
	input := "help me design a microservices architecture"
	tools := []string{"View", "Glob", "Grep", "LS"}

	agent, prompt, err := executor.ExecuteForInput(context.Background(), input, tools)
	require.NoError(t, err)
	require.NotNil(t, agent)
	require.Equal(t, "code-architect", agent.Name)
	require.Contains(t, prompt, agent.SystemPrompt)
	require.Contains(t, prompt, input)

	// Test that tool permissions are enforced
	restrictedTools := []string{"Bash"} // Not in code-architect's tools
	_, _, err = executor.ExecuteForInput(context.Background(), input, restrictedTools)
	require.Error(t, err)
	require.Contains(t, strings.ToLower(err.Error()), "not allowed to use tool")
}

func TestAgentValidation(t *testing.T) {
	t.Parallel()

	// Test valid agent
	validAgent := &AgentConfig{
		Name:         "test-agent",
		Description:  "Test agent",
		Model:        "inherit",
		Color:        "blue",
		Tools:        []string{"View"},
		Triggers:     []string{"test"},
		SystemPrompt: "You are a test agent",
	}

	require.NoError(t, ValidateAgent(validAgent))

	// Test invalid agents
	invalidAgents := []struct {
		name   string
		agent  *AgentConfig
		reason string
	}{
		{
			name: "no name",
			agent: &AgentConfig{
				Description:  "Test agent",
				Triggers:     []string{"test"},
				SystemPrompt: "You are a test agent",
			},
			reason: "should require name",
		},
		{
			name: "no triggers",
			agent: &AgentConfig{
				Name:         "test-agent",
				Description:  "Test agent",
				SystemPrompt: "You are a test agent",
			},
			reason: "should require triggers",
		},
		{
			name: "no system prompt",
			agent: &AgentConfig{
				Name:        "test-agent",
				Description: "Test agent",
				Triggers:    []string{"test"},
			},
			reason: "should require system prompt",
		},
	}

	for _, tc := range invalidAgents {
		err := ValidateAgent(tc.agent)
		require.Error(t, err, tc.reason)
	}
}

func TestExplainTriggering(t *testing.T) {
	t.Parallel()

	registry := NewAgentRegistry()
	err := registry.LoadAgentsFromDirectory(".")
	require.NoError(t, err)

	executor := NewAgentExecutor(registry)

	// Test clear trigger match
	explanation := executor.ExplainTriggering("I have a performance bottleneck")
	require.Contains(t, strings.ToLower(explanation), "performance-optimizer")
	require.Contains(t, strings.ToLower(explanation), "performance")

	// Test multiple potential matches
	explanation = executor.ExplainTriggering("help me fix this bug")
	require.Contains(t, strings.ToLower(explanation), "bug-specialist")
	require.Contains(t, strings.ToLower(explanation), "bug")
}
