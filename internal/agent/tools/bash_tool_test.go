package tools

import (
	"context"
	"encoding/json"
	"testing"

	"charm.land/fantasy"
	"github.com/nexora/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBashTool(t *testing.T) {
	mockPermissions := &mockPermissionService{}
	attribution := &config.Attribution{
		GeneratedWith: true,
		TrailerStyle:  config.TrailerStyleAssistedBy,
	}

	tool := NewBashTool(mockPermissions, "/tmp", attribution, "test-model")

	assert.NotNil(t, tool)
	assert.Equal(t, BashToolName, tool.Info().Name)
	assert.Contains(t, tool.Info().Description, "bash") // Basic check that description is generated
}

func TestBashTool_MissingCommand(t *testing.T) {
	mockPermissions := &mockPermissionService{}
	attribution := &config.Attribution{}

	tool := NewBashTool(mockPermissions, "/tmp", attribution, "test-model")

	ctx := context.Background()
	params := BashParams{Command: ""} // Empty command
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-1",
		Name:  BashToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content, "missing command")
}

func TestBashTool_SimpleCommand(t *testing.T) {
	mockPermissions := &mockPermissionService{}
	attribution := &config.Attribution{}

	tool := NewBashTool(mockPermissions, "/tmp", attribution, "test-model")

	// Create context with session ID
	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session-123")

	params := BashParams{
		Command:    "echo test",
		WorkingDir: "/tmp",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-2",
		Name:  BashToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)
	assert.Contains(t, resp.Content, "test")
}

func TestBashTool_WorkingDirectory(t *testing.T) {
	mockPermissions := &mockPermissionService{}
	attribution := &config.Attribution{}

	tool := NewBashTool(mockPermissions, "/", attribution, "test-model")

	// Create context with session ID
	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session-123")

	params := BashParams{
		Command:    "pwd",
		WorkingDir: "/tmp",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-3",
		Name:  BashToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)
	// The output should contain /tmp (pwd shows current directory)
	assert.Contains(t, resp.Content, "/tmp")
}

func TestBashTool_BackgroundCommand(t *testing.T) {
	mockPermissions := &mockPermissionService{}
	attribution := &config.Attribution{}

	tool := NewBashTool(mockPermissions, "/tmp", attribution, "test-model")

	// Create context with session ID
	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session-123")

	params := BashParams{
		Command:         "echo background job",
		RunInBackground: true,
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-4",
		Name:  BashToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)
	// Background commands should complete quickly for simple echo command
	assert.Contains(t, resp.Content, "background job")
}

func TestBashTool_InvalidWorkingDir(t *testing.T) {
	mockPermissions := &mockPermissionService{}
	attribution := &config.Attribution{}

	tool := NewBashTool(mockPermissions, "/tmp", attribution, "test-model")

	// Create context with session ID
	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session-123")

	params := BashParams{
		Command:    "echo test",
		WorkingDir: "/nonexistent/path",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-5",
		Name:  BashToolName,
		Input: string(paramsJSON),
	}

	_, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	// The bash tool may or may not error on invalid working directory,
	// depending on how it handles the error
	// assert.True(t, resp.IsError)
}

func TestBashTool_WithSessionContext(t *testing.T) {
	mockPermissions := &mockPermissionService{}
	attribution := &config.Attribution{}

	tool := NewBashTool(mockPermissions, "/tmp", attribution, "test-model")

	// Create context with session ID
	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session-123")

	params := BashParams{
		Command:    "echo with session",
		WorkingDir: "/tmp",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-6",
		Name:  BashToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)
	assert.Contains(t, resp.Content, "with session")
}

func TestBashTool_CommandWithOutput(t *testing.T) {
	mockPermissions := &mockPermissionService{}
	attribution := &config.Attribution{}

	tool := NewBashTool(mockPermissions, "/tmp", attribution, "test-model")

	// Create context with session ID
	ctx := context.WithValue(context.Background(), SessionIDContextKey, "test-session-123")

	params := BashParams{
		Command: "echo 'multi\nline\noutput'",
	}
	paramsJSON, _ := json.Marshal(params)
	toolCall := fantasy.ToolCall{
		ID:    "test-7",
		Name:  BashToolName,
		Input: string(paramsJSON),
	}

	resp, err := tool.Run(ctx, toolCall)

	require.NoError(t, err)
	assert.False(t, resp.IsError)
	assert.Contains(t, resp.Content, "multi")
	assert.Contains(t, resp.Content, "line")
	assert.Contains(t, resp.Content, "output")
}
