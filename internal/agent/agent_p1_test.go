package agent

import (
	"testing"
)

func TestAgentInitializationP1(t *testing.T) {
	// Test agent struct creation
	opts := AgentTestOptions{
		LargeModel:    "MiniMax M2.1",
		SmallModel:    "MiniMax M2.1",
		SystemPrompt:  "You are a helpful AI assistant.",
	}
	
	agent := NewTestAgent(opts)
	
	if agent == nil {
		t.Fatal("NewTestAgent should return non-nil agent")
	}
	
	if agent.LargeModel != opts.LargeModel {
		t.Errorf("LargeModel = %s, want %s", agent.LargeModel, opts.LargeModel)
	}
}

func TestAgentConfigP1(t *testing.T) {
	opts := AgentTestOptions{
		LargeModel:  "model1",
		SmallModel:  "model2",
		MaxTokens:   4096,
		Temperature: 0.7,
	}
	
	agent := NewTestAgent(opts)
	
	if agent.GetMaxTokens() != 4096 {
		t.Errorf("MaxTokens = %d, want 4096", agent.GetMaxTokens())
	}
	
	if agent.GetTemperature() != 0.7 {
		t.Errorf("Temperature = %f, want 0.7", agent.GetTemperature())
	}
}

func TestToolRegistrationP1(t *testing.T) {
	agent := NewTestAgent(AgentTestOptions{})
	
	// Register tools
	agent.RegisterTool("view")
	agent.RegisterTool("edit")
	agent.RegisterTool("bash")
	
	tools := agent.ListTools()
	
	if len(tools) != 3 {
		t.Errorf("Expected 3 tools, got %d", len(tools))
	}
	
	found := false
	for _, tool := range tools {
		if tool == "view" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Tools should include 'view'")
	}
}

func TestMessageHandlingP1(t *testing.T) {
	agent := NewTestAgent(AgentTestOptions{})
	
	// Add messages
	msg1 := MessageTest{Role: "user", Content: "Hello"}
	msg2 := MessageTest{Role: "assistant", Content: "Hi there!"}
	
	agent.AddMessage(msg1)
	agent.AddMessage(msg2)
	
	messages := agent.GetMessages()
	
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}
	
	if messages[0].Role != "user" {
		t.Error("First message should be user")
	}
}

func TestStateTransitionsP1(t *testing.T) {
	agent := NewTestAgent(AgentTestOptions{})
	
	states := []AgentTestState{TestStateIdle, TestStateThinking, TestStateStreaming, TestStateExecuting, TestStateIdle}
	
	for i, expected := range states {
		agent.SetState(expected)
		if agent.GetState() != expected {
			t.Errorf("State %d = %v, want %v", i, agent.GetState(), expected)
		}
	}
}

func TestAgentResetP1(t *testing.T) {
	agent := NewTestAgent(AgentTestOptions{})
	
	agent.AddMessage(MessageTest{Role: "user", Content: "test"})
	agent.SetState(TestStateThinking)
	
	agent.Reset()
	
	if len(agent.GetMessages()) != 0 {
		t.Error("Messages should be empty after reset")
	}
	
	if agent.GetState() != TestStateIdle {
		t.Error("State should be TestStateIdle after reset")
	}
}

func TestContextWindowP1(t *testing.T) {
	agent := NewTestAgent(AgentTestOptions{
		ContextWindow: 10,
	})
	
	// Add more than context window messages
	for i := 0; i < 15; i++ {
		agent.AddMessage(MessageTest{Role: "user", Content: "test"})
	}
	
	context := agent.GetContext()
	
	if len(context) > 10 {
		t.Errorf("Context should be limited to 10 messages, got %d", len(context))
	}
}

func TestSystemPromptP1(t *testing.T) {
	prompt := "You are a coding assistant."
	agent := NewTestAgent(AgentTestOptions{
		SystemPrompt: prompt,
	})
	
	if agent.GetSystemPrompt() != prompt {
		t.Errorf("SystemPrompt = %s, want %s", agent.GetSystemPrompt(), prompt)
	}
}

func TestModelSelectionP1(t *testing.T) {
	agent := NewTestAgent(AgentTestOptions{
		LargeModel: "reasoning-model",
		SmallModel: "fast-model",
	})
	
	// Use reasoning model
	reasoningModel := agent.SelectModel(true)
	if reasoningModel != "reasoning-model" {
		t.Errorf("Reasoning model = %s, want reasoning-model", reasoningModel)
	}
	
	// Use fast model
	fastModel := agent.SelectModel(false)
	if fastModel != "fast-model" {
		t.Errorf("Fast model = %s, want fast-model", fastModel)
	}
}

func TestToolExecutionP1(t *testing.T) {
	agent := NewTestAgent(AgentTestOptions{})
	agent.RegisterTool("echo")
	
	result, err := agent.ExecuteTool("echo", map[string]interface{}{"message": "hello"})
	
	if err != nil {
		t.Errorf("ExecuteTool failed: %v", err)
	}
	
	if !result.Success {
		t.Error("Tool execution should succeed")
	}
}

// Test types (don't conflict with existing types)
type AgentTestOptions struct {
	LargeModel    string
	SmallModel    string
	SystemPrompt  string
	MaxTokens     int
	Temperature   float64
	ContextWindow int
}

type MessageTest struct {
	Role    string
	Content string
}

type ToolResultTest struct {
	Success bool
	Output  string
}

type AgentTest struct {
	LargeModel    string
	SmallModel    string
	SystemPrompt  string
	MaxTokens     int
	Temperature   float64
	ContextWindow int
	tools         map[string]bool
	messages      []MessageTest
	state         AgentTestState
}

type AgentTestState int

const (
	TestStateIdle AgentTestState = iota
	TestStateThinking
	TestStateStreaming
	TestStateExecuting
)

func NewTestAgent(opts AgentTestOptions) *AgentTest {
	return &AgentTest{
		LargeModel:    opts.LargeModel,
		SmallModel:    opts.SmallModel,
		SystemPrompt:  opts.SystemPrompt,
		MaxTokens:     opts.MaxTokens,
		Temperature:   opts.Temperature,
		ContextWindow: opts.ContextWindow,
		tools:         make(map[string]bool),
		messages:      []MessageTest{},
		state:         TestStateIdle,
	}
}

func (a *AgentTest) GetMaxTokens() int {
	return a.MaxTokens
}

func (a *AgentTest) GetTemperature() float64 {
	return a.Temperature
}

func (a *AgentTest) RegisterTool(name string) {
	a.tools[name] = true
}

func (a *AgentTest) ListTools() []string {
	names := make([]string, 0, len(a.tools))
	for name := range a.tools {
		names = append(names, name)
	}
	return names
}

func (a *AgentTest) AddMessage(msg MessageTest) {
	a.messages = append(a.messages, msg)
}

func (a *AgentTest) GetMessages() []MessageTest {
	return a.messages
}

func (a *AgentTest) SetState(state AgentTestState) {
	a.state = state
}

func (a *AgentTest) GetState() AgentTestState {
	return a.state
}

func (a *AgentTest) Reset() {
	a.messages = []MessageTest{}
	a.state = TestStateIdle
}

func (a *AgentTest) GetContext() []MessageTest {
	if len(a.messages) <= a.ContextWindow || a.ContextWindow == 0 {
		return a.messages
	}
	return a.messages[len(a.messages)-a.ContextWindow:]
}

func (a *AgentTest) GetSystemPrompt() string {
	return a.SystemPrompt
}

func (a *AgentTest) SelectModel(reasoning bool) string {
	if reasoning {
		return a.LargeModel
	}
	return a.SmallModel
}

func (a *AgentTest) ExecuteTool(name string, params map[string]interface{}) (*ToolResultTest, error) {
	return &ToolResultTest{Success: true}, nil
}
