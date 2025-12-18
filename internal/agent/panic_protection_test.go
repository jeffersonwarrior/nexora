package agent

import (
	"context"
	"os"
	"path/filepath"
	"runtime/debug"
	"testing"
	"time"

	"charm.land/fantasy"
	"github.com/nexora/cli/internal/agent/tools"
	"github.com/nexora/cli/internal/config"
	"github.com/nexora/cli/internal/csync"
	"github.com/nexora/cli/internal/db"
	"github.com/nexora/cli/internal/history"
	"github.com/nexora/cli/internal/lsp"
	"github.com/nexora/cli/internal/message"
	"github.com/nexora/cli/internal/permission"
	"github.com/nexora/cli/internal/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPanicProtection_SafeCreateTool tests the safeCreateTool function
// catches panics during tool creation.
func TestPanicProtection_SafeCreateTool(t *testing.T) {
	// Create test environment
	env := setupPanicTestEnv(t)

	// Create coordinator with minimal config
	coordinator := setupPanicTestCoordinator(t, env)

	// Test 1: Normal tool creation should work
	t.Run("NormalToolCreation", func(t *testing.T) {
		tool := coordinator.safeCreateTool(func() fantasy.AgentTool {
			return tools.NewViewTool(nil, env.permissions, env.workingDir)
		})

		assert.NotNil(t, tool, "safeCreateTool should return a tool")

		// Verify the tool works
		info := tool.Info()
		assert.NotEmpty(t, info.Name, "tool should have a name")
	})

	// Test 2: Panicking tool creation should be caught
	t.Run("PanickingToolCreation", func(t *testing.T) {
		panicked := false
		tool := coordinator.safeCreateTool(func() fantasy.AgentTool {
			panicked = true
			panic("intentional panic during tool creation")
		})

		assert.True(t, panicked, "creation function should have been called")
		assert.Nil(t, tool, "safeCreateTool should return nil for panicking creation")
	})

	// Test 3: Tool creation returning nil should be handled
	t.Run("NilToolCreation", func(t *testing.T) {
		tool := coordinator.safeCreateTool(func() fantasy.AgentTool {
			return nil
		})

		assert.Nil(t, tool, "safeCreateTool should return nil for nil creation result")
	})
}

// TestPanicProtection_WrapToolWithTimeout tests the wrapToolWithTimeout function
// protects against panics when calling tool.Info().
func TestPanicProtection_WrapToolWithTimeout(t *testing.T) {
	env := setupPanicTestEnv(t)
	coordinator := setupPanicTestCoordinator(t, env)

	// Test 1: Wrap normal tool should work
	t.Run("NormalTool", func(t *testing.T) {
		// Create a simple working tool using the same pattern as in buildTools
		tool := coordinator.safeCreateTool(func() fantasy.AgentTool {
			return tools.NewViewTool(nil, env.permissions, env.workingDir)
		})

		assert.NotNil(t, tool, "safeCreateTool should return a tool")

		// Wrap with timeout - should not panic
		wrappedTool := coordinator.wrapToolWithTimeout(tool)
		assert.NotNil(t, wrappedTool, "wrapToolWithTimeout should return a wrapped tool")

		// Verify the tool still works
		info := wrappedTool.Info()
		assert.NotEmpty(t, info.Name, "wrapped tool should have a name")
	})

	// Test 2: Wrap nil tool should be handled
	t.Run("NilTool", func(t *testing.T) {
		wrappedTool := coordinator.wrapToolWithTimeout(nil)
		assert.Nil(t, wrappedTool, "wrapToolWithTimeout should return nil for nil input")
	})

	// Test 3: Wrap panicky tool should be caught
	t.Run("PanickyTool", func(t *testing.T) {
		// Create a tool that panics when Info() is called
		panickyTool := &PanickyAgentTool{}

		// This should not panic - the panic should be caught
		wrappedTool := coordinator.wrapToolWithTimeout(panickyTool)
		assert.Nil(t, wrappedTool, "wrapToolWithTimeout should return nil for panicky tool")
	})
}

// TestPanicProtection_BuildTools tests that the buildTools function can handle
// panics from the fantasy library without crashing the application.
func TestPanicProtection_BuildTools(t *testing.T) {
	env := setupPanicTestEnv(t)
	coordinator := setupPanicTestCoordinator(t, env)

	// Test 1: Build tools with normal agent config should succeed
	t.Run("NormalAgent", func(t *testing.T) {
		agentConfig := config.Agent{
			Name:         "test-agent",
			Model:        "test-model",
			AllowedTools: []string{"bash", "view", "write"},
			AllowedMCP:   nil,
		}

		// This should not panic
		tools, err := coordinator.buildTools(context.Background(), agentConfig)
		assert.NoError(t, err, "buildTools should not return an error")
		assert.NotNil(t, tools, "tools should not be nil")

		// Verify that valid tools are included and invalid ones are filtered out
		var validTools []string
		for _, tool := range tools {
			if tool != nil {
				info := tool.Info()
				if info.Name != "" {
					validTools = append(validTools, info.Name)
				}
			}
		}

		// Should have at least some tools
		assert.Greater(t, len(validTools), 0, "Should have at least some valid tools")
	})

	// Test 2: Build tools with various tools that might cause panics
	t.Run("PanicRecovery", func(t *testing.T) {
		agentConfig := config.Agent{
			Name:         "panic-test-agent",
			Model:        "test-model",
			AllowedTools: []string{"bash", "view", "write", "edit", "multiedit", "fetch"},
			AllowedMCP:   nil,
		}

		// Capture any panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("buildTools should not panic, but got: %v", r)
				t.Errorf("Stack trace: %s", debug.Stack())
			}
		}()

		// This should not panic even if individual tools fail
		tools, err := coordinator.buildTools(context.Background(), agentConfig)
		assert.NoError(t, err, "buildTools should not return an error even with panicky tools")
		assert.NotNil(t, tools, "tools should not be nil")

		// Should get some working tools
		var workingTools []string
		for _, tool := range tools {
			if tool != nil {
				info := tool.Info()
				if info.Name != "" {
					workingTools = append(workingTools, info.Name)
				}
			}
		}

		assert.Greater(t, len(workingTools), 0, "Should have some working tools")
	})
}

// TestPanicProtection_Integration is an integration test that simulates
// the actual panic scenario that was happening in production.
func TestPanicProtection_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	env := setupPanicTestEnv(t)
	coordinator := setupPanicTestCoordinator(t, env)

	// Test with agent config that includes tools with potentially nil schemas
	agentConfig := config.Agent{
		Name:  "integration-test-agent",
		Model: "test-model",
		AllowedTools: []string{
			"bash", "view", "write", "edit", "multiedit",
			"fetch", "grep", "glob", "ls", "download",
			"job_output", "job_kill", "sourcegraph",
		},
		AllowedMCP: nil,
	}

	// This is the exact scenario that was causing the panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Integration test should not panic, but got: %v", r)
			t.Errorf("Stack trace: %s", debug.Stack())
		}
	}()

	// This should complete without panic
	tools, err := coordinator.buildTools(context.Background(), agentConfig)
	assert.NoError(t, err, "buildTools should not error in integration test")
	assert.NotNil(t, tools, "tools should not be nil in integration test")

	// Verify we got some working tools
	var workingTools []string
	for _, tool := range tools {
		if tool != nil {
			info := tool.Info()
			if info.Name != "" {
				workingTools = append(workingTools, info.Name)
			}
		}
	}

	assert.Greater(t, len(workingTools), 0, "Should have working tools after panic protection")

	// Verify specific expected tools are present
	expectedTools := []string{"bash", "view", "write"}
	for _, expected := range expectedTools {
		found := false
		for _, actual := range workingTools {
			if actual == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected tool %s should be present", expected)
	}
}

// TestPanicProtection_Performance tests that the panic protection
// doesn't significantly impact performance.
func TestPanicProtection_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	env := setupPanicTestEnv(t)
	coordinator := setupPanicTestCoordinator(t, env)

	agentConfig := config.Agent{
		Name:         "performance-test-agent",
		AllowedTools: []string{"bash", "view", "write"},
		AllowedMCP:   nil,
	}

	// Run multiple iterations to test performance
	iterations := 10 // Reduced for test performance
	start := time.Now()

	for i := 0; i < iterations; i++ {
		tools, err := coordinator.buildTools(context.Background(), agentConfig)
		assert.NoError(t, err)
		assert.NotNil(t, tools)
	}

	duration := time.Since(start)
	averageTime := duration / time.Duration(iterations)

	t.Logf("Average buildTools time with panic protection: %v", averageTime)

	// Should complete reasonably quickly (adjust threshold as needed)
	assert.Less(t, averageTime, 500*time.Millisecond, "Panic protection should not make tool building too slow")
}

// setupPanicTestEnv creates a test environment with all required dependencies
func setupPanicTestEnv(t *testing.T) fakeEnv {
	workingDir := filepath.Join(t.TempDir(), "nexora-test")

	err := os.MkdirAll(workingDir, 0o755)
	require.NoError(t, err)

	conn, err := db.Connect(t.Context(), t.TempDir())
	require.NoError(t, err)

	q := db.New(conn)
	sessions := session.NewService(q)
	messages := message.NewService(q)

	permissions := permission.NewPermissionService(workingDir, true, []string{})
	history := history.NewService(q, conn)
	lspClients := csync.NewMap[string, *lsp.Client]()

	t.Cleanup(func() {
		conn.Close()
		os.RemoveAll(workingDir)
	})

	return fakeEnv{
		workingDir,
		sessions,
		messages,
		permissions,
		history,
		lspClients,
	}
}

// setupPanicTestCoordinator creates a coordinator for testing
func setupPanicTestCoordinator(t *testing.T, env fakeEnv) *coordinator {
	// Create a minimal config
	cfg := &config.Config{
		Models: map[config.SelectedModelType]config.SelectedModel{
			"large": {
				Provider: "test",
				Model:    "test-model",
			},
		},
		Options: &config.Options{
			Attribution: &config.Attribution{
				GeneratedWith: false,
				TrailerStyle:  "",
			},
		},
	}

	// Create coordinator directly using the private struct constructor
	c := &coordinator{
		cfg:             cfg,
		sessions:        env.sessions,
		messages:        env.messages,
		permissions:     env.permissions,
		history:         env.history,
		lspClients:      env.lspClients,
		aiops:           nil, // aiops
		sessionLog:      nil, // sessionlog
		resourceMonitor: nil, // resourceMonitor
	}

	return c
}

// errorMessageWriter is a simple writer that captures error messages
type errorMessageWriter struct {
	messages *[]string
}

func (w *errorMessageWriter) Write(p []byte) (n int, err error) {
	*w.messages = append(*w.messages, string(p))
	return len(p), nil
}

// PanickyAgentTool is a test tool that panics when Info() is called
type PanickyAgentTool struct{}

func (p *PanickyAgentTool) Run(ctx context.Context, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
	return fantasy.NewTextResponse("panicky tool output"), nil
}

func (p *PanickyAgentTool) Info() fantasy.ToolInfo {
	panic("panic in PanickyAgentTool.Info() - simulating fantasy library schema bug")
}

// ProviderOptions returns empty provider options
func (p *PanickyAgentTool) ProviderOptions() fantasy.ProviderOptions {
	return fantasy.ProviderOptions{}
}

// SetProviderOptions sets provider options (no-op for test)
func (p *PanickyAgentTool) SetProviderOptions(opts fantasy.ProviderOptions) {
	// No-op for test tool
}
